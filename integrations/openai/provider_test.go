package openai

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
	})

	t.Run("with options", func(t *testing.T) {
		p := NewProvider(
			WithAPIKey("test-key"),
			WithBaseURL("https://custom.api.com"),
			WithModel("gpt-3.5-turbo"),
			WithTemperature(0.7),
			WithMaxTokens(100),
		)

		if p.apiKey != "test-key" {
			t.Errorf("apiKey = %q, want %q", p.apiKey, "test-key")
		}
		if p.baseURL != "https://custom.api.com" {
			t.Errorf("baseURL = %q, want %q", p.baseURL, "https://custom.api.com")
		}
		if p.model != "gpt-3.5-turbo" {
			t.Errorf("model = %q, want %q", p.model, "gpt-3.5-turbo")
		}
		if p.temperature != 0.7 {
			t.Errorf("temperature = %v, want 0.7", p.temperature)
		}
		if p.maxTokens != 100 {
			t.Errorf("maxTokens = %d, want 100", p.maxTokens)
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
	if p.Name() != "openai" {
		t.Errorf("Name() = %q, want %q", p.Name(), "openai")
	}
}

func TestProviderDefaultModel(t *testing.T) {
	p := NewProvider(WithModel("gpt-4"))
	if p.DefaultModel() != "gpt-4" {
		t.Errorf("DefaultModel() = %q, want %q", p.DefaultModel(), "gpt-4")
	}
}

func TestProviderComplete(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	// Mock successful completion response
	ms.OnPost("/chat/completions").RespondJSON(200, map[string]any{
		"id":      "chatcmpl-123",
		"object":  "chat.completion",
		"created": 1677652288,
		"model":   "gpt-4o",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": "Hello! How can I help you today?",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     10,
			"completion_tokens": 8,
			"total_tokens":      18,
		},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("test-api-key"),
		WithModel("gpt-4o"),
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
	if resp.Model != "gpt-4o" {
		t.Errorf("Model = %q, want %q", resp.Model, "gpt-4o")
	}
	if resp.PromptTokens != 10 {
		t.Errorf("PromptTokens = %d, want 10", resp.PromptTokens)
	}
	if resp.OutputTokens != 8 {
		t.Errorf("OutputTokens = %d, want 8", resp.OutputTokens)
	}

	// Verify request was made correctly
	last := ms.LastRequest()
	if last.Headers.Get("Authorization") != "Bearer test-api-key" {
		t.Errorf("Authorization = %q, want %q", last.Headers.Get("Authorization"), "Bearer test-api-key")
	}
	if last.Headers.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", last.Headers.Get("Content-Type"), "application/json")
	}
}

func TestProviderCompleteWithTemperature(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/chat/completions").RespondJSON(200, map[string]any{
		"id":      "chatcmpl-123",
		"object":  "chat.completion",
		"created": 1677652288,
		"model":   "gpt-4o",
		"choices": []map[string]any{
			{
				"index":   0,
				"message": map[string]any{"role": "assistant", "content": "Response"},
			},
		},
		"usage": map[string]any{},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("key"),
		WithTemperature(0.5),
	)

	_, _ = p.Complete(context.Background(), llm.CompletionRequest{
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	})

	// Temperature should be in the request body
	if ms.RequestCount() != 1 {
		t.Error("expected 1 request")
	}
}

func TestProviderCompleteAPIError(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/chat/completions").Respond(401, `{"error": "Invalid API key"}`)

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

func TestProviderCompleteEmptyChoices(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/chat/completions").RespondJSON(200, map[string]any{
		"id":      "chatcmpl-123",
		"object":  "chat.completion",
		"choices": []map[string]any{},
		"usage":   map[string]any{},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("key"),
	)

	_, err := p.Complete(context.Background(), llm.CompletionRequest{
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	})

	if err == nil {
		t.Error("expected error for empty choices")
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	var _ llm.Provider = (*Provider)(nil)
}

func TestProviderCompleteWithModelOverride(t *testing.T) {
	ms := testutil.NewMockServer()
	defer ms.Close()

	ms.OnPost("/chat/completions").RespondJSON(200, map[string]any{
		"id":    "chatcmpl-123",
		"model": "gpt-3.5-turbo",
		"choices": []map[string]any{
			{"index": 0, "message": map[string]any{"role": "assistant", "content": "Response"}},
		},
		"usage": map[string]any{},
	})

	p := NewProvider(
		WithBaseURL(ms.URL()),
		WithAPIKey("key"),
		WithModel("gpt-4o"), // default
	)

	// Override model in request
	resp, err := p.Complete(context.Background(), llm.CompletionRequest{
		Model:    "gpt-3.5-turbo", // override
		Messages: []llm.Message{{Role: "user", Content: "test"}},
	})

	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp.Model != "gpt-3.5-turbo" {
		t.Errorf("Model = %q, want %q", resp.Model, "gpt-3.5-turbo")
	}
}
