package yandexAi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zenrot/QuestMakerAI/internal/llmClient"
	"github.com/zenrot/QuestMakerAI/internal/schemas"
)

type ServerYandexAI struct {
	URLPOST string
	URLGET  string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	ApiKey         string
	ModelURI       string
	HttpClient     HTTPClient
	ServerYandexAI *ServerYandexAI
}

type CompletionOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   string  `json:"maxTokens,omitempty"`
}

type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type CompletionRequest struct {
	ModelURI          string             `json:"modelUri"`
	CompletionOptions *CompletionOptions `json:"completionOptions,omitempty"`
	Messages          []Message          `json:"messages"`
	JsonSchema        struct {
		Schema json.RawMessage `json:"schema"`
	} `json:"jsonSchema"`
}

type Alternative struct {
	Message Message `json:"message"`
	Status  string  `json:"status"`
}

type CompletionTokensDetails struct {
	ReasoningTokens string `json:"reasoningTokens"`
}

type Usage struct {
	InputTextTokens      string                  `json:"inputTextTokens"`
	CompletionTokens     string                  `json:"completionTokens"`
	TotalTokens          string                  `json:"totalTokens"`
	CompletionTokensInfo CompletionTokensDetails `json:"completionTokensDetails"`
}

type OperationResponse struct {
	Done     bool   `json:"done"`
	ID       string `json:"id"`
	Response struct {
		Type         string        `json:"@type"`
		Alternatives []Alternative `json:"alternatives"`
		Usage        Usage         `json:"usage"`
		ModelVersion string        `json:"modelVersion"`
	} `json:"response"`
	Error json.RawMessage `json:"error,omitempty"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		ApiKey:     apiKey,
		ModelURI:   "gpt://b1gnrl1oguq548jflgka/yandexgpt/rc",
		HttpClient: http.DefaultClient,
		ServerYandexAI: &ServerYandexAI{
			URLPOST: "https://llm.api.cloud.yandex.net/foundationModels/v1/completionAsync",
			URLGET:  "https://operation.api.cloud.yandex.net/operations/",
		},
	}
}

func (cl *Client) RequestTest(ctx context.Context, topic string, questionNumber int) (*schemas.QuestionsSchemaJson, error) {

	id, err := cl.sendData(ctx, topic, questionNumber)
	if err != nil {
		return nil, err
	}
	resp, err := cl.waitForData(ctx, id)
	if err != nil {
		return nil, err
	}
	text := resp.Response.Alternatives[0].Message.Text
	if cor, err := schemas.ValidateSchema(schemas.QuestionSchema, []byte(text)); err != nil || !cor {
		return nil, err
	}
	var result schemas.QuestionsSchemaJson
	if err = json.Unmarshal([]byte(text), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (cl *Client) sendData(ctx context.Context, topic string, questionNumber int) (ID string, error error) {
	reqBody := CompletionRequest{
		ModelURI: cl.ModelURI,
		CompletionOptions: &CompletionOptions{
			Temperature: llmClient.Temperature,
			MaxTokens:   llmClient.MaxTokens,
		},
		Messages: []Message{
			{
				Role: "system",
				Text: fmt.Sprintf(llmClient.SystemPrompt, questionNumber),
			},
			{
				Role: "user",
				Text: topic,
			},
		},
		JsonSchema: struct {
			Schema json.RawMessage `json:"schema"`
		}{Schema: schemas.QuestionSchema},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", cl.ServerYandexAI.URLPOST, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprint("Api-Key ", cl.ApiKey))

	resp, err := cl.HttpClient.Do(req)
	if err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	var extractedData map[string]interface{}
	err = json.Unmarshal(bodyBytes, &extractedData)
	if err != nil {
		return "", err
	}
	errVal := extractedData["error"]
	if errVal != nil {
		return "", errors.New(errVal.(string))
	}
	idVal, ok := extractedData["id"]
	if !ok {
		return "", fmt.Errorf("response has no id field: %s", string(bodyBytes))
	}
	id, ok := idVal.(string)
	if !ok {
		return "", fmt.Errorf("id is not string: %v", idVal)
	}
	id = strings.TrimSpace(id)
	return id, nil
}

func (cl *Client) waitForData(ctx context.Context, id string) (*OperationResponse, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s%s", cl.ServerYandexAI.URLGET, id), nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprint("Api-Key ", cl.ApiKey))

			resp, err := cl.HttpClient.Do(req)
			if err != nil {
				return nil, err
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body.Close()

			var opResp OperationResponse
			if err := json.Unmarshal(bodyBytes, &opResp); err != nil {
				return nil, err
			}

			if opResp.Error != nil {
				return nil, fmt.Errorf("llm error: %s", opResp.Error)
			}

			if opResp.Done {
				return &opResp, nil
			}

			select {
			case <-time.After(1 * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}
	}
}
