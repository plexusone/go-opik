package opik

import (
	"testing"
	"time"
)

func TestSpanID(t *testing.T) {
	span := &Span{
		id: "test-span-id-123",
	}

	if span.ID() != "test-span-id-123" {
		t.Errorf("ID() = %q, want %q", span.ID(), "test-span-id-123")
	}
}

func TestSpanTraceID(t *testing.T) {
	span := &Span{
		traceID: "trace-id-456",
	}

	if span.TraceID() != "trace-id-456" {
		t.Errorf("TraceID() = %q, want %q", span.TraceID(), "trace-id-456")
	}
}

func TestSpanParentSpanID(t *testing.T) {
	t.Run("with parent", func(t *testing.T) {
		span := &Span{
			parentSpanID: "parent-span-789",
		}
		if span.ParentSpanID() != "parent-span-789" {
			t.Errorf("ParentSpanID() = %q, want %q", span.ParentSpanID(), "parent-span-789")
		}
	})

	t.Run("without parent", func(t *testing.T) {
		span := &Span{}
		if span.ParentSpanID() != "" {
			t.Errorf("ParentSpanID() = %q, want empty", span.ParentSpanID())
		}
	})
}

func TestSpanName(t *testing.T) {
	span := &Span{
		name: "llm-call",
	}

	if span.Name() != "llm-call" {
		t.Errorf("Name() = %q, want %q", span.Name(), "llm-call")
	}
}

func TestSpanType(t *testing.T) {
	tests := []struct {
		spanType string
	}{
		{SpanTypeLLM},
		{SpanTypeTool},
		{SpanTypeGeneral},
		{SpanTypeGuardrail},
	}

	for _, tt := range tests {
		t.Run(tt.spanType, func(t *testing.T) {
			span := &Span{spanType: tt.spanType}
			if span.Type() != tt.spanType {
				t.Errorf("Type() = %q, want %q", span.Type(), tt.spanType)
			}
		})
	}
}

func TestSpanStartTime(t *testing.T) {
	now := time.Now()
	span := &Span{
		startTime: now,
	}

	if !span.StartTime().Equal(now) {
		t.Errorf("StartTime() = %v, want %v", span.StartTime(), now)
	}
}

func TestSpanEndTime(t *testing.T) {
	t.Run("not ended", func(t *testing.T) {
		span := &Span{}
		if span.EndTime() != nil {
			t.Error("EndTime() should be nil for unended span")
		}
	})

	t.Run("ended", func(t *testing.T) {
		endTime := time.Now()
		span := &Span{
			endTime: &endTime,
		}
		if span.EndTime() == nil {
			t.Error("EndTime() should not be nil for ended span")
		}
		if !span.EndTime().Equal(endTime) {
			t.Errorf("EndTime() = %v, want %v", span.EndTime(), endTime)
		}
	})
}

func TestSpanFields(t *testing.T) {
	now := time.Now()
	span := &Span{
		id:           "span-123",
		traceID:      "trace-456",
		parentSpanID: "parent-789",
		name:         "test-span",
		spanType:     SpanTypeLLM,
		startTime:    now,
		input:        map[string]any{"prompt": "hello"},
		output:       map[string]any{"response": "world"},
		metadata:     map[string]any{"model": "gpt-4"},
		tags:         []string{"tag1"},
		model:        "gpt-4",
		provider:     "openai",
		usage:        map[string]int{"tokens": 100},
		ended:        false,
	}

	if span.ID() != "span-123" {
		t.Errorf("ID() = %q, want %q", span.ID(), "span-123")
	}
	if span.TraceID() != "trace-456" {
		t.Errorf("TraceID() = %q, want %q", span.TraceID(), "trace-456")
	}
	if span.ParentSpanID() != "parent-789" {
		t.Errorf("ParentSpanID() = %q, want %q", span.ParentSpanID(), "parent-789")
	}
	if span.Name() != "test-span" {
		t.Errorf("Name() = %q, want %q", span.Name(), "test-span")
	}
	if span.Type() != SpanTypeLLM {
		t.Errorf("Type() = %q, want %q", span.Type(), SpanTypeLLM)
	}
}

func TestSpanEnded(t *testing.T) {
	t.Run("initially not ended", func(t *testing.T) {
		span := &Span{}
		if span.ended {
			t.Error("span should not be ended initially")
		}
	})

	t.Run("ended flag set", func(t *testing.T) {
		span := &Span{ended: true}
		if !span.ended {
			t.Error("span should be ended")
		}
	})
}

func TestSpanWithAllFields(t *testing.T) {
	now := time.Now()
	endTime := now.Add(2 * time.Second)

	span := &Span{
		id:           "full-span-id",
		traceID:      "full-trace-id",
		parentSpanID: "full-parent-id",
		name:         "full-span",
		spanType:     SpanTypeLLM,
		startTime:    now,
		endTime:      &endTime,
		input: map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		},
		output: map[string]any{
			"message": map[string]string{
				"role":    "assistant",
				"content": "Hi!",
			},
		},
		metadata: map[string]any{
			"temperature": 0.7,
			"max_tokens":  1000,
		},
		tags:     []string{"production", "chat"},
		model:    "gpt-4-turbo",
		provider: "openai",
		usage: map[string]int{
			"prompt_tokens":     50,
			"completion_tokens": 100,
			"total_tokens":      150,
		},
		ended: true,
	}

	// Verify all fields
	if span.ID() != "full-span-id" {
		t.Errorf("ID() = %q, want %q", span.ID(), "full-span-id")
	}
	if span.TraceID() != "full-trace-id" {
		t.Errorf("TraceID() = %q, want %q", span.TraceID(), "full-trace-id")
	}
	if span.ParentSpanID() != "full-parent-id" {
		t.Errorf("ParentSpanID() = %q, want %q", span.ParentSpanID(), "full-parent-id")
	}
	if span.Type() != SpanTypeLLM {
		t.Errorf("Type() = %q, want %q", span.Type(), SpanTypeLLM)
	}
	if span.EndTime() == nil {
		t.Error("EndTime() should not be nil")
	}
}

func TestSpanEmptyFields(t *testing.T) {
	span := &Span{}

	if span.ID() != "" {
		t.Errorf("ID() = %q, want empty", span.ID())
	}
	if span.TraceID() != "" {
		t.Errorf("TraceID() = %q, want empty", span.TraceID())
	}
	if span.ParentSpanID() != "" {
		t.Errorf("ParentSpanID() = %q, want empty", span.ParentSpanID())
	}
	if span.Name() != "" {
		t.Errorf("Name() = %q, want empty", span.Name())
	}
	if span.Type() != "" {
		t.Errorf("Type() = %q, want empty", span.Type())
	}
}

func TestSpanSetUsage(t *testing.T) {
	span := &Span{}

	usage := map[string]int{
		"prompt_tokens":     100,
		"completion_tokens": 50,
		"total_tokens":      150,
	}

	span.SetUsage(usage)

	if span.usage == nil {
		t.Fatal("usage should not be nil after SetUsage")
	}
	if span.usage["prompt_tokens"] != 100 {
		t.Errorf("usage[prompt_tokens] = %d, want 100", span.usage["prompt_tokens"])
	}
	if span.usage["completion_tokens"] != 50 {
		t.Errorf("usage[completion_tokens] = %d, want 50", span.usage["completion_tokens"])
	}
	if span.usage["total_tokens"] != 150 {
		t.Errorf("usage[total_tokens] = %d, want 150", span.usage["total_tokens"])
	}
}

func TestSpanInputOutput(t *testing.T) {
	span := &Span{
		input: map[string]any{
			"messages": []map[string]string{
				{"role": "system", "content": "You are helpful"},
				{"role": "user", "content": "Hello"},
			},
		},
		output: map[string]any{
			"message": map[string]string{
				"role":    "assistant",
				"content": "Hello! How can I help?",
			},
			"finish_reason": "stop",
		},
	}

	if span.input == nil {
		t.Error("input should not be nil")
	}
	if span.output == nil {
		t.Error("output should not be nil")
	}
}

func TestSpanMetadata(t *testing.T) {
	span := &Span{
		metadata: map[string]any{
			"temperature":       0.7,
			"max_tokens":        2048,
			"presence_penalty":  0.0,
			"frequency_penalty": 0.0,
			"user":              "user-123",
		},
	}

	if span.metadata == nil {
		t.Error("metadata should not be nil")
	}
	if span.metadata["temperature"] != 0.7 {
		t.Errorf("metadata[temperature] = %v, want 0.7", span.metadata["temperature"])
	}
	if span.metadata["max_tokens"] != 2048 {
		t.Errorf("metadata[max_tokens] = %v, want 2048", span.metadata["max_tokens"])
	}
}

func TestSpanTags(t *testing.T) {
	span := &Span{
		tags: []string{"chat", "production", "gpt-4"},
	}

	if len(span.tags) != 3 {
		t.Errorf("tags length = %d, want 3", len(span.tags))
	}
	if span.tags[0] != "chat" {
		t.Errorf("tags[0] = %q, want %q", span.tags[0], "chat")
	}
}

func TestSpanModelProvider(t *testing.T) {
	span := &Span{
		model:    "claude-3-opus",
		provider: "anthropic",
	}

	if span.model != "claude-3-opus" {
		t.Errorf("model = %q, want %q", span.model, "claude-3-opus")
	}
	if span.provider != "anthropic" {
		t.Errorf("provider = %q, want %q", span.provider, "anthropic")
	}
}

func TestSpanNilClient(t *testing.T) {
	// Test that a span with nil client doesn't panic on getters
	span := &Span{
		id:      "test-id",
		traceID: "trace-id",
		name:    "test-name",
	}

	// These should not panic
	_ = span.ID()
	_ = span.TraceID()
	_ = span.ParentSpanID()
	_ = span.Name()
	_ = span.Type()
	_ = span.StartTime()
	_ = span.EndTime()
}

func TestSpanHierarchy(t *testing.T) {
	// Test span hierarchy structure
	rootSpan := &Span{
		id:       "root-span",
		traceID:  "trace-1",
		name:     "root",
		spanType: SpanTypeGeneral,
	}

	childSpan := &Span{
		id:           "child-span",
		traceID:      "trace-1",
		parentSpanID: "root-span",
		name:         "child",
		spanType:     SpanTypeLLM,
	}

	grandchildSpan := &Span{
		id:           "grandchild-span",
		traceID:      "trace-1",
		parentSpanID: "child-span",
		name:         "grandchild",
		spanType:     SpanTypeTool,
	}

	// Verify hierarchy
	if rootSpan.ParentSpanID() != "" {
		t.Error("root span should have no parent")
	}
	if childSpan.ParentSpanID() != rootSpan.ID() {
		t.Errorf("child parent = %q, want %q", childSpan.ParentSpanID(), rootSpan.ID())
	}
	if grandchildSpan.ParentSpanID() != childSpan.ID() {
		t.Errorf("grandchild parent = %q, want %q", grandchildSpan.ParentSpanID(), childSpan.ID())
	}

	// All should have same trace ID
	if childSpan.TraceID() != rootSpan.TraceID() {
		t.Error("all spans should have same trace ID")
	}
	if grandchildSpan.TraceID() != rootSpan.TraceID() {
		t.Error("all spans should have same trace ID")
	}
}
