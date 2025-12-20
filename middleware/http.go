package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	opik "github.com/grokify/go-comet-ml-opik"
)

// TracingMiddleware wraps an http.Handler to automatically create spans for incoming requests.
func TracingMiddleware(client *opik.Client, traceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a trace for this request
			ctx, trace, err := opik.StartTrace(r.Context(), client, traceName,
				opik.WithTraceInput(map[string]any{
					"method":     r.Method,
					"path":       r.URL.Path,
					"query":      r.URL.RawQuery,
					"host":       r.Host,
					"user_agent": r.UserAgent(),
				}),
			)
			if err != nil {
				// If tracing fails, continue without it
				next.ServeHTTP(w, r)
				return
			}

			// Wrap the response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Create a span for the handler
			ctx, span, err := opik.StartSpan(ctx, r.URL.Path,
				opik.WithSpanType(opik.SpanTypeGeneral),
				opik.WithSpanInput(map[string]any{
					"method": r.Method,
					"path":   r.URL.Path,
				}),
			)
			if err != nil {
				_ = trace.End(ctx)
				next.ServeHTTP(w, r)
				return
			}

			// Call the next handler
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// End span with response info
			_ = span.End(ctx,
				opik.WithSpanOutput(map[string]any{
					"status_code": wrapped.statusCode,
				}),
			)

			// End trace with response info
			_ = trace.End(ctx,
				opik.WithTraceOutput(map[string]any{
					"status_code": wrapped.statusCode,
				}),
			)
		})
	}
}

// TracingRoundTripper wraps an http.RoundTripper to automatically create spans for outgoing requests.
type TracingRoundTripper struct {
	transport http.RoundTripper
	spanName  string
}

// NewTracingRoundTripper creates a new TracingRoundTripper.
// If transport is nil, http.DefaultTransport is used.
func NewTracingRoundTripper(transport http.RoundTripper, spanName string) *TracingRoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	if spanName == "" {
		spanName = "http-request"
	}
	return &TracingRoundTripper{
		transport: transport,
		spanName:  spanName,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *TracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Only create span if there's an active trace/span in context
	parentSpan := opik.SpanFromContext(ctx)
	parentTrace := opik.TraceFromContext(ctx)
	if parentSpan == nil && parentTrace == nil {
		return t.transport.RoundTrip(req)
	}

	// Create a span for this HTTP request
	spanName := t.spanName + " " + req.Method + " " + req.URL.Host
	ctx, span, err := opik.StartSpan(ctx, spanName,
		opik.WithSpanType(opik.SpanTypeTool),
		opik.WithSpanInput(map[string]any{
			"method": req.Method,
			"url":    req.URL.String(),
			"host":   req.URL.Host,
		}),
	)
	if err != nil {
		return t.transport.RoundTrip(req)
	}

	start := time.Now()

	// Make the request
	resp, err := t.transport.RoundTrip(req.WithContext(ctx))

	duration := time.Since(start)

	// End the span
	if err != nil {
		_ = span.End(ctx,
			opik.WithSpanOutput(map[string]any{
				"error":    err.Error(),
				"duration": duration.String(),
			}),
		)
	} else {
		_ = span.End(ctx,
			opik.WithSpanOutput(map[string]any{
				"status_code":    resp.StatusCode,
				"status":         resp.Status,
				"content_length": resp.ContentLength,
				"duration":       duration.String(),
			}),
		)
	}

	return resp, err
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// TracingHTTPClient returns an http.Client with tracing enabled.
func TracingHTTPClient(spanName string) *http.Client {
	return &http.Client{
		Transport: NewTracingRoundTripper(nil, spanName),
	}
}

// InjectTraceHeaders adds trace context headers to an HTTP request.
// This can be used for distributed tracing across services.
func InjectTraceHeaders(ctx context.Context, req *http.Request) {
	traceID := opik.CurrentTraceID(ctx)
	spanID := opik.CurrentSpanID(ctx)

	if traceID != "" {
		req.Header.Set("X-Opik-Trace-ID", traceID)
	}
	if spanID != "" {
		req.Header.Set("X-Opik-Span-ID", spanID)
	}
}

// ExtractTraceContext extracts trace context from HTTP request headers.
// Returns the trace ID and span ID if present.
func ExtractTraceContext(r *http.Request) (traceID, spanID string) {
	traceID = r.Header.Get("X-Opik-Trace-ID")
	spanID = r.Header.Get("X-Opik-Span-ID")
	return
}

// RequestMetadata extracts common metadata from an HTTP request.
func RequestMetadata(r *http.Request) map[string]any {
	return map[string]any{
		"method":         r.Method,
		"path":           r.URL.Path,
		"query":          r.URL.RawQuery,
		"host":           r.Host,
		"remote_addr":    r.RemoteAddr,
		"user_agent":     r.UserAgent(),
		"content_length": r.ContentLength,
		"protocol":       r.Proto,
	}
}

// ResponseMetadata creates metadata from an HTTP response.
func ResponseMetadata(resp *http.Response, duration time.Duration) map[string]any {
	return map[string]any{
		"status_code":    resp.StatusCode,
		"status":         resp.Status,
		"content_length": resp.ContentLength,
		"duration_ms":    duration.Milliseconds(),
	}
}

// StatusCodeCategory returns a category string for the status code.
func StatusCodeCategory(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "success"
	case code >= 400 && code < 500:
		return "client_error"
	case code >= 500:
		return "server_error"
	default:
		return "other"
	}
}

// FormatDuration formats a duration for display.
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return strconv.FormatInt(d.Microseconds(), 10) + "µs"
	}
	if d < time.Second {
		return strconv.FormatInt(d.Milliseconds(), 10) + "ms"
	}
	return d.Round(time.Millisecond).String()
}
