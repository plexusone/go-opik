package opik

import (
	"context"
)

// Context keys for storing trace and span data.
type contextKey int

const (
	traceContextKey contextKey = iota
	spanContextKey
	clientContextKey
)

// ContextWithTrace returns a new context with the trace attached.
func ContextWithTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, traceContextKey, trace)
}

// TraceFromContext returns the trace from the context, or nil if none.
func TraceFromContext(ctx context.Context) *Trace {
	if trace, ok := ctx.Value(traceContextKey).(*Trace); ok {
		return trace
	}
	return nil
}

// ContextWithSpan returns a new context with the span attached.
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanContextKey, span)
}

// SpanFromContext returns the span from the context, or nil if none.
func SpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanContextKey).(*Span); ok {
		return span
	}
	return nil
}

// ContextWithClient returns a new context with the client attached.
func ContextWithClient(ctx context.Context, client *Client) context.Context {
	return context.WithValue(ctx, clientContextKey, client)
}

// ClientFromContext returns the client from the context, or nil if none.
func ClientFromContext(ctx context.Context) *Client {
	if client, ok := ctx.Value(clientContextKey).(*Client); ok {
		return client
	}
	return nil
}

// StartTrace creates a new trace and attaches it to the context.
// Returns the new context and the trace.
func StartTrace(ctx context.Context, client *Client, name string, opts ...TraceOption) (context.Context, *Trace, error) {
	trace, err := client.Trace(ctx, name, opts...)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := ContextWithTrace(ctx, trace)
	newCtx = ContextWithClient(newCtx, client)
	return newCtx, trace, nil
}

// StartSpan creates a new span and attaches it to the context.
// If there is no parent span in the context, it uses the trace from context.
// Returns the new context and the span.
func StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, *Span, error) {
	// Try to get parent span first
	parentSpan := SpanFromContext(ctx)
	if parentSpan != nil {
		span, err := parentSpan.Span(ctx, name, opts...)
		if err != nil {
			return ctx, nil, err
		}
		newCtx := ContextWithSpan(ctx, span)
		return newCtx, span, nil
	}

	// Fall back to trace
	trace := TraceFromContext(ctx)
	if trace != nil {
		span, err := trace.Span(ctx, name, opts...)
		if err != nil {
			return ctx, nil, err
		}
		newCtx := ContextWithSpan(ctx, span)
		return newCtx, span, nil
	}

	return ctx, nil, ErrNoActiveTrace
}

// EndTrace ends the current trace in the context.
func EndTrace(ctx context.Context, opts ...TraceOption) error {
	trace := TraceFromContext(ctx)
	if trace == nil {
		return ErrNoActiveTrace
	}
	return trace.End(ctx, opts...)
}

// EndSpan ends the current span in the context.
func EndSpan(ctx context.Context, opts ...SpanOption) error {
	span := SpanFromContext(ctx)
	if span == nil {
		return ErrNoActiveSpan
	}
	return span.End(ctx, opts...)
}

// CurrentTraceID returns the current trace ID from the context, or empty string if none.
func CurrentTraceID(ctx context.Context) string {
	if trace := TraceFromContext(ctx); trace != nil {
		return trace.ID()
	}
	if span := SpanFromContext(ctx); span != nil {
		return span.TraceID()
	}
	return ""
}

// CurrentSpanID returns the current span ID from the context, or empty string if none.
func CurrentSpanID(ctx context.Context) string {
	if span := SpanFromContext(ctx); span != nil {
		return span.ID()
	}
	return ""
}
