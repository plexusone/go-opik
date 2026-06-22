package agentobs

import (
	"context"
	"errors"
	"testing"
)

type recordingObserver struct {
	events []AgentEvent
	err    error
}

func (r *recordingObserver) OnEvent(_ context.Context, event AgentEvent) error {
	r.events = append(r.events, event)
	return r.err
}

func (r *recordingObserver) Flush(_ context.Context) error {
	return nil
}

func (r *recordingObserver) Close() error {
	return nil
}

func TestFilteredObserver(t *testing.T) {
	recorder := &recordingObserver{}
	filtered := NewFilteredObserver(recorder, FilterByEventType(EventMessageReceived, EventMessageSent))

	ctx := context.Background()

	// Should pass through message events
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived))
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageSent))

	// Should filter out other events
	_ = filtered.OnEvent(ctx, NewEvent(EventToolCalled))
	_ = filtered.OnEvent(ctx, NewEvent(EventSessionCreated))

	if len(recorder.events) != 2 {
		t.Errorf("expected 2 events, got %d", len(recorder.events))
	}
}

func TestFilterBySessionID(t *testing.T) {
	recorder := &recordingObserver{}
	filtered := NewFilteredObserver(recorder, FilterBySessionID("session-1"))

	ctx := context.Background()

	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived).WithSession("session-1"))
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived).WithSession("session-2"))
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived).WithSession("session-1"))

	if len(recorder.events) != 2 {
		t.Errorf("expected 2 events for session-1, got %d", len(recorder.events))
	}
}

func TestFilterExcludeEventType(t *testing.T) {
	recorder := &recordingObserver{}
	filtered := NewFilteredObserver(recorder, FilterExcludeEventType(EventJobExecuted))

	ctx := context.Background()

	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived))
	_ = filtered.OnEvent(ctx, NewEvent(EventJobExecuted))
	_ = filtered.OnEvent(ctx, NewEvent(EventToolCalled))

	if len(recorder.events) != 2 {
		t.Errorf("expected 2 events (excluding job), got %d", len(recorder.events))
	}
}

func TestFilterHasSessionID(t *testing.T) {
	recorder := &recordingObserver{}
	filtered := NewFilteredObserver(recorder, FilterHasSessionID())

	ctx := context.Background()

	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived).WithSession("session-1"))
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived)) // No session ID
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived).WithSession("session-2"))

	if len(recorder.events) != 2 {
		t.Errorf("expected 2 events with session IDs, got %d", len(recorder.events))
	}
}

func TestFilterHasTraceID(t *testing.T) {
	recorder := &recordingObserver{}
	filtered := NewFilteredObserver(recorder, FilterHasTraceID())

	ctx := context.Background()

	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived).WithTrace("trace-1", ""))
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived)) // No trace ID
	_ = filtered.OnEvent(ctx, NewEvent(EventMessageReceived).WithTrace("trace-2", "span-1"))

	if len(recorder.events) != 2 {
		t.Errorf("expected 2 events with trace IDs, got %d", len(recorder.events))
	}
}

func TestMultiObserver(t *testing.T) {
	r1 := &recordingObserver{}
	r2 := &recordingObserver{}
	multi := NewMultiObserver(r1, r2)

	ctx := context.Background()

	_ = multi.OnEvent(ctx, NewEvent(EventMessageReceived))
	_ = multi.OnEvent(ctx, NewEvent(EventToolCalled))

	if len(r1.events) != 2 {
		t.Errorf("r1: expected 2 events, got %d", len(r1.events))
	}
	if len(r2.events) != 2 {
		t.Errorf("r2: expected 2 events, got %d", len(r2.events))
	}
}

func TestMultiObserverWithError(t *testing.T) {
	testErr := errors.New("observer error")
	r1 := &recordingObserver{err: testErr}
	r2 := &recordingObserver{}
	multi := NewMultiObserver(r1, r2)

	ctx := context.Background()

	err := multi.OnEvent(ctx, NewEvent(EventMessageReceived))

	// Should return first error but still call all observers
	if err != testErr {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
	if len(r2.events) != 1 {
		t.Error("r2 should still receive events even when r1 errors")
	}
}

func TestNoOpObserver(t *testing.T) {
	noop := NoOpObserver{}
	ctx := context.Background()

	// Should not error
	if err := noop.OnEvent(ctx, NewEvent(EventMessageReceived)); err != nil {
		t.Errorf("OnEvent error: %v", err)
	}
	if err := noop.Flush(ctx); err != nil {
		t.Errorf("Flush error: %v", err)
	}
	if err := noop.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestFilteredObserverDelegates(t *testing.T) {
	recorder := &recordingObserver{}
	filtered := NewFilteredObserver(recorder)

	ctx := context.Background()

	// Test Flush and Close are delegated
	if err := filtered.Flush(ctx); err != nil {
		t.Errorf("Flush error: %v", err)
	}
	if err := filtered.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestMultiObserverDelegates(t *testing.T) {
	r1 := &recordingObserver{}
	r2 := &recordingObserver{}
	multi := NewMultiObserver(r1, r2)

	ctx := context.Background()

	if err := multi.Flush(ctx); err != nil {
		t.Errorf("Flush error: %v", err)
	}
	if err := multi.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}
