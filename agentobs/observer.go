package agentobs

import "context"

// AgentObserver defines the interface for agent observability backends.
// Implementations receive generic agent events and translate them to
// backend-specific telemetry data.
type AgentObserver interface {
	// OnEvent processes an agent event.
	// The implementation should translate the event to the appropriate
	// backend-specific operations (e.g., creating traces, spans, etc.)
	OnEvent(ctx context.Context, event AgentEvent) error

	// Flush ensures all pending events are sent to the backend.
	Flush(ctx context.Context) error

	// Close releases resources and stops the observer.
	Close() error
}

// EventFilter defines a function that filters events.
// Returns true if the event should be processed, false to skip.
type EventFilter func(event AgentEvent) bool

// FilteredObserver wraps an observer with event filtering.
type FilteredObserver struct {
	observer AgentObserver
	filters  []EventFilter
}

// NewFilteredObserver creates an observer that only processes events
// matching all provided filters.
func NewFilteredObserver(observer AgentObserver, filters ...EventFilter) *FilteredObserver {
	return &FilteredObserver{
		observer: observer,
		filters:  filters,
	}
}

// OnEvent processes the event if it passes all filters.
func (f *FilteredObserver) OnEvent(ctx context.Context, event AgentEvent) error {
	for _, filter := range f.filters {
		if !filter(event) {
			return nil // Event filtered out
		}
	}
	return f.observer.OnEvent(ctx, event)
}

// Flush delegates to the underlying observer.
func (f *FilteredObserver) Flush(ctx context.Context) error {
	return f.observer.Flush(ctx)
}

// Close delegates to the underlying observer.
func (f *FilteredObserver) Close() error {
	return f.observer.Close()
}

// Common event filters.

// FilterByEventType creates a filter that only passes specified event types.
func FilterByEventType(types ...EventType) EventFilter {
	typeSet := make(map[EventType]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	return func(event AgentEvent) bool {
		return typeSet[event.Type]
	}
}

// FilterBySessionID creates a filter that only passes events for a specific session.
func FilterBySessionID(sessionID string) EventFilter {
	return func(event AgentEvent) bool {
		return event.SessionID == sessionID
	}
}

// FilterExcludeEventType creates a filter that excludes specified event types.
func FilterExcludeEventType(types ...EventType) EventFilter {
	typeSet := make(map[EventType]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	return func(event AgentEvent) bool {
		return !typeSet[event.Type]
	}
}

// FilterHasSessionID filters to events that have a session ID.
func FilterHasSessionID() EventFilter {
	return func(event AgentEvent) bool {
		return event.SessionID != ""
	}
}

// FilterHasTraceID filters to events that have a trace ID.
func FilterHasTraceID() EventFilter {
	return func(event AgentEvent) bool {
		return event.TraceID != ""
	}
}

// MultiObserver broadcasts events to multiple observers.
type MultiObserver struct {
	observers []AgentObserver
}

// NewMultiObserver creates an observer that sends events to all provided observers.
func NewMultiObserver(observers ...AgentObserver) *MultiObserver {
	return &MultiObserver{observers: observers}
}

// OnEvent sends the event to all observers.
// Returns the first error encountered, but continues sending to remaining observers.
func (m *MultiObserver) OnEvent(ctx context.Context, event AgentEvent) error {
	var firstErr error
	for _, obs := range m.observers {
		if err := obs.OnEvent(ctx, event); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Flush flushes all observers.
func (m *MultiObserver) Flush(ctx context.Context) error {
	var firstErr error
	for _, obs := range m.observers {
		if err := obs.Flush(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Close closes all observers.
func (m *MultiObserver) Close() error {
	var firstErr error
	for _, obs := range m.observers {
		if err := obs.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// NoOpObserver is an observer that does nothing.
// Useful for testing or disabling observability.
type NoOpObserver struct{}

// OnEvent does nothing.
func (NoOpObserver) OnEvent(_ context.Context, _ AgentEvent) error { return nil }

// Flush does nothing.
func (NoOpObserver) Flush(_ context.Context) error { return nil }

// Close does nothing.
func (NoOpObserver) Close() error { return nil }
