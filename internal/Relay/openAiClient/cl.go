package openAiClient

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func Request(token string, ctx context.Context) (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(token),
	)
	params := openai.ChatCompletionNewParams{
		Model: "gpt-3.5-turbo-1106",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a concise and professional assistant"),
			openai.UserMessage("расскажи о себе"),
		},
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
