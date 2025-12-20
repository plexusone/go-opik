package opik

import (
	"testing"
	"time"
)

func TestDefaultBatcherConfig(t *testing.T) {
	config := DefaultBatcherConfig()

	if config.MaxBatchSize != 100 {
		t.Errorf("MaxBatchSize = %d, want 100", config.MaxBatchSize)
	}
	if config.FlushInterval != 5*time.Second {
		t.Errorf("FlushInterval = %v, want 5s", config.FlushInterval)
	}
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.RetryDelay != 100*time.Millisecond {
		t.Errorf("RetryDelay = %v, want 100ms", config.RetryDelay)
	}
	if config.Workers != 2 {
		t.Errorf("Workers = %d, want 2", config.Workers)
	}
}

func TestBatcherConfigCustom(t *testing.T) {
	config := BatcherConfig{
		MaxBatchSize:  50,
		FlushInterval: 2 * time.Second,
		MaxRetries:    5,
		RetryDelay:    200 * time.Millisecond,
		Workers:       4,
	}

	if config.MaxBatchSize != 50 {
		t.Errorf("MaxBatchSize = %d, want 50", config.MaxBatchSize)
	}
	if config.Workers != 4 {
		t.Errorf("Workers = %d, want 4", config.Workers)
	}
}

func TestTraceBatchItemType(t *testing.T) {
	item := TraceBatchItem{Trace: nil}
	if item.Type() != "trace" {
		t.Errorf("Type() = %q, want %q", item.Type(), "trace")
	}
}

func TestSpanBatchItemType(t *testing.T) {
	item := SpanBatchItem{Span: nil}
	if item.Type() != "span" {
		t.Errorf("Type() = %q, want %q", item.Type(), "span")
	}
}

func TestFeedbackBatchItemType(t *testing.T) {
	item := FeedbackBatchItem{
		EntityType: "trace",
		EntityID:   "trace-123",
		Name:       "accuracy",
		Value:      0.95,
		Reason:     "high quality response",
	}

	if item.Type() != "feedback" {
		t.Errorf("Type() = %q, want %q", item.Type(), "feedback")
	}
	if item.EntityType != "trace" {
		t.Errorf("EntityType = %q, want %q", item.EntityType, "trace")
	}
	if item.EntityID != "trace-123" {
		t.Errorf("EntityID = %q, want %q", item.EntityID, "trace-123")
	}
	if item.Name != "accuracy" {
		t.Errorf("Name = %q, want %q", item.Name, "accuracy")
	}
	if item.Value != 0.95 {
		t.Errorf("Value = %v, want 0.95", item.Value)
	}
	if item.Reason != "high quality response" {
		t.Errorf("Reason = %q, want %q", item.Reason, "high quality response")
	}
}

func TestFeedbackBatchItemForSpan(t *testing.T) {
	item := FeedbackBatchItem{
		EntityType: "span",
		EntityID:   "span-456",
		Name:       "relevance",
		Value:      0.8,
	}

	if item.EntityType != "span" {
		t.Errorf("EntityType = %q, want %q", item.EntityType, "span")
	}
	if item.EntityID != "span-456" {
		t.Errorf("EntityID = %q, want %q", item.EntityID, "span-456")
	}
}

func TestBatchItemInterface(t *testing.T) {
	// Test that all batch item types implement the interface
	var items []BatchItem

	items = append(items, TraceBatchItem{})
	items = append(items, SpanBatchItem{})
	items = append(items, FeedbackBatchItem{})

	expectedTypes := []string{"trace", "span", "feedback"}
	for i, item := range items {
		if item.Type() != expectedTypes[i] {
			t.Errorf("item[%d].Type() = %q, want %q", i, item.Type(), expectedTypes[i])
		}
	}
}

func TestBatcherConfigZeroValues(t *testing.T) {
	config := BatcherConfig{}

	if config.MaxBatchSize != 0 {
		t.Errorf("zero MaxBatchSize = %d, want 0", config.MaxBatchSize)
	}
	if config.FlushInterval != 0 {
		t.Errorf("zero FlushInterval = %v, want 0", config.FlushInterval)
	}
	if config.MaxRetries != 0 {
		t.Errorf("zero MaxRetries = %d, want 0", config.MaxRetries)
	}
	if config.Workers != 0 {
		t.Errorf("zero Workers = %d, want 0", config.Workers)
	}
}
