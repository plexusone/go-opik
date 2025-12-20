package opik

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/grokify/go-comet-ml-opik/internal/api"
)

// Span represents a span within a trace in Opik.
type Span struct {
	client       *Client
	id           string
	traceID      string
	parentSpanID string
	name         string
	spanType     string
	startTime    time.Time
	endTime      *time.Time
	input        any
	output       any
	metadata     map[string]any
	tags         []string
	model        string
	provider     string
	usage        map[string]int
	ended        bool
}

// ID returns the span ID.
func (s *Span) ID() string {
	return s.id
}

// TraceID returns the trace ID.
func (s *Span) TraceID() string {
	return s.traceID
}

// ParentSpanID returns the parent span ID, or empty string if none.
func (s *Span) ParentSpanID() string {
	return s.parentSpanID
}

// Name returns the span name.
func (s *Span) Name() string {
	return s.name
}

// Type returns the span type.
func (s *Span) Type() string {
	return s.spanType
}

// StartTime returns the start time.
func (s *Span) StartTime() time.Time {
	return s.startTime
}

// EndTime returns the end time, or nil if not ended.
func (s *Span) EndTime() *time.Time {
	return s.endTime
}

// End ends the span with optional output.
func (s *Span) End(ctx context.Context, opts ...SpanOption) error {
	if s.ended {
		return nil
	}

	options := &spanOptions{
		metadata: make(map[string]any),
	}
	for _, opt := range opts {
		opt(options)
	}

	endTime := time.Now()
	s.endTime = &endTime
	s.ended = true

	// Merge output and metadata
	if options.output != nil {
		s.output = options.output
	}
	for k, v := range options.metadata {
		s.metadata[k] = v
	}
	if options.model != "" {
		s.model = options.model
	}
	if options.provider != "" {
		s.provider = options.provider
	}

	// Prepare update request
	spanUUID, err := uuid.Parse(s.id)
	if err != nil {
		return err
	}

	traceUUID, err := uuid.Parse(s.traceID)
	if err != nil {
		return err
	}

	var outputJSON api.JsonListString
	if s.output != nil {
		data, _ := json.Marshal(s.output)
		outputJSON = api.JsonListString(data)
	}

	var metadataJSON api.JsonListString
	if len(s.metadata) > 0 {
		data, _ := json.Marshal(s.metadata)
		metadataJSON = api.JsonListString(data)
	}

	// Create update request - SpanBatchUpdate uses Ids + single Update
	req := api.SpanBatchUpdate{
		Ids: []uuid.UUID{spanUUID},
		Update: api.SpanUpdate{
			TraceID:  traceUUID,
			EndTime:  api.NewOptDateTime(endTime),
			Output:   outputJSON,
			Metadata: metadataJSON,
			Model:    api.NewOptString(s.model),
			Provider: api.NewOptString(s.provider),
		},
	}

	_, err = s.client.apiClient.BatchUpdateSpans(ctx, api.NewOptSpanBatchUpdate(req))
	return err
}

// Update updates the span with new data.
func (s *Span) Update(ctx context.Context, opts ...SpanOption) error {
	options := &spanOptions{
		metadata: make(map[string]any),
	}
	for _, opt := range opts {
		opt(options)
	}

	spanUUID, err := uuid.Parse(s.id)
	if err != nil {
		return err
	}

	traceUUID, err := uuid.Parse(s.traceID)
	if err != nil {
		return err
	}

	// Merge metadata
	for k, v := range options.metadata {
		s.metadata[k] = v
	}
	if options.output != nil {
		s.output = options.output
	}
	if options.model != "" {
		s.model = options.model
	}
	if options.provider != "" {
		s.provider = options.provider
	}

	var outputJSON api.JsonListString
	if s.output != nil {
		data, _ := json.Marshal(s.output)
		outputJSON = api.JsonListString(data)
	}

	var metadataJSON api.JsonListString
	if len(s.metadata) > 0 {
		data, _ := json.Marshal(s.metadata)
		metadataJSON = api.JsonListString(data)
	}

	req := api.SpanBatchUpdate{
		Ids: []uuid.UUID{spanUUID},
		Update: api.SpanUpdate{
			TraceID:  traceUUID,
			Output:   outputJSON,
			Metadata: metadataJSON,
			Model:    api.NewOptString(s.model),
			Provider: api.NewOptString(s.provider),
			Tags:     options.tags,
		},
	}

	_, err = s.client.apiClient.BatchUpdateSpans(ctx, api.NewOptSpanBatchUpdate(req))
	return err
}

// Span creates a child span within this span.
func (s *Span) Span(ctx context.Context, name string, opts ...SpanOption) (*Span, error) {
	return s.client.createSpan(ctx, s.traceID, s.id, name, opts...)
}

// AddFeedbackScore adds a feedback score to this span.
func (s *Span) AddFeedbackScore(ctx context.Context, name string, value float64, reason string) error {
	spanUUID, err := uuid.Parse(s.id)
	if err != nil {
		return err
	}

	req := api.FeedbackScore{
		Name:   name,
		Value:  value,
		Reason: api.NewOptString(reason),
		Source: api.FeedbackScoreSourceSdk,
	}

	return s.client.apiClient.AddSpanFeedbackScore(ctx, api.NewOptFeedbackScore(req), api.AddSpanFeedbackScoreParams{
		ID: spanUUID,
	})
}

// SetUsage sets LLM usage metrics for this span.
func (s *Span) SetUsage(usage map[string]int) {
	s.usage = usage
}

// createSpan is a helper to create spans (used by both Client and Trace).
func (c *Client) createSpan(ctx context.Context, traceID, parentSpanID, name string, opts ...SpanOption) (*Span, error) {
	if c.config.TracingDisabled {
		return nil, ErrTracingDisabled
	}

	options := defaultSpanOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Generate span ID
	spanUUID := uuid.New()
	spanID := spanUUID.String()
	traceUUID, err := uuid.Parse(traceID)
	if err != nil {
		return nil, err
	}

	// Prepare input/output/metadata as JSON
	var inputJSON, outputJSON, metadataJSON api.JsonListStringWrite
	if options.input != nil {
		data, _ := json.Marshal(options.input)
		inputJSON = api.JsonListStringWrite(data)
	}
	if options.output != nil {
		data, _ := json.Marshal(options.output)
		outputJSON = api.JsonListStringWrite(data)
	}
	if len(options.metadata) > 0 {
		data, _ := json.Marshal(options.metadata)
		metadataJSON = api.JsonListStringWrite(data)
	}

	startTime := time.Now()

	// Determine span type
	spanType := api.SpanWriteType(options.spanType)

	// Create span request
	spanWrite := api.SpanWrite{
		ID:        api.NewOptUUID(spanUUID),
		TraceID:   api.NewOptUUID(traceUUID),
		Name:      api.NewOptString(name),
		Type:      api.NewOptSpanWriteType(spanType),
		StartTime: startTime,
		Input:     inputJSON,
		Output:    outputJSON,
		Metadata:  metadataJSON,
		Tags:      options.tags,
		Model:     api.NewOptString(options.model),
		Provider:  api.NewOptString(options.provider),
	}

	if parentSpanID != "" {
		parentUUID, err := uuid.Parse(parentSpanID)
		if err != nil {
			return nil, err
		}
		spanWrite.ParentSpanID = api.NewOptUUID(parentUUID)
	}

	req := api.SpanBatchWrite{
		Spans: []api.SpanWrite{spanWrite},
	}

	// Send to API
	err = c.apiClient.CreateSpans(ctx, api.NewOptSpanBatchWrite(req))
	if err != nil {
		return nil, err
	}

	return &Span{
		client:       c,
		id:           spanID,
		traceID:      traceID,
		parentSpanID: parentSpanID,
		name:         name,
		spanType:     options.spanType,
		startTime:    startTime,
		input:        options.input,
		output:       options.output,
		metadata:     options.metadata,
		tags:         options.tags,
		model:        options.model,
		provider:     options.provider,
	}, nil
}
