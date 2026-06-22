package agentobs

import (
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent(EventMessageReceived)

	if event.Type != EventMessageReceived {
		t.Errorf("expected type %s, got %s", EventMessageReceived, event.Type)
	}

	if event.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	if event.Data == nil {
		t.Error("expected non-nil data map")
	}
}

func TestAgentEventWithMethods(t *testing.T) {
	event := NewEvent(EventToolCalled).
		WithSession("session-123").
		WithTrace("trace-456", "span-789").
		WithParentSpan("parent-000").
		WithData("name", "test-tool").
		WithData("params", map[string]any{"key": "value"})

	if event.SessionID != "session-123" {
		t.Errorf("expected session ID session-123, got %s", event.SessionID)
	}
	if event.TraceID != "trace-456" {
		t.Errorf("expected trace ID trace-456, got %s", event.TraceID)
	}
	if event.SpanID != "span-789" {
		t.Errorf("expected span ID span-789, got %s", event.SpanID)
	}
	if event.ParentSpanID != "parent-000" {
		t.Errorf("expected parent span ID parent-000, got %s", event.ParentSpanID)
	}
	if event.GetString("name") != "test-tool" {
		t.Errorf("expected name test-tool, got %s", event.GetString("name"))
	}
}

func TestAgentEventGetters(t *testing.T) {
	event := AgentEvent{
		Type:      EventMessageSent,
		Timestamp: time.Now(),
		Data: map[string]any{
			"string_val": "hello",
			"bool_val":   true,
			"map_val":    map[string]any{"nested": "value"},
			"int_val":    42,
			"empty_val":  "",
			"false_val":  false,
		},
	}

	// GetString tests
	if got := event.GetString("string_val"); got != "hello" {
		t.Errorf("GetString(string_val): expected hello, got %s", got)
	}
	if got := event.GetString("int_val"); got != "" {
		t.Errorf("GetString(int_val): expected empty, got %s", got)
	}
	if got := event.GetString("missing"); got != "" {
		t.Errorf("GetString(missing): expected empty, got %s", got)
	}

	// GetBool tests
	if got := event.GetBool("bool_val"); !got {
		t.Error("GetBool(bool_val): expected true")
	}
	if got := event.GetBool("false_val"); got {
		t.Error("GetBool(false_val): expected false")
	}
	if got := event.GetBool("string_val"); got {
		t.Error("GetBool(string_val): expected false for non-bool")
	}
	if got := event.GetBool("missing"); got {
		t.Error("GetBool(missing): expected false")
	}

	// GetMap tests
	if got := event.GetMap("map_val"); got == nil || got["nested"] != "value" {
		t.Errorf("GetMap(map_val): expected nested map, got %v", got)
	}
	if got := event.GetMap("string_val"); got != nil {
		t.Errorf("GetMap(string_val): expected nil, got %v", got)
	}
	if got := event.GetMap("missing"); got != nil {
		t.Errorf("GetMap(missing): expected nil, got %v", got)
	}
}

func TestAgentEventWithDataMap(t *testing.T) {
	event := NewEvent(EventSessionCreated).
		WithData("existing", "value").
		WithDataMap(map[string]any{
			"new1": "val1",
			"new2": "val2",
		})

	if event.GetString("existing") != "value" {
		t.Error("existing key should be preserved")
	}
	if event.GetString("new1") != "val1" {
		t.Error("new1 should be added")
	}
	if event.GetString("new2") != "val2" {
		t.Error("new2 should be added")
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventSessionCreated, "session.created"},
		{EventMessageReceived, "message.received"},
		{EventToolCalled, "tool.called"},
		{EventToolCompleted, "tool.completed"},
	}

	for _, tc := range tests {
		if got := tc.eventType.String(); got != tc.expected {
			t.Errorf("EventType.String(): expected %s, got %s", tc.expected, got)
		}
	}
}

func TestAgentEventNilData(t *testing.T) {
	event := AgentEvent{
		Type: EventError,
	}

	// Should not panic with nil data
	if got := event.GetString("any"); got != "" {
		t.Error("GetString on nil Data should return empty")
	}
	if got := event.GetBool("any"); got {
		t.Error("GetBool on nil Data should return false")
	}
	if got := event.GetMap("any"); got != nil {
		t.Error("GetMap on nil Data should return nil")
	}

	// WithData should initialize the map
	event = event.WithData("key", "value")
	if event.Data == nil {
		t.Error("WithData should initialize Data map")
	}
	if event.GetString("key") != "value" {
		t.Error("value should be set after WithData")
	}
}
