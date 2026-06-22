package agentobs

import (
	"context"
	"sync"
	"time"
)

// TraceClient is the interface that trace providers must implement.
// This allows the TraceManager to work with different backends.
type TraceClient interface {
	// CreateTrace creates a new trace.
	CreateTrace(ctx context.Context, name string, input any, tags []string) (Trace, error)

	// IsTracingEnabled returns whether tracing is active.
	IsTracingEnabled() bool
}

// Trace represents a trace in the observability backend.
type Trace interface {
	// ID returns the trace ID.
	ID() string

	// Name returns the trace name.
	Name() string

	// End ends the trace with optional output.
	End(ctx context.Context, output any) error

	// Update updates the trace metadata.
	Update(ctx context.Context, metadata map[string]any) error

	// CreateSpan creates a new span within this trace.
	CreateSpan(ctx context.Context, name string, spanType string, input any) (Span, error)
}

// Span represents a span within a trace.
type Span interface {
	// ID returns the span ID.
	ID() string

	// TraceID returns the parent trace ID.
	TraceID() string

	// Name returns the span name.
	Name() string

	// End ends the span with optional output.
	End(ctx context.Context, output any, err error) error

	// CreateChildSpan creates a child span.
	CreateChildSpan(ctx context.Context, name string, spanType string, input any) (Span, error)
}

// TraceState holds the state of an active trace.
type TraceState struct {
	Trace        Trace
	SessionID    string
	LastActivity time.Time
	OpenSpans    map[string]Span // spanID -> Span
	SpanStack    []string        // Stack of span IDs for nesting
	mu           sync.RWMutex
}

// NewTraceState creates a new trace state.
func NewTraceState(trace Trace, sessionID string) *TraceState {
	return &TraceState{
		Trace:        trace,
		SessionID:    sessionID,
		LastActivity: time.Now(),
		OpenSpans:    make(map[string]Span),
		SpanStack:    make([]string, 0),
	}
}

// Touch updates the last activity timestamp.
func (ts *TraceState) Touch() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.LastActivity = time.Now()
}

// IsStale returns true if the trace has been inactive longer than timeout.
func (ts *TraceState) IsStale(timeout time.Duration) bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return time.Since(ts.LastActivity) > timeout
}

// PushSpan adds a span to the stack.
func (ts *TraceState) PushSpan(span Span) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.OpenSpans[span.ID()] = span
	ts.SpanStack = append(ts.SpanStack, span.ID())
	ts.LastActivity = time.Now()
}

// PopSpan removes and returns the top span from the stack.
func (ts *TraceState) PopSpan() Span {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if len(ts.SpanStack) == 0 {
		return nil
	}
	spanID := ts.SpanStack[len(ts.SpanStack)-1]
	ts.SpanStack = ts.SpanStack[:len(ts.SpanStack)-1]
	span := ts.OpenSpans[spanID]
	delete(ts.OpenSpans, spanID)
	ts.LastActivity = time.Now()
	return span
}

// CurrentSpan returns the current (top) span without removing it.
func (ts *TraceState) CurrentSpan() Span {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if len(ts.SpanStack) == 0 {
		return nil
	}
	spanID := ts.SpanStack[len(ts.SpanStack)-1]
	return ts.OpenSpans[spanID]
}

// GetSpan returns a span by ID.
func (ts *TraceState) GetSpan(spanID string) Span {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.OpenSpans[spanID]
}

// RemoveSpan removes a specific span by ID.
func (ts *TraceState) RemoveSpan(spanID string) Span {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	span := ts.OpenSpans[spanID]
	delete(ts.OpenSpans, spanID)
	// Also remove from stack
	for i, id := range ts.SpanStack {
		if id == spanID {
			ts.SpanStack = append(ts.SpanStack[:i], ts.SpanStack[i+1:]...)
			break
		}
	}
	ts.LastActivity = time.Now()
	return span
}

// HasOpenSpans returns true if there are open spans.
func (ts *TraceState) HasOpenSpans() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.OpenSpans) > 0
}

// OpenSpanCount returns the number of open spans.
func (ts *TraceState) OpenSpanCount() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.OpenSpans)
}

// TraceManager manages the lifecycle of traces associated with agent sessions.
type TraceManager struct {
	client      TraceClient
	config      TraceManagerConfig
	traces      map[string]*TraceState // sessionID -> TraceState
	mu          sync.RWMutex
	sweepTicker *time.Ticker
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewTraceManager creates a new TraceManager.
func NewTraceManager(client TraceClient, opts ...TraceManagerOption) *TraceManager {
	cfg := DefaultTraceManagerConfig()
	cfg.ApplyOptions(opts...)

	tm := &TraceManager{
		client: client,
		config: cfg,
		traces: make(map[string]*TraceState),
		stopCh: make(chan struct{}),
	}

	// Start stale trace sweeper
	if cfg.SweepInterval > 0 {
		tm.sweepTicker = time.NewTicker(cfg.SweepInterval)
		tm.wg.Add(1)
		go tm.sweepLoop()
	}

	return tm
}

// sweepLoop periodically checks for and closes stale traces.
func (tm *TraceManager) sweepLoop() {
	defer tm.wg.Done()
	for {
		select {
		case <-tm.sweepTicker.C:
			tm.sweepStaleTraces()
		case <-tm.stopCh:
			return
		}
	}
}

// sweepStaleTraces closes traces that have been inactive too long.
func (tm *TraceManager) sweepStaleTraces() {
	tm.mu.Lock()
	var staleSessionIDs []string
	for sessionID, state := range tm.traces {
		if state.IsStale(tm.config.StaleTimeout) {
			staleSessionIDs = append(staleSessionIDs, sessionID)
		}
	}
	tm.mu.Unlock()

	// Close stale traces outside the lock
	ctx := context.Background()
	for _, sessionID := range staleSessionIDs {
		_ = tm.EndTrace(ctx, sessionID, map[string]any{
			"closed_reason": "stale",
			"stale_timeout": tm.config.StaleTimeout.String(),
		})
	}
}

// StartTrace creates a new trace for a session.
func (tm *TraceManager) StartTrace(ctx context.Context, sessionID string, agentName string, input any) (*TraceState, error) {
	if tm.client == nil || !tm.client.IsTracingEnabled() {
		return nil, nil
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check for existing trace
	if existing, ok := tm.traces[sessionID]; ok {
		existing.Touch()
		return existing, nil
	}

	// Create new trace
	traceName := tm.config.TraceNamePrefix + sessionID
	if agentName != "" {
		traceName = tm.config.TraceNamePrefix + agentName + "." + sessionID
	}

	trace, err := tm.client.CreateTrace(ctx, traceName, input, tm.config.DefaultTags)
	if err != nil {
		return nil, err
	}

	state := NewTraceState(trace, sessionID)
	tm.traces[sessionID] = state
	return state, nil
}

// GetTrace returns the trace state for a session.
func (tm *TraceManager) GetTrace(sessionID string) *TraceState {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.traces[sessionID]
}

// EndTrace ends a trace and removes it from management.
func (tm *TraceManager) EndTrace(ctx context.Context, sessionID string, output any) error {
	tm.mu.Lock()
	state, ok := tm.traces[sessionID]
	if !ok {
		tm.mu.Unlock()
		return nil
	}
	delete(tm.traces, sessionID)
	tm.mu.Unlock()

	// Close all open spans first
	for _, span := range state.OpenSpans {
		if err := span.End(ctx, nil, nil); err != nil {
			// Log but continue
		}
	}

	// End the trace
	return state.Trace.End(ctx, output)
}

// TouchTrace updates the last activity for a session's trace.
func (tm *TraceManager) TouchTrace(sessionID string) {
	tm.mu.RLock()
	state := tm.traces[sessionID]
	tm.mu.RUnlock()
	if state != nil {
		state.Touch()
	}
}

// StartSpan creates a new span within a session's trace.
func (tm *TraceManager) StartSpan(ctx context.Context, sessionID, spanName, spanType string, input any) (Span, error) {
	state := tm.GetTrace(sessionID)
	if state == nil {
		return nil, nil
	}

	state.Touch()

	var span Span
	var err error

	// If there's a current span, create a child span
	if currentSpan := state.CurrentSpan(); currentSpan != nil {
		span, err = currentSpan.CreateChildSpan(ctx, spanName, spanType, input)
	} else {
		// Otherwise create a span on the trace
		span, err = state.Trace.CreateSpan(ctx, spanName, spanType, input)
	}

	if err != nil {
		return nil, err
	}

	state.PushSpan(span)
	return span, nil
}

// EndSpan ends the current span for a session.
func (tm *TraceManager) EndSpan(ctx context.Context, sessionID string, output any, spanErr error) error {
	state := tm.GetTrace(sessionID)
	if state == nil {
		return nil
	}

	span := state.PopSpan()
	if span == nil {
		return nil
	}

	return span.End(ctx, output, spanErr)
}

// EndSpanByID ends a specific span by its ID.
func (tm *TraceManager) EndSpanByID(ctx context.Context, sessionID, spanID string, output any, spanErr error) error {
	state := tm.GetTrace(sessionID)
	if state == nil {
		return nil
	}

	span := state.RemoveSpan(spanID)
	if span == nil {
		return nil
	}

	return span.End(ctx, output, spanErr)
}

// GetCurrentSpan returns the current span for a session.
func (tm *TraceManager) GetCurrentSpan(sessionID string) Span {
	state := tm.GetTrace(sessionID)
	if state == nil {
		return nil
	}
	return state.CurrentSpan()
}

// ActiveTraceCount returns the number of active traces.
func (tm *TraceManager) ActiveTraceCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.traces)
}

// ActiveSessionIDs returns the session IDs with active traces.
func (tm *TraceManager) ActiveSessionIDs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	ids := make([]string, 0, len(tm.traces))
	for id := range tm.traces {
		ids = append(ids, id)
	}
	return ids
}

// Close stops the manager and closes all active traces.
func (tm *TraceManager) Close() error {
	// Stop the sweep loop
	close(tm.stopCh)
	if tm.sweepTicker != nil {
		tm.sweepTicker.Stop()
	}
	tm.wg.Wait()

	// Close all active traces
	ctx := context.Background()
	tm.mu.Lock()
	sessionIDs := make([]string, 0, len(tm.traces))
	for id := range tm.traces {
		sessionIDs = append(sessionIDs, id)
	}
	tm.mu.Unlock()

	for _, sessionID := range sessionIDs {
		_ = tm.EndTrace(ctx, sessionID, map[string]any{
			"closed_reason": "shutdown",
		})
	}

	return nil
}

// SanitizeContent sanitizes content if configured.
func (tm *TraceManager) SanitizeContent(content string) string {
	if !tm.config.SanitizeContent {
		return content
	}
	return Sanitize(content, tm.config.SanitizeConfig)
}

// SanitizeData sanitizes map data if configured.
func (tm *TraceManager) SanitizeData(data map[string]any) map[string]any {
	if !tm.config.SanitizeContent {
		return data
	}
	return SanitizeMap(data, tm.config.SanitizeConfig)
}

// ExtractMedia extracts media references if configured.
func (tm *TraceManager) ExtractMedia(content string) []MediaRef {
	if !tm.config.ExtractMedia {
		return nil
	}
	return ExtractMedia(content)
}
