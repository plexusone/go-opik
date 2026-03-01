package llmops

import (
	"context"
	"time"

	"github.com/plexusone/omniobserve/llmops"
	opik "github.com/plexusone/opik-go"
)

// traceAdapter adapts opik.Trace to llmops.Trace.
type traceAdapter struct {
	trace *opik.Trace
}

// ID returns the trace ID.
func (t *traceAdapter) ID() string {
	return t.trace.ID()
}

// Name returns the trace name.
func (t *traceAdapter) Name() string {
	return t.trace.Name()
}

// StartSpan creates a child span.
func (t *traceAdapter) StartSpan(ctx context.Context, name string, opts ...llmops.SpanOption) (context.Context, llmops.Span, error) {
	cfg := llmops.ApplySpanOptions(opts...)
	opikOpts := mapSpanOptions(cfg)

	span, err := t.trace.Span(ctx, name, opikOpts...)
	if err != nil {
		return ctx, nil, err
	}

	newCtx := opik.ContextWithSpan(ctx, span)
	return newCtx, &spanAdapter{span: span}, nil
}

// SetInput sets the trace input.
func (t *traceAdapter) SetInput(input any) error {
	return t.trace.Update(context.Background(), opik.WithTraceInput(input))
}

// SetOutput sets the trace output.
func (t *traceAdapter) SetOutput(output any) error {
	return t.trace.Update(context.Background(), opik.WithTraceOutput(output))
}

// SetMetadata sets the trace metadata.
func (t *traceAdapter) SetMetadata(metadata map[string]any) error {
	return t.trace.Update(context.Background(), opik.WithTraceMetadata(metadata))
}

// AddTag adds a tag to the trace.
func (t *traceAdapter) AddTag(tag string) error {
	return t.trace.Update(context.Background(), opik.WithTraceTags(tag))
}

// AddFeedbackScore adds a feedback score.
func (t *traceAdapter) AddFeedbackScore(ctx context.Context, name string, score float64, opts ...llmops.FeedbackOption) error {
	cfg := &llmops.FeedbackOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	reason := cfg.Reason
	if reason == "" {
		reason = "feedback"
	}
	return t.trace.AddFeedbackScore(ctx, name, score, reason)
}

// End ends the trace.
func (t *traceAdapter) End(opts ...llmops.EndOption) error {
	cfg := &llmops.EndOptions{}
	for _, opt := range opts {
		opt(cfg)
	}

	opikOpts := []opik.TraceOption{}
	if cfg.Output != nil {
		opikOpts = append(opikOpts, opik.WithTraceOutput(cfg.Output))
	}
	if cfg.Metadata != nil {
		opikOpts = append(opikOpts, opik.WithTraceMetadata(cfg.Metadata))
	}

	return t.trace.End(context.Background(), opikOpts...)
}

// EndTime returns when the trace ended.
func (t *traceAdapter) EndTime() *time.Time {
	return t.trace.EndTime()
}

// Duration returns the trace duration.
func (t *traceAdapter) Duration() time.Duration {
	startTime := t.trace.StartTime()
	if endTime := t.trace.EndTime(); endTime != nil {
		return endTime.Sub(startTime)
	}
	return time.Since(startTime)
}

// spanAdapter adapts opik.Span to llmops.Span.
type spanAdapter struct {
	span *opik.Span
}

// ID returns the span ID.
func (s *spanAdapter) ID() string {
	return s.span.ID()
}

// TraceID returns the parent trace ID.
func (s *spanAdapter) TraceID() string {
	return s.span.TraceID()
}

// ParentSpanID returns the parent span ID.
func (s *spanAdapter) ParentSpanID() string {
	return s.span.ParentSpanID()
}

// Name returns the span name.
func (s *spanAdapter) Name() string {
	return s.span.Name()
}

// Type returns the span type.
func (s *spanAdapter) Type() llmops.SpanType {
	return llmops.SpanType(s.span.Type())
}

// StartSpan creates a child span.
func (s *spanAdapter) StartSpan(ctx context.Context, name string, opts ...llmops.SpanOption) (context.Context, llmops.Span, error) {
	cfg := llmops.ApplySpanOptions(opts...)
	opikOpts := mapSpanOptions(cfg)

	span, err := s.span.Span(ctx, name, opikOpts...)
	if err != nil {
		return ctx, nil, err
	}

	newCtx := opik.ContextWithSpan(ctx, span)
	return newCtx, &spanAdapter{span: span}, nil
}

// SetInput sets the span input.
func (s *spanAdapter) SetInput(input any) error {
	return s.span.Update(context.Background(), opik.WithSpanInput(input))
}

// SetOutput sets the span output.
func (s *spanAdapter) SetOutput(output any) error {
	return s.span.Update(context.Background(), opik.WithSpanOutput(output))
}

// SetMetadata sets the span metadata.
func (s *spanAdapter) SetMetadata(metadata map[string]any) error {
	return s.span.Update(context.Background(), opik.WithSpanMetadata(metadata))
}

// SetModel sets the model name.
func (s *spanAdapter) SetModel(model string) error {
	return s.span.Update(context.Background(), opik.WithSpanModel(model))
}

// SetProvider sets the provider name.
func (s *spanAdapter) SetProvider(provider string) error {
	return s.span.Update(context.Background(), opik.WithSpanProvider(provider))
}

// SetUsage sets token usage.
func (s *spanAdapter) SetUsage(usage llmops.TokenUsage) error {
	usageMap := map[string]int{
		"prompt_tokens":     usage.PromptTokens,
		"completion_tokens": usage.CompletionTokens,
		"total_tokens":      usage.TotalTokens,
	}
	s.span.SetUsage(usageMap)
	return nil
}

// AddTag adds a tag.
func (s *spanAdapter) AddTag(tag string) error {
	return s.span.Update(context.Background(), opik.WithSpanTags(tag))
}

// AddFeedbackScore adds a feedback score.
func (s *spanAdapter) AddFeedbackScore(ctx context.Context, name string, score float64, opts ...llmops.FeedbackOption) error {
	cfg := &llmops.FeedbackOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	reason := cfg.Reason
	if reason == "" {
		reason = "feedback"
	}
	return s.span.AddFeedbackScore(ctx, name, score, reason)
}

// End ends the span.
func (s *spanAdapter) End(opts ...llmops.EndOption) error {
	cfg := &llmops.EndOptions{}
	for _, opt := range opts {
		opt(cfg)
	}

	opikOpts := []opik.SpanOption{}
	if cfg.Output != nil {
		opikOpts = append(opikOpts, opik.WithSpanOutput(cfg.Output))
	}
	if cfg.Metadata != nil {
		opikOpts = append(opikOpts, opik.WithSpanMetadata(cfg.Metadata))
	}

	return s.span.End(context.Background(), opikOpts...)
}

// EndTime returns when the span ended.
func (s *spanAdapter) EndTime() *time.Time {
	return s.span.EndTime()
}

// Duration returns the span duration.
func (s *spanAdapter) Duration() time.Duration {
	startTime := s.span.StartTime()
	if endTime := s.span.EndTime(); endTime != nil {
		return endTime.Sub(startTime)
	}
	return time.Since(startTime)
}
