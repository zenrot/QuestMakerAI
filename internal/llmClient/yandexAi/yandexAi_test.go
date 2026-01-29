package yandexAi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendData_OK(t *testing.T) {
	// mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// примитивные проверки
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/completionAsync") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"test-id"}`))
	}))
	defer ts.Close()

	cl := NewClient("fake-key")
	cl.HttpClient = ts.Client()
	cl.ServerYandexAI.URLPOST = ts.URL + "/completionAsync"

	id, err := cl.sendData(context.Background(), "topic", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "test-id" {
		t.Fatalf("expected id 'test-id', got %s", id)
	}
}

func TestWaitForData_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`
		{
			"done": true,
			"response": {
				"@type": "type",
				"alternatives": [
					{
						"message": {
							"role": "assistant",
							"text": "{}"
						},
						"status": "SUCCESS"
					}
				],
				"usage": {},
				"modelVersion": "test"
			}
		}`))
	}))
	defer ts.Close()

	cl := NewClient("fake-key")
	cl.HttpClient = ts.Client()
	cl.ServerYandexAI.URLGET = ts.URL + "/"

	resp, err := cl.waitForData(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Done {
		t.Fatalf("expected Done=true")
	}
}

func TestWaitForData_ContextCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"done": false}`))
	}))
	defer ts.Close()

	cl := NewClient("fake-key")
	cl.HttpClient = ts.Client()
	cl.ServerYandexAI.URLGET = ts.URL + "/"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cl.waitForData(ctx, "id")
	if err == nil {
		t.Fatalf("expected context error")
	}
}
