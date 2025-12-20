package anthropic

import (
	"context"
	"net/http"
	"testing"

	"github.com/grokify/go-comet-ml-opik/evaluation/llm"
	"github.com/grokify/go-comet-ml-opik/testutil"
)

func TestNewProvider(t *testing.T) {
	t.Run("with defaults", func(t *testing.T) {
		p := NewProvider()
		if p.model != defaultModel {
			t.Errorf("model = %q, want %q", p.model, defaultModel)
		}
		if p.baseURL != defaultBaseURL {
			t.Errorf("baseURL = %q, want %q", p.baseURL, defaultBaseURL)
		}
		if p.maxTokens != 4096 {
			t.Errorf("maxTokens = %d, want 4096", p.maxTokens)
		}
	})

	t.Run("with options", func(t *testing.T) {
		p := NewProvider(
			WithAPIKey("test-key"),
			WithBaseURL("https://custom.api.com"),
			WithModel("claude-3-opus-20240229"),
			WithTemperature(0.7),
			WithMaxTokens(8192),
		)

		if p.apiKey != "test-key" {
			t.Errorf("apiKey = %q, want %q", p.apiKey, "test-key")
		}
		if p.baseURL != "https://custom.api.com" {
			t.Errorf("baseURL = %q, want %q", p.baseURL, "https://custom.api.com")
		}
		if p.model != "claude-3-opus-20240229" {
			t.Errorf("model = %q, want %q", p.model, "claude-3-opus-20240229")
		}
		if p.temperature != 0.7 {
			t.Errorf("temperature = %v, want 0.7", p.temperature)
		}
		if p.maxTokens != 8192 {
			t.Errorf("maxTokens = %d, want 8192", p.maxTokens)
		}
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30}
		p := NewProvider(WithHTTPClient(customClient))
		if p.client != customClient {
			t.Error("custom HTTP client not set")
		}
	})
}

func TestProviderName(t *testing.T) {
	p := NewProvider()
	if p.Name() != "anthropic" {
		t.Errorf("Name() = %q, want %q", p.Name(), "anthropic")
	}
}

func TestProviderDefaultModel(t *testing.T) {
	p := NewProvider(WithModel("claude-3-haiku-20240307"))
	if p.DefaultModel() != "claude-3-haiku-20240307" {
		t.Errorf("DefaultModel() = %q, want %q", p.DefaultModel(), "claude-3-haiku-20240307")
	}
}

func TestProviderComplete(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Mock successful completion response
	ms.OnPost("/messages").RespondJSON(200, map[string]any{
		"id":    "msg_123",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-sonnet-4-20250514",
		"content": []map[string]any{
			{
				"type": "text",
				"text": "Hello! How can I help you today?",
			},
		},
		"stop_reason": "end_turn",
		"usage": map[string]any{
			"input_tokens":  10,
			"output_tokens": 8,
		},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("test-api-key"),
		WithModel("claude-sonnet-4-20250514"),
	)

	resp, err := p.Complete(context.Background(), llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Hello!"},
		},
	})

	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}

	if resp.Content != "Hello! How can I help you today?" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello! How can I help you today?")
	}
	if resp.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model = %q, want %q", resp.Model, "claude-sonnet-4-20250514")
	}
	if resp.PromptTokens != 10 {
		t.Errorf("PromptTokens = %d, want 10", resp.PromptTokens)
	}
	if resp.OutputTokens != 8 {
		t.Errorf("OutputTokens = %d, want 8", resp.OutputTokens)
	}

	// Verify request headers
	last := ms.LastRequest()
	if last.Headers.Get("x-api-key") != "test-api-key" {
		t.Errorf("x-api-key = %q, want %q", last.Headers.Get("x-api-key"), "test-api-key")
	}
	if last.Headers.Get("anthropic-version") != apiVersion {
		t.Errorf("anthropic-version = %q, want %q", last.Headers.Get("anthropic-version"), apiVersion)
	}
	if last.Headers.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", last.Headers.Get("Content-Type"), "application/json")
	}
}

func TestProviderCompleteWithSystemMessage(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/messages").RespondJSON(200, map[string]any{
		"id":    "msg_123",
		"type":  "message",
		"role":  "assistant",
		"model": "claude-sonnet-4-20250514",
		"content": []map[string]any{
			{"type": "text", "text": "Response"},
		},
		"usage": map[string]any{},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("key"),
	)

	// System message should be extracted and sent separately
	_, err := p.Complete(context.Background(), llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello!"},
		},
	})

	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}

	// System message should be in the request body as "system" field
	if ms.RequestCount() != 1 {
		t.Error("expected 1 request")
	}
}

func TestProviderCompleteAPIError(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/messages").Respond(401, `{"error": {"message": "Invalid API key"}}`)

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("invalid-key"),
	)

	_, err := p.Complete(context.Background(), llm.CompletionRequest{
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	})

	if err == nil {
		t.Error("expected error for API error")
	}
}

func TestProviderCompleteEmptyContent(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/messages").RespondJSON(200, map[string]any{
		"id":      "msg_123",
		"type":    "message",
		"content": []map[string]any{},
		"usage":   map[string]any{},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("key"),
	)

	resp, err := p.Complete(context.Background(), llm.CompletionRequest{
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	})

	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	// Empty content should result in empty string
	if resp.Content != "" {
		t.Errorf("Content = %q, want empty", resp.Content)
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	var _ llm.Provider = (*Provider)(nil)
}

func TestProviderCompleteWithModelOverride(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/messages").RespondJSON(200, map[string]any{
		"id":    "msg_123",
		"model": "claude-3-haiku-20240307",
		"content": []map[string]any{
			{"type": "text", "text": "Response"},
		},
		"usage": map[string]any{},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("key"),
		WithModel("claude-sonnet-4-20250514"), // default
	)

	// Override model in request
	resp, err := p.Complete(context.Background(), llm.CompletionRequest{
		Model:    "claude-3-haiku-20240307", // override
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	})

	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp.Model != "claude-3-haiku-20240307" {
		t.Errorf("Model = %q, want %q", resp.Model, "claude-3-haiku-20240307")
	}
}
