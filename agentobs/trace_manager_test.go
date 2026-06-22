package agentobs

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// Mock implementations for testing

type mockTrace struct {
	id       string
	name     string
	ended    bool
	output   any
	metadata map[string]any
	spans    []*mockSpan
	mu       sync.Mutex
}

func (t *mockTrace) ID() string   { return t.id }
func (t *mockTrace) Name() string { return t.name }

func (t *mockTrace) End(_ context.Context, output any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ended = true
	t.output = output
	return nil
}

func (t *mockTrace) Update(_ context.Context, metadata map[string]any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.metadata == nil {
		t.metadata = make(map[string]any)
	}
	for k, v := range metadata {
		t.metadata[k] = v
	}
	return nil
}

func (t *mockTrace) CreateSpan(_ context.Context, name string, spanType string, input any) (Span, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	span := &mockSpan{
		id:       "span-" + name,
		traceID:  t.id,
		name:     name,
		spanType: spanType,
		input:    input,
	}
	t.spans = append(t.spans, span)
	return span, nil
}

type mockSpan struct {
	id       string
	traceID  string
	name     string
	spanType string
	input    any
	output   any
	ended    bool
	err      error
	children []*mockSpan
	mu       sync.Mutex
}

func (s *mockSpan) ID() string      { return s.id }
func (s *mockSpan) TraceID() string { return s.traceID }
func (s *mockSpan) Name() string    { return s.name }

func (s *mockSpan) End(_ context.Context, output any, spanErr error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ended = true
	s.output = output
	s.err = spanErr
	return nil
}

func (s *mockSpan) CreateChildSpan(_ context.Context, name string, spanType string, input any) (Span, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	child := &mockSpan{
		id:       "child-" + name,
		traceID:  s.traceID,
		name:     name,
		spanType: spanType,
		input:    input,
	}
	s.children = append(s.children, child)
	return child, nil
}

type mockTraceClient struct {
	traces         []*mockTrace
	tracingEnabled bool
	mu             sync.Mutex
}

func newMockTraceClient() *mockTraceClient {
	return &mockTraceClient{
		tracingEnabled: true,
	}
}

func (c *mockTraceClient) CreateTrace(_ context.Context, name string, input any, _ []string) (Trace, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	trace := &mockTrace{
		id:   "trace-" + name,
		name: name,
	}
	c.traces = append(c.traces, trace)
	return trace, nil
}

func (c *mockTraceClient) IsTracingEnabled() bool {
	return c.tracingEnabled
}

// Tests

func TestTraceManagerStartTrace(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client,
		WithStaleTimeout(5*time.Minute),
		WithSweepInterval(0), // Disable sweep for testing
	)
	defer manager.Close()

	ctx := context.Background()
	state, err := manager.StartTrace(ctx, "session-1", "test-agent", nil)
	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	if state == nil {
		t.Fatal("expected non-nil trace state")
	}
	if state.SessionID != "session-1" {
		t.Errorf("expected session ID session-1, got %s", state.SessionID)
	}
	if manager.ActiveTraceCount() != 1 {
		t.Errorf("expected 1 active trace, got %d", manager.ActiveTraceCount())
	}

	// Starting another trace for same session should return existing
	state2, err := manager.StartTrace(ctx, "session-1", "test-agent", nil)
	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}
	if state2 != state {
		t.Error("expected same trace state for same session")
	}
	if manager.ActiveTraceCount() != 1 {
		t.Errorf("expected still 1 active trace, got %d", manager.ActiveTraceCount())
	}
}

func TestTraceManagerEndTrace(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client, WithSweepInterval(0))
	defer manager.Close()

	ctx := context.Background()
	_, err := manager.StartTrace(ctx, "session-1", "agent", nil)
	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	err = manager.EndTrace(ctx, "session-1", map[string]any{"result": "done"})
	if err != nil {
		t.Fatalf("EndTrace failed: %v", err)
	}

	if manager.ActiveTraceCount() != 0 {
		t.Errorf("expected 0 active traces, got %d", manager.ActiveTraceCount())
	}

	// Verify trace was ended
	if len(client.traces) != 1 {
		t.Fatal("expected 1 trace")
	}
	if !client.traces[0].ended {
		t.Error("expected trace to be ended")
	}
}

func TestTraceManagerSpans(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client, WithSweepInterval(0))
	defer manager.Close()

	ctx := context.Background()
	_, err := manager.StartTrace(ctx, "session-1", "agent", nil)
	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	// Start a span
	span1, err := manager.StartSpan(ctx, "session-1", "tool-1", "tool", map[string]any{"input": "test"})
	if err != nil {
		t.Fatalf("StartSpan failed: %v", err)
	}
	if span1 == nil {
		t.Fatal("expected non-nil span")
	}

	// Check current span
	current := manager.GetCurrentSpan("session-1")
	if current != span1 {
		t.Error("expected span1 to be current")
	}

	// Start nested span
	span2, err := manager.StartSpan(ctx, "session-1", "nested", "general", nil)
	if err != nil {
		t.Fatalf("StartSpan failed: %v", err)
	}

	current = manager.GetCurrentSpan("session-1")
	if current != span2 {
		t.Error("expected span2 to be current")
	}

	// End nested span
	err = manager.EndSpan(ctx, "session-1", "nested-output", nil)
	if err != nil {
		t.Fatalf("EndSpan failed: %v", err)
	}

	// Current should be back to span1
	current = manager.GetCurrentSpan("session-1")
	if current != span1 {
		t.Error("expected span1 to be current after ending span2")
	}

	// End span1
	err = manager.EndSpan(ctx, "session-1", "tool-output", nil)
	if err != nil {
		t.Fatalf("EndSpan failed: %v", err)
	}

	// No current span
	current = manager.GetCurrentSpan("session-1")
	if current != nil {
		t.Error("expected no current span")
	}
}

func TestTraceManagerEndSpanByID(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client, WithSweepInterval(0))
	defer manager.Close()

	ctx := context.Background()
	_, err := manager.StartTrace(ctx, "session-1", "agent", nil)
	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	span, err := manager.StartSpan(ctx, "session-1", "tool-1", "tool", nil)
	if err != nil {
		t.Fatalf("StartSpan failed: %v", err)
	}

	spanID := span.ID()
	err = manager.EndSpanByID(ctx, "session-1", spanID, "output", nil)
	if err != nil {
		t.Fatalf("EndSpanByID failed: %v", err)
	}

	// Span should be ended and removed
	mockSpan := span.(*mockSpan)
	if !mockSpan.ended {
		t.Error("expected span to be ended")
	}

	state := manager.GetTrace("session-1")
	if state.HasOpenSpans() {
		t.Error("expected no open spans")
	}
}

func TestTraceStateIsStale(t *testing.T) {
	trace := &mockTrace{id: "test"}
	state := NewTraceState(trace, "session-1")

	// Should not be stale immediately
	if state.IsStale(100 * time.Millisecond) {
		t.Error("should not be stale immediately")
	}

	// Wait and check
	time.Sleep(150 * time.Millisecond)
	if !state.IsStale(100 * time.Millisecond) {
		t.Error("should be stale after timeout")
	}

	// Touch should reset
	state.Touch()
	if state.IsStale(100 * time.Millisecond) {
		t.Error("should not be stale after touch")
	}
}

func TestTraceManagerStaleCleanup(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client,
		WithStaleTimeout(50*time.Millisecond),
		WithSweepInterval(25*time.Millisecond),
	)
	defer manager.Close()

	ctx := context.Background()
	_, err := manager.StartTrace(ctx, "session-1", "agent", nil)
	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	if manager.ActiveTraceCount() != 1 {
		t.Fatal("expected 1 active trace")
	}

	// Wait for stale cleanup
	time.Sleep(150 * time.Millisecond)

	if manager.ActiveTraceCount() != 0 {
		t.Errorf("expected 0 active traces after cleanup, got %d", manager.ActiveTraceCount())
	}

	// Verify trace was ended with stale reason
	if len(client.traces) != 1 {
		t.Fatal("expected 1 trace")
	}
	if !client.traces[0].ended {
		t.Error("expected trace to be ended")
	}
}

func TestTraceManagerClose(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client, WithSweepInterval(0))

	ctx := context.Background()
	_, _ = manager.StartTrace(ctx, "session-1", "agent", nil)
	_, _ = manager.StartTrace(ctx, "session-2", "agent", nil)

	if manager.ActiveTraceCount() != 2 {
		t.Fatal("expected 2 active traces")
	}

	err := manager.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// All traces should be ended
	for _, trace := range client.traces {
		if !trace.ended {
			t.Error("expected all traces to be ended on close")
		}
	}
}

func TestTraceManagerTracingDisabled(t *testing.T) {
	client := newMockTraceClient()
	client.tracingEnabled = false
	manager := NewTraceManager(client, WithSweepInterval(0))
	defer manager.Close()

	ctx := context.Background()
	state, err := manager.StartTrace(ctx, "session-1", "agent", nil)
	if err != nil {
		t.Fatalf("StartTrace failed: %v", err)
	}

	// Should return nil when tracing disabled
	if state != nil {
		t.Error("expected nil state when tracing disabled")
	}
}

func TestTraceManagerSanitization(t *testing.T) {
	manager := NewTraceManager(nil,
		WithSanitization(true),
		WithSanitizeConfig(SanitizeConfig{
			RemoveSecrets: true,
		}),
	)
	defer manager.Close()

	content := "api_key: sk-secret123456789012"
	sanitized := manager.SanitizeContent(content)

	if sanitized == content {
		t.Error("expected content to be sanitized")
	}
}

func TestTraceManagerMediaExtraction(t *testing.T) {
	manager := NewTraceManager(nil,
		WithMediaExtraction(true),
	)
	defer manager.Close()

	content := "![test](https://example.com/image.png)"
	refs := manager.ExtractMedia(content)

	if len(refs) != 1 {
		t.Errorf("expected 1 media ref, got %d", len(refs))
	}

	// Disable extraction
	manager.config.ExtractMedia = false
	refs = manager.ExtractMedia(content)
	if refs != nil {
		t.Error("expected nil when extraction disabled")
	}
}

func TestTraceStateSpanStack(t *testing.T) {
	trace := &mockTrace{id: "test"}
	state := NewTraceState(trace, "session")

	span1 := &mockSpan{id: "span-1", name: "first"}
	span2 := &mockSpan{id: "span-2", name: "second"}
	span3 := &mockSpan{id: "span-3", name: "third"}

	state.PushSpan(span1)
	state.PushSpan(span2)
	state.PushSpan(span3)

	if state.OpenSpanCount() != 3 {
		t.Errorf("expected 3 open spans, got %d", state.OpenSpanCount())
	}

	// Pop should return in LIFO order
	popped := state.PopSpan()
	if popped != span3 {
		t.Error("expected span3 from pop")
	}

	popped = state.PopSpan()
	if popped != span2 {
		t.Error("expected span2 from pop")
	}

	// Get specific span
	got := state.GetSpan("span-1")
	if got != span1 {
		t.Error("expected span1 from GetSpan")
	}

	// Remove specific span
	removed := state.RemoveSpan("span-1")
	if removed != span1 {
		t.Error("expected span1 from RemoveSpan")
	}

	if state.HasOpenSpans() {
		t.Error("expected no open spans")
	}

	// Pop from empty should return nil
	popped = state.PopSpan()
	if popped != nil {
		t.Error("expected nil from empty stack")
	}
}

func TestTraceManagerActiveSessionIDs(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client, WithSweepInterval(0))
	defer manager.Close()

	ctx := context.Background()
	_, _ = manager.StartTrace(ctx, "alpha", "agent", nil)
	_, _ = manager.StartTrace(ctx, "beta", "agent", nil)
	_, _ = manager.StartTrace(ctx, "gamma", "agent", nil)

	ids := manager.ActiveSessionIDs()
	if len(ids) != 3 {
		t.Fatalf("expected 3 session IDs, got %d", len(ids))
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	if !idSet["alpha"] || !idSet["beta"] || !idSet["gamma"] {
		t.Error("missing expected session IDs")
	}
}

func TestTraceManagerMissingSession(t *testing.T) {
	client := newMockTraceClient()
	manager := NewTraceManager(client, WithSweepInterval(0))
	defer manager.Close()

	ctx := context.Background()

	// Operations on non-existent session should not error
	span, err := manager.StartSpan(ctx, "missing", "test", "general", nil)
	if err != nil {
		t.Errorf("StartSpan on missing session should not error: %v", err)
	}
	if span != nil {
		t.Error("expected nil span for missing session")
	}

	err = manager.EndSpan(ctx, "missing", nil, nil)
	if err != nil {
		t.Errorf("EndSpan on missing session should not error: %v", err)
	}

	err = manager.EndTrace(ctx, "missing", nil)
	if err != nil {
		t.Errorf("EndTrace on missing session should not error: %v", err)
	}

	current := manager.GetCurrentSpan("missing")
	if current != nil {
		t.Error("expected nil current span for missing session")
	}
}

func TestTraceStateSpanWithError(t *testing.T) {
	trace := &mockTrace{id: "test"}
	state := NewTraceState(trace, "session")

	span := &mockSpan{id: "span-1"}
	state.PushSpan(span)

	// Simulate ending with error
	testErr := errors.New("tool failed")
	state.PopSpan()

	// The span itself would be ended by the manager
	_ = span.End(context.Background(), nil, testErr)

	if span.err != testErr {
		t.Error("expected error to be recorded on span")
	}
}
