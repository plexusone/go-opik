package opik

import (
	"context"
	"net/http"
)

// DistributedTraceHeaders contains trace context for cross-service propagation.
type DistributedTraceHeaders struct {
	TraceID      string `json:"opik_trace_id"`
	ParentSpanID string `json:"opik_parent_span_id"`
}

// Header names for distributed tracing.
const (
	HeaderTraceID      = "X-Opik-Trace-ID"
	HeaderParentSpanID = "X-Opik-Parent-Span-ID"
)

// GetDistributedTraceHeaders returns the current trace context from the context.
// This can be used to propagate trace context across service boundaries.
func GetDistributedTraceHeaders(ctx context.Context) DistributedTraceHeaders {
	headers := DistributedTraceHeaders{}

	if trace := TraceFromContext(ctx); trace != nil {
		headers.TraceID = trace.ID()
	}

	if span := SpanFromContext(ctx); span != nil {
		headers.ParentSpanID = span.ID()
		if headers.TraceID == "" {
			headers.TraceID = span.TraceID()
		}
	}

	return headers
}

// InjectDistributedTraceHeaders adds trace context headers to an HTTP request.
func InjectDistributedTraceHeaders(ctx context.Context, req *http.Request) {
	headers := GetDistributedTraceHeaders(ctx)

	if headers.TraceID != "" {
		req.Header.Set(HeaderTraceID, headers.TraceID)
	}
	if headers.ParentSpanID != "" {
		req.Header.Set(HeaderParentSpanID, headers.ParentSpanID)
	}
}

// ExtractDistributedTraceHeaders extracts trace context from HTTP request headers.
func ExtractDistributedTraceHeaders(req *http.Request) DistributedTraceHeaders {
	return DistributedTraceHeaders{
		TraceID:      req.Header.Get(HeaderTraceID),
		ParentSpanID: req.Header.Get(HeaderParentSpanID),
	}
}

// ContinueTrace creates a new span that continues from distributed trace headers.
// This is used when receiving a request from another service that has trace context.
func (c *Client) ContinueTrace(ctx context.Context, headers DistributedTraceHeaders, spanName string, opts ...SpanOption) (context.Context, *Span, error) {
	if headers.TraceID == "" {
		// No trace context, start a new trace and return the first span
		ctx, trace, err := StartTrace(ctx, c, spanName)
		if err != nil {
			return ctx, nil, err
		}
		span, err := trace.Span(ctx, spanName, opts...)
		if err != nil {
			return ctx, nil, err
		}
		newCtx := ContextWithSpan(ctx, span)
		return newCtx, span, nil
	}

	// Create a span that continues the distributed trace
	span, err := c.createSpanWithParent(ctx, headers.TraceID, headers.ParentSpanID, spanName, opts...)
	if err != nil {
		return ctx, nil, err
	}

	newCtx := ContextWithSpan(ctx, span)
	return newCtx, span, nil
}

// createSpanWithParent creates a span with explicit trace and parent span IDs.
func (c *Client) createSpanWithParent(ctx context.Context, traceID, parentSpanID, name string, opts ...SpanOption) (*Span, error) {
	return c.createSpan(ctx, traceID, parentSpanID, name, opts...)
}

// PropagatingRoundTripper wraps an http.RoundTripper to automatically inject
// distributed trace headers into outgoing requests.
type PropagatingRoundTripper struct {
	transport http.RoundTripper
}

// NewPropagatingRoundTripper creates a new PropagatingRoundTripper.
func NewPropagatingRoundTripper(transport http.RoundTripper) *PropagatingRoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &PropagatingRoundTripper{transport: transport}
}

// RoundTrip implements http.RoundTripper.
func (t *PropagatingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	InjectDistributedTraceHeaders(req.Context(), req)
	return t.transport.RoundTrip(req)
}

// PropagatingHTTPClient returns an http.Client that automatically propagates
// distributed trace headers to all outgoing requests.
func PropagatingHTTPClient() *http.Client {
	return &http.Client{
		Transport: NewPropagatingRoundTripper(nil),
	}
}
