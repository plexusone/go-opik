package opik

import (
	"context"
	"strings"
	"sync"
	"time"
)

// StreamChunk represents a chunk of streaming data.
type StreamChunk struct {
	Content      string
	Index        int
	IsFirst      bool
	IsLast       bool
	Timestamp    time.Time
	TokenCount   int
	FinishReason string
	Metadata     map[string]any
}

// StreamAccumulator accumulates streaming chunks into a complete response.
type StreamAccumulator struct {
	mu           sync.Mutex
	chunks       []StreamChunk
	content      strings.Builder
	totalTokens  int
	firstChunk   time.Time
	lastChunk    time.Time
	finishReason string
	metadata     map[string]any
}

// NewStreamAccumulator creates a new stream accumulator.
func NewStreamAccumulator() *StreamAccumulator {
	return &StreamAccumulator{
		metadata: make(map[string]any),
	}
}

// AddChunk adds a chunk to the accumulator.
func (a *StreamAccumulator) AddChunk(chunk StreamChunk) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.chunks) == 0 {
		a.firstChunk = chunk.Timestamp
	}
	a.lastChunk = chunk.Timestamp

	a.chunks = append(a.chunks, chunk)
	a.content.WriteString(chunk.Content)
	a.totalTokens += chunk.TokenCount

	if chunk.FinishReason != "" {
		a.finishReason = chunk.FinishReason
	}

	for k, v := range chunk.Metadata {
		a.metadata[k] = v
	}
}

// Content returns the accumulated content.
func (a *StreamAccumulator) Content() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.content.String()
}

// TotalTokens returns the total token count.
func (a *StreamAccumulator) TotalTokens() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.totalTokens
}

// ChunkCount returns the number of chunks received.
func (a *StreamAccumulator) ChunkCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.chunks)
}

// Duration returns the duration from first to last chunk.
func (a *StreamAccumulator) Duration() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.firstChunk.IsZero() {
		return 0
	}
	return a.lastChunk.Sub(a.firstChunk)
}

// TimeToFirstChunk returns the time to first chunk (should be set externally).
func (a *StreamAccumulator) TimeToFirstChunk(streamStart time.Time) time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.firstChunk.IsZero() {
		return 0
	}
	return a.firstChunk.Sub(streamStart)
}

// FinishReason returns the finish reason.
func (a *StreamAccumulator) FinishReason() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.finishReason
}

// Metadata returns the accumulated metadata.
func (a *StreamAccumulator) Metadata() map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make(map[string]any, len(a.metadata))
	for k, v := range a.metadata {
		result[k] = v
	}
	return result
}

// ToOutput returns the accumulator as a span output.
func (a *StreamAccumulator) ToOutput() map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()

	return map[string]any{
		"content":       a.content.String(),
		"chunk_count":   len(a.chunks),
		"total_tokens":  a.totalTokens,
		"finish_reason": a.finishReason,
	}
}

// StreamingSpan wraps a span for streaming operations.
type StreamingSpan struct {
	span        *Span
	accumulator *StreamAccumulator
	startTime   time.Time
	onChunk     func(chunk StreamChunk)
}

// NewStreamingSpan creates a new streaming span wrapper.
func NewStreamingSpan(span *Span) *StreamingSpan {
	return &StreamingSpan{
		span:        span,
		accumulator: NewStreamAccumulator(),
		startTime:   time.Now(),
	}
}

// OnChunk sets a callback for when chunks are received.
func (s *StreamingSpan) OnChunk(fn func(chunk StreamChunk)) {
	s.onChunk = fn
}

// AddChunk adds a chunk and tracks it.
func (s *StreamingSpan) AddChunk(content string, opts ...StreamChunkOption) {
	chunk := StreamChunk{
		Content:   content,
		Index:     s.accumulator.ChunkCount(),
		IsFirst:   s.accumulator.ChunkCount() == 0,
		Timestamp: time.Now(),
		Metadata:  make(map[string]any),
	}

	for _, opt := range opts {
		opt(&chunk)
	}

	s.accumulator.AddChunk(chunk)

	if s.onChunk != nil {
		s.onChunk(chunk)
	}
}

// StreamChunkOption configures a stream chunk.
type StreamChunkOption func(*StreamChunk)

// WithChunkTokenCount sets the token count for a chunk.
func WithChunkTokenCount(count int) StreamChunkOption {
	return func(c *StreamChunk) {
		c.TokenCount = count
	}
}

// WithChunkFinishReason sets the finish reason.
func WithChunkFinishReason(reason string) StreamChunkOption {
	return func(c *StreamChunk) {
		c.FinishReason = reason
		c.IsLast = true
	}
}

// WithChunkMetadata adds metadata to a chunk.
func WithChunkMetadata(key string, value any) StreamChunkOption {
	return func(c *StreamChunk) {
		c.Metadata[key] = value
	}
}

// End ends the streaming span with accumulated data.
func (s *StreamingSpan) End(ctx context.Context, opts ...SpanOption) error {
	// Add streaming metadata
	metadata := map[string]any{
		"streaming":           true,
		"chunk_count":         s.accumulator.ChunkCount(),
		"time_to_first_chunk": s.accumulator.TimeToFirstChunk(s.startTime).Milliseconds(),
		"stream_duration_ms":  s.accumulator.Duration().Milliseconds(),
		"total_tokens":        s.accumulator.TotalTokens(),
	}

	allOpts := append([]SpanOption{
		WithSpanOutput(s.accumulator.ToOutput()),
		WithSpanMetadata(metadata),
	}, opts...)

	return s.span.End(ctx, allOpts...)
}

// Span returns the underlying span.
func (s *StreamingSpan) Span() *Span {
	return s.span
}

// Accumulator returns the stream accumulator.
func (s *StreamingSpan) Accumulator() *StreamAccumulator {
	return s.accumulator
}

// ID returns the span ID.
func (s *StreamingSpan) ID() string {
	return s.span.ID()
}

// TraceID returns the trace ID.
func (s *StreamingSpan) TraceID() string {
	return s.span.TraceID()
}

// StartStreamingSpan starts a new span configured for streaming.
func StartStreamingSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, *StreamingSpan, error) {
	// Add streaming metadata to options
	allOpts := append([]SpanOption{
		WithSpanMetadata(map[string]any{"streaming": true}),
	}, opts...)

	newCtx, span, err := StartSpan(ctx, name, allOpts...)
	if err != nil {
		return ctx, nil, err
	}

	streamingSpan := NewStreamingSpan(span)
	return newCtx, streamingSpan, nil
}

// StreamHandler is a generic interface for processing streams.
type StreamHandler interface {
	// HandleChunk processes a chunk of data.
	HandleChunk(chunk StreamChunk) error
	// Finalize is called when the stream ends.
	Finalize() error
}

// TracingStreamHandler wraps a StreamHandler to add tracing.
type TracingStreamHandler struct {
	inner         StreamHandler
	streamingSpan *StreamingSpan
}

// NewTracingStreamHandler creates a tracing stream handler.
func NewTracingStreamHandler(inner StreamHandler, span *Span) *TracingStreamHandler {
	return &TracingStreamHandler{
		inner:         inner,
		streamingSpan: NewStreamingSpan(span),
	}
}

// HandleChunk processes a chunk and tracks it.
func (h *TracingStreamHandler) HandleChunk(chunk StreamChunk) error {
	h.streamingSpan.accumulator.AddChunk(chunk)
	return h.inner.HandleChunk(chunk)
}

// Finalize ends the stream and the span.
func (h *TracingStreamHandler) Finalize() error {
	return h.inner.Finalize()
}

// StreamingSpan returns the streaming span for ending.
func (h *TracingStreamHandler) StreamingSpan() *StreamingSpan {
	return h.streamingSpan
}

// BufferingStreamHandler buffers chunks and calls a callback with complete content.
type BufferingStreamHandler struct {
	accumulator *StreamAccumulator
	onComplete  func(content string) error
	onChunk     func(chunk StreamChunk) error
}

// NewBufferingStreamHandler creates a buffering handler.
func NewBufferingStreamHandler(onComplete func(content string) error) *BufferingStreamHandler {
	return &BufferingStreamHandler{
		accumulator: NewStreamAccumulator(),
		onComplete:  onComplete,
	}
}

// OnChunk sets a callback for each chunk.
func (h *BufferingStreamHandler) OnChunk(fn func(chunk StreamChunk) error) {
	h.onChunk = fn
}

// HandleChunk adds a chunk to the buffer.
func (h *BufferingStreamHandler) HandleChunk(chunk StreamChunk) error {
	h.accumulator.AddChunk(chunk)
	if h.onChunk != nil {
		return h.onChunk(chunk)
	}
	return nil
}

// Finalize calls the completion callback with buffered content.
func (h *BufferingStreamHandler) Finalize() error {
	if h.onComplete != nil {
		return h.onComplete(h.accumulator.Content())
	}
	return nil
}

// Content returns the current buffered content.
func (h *BufferingStreamHandler) Content() string {
	return h.accumulator.Content()
}

// Accumulator returns the stream accumulator.
func (h *BufferingStreamHandler) Accumulator() *StreamAccumulator {
	return h.accumulator
}
