package gollm

import (
	"testing"

	"github.com/grokify/gollm/provider"
)

func TestNewProvider(t *testing.T) {
	t.Run("with nil client", func(t *testing.T) {
		p := NewProvider(nil)
		if p.client != nil {
			t.Error("client should be nil")
		}
	})

	t.Run("with options", func(t *testing.T) {
		p := NewProvider(nil,
			WithModel("gpt-4"),
			WithTemperature(0.7),
			WithMaxTokens(1000),
		)

		if p.model != "gpt-4" {
			t.Errorf("model = %q, want %q", p.model, "gpt-4")
		}
		if p.temperature != 0.7 {
			t.Errorf("temperature = %v, want 0.7", p.temperature)
		}
		if p.maxTokens != 1000 {
			t.Errorf("maxTokens = %d, want 1000", p.maxTokens)
		}
	})
}

func TestProviderName(t *testing.T) {
	p := NewProvider(nil)
	if p.Name() != "gollm" {
		t.Errorf("Name() = %q, want %q", p.Name(), "gollm")
	}
}

func TestProviderDefaultModel(t *testing.T) {
	p := NewProvider(nil, WithModel("claude-3-opus"))
	if p.DefaultModel() != "claude-3-opus" {
		t.Errorf("DefaultModel() = %q, want %q", p.DefaultModel(), "claude-3-opus")
	}
}

func TestNewTracingClient(t *testing.T) {
	tc := NewTracingClient(nil, nil)
	if tc == nil {
		t.Fatal("TracingClient should not be nil")
	}
	if tc.client != nil {
		t.Error("client should be nil")
	}
	if tc.opikClient != nil {
		t.Error("opikClient should be nil")
	}
}

func TestTracingClientClient(t *testing.T) {
	tc := NewTracingClient(nil, nil)
	if tc.Client() != nil {
		t.Error("Client() should return nil when underlying client is nil")
	}
}

func TestRequestToMap(t *testing.T) {
	temp := 0.5
	maxTokens := 100

	req := &provider.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: "You are helpful."},
			{Role: provider.RoleUser, Content: "Hello"},
		},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Stop:        []string{"END"},
	}

	m := requestToMap(req)

	if m["model"] != "gpt-4" {
		t.Errorf("model = %v, want %q", m["model"], "gpt-4")
	}

	messages, ok := m["messages"].([]map[string]any)
	if !ok {
		t.Fatal("messages should be []map[string]any")
	}
	if len(messages) != 2 {
		t.Errorf("messages length = %d, want 2", len(messages))
	}
	if messages[0]["role"] != "system" {
		t.Errorf("messages[0].role = %v, want system", messages[0]["role"])
	}
	if messages[1]["content"] != "Hello" {
		t.Errorf("messages[1].content = %v, want Hello", messages[1]["content"])
	}

	if m["temperature"] != temp {
		t.Errorf("temperature = %v, want %v", m["temperature"], temp)
	}
	if m["max_tokens"] != maxTokens {
		t.Errorf("max_tokens = %v, want %v", m["max_tokens"], maxTokens)
	}

	stop, ok := m["stop"].([]string)
	if !ok || len(stop) != 1 || stop[0] != "END" {
		t.Errorf("stop = %v, want [END]", m["stop"])
	}
}

func TestRequestToMapMinimal(t *testing.T) {
	req := &provider.ChatCompletionRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []provider.Message{},
	}

	m := requestToMap(req)

	if m["model"] != "gpt-3.5-turbo" {
		t.Errorf("model = %v, want %q", m["model"], "gpt-3.5-turbo")
	}

	// Optional fields should not be present
	if _, ok := m["temperature"]; ok {
		t.Error("temperature should not be present")
	}
	if _, ok := m["max_tokens"]; ok {
		t.Error("max_tokens should not be present")
	}
	if _, ok := m["stop"]; ok {
		t.Error("stop should not be present")
	}
}

func TestResponseToMap(t *testing.T) {
	finishReason := "stop"
	resp := &provider.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Model:   "gpt-4",
		Created: 1677652288,
		Choices: []provider.ChatCompletionChoice{
			{
				Message:      provider.Message{Role: provider.RoleAssistant, Content: "Hello!"},
				FinishReason: &finishReason,
			},
		},
		Usage: provider.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	m := responseToMap(resp)

	if m["id"] != "chatcmpl-123" {
		t.Errorf("id = %v, want %q", m["id"], "chatcmpl-123")
	}
	if m["model"] != "gpt-4" {
		t.Errorf("model = %v, want %q", m["model"], "gpt-4")
	}
	if m["content"] != "Hello!" {
		t.Errorf("content = %v, want %q", m["content"], "Hello!")
	}
	if m["finish_reason"] != "stop" {
		t.Errorf("finish_reason = %v, want %q", m["finish_reason"], "stop")
	}

	usage, ok := m["usage"].(map[string]any)
	if !ok {
		t.Fatal("usage should be map[string]any")
	}
	if usage["prompt_tokens"] != 10 {
		t.Errorf("usage.prompt_tokens = %v, want 10", usage["prompt_tokens"])
	}
	if usage["completion_tokens"] != 5 {
		t.Errorf("usage.completion_tokens = %v, want 5", usage["completion_tokens"])
	}
	if usage["total_tokens"] != 15 {
		t.Errorf("usage.total_tokens = %v, want 15", usage["total_tokens"])
	}
}

func TestResponseToMapEmptyChoices(t *testing.T) {
	resp := &provider.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Model:   "gpt-4",
		Choices: []provider.ChatCompletionChoice{},
		Usage:   provider.Usage{},
	}

	m := responseToMap(resp)

	// Content should not be present when no choices
	if _, ok := m["content"]; ok {
		t.Error("content should not be present when no choices")
	}
	if _, ok := m["finish_reason"]; ok {
		t.Error("finish_reason should not be present when no choices")
	}
}

func TestTracingStreamInit(t *testing.T) {
	// Test that tracingStream fields initialize correctly
	ts := &tracingStream{
		closed: false,
	}

	if ts.closed {
		t.Error("closed should be false initially")
	}
	if ts.responseBuffer.Len() != 0 {
		t.Error("responseBuffer should be empty initially")
	}
	if ts.model != "" {
		t.Error("model should be empty initially")
	}
	if ts.usage != nil {
		t.Error("usage should be nil initially")
	}
}
