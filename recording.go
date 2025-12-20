package opik

import (
	"context"
	"sync"
	"time"
)

// RecordedTrace represents a trace captured during local recording.
type RecordedTrace struct {
	ID        string
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Input     any
	Output    any
	Metadata  map[string]any
	Tags      []string
	Spans     []*RecordedSpan
	Feedback  []*RecordedFeedback
}

// RecordedSpan represents a span captured during local recording.
type RecordedSpan struct {
	ID           string
	TraceID      string
	ParentSpanID string
	Name         string
	Type         string
	StartTime    time.Time
	EndTime      time.Time
	Input        any
	Output       any
	Metadata     map[string]any
	Tags         []string
	Model        string
	Provider     string
	Children     []*RecordedSpan
	Feedback     []*RecordedFeedback
}

// RecordedFeedback represents a feedback score captured during local recording.
type RecordedFeedback struct {
	Name   string
	Value  float64
	Reason string
}

// LocalRecording captures traces and spans locally without sending to the server.
type LocalRecording struct {
	mu       sync.RWMutex
	traces   map[string]*RecordedTrace
	spans    map[string]*RecordedSpan
	feedback []RecordedFeedback
}

// NewLocalRecording creates a new local recording storage.
func NewLocalRecording() *LocalRecording {
	return &LocalRecording{
		traces: make(map[string]*RecordedTrace),
		spans:  make(map[string]*RecordedSpan),
	}
}

// AddTrace adds a trace to the recording.
func (r *LocalRecording) AddTrace(trace *RecordedTrace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.traces[trace.ID] = trace
}

// AddSpan adds a span to the recording.
func (r *LocalRecording) AddSpan(span *RecordedSpan) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans[span.ID] = span

	// Also add to parent trace
	if trace, ok := r.traces[span.TraceID]; ok {
		if span.ParentSpanID == "" {
			trace.Spans = append(trace.Spans, span)
		}
	}

	// Add to parent span if exists
	if span.ParentSpanID != "" {
		if parent, ok := r.spans[span.ParentSpanID]; ok {
			parent.Children = append(parent.Children, span)
		}
	}
}

// AddFeedback adds feedback to the recording.
func (r *LocalRecording) AddFeedback(entityID string, feedback RecordedFeedback) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Try trace first
	if trace, ok := r.traces[entityID]; ok {
		trace.Feedback = append(trace.Feedback, &feedback)
		return
	}

	// Try span
	if span, ok := r.spans[entityID]; ok {
		span.Feedback = append(span.Feedback, &feedback)
		return
	}

	// Store as orphan
	r.feedback = append(r.feedback, feedback)
}

// Traces returns all recorded traces.
func (r *LocalRecording) Traces() []*RecordedTrace {
	r.mu.RLock()
	defer r.mu.RUnlock()

	traces := make([]*RecordedTrace, 0, len(r.traces))
	for _, t := range r.traces {
		traces = append(traces, t)
	}
	return traces
}

// Spans returns all recorded spans.
func (r *LocalRecording) Spans() []*RecordedSpan {
	r.mu.RLock()
	defer r.mu.RUnlock()

	spans := make([]*RecordedSpan, 0, len(r.spans))
	for _, s := range r.spans {
		spans = append(spans, s)
	}
	return spans
}

// GetTrace returns a specific trace by ID.
func (r *LocalRecording) GetTrace(id string) *RecordedTrace {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.traces[id]
}

// GetSpan returns a specific span by ID.
func (r *LocalRecording) GetSpan(id string) *RecordedSpan {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.spans[id]
}

// Clear clears all recorded data.
func (r *LocalRecording) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.traces = make(map[string]*RecordedTrace)
	r.spans = make(map[string]*RecordedSpan)
	r.feedback = nil
}

// TraceCount returns the number of recorded traces.
func (r *LocalRecording) TraceCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.traces)
}

// SpanCount returns the number of recorded spans.
func (r *LocalRecording) SpanCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.spans)
}

// RecordingClient is a client that records traces locally instead of sending to server.
type RecordingClient struct {
	recording *LocalRecording
	project   string
}

// NewRecordingClient creates a new recording client for local testing.
func NewRecordingClient(projectName string) *RecordingClient {
	return &RecordingClient{
		recording: NewLocalRecording(),
		project:   projectName,
	}
}

// Recording returns the local recording storage.
func (c *RecordingClient) Recording() *LocalRecording {
	return c.recording
}

// Trace creates a new trace and records it locally.
func (c *RecordingClient) Trace(ctx context.Context, name string, opts ...TraceOption) (*RecordingTrace, error) {
	options := &traceOptions{
		metadata: make(map[string]any),
	}
	for _, opt := range opts {
		opt(options)
	}

	trace := &RecordedTrace{
		ID:        generateID(),
		Name:      name,
		StartTime: time.Now(),
		Input:     options.input,
		Metadata:  options.metadata,
		Tags:      options.tags,
		Spans:     make([]*RecordedSpan, 0),
		Feedback:  make([]*RecordedFeedback, 0),
	}

	c.recording.AddTrace(trace)

	return &RecordingTrace{
		client: c,
		trace:  trace,
	}, nil
}

// RecordingTrace is a trace that records locally.
type RecordingTrace struct {
	client *RecordingClient
	trace  *RecordedTrace
}

// ID returns the trace ID.
func (t *RecordingTrace) ID() string {
	return t.trace.ID
}

// Name returns the trace name.
func (t *RecordingTrace) Name() string {
	return t.trace.Name
}

// End ends the trace.
func (t *RecordingTrace) End(ctx context.Context, opts ...TraceOption) error {
	options := &traceOptions{}
	for _, opt := range opts {
		opt(options)
	}

	t.trace.EndTime = time.Now()
	if options.output != nil {
		t.trace.Output = options.output
	}

	return nil
}

// Span creates a new span under this trace.
func (t *RecordingTrace) Span(ctx context.Context, name string, opts ...SpanOption) (*RecordingSpan, error) {
	options := &spanOptions{
		spanType: SpanTypeGeneral,
		metadata: make(map[string]any),
	}
	for _, opt := range opts {
		opt(options)
	}

	span := &RecordedSpan{
		ID:        generateID(),
		TraceID:   t.trace.ID,
		Name:      name,
		Type:      options.spanType,
		StartTime: time.Now(),
		Input:     options.input,
		Metadata:  options.metadata,
		Tags:      options.tags,
		Model:     options.model,
		Provider:  options.provider,
		Children:  make([]*RecordedSpan, 0),
		Feedback:  make([]*RecordedFeedback, 0),
	}

	t.client.recording.AddSpan(span)

	return &RecordingSpan{
		client: t.client,
		span:   span,
	}, nil
}

// AddFeedbackScore adds a feedback score to this trace.
func (t *RecordingTrace) AddFeedbackScore(ctx context.Context, name string, value float64, reason string) error {
	t.client.recording.AddFeedback(t.trace.ID, RecordedFeedback{
		Name:   name,
		Value:  value,
		Reason: reason,
	})
	return nil
}

// RecordingSpan is a span that records locally.
type RecordingSpan struct {
	client *RecordingClient
	span   *RecordedSpan
}

// ID returns the span ID.
func (s *RecordingSpan) ID() string {
	return s.span.ID
}

// TraceID returns the trace ID.
func (s *RecordingSpan) TraceID() string {
	return s.span.TraceID
}

// Name returns the span name.
func (s *RecordingSpan) Name() string {
	return s.span.Name
}

// End ends the span.
func (s *RecordingSpan) End(ctx context.Context, opts ...SpanOption) error {
	options := &spanOptions{}
	for _, opt := range opts {
		opt(options)
	}

	s.span.EndTime = time.Now()
	if options.output != nil {
		s.span.Output = options.output
	}

	return nil
}

// Span creates a child span.
func (s *RecordingSpan) Span(ctx context.Context, name string, opts ...SpanOption) (*RecordingSpan, error) {
	options := &spanOptions{
		spanType: SpanTypeGeneral,
		metadata: make(map[string]any),
	}
	for _, opt := range opts {
		opt(options)
	}

	span := &RecordedSpan{
		ID:           generateID(),
		TraceID:      s.span.TraceID,
		ParentSpanID: s.span.ID,
		Name:         name,
		Type:         options.spanType,
		StartTime:    time.Now(),
		Input:        options.input,
		Metadata:     options.metadata,
		Tags:         options.tags,
		Model:        options.model,
		Provider:     options.provider,
		Children:     make([]*RecordedSpan, 0),
		Feedback:     make([]*RecordedFeedback, 0),
	}

	s.client.recording.AddSpan(span)

	return &RecordingSpan{
		client: s.client,
		span:   span,
	}, nil
}

// AddFeedbackScore adds a feedback score to this span.
func (s *RecordingSpan) AddFeedbackScore(ctx context.Context, name string, value float64, reason string) error {
	s.client.recording.AddFeedback(s.span.ID, RecordedFeedback{
		Name:   name,
		Value:  value,
		Reason: reason,
	})
	return nil
}

// Helper function to generate IDs
func generateID() string {
	return generateUUID()
}

func generateUUID() string {
	// Use google/uuid if available, otherwise simple implementation
	b := make([]byte, 16)
	// This is a simple implementation - in production use crypto/rand
	for i := range b {
		b[i] = byte(time.Now().UnixNano() >> (i * 4))
	}
	return formatUUID(b)
}

func formatUUID(b []byte) string {
	return string(hexEncode(b[0:4])) + "-" +
		string(hexEncode(b[4:6])) + "-" +
		string(hexEncode(b[6:8])) + "-" +
		string(hexEncode(b[8:10])) + "-" +
		string(hexEncode(b[10:16]))
}

func hexEncode(b []byte) []byte {
	const hex = "0123456789abcdef"
	dst := make([]byte, len(b)*2)
	for i, v := range b {
		dst[i*2] = hex[v>>4]
		dst[i*2+1] = hex[v&0x0f]
	}
	return dst
}

// RecordTracesLocally returns a recording client for local testing.
// Usage:
//
//	client := opik.RecordTracesLocally("my-project")
//	trace, _ := client.Trace(ctx, "test-trace")
//	// ... do work ...
//	trace.End(ctx)
//	traces := client.Recording().Traces()
func RecordTracesLocally(projectName string) *RecordingClient {
	return NewRecordingClient(projectName)
}
