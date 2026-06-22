package agentobs

import "time"

// EventType defines the type of agent lifecycle event.
type EventType string

// Event type constants matching common agent framework events.
const (
	// EventSessionCreated indicates a new agent session was created.
	EventSessionCreated EventType = "session.created"

	// EventSessionUpdated indicates an existing session was updated.
	EventSessionUpdated EventType = "session.updated"

	// EventSessionClosed indicates a session was explicitly closed.
	EventSessionClosed EventType = "session.closed"

	// EventMessageReceived indicates a user message was received.
	EventMessageReceived EventType = "message.received"

	// EventMessageSent indicates an assistant message was sent.
	EventMessageSent EventType = "message.sent"

	// EventToolCalled indicates a tool invocation started.
	EventToolCalled EventType = "tool.called"

	// EventToolCompleted indicates a tool invocation completed.
	EventToolCompleted EventType = "tool.completed"

	// EventSubagentStarted indicates a subagent was spawned.
	EventSubagentStarted EventType = "subagent.started"

	// EventSubagentCompleted indicates a subagent completed its task.
	EventSubagentCompleted EventType = "subagent.completed"

	// EventJobExecuted indicates a scheduled job was executed.
	EventJobExecuted EventType = "job.executed"

	// EventError indicates an error occurred.
	EventError EventType = "error"
)

// String returns the string representation of the event type.
func (e EventType) String() string {
	return string(e)
}

// AgentEvent represents a generic agent lifecycle event.
// Framework integrations convert their native events into this generic format.
type AgentEvent struct {
	// Type identifies the kind of event.
	Type EventType `json:"type"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// SessionID identifies the agent session this event belongs to.
	SessionID string `json:"session_id,omitempty"`

	// TraceID is the Opik trace ID for correlation.
	TraceID string `json:"trace_id,omitempty"`

	// SpanID is the Opik span ID for this specific event.
	SpanID string `json:"span_id,omitempty"`

	// ParentSpanID is the parent span ID for nested operations.
	ParentSpanID string `json:"parent_span_id,omitempty"`

	// Data contains event-specific payload.
	// The structure depends on the event type.
	Data map[string]any `json:"data,omitempty"`
}

// NewEvent creates a new AgentEvent with the given type and current timestamp.
func NewEvent(eventType EventType) AgentEvent {
	return AgentEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      make(map[string]any),
	}
}

// WithSession sets the session ID and returns the event.
func (e AgentEvent) WithSession(sessionID string) AgentEvent {
	e.SessionID = sessionID
	return e
}

// WithTrace sets the trace and span IDs and returns the event.
func (e AgentEvent) WithTrace(traceID, spanID string) AgentEvent {
	e.TraceID = traceID
	e.SpanID = spanID
	return e
}

// WithParentSpan sets the parent span ID and returns the event.
func (e AgentEvent) WithParentSpan(parentSpanID string) AgentEvent {
	e.ParentSpanID = parentSpanID
	return e
}

// WithData sets a data field and returns the event.
func (e AgentEvent) WithData(key string, value any) AgentEvent {
	if e.Data == nil {
		e.Data = make(map[string]any)
	}
	e.Data[key] = value
	return e
}

// WithDataMap merges a map into the event data and returns the event.
func (e AgentEvent) WithDataMap(data map[string]any) AgentEvent {
	if e.Data == nil {
		e.Data = make(map[string]any)
	}
	for k, v := range data {
		e.Data[k] = v
	}
	return e
}

// GetString returns a string value from Data, or empty string if not found or not a string.
func (e AgentEvent) GetString(key string) string {
	if e.Data == nil {
		return ""
	}
	if v, ok := e.Data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetBool returns a bool value from Data, or false if not found or not a bool.
func (e AgentEvent) GetBool(key string) bool {
	if e.Data == nil {
		return false
	}
	if v, ok := e.Data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetMap returns a map value from Data, or nil if not found or not a map.
func (e AgentEvent) GetMap(key string) map[string]any {
	if e.Data == nil {
		return nil
	}
	if v, ok := e.Data[key]; ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return nil
}
