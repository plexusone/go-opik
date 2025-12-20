package opik

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/grokify/go-comet-ml-opik/internal/api"
)

// Trace represents an execution trace in Opik.
type Trace struct {
	client      *Client
	id          string
	name        string
	projectName string
	startTime   time.Time
	endTime     *time.Time
	input       any
	output      any
	metadata    map[string]any
	tags        []string
	ended       bool
}

// ID returns the trace ID.
func (t *Trace) ID() string {
	return t.id
}

// Name returns the trace name.
func (t *Trace) Name() string {
	return t.name
}

// ProjectName returns the project name.
func (t *Trace) ProjectName() string {
	return t.projectName
}

// StartTime returns the start time.
func (t *Trace) StartTime() time.Time {
	return t.startTime
}

// EndTime returns the end time, or nil if not ended.
func (t *Trace) EndTime() *time.Time {
	return t.endTime
}

// End ends the trace with optional output.
func (t *Trace) End(ctx context.Context, opts ...TraceOption) error {
	if t.ended {
		return nil
	}

	options := &traceOptions{
		metadata: make(map[string]any),
	}
	for _, opt := range opts {
		opt(options)
	}

	endTime := time.Now()
	t.endTime = &endTime
	t.ended = true

	// Merge output and metadata
	if options.output != nil {
		t.output = options.output
	}
	for k, v := range options.metadata {
		t.metadata[k] = v
	}

	// Prepare update request
	traceUUID, err := uuid.Parse(t.id)
	if err != nil {
		return err
	}

	var outputJSON api.JsonListString
	if t.output != nil {
		data, _ := json.Marshal(t.output)
		outputJSON = api.JsonListString(data)
	}

	var metadataJSON api.JsonListString
	if len(t.metadata) > 0 {
		data, _ := json.Marshal(t.metadata)
		metadataJSON = api.JsonListString(data)
	}

	// Create update request - TraceBatchUpdate uses Ids + single Update
	req := api.TraceBatchUpdate{
		Ids: []uuid.UUID{traceUUID},
		Update: api.TraceUpdate{
			EndTime:  api.NewOptDateTime(endTime),
			Output:   outputJSON,
			Metadata: metadataJSON,
		},
	}

	_, err = t.client.apiClient.BatchUpdateTraces(ctx, api.NewOptTraceBatchUpdate(req))
	return err
}

// Update updates the trace with new data.
func (t *Trace) Update(ctx context.Context, opts ...TraceOption) error {
	options := &traceOptions{
		metadata: make(map[string]any),
	}
	for _, opt := range opts {
		opt(options)
	}

	traceUUID, err := uuid.Parse(t.id)
	if err != nil {
		return err
	}

	// Merge metadata
	for k, v := range options.metadata {
		t.metadata[k] = v
	}
	if options.output != nil {
		t.output = options.output
	}

	var outputJSON api.JsonListString
	if t.output != nil {
		data, _ := json.Marshal(t.output)
		outputJSON = api.JsonListString(data)
	}

	var metadataJSON api.JsonListString
	if len(t.metadata) > 0 {
		data, _ := json.Marshal(t.metadata)
		metadataJSON = api.JsonListString(data)
	}

	req := api.TraceBatchUpdate{
		Ids: []uuid.UUID{traceUUID},
		Update: api.TraceUpdate{
			Output:   outputJSON,
			Metadata: metadataJSON,
			Tags:     options.tags,
		},
	}

	_, err = t.client.apiClient.BatchUpdateTraces(ctx, api.NewOptTraceBatchUpdate(req))
	return err
}

// Span creates a new span within this trace.
func (t *Trace) Span(ctx context.Context, name string, opts ...SpanOption) (*Span, error) {
	return t.client.createSpan(ctx, t.id, "", name, opts...)
}

// AddFeedbackScore adds a feedback score to this trace.
func (t *Trace) AddFeedbackScore(ctx context.Context, name string, value float64, reason string) error {
	traceUUID, err := uuid.Parse(t.id)
	if err != nil {
		return err
	}

	req := api.FeedbackScore{
		Name:   name,
		Value:  value,
		Reason: api.NewOptString(reason),
		Source: api.FeedbackScoreSourceSdk,
	}

	return t.client.apiClient.AddTraceFeedbackScore(ctx, api.NewOptFeedbackScore(req), api.AddTraceFeedbackScoreParams{
		ID: traceUUID,
	})
}
