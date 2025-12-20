package anthropic

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	opik "github.com/grokify/go-comet-ml-opik"
)

// TracingTransport wraps an http.RoundTripper to automatically trace Anthropic API calls.
type TracingTransport struct {
	inner       http.RoundTripper
	opikClient  *opik.Client
	spanOptions []opik.SpanOption
}

// NewTracingTransport creates a new tracing transport.
func NewTracingTransport(inner http.RoundTripper, opikClient *opik.Client, opts ...opik.SpanOption) *TracingTransport {
	if inner == nil {
		inner = http.DefaultTransport
	}
	return &TracingTransport{
		inner:       inner,
		opikClient:  opikClient,
		spanOptions: opts,
	}
}

// RoundTrip implements http.RoundTripper with tracing.
func (t *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only trace Anthropic API calls
	if !isAnthropicRequest(req) {
		return t.inner.RoundTrip(req)
	}

	ctx := req.Context()

	// Parse request body for span metadata
	var reqData map[string]any
	var model string
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(body))
		_ = json.Unmarshal(body, &reqData)
		if m, ok := reqData["model"].(string); ok {
			model = m
		}
	}

	// Determine operation name from URL
	operation := getOperationName(req.URL.Path)

	// Start span
	opts := append([]opik.SpanOption{
		opik.WithSpanType(opik.SpanTypeLLM),
		opik.WithSpanProvider("anthropic"),
		opik.WithSpanInput(reqData),
	}, t.spanOptions...)

	if model != "" {
		opts = append(opts, opik.WithSpanModel(model))
	}

	var span *opik.Span
	var err error

	// Try to get parent from context
	if parentSpan := opik.SpanFromContext(ctx); parentSpan != nil {
		span, err = parentSpan.Span(ctx, operation, opts...)
	} else if trace := opik.TraceFromContext(ctx); trace != nil {
		span, err = trace.Span(ctx, operation, opts...)
	}

	// Execute request
	startTime := time.Now()
	resp, respErr := t.inner.RoundTrip(req)
	duration := time.Since(startTime)

	// End span with results
	if span != nil && err == nil {
		endOpts := []opik.SpanOption{}

		if resp != nil && resp.Body != nil {
			// Read response body for output
			body, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewReader(body))

			var respData map[string]any
			if json.Unmarshal(body, &respData) == nil {
				endOpts = append(endOpts, opik.WithSpanOutput(respData))

				// Extract usage info
				if usage, ok := respData["usage"].(map[string]any); ok {
					metadata := map[string]any{
						"duration_ms": duration.Milliseconds(),
					}
					if it, ok := usage["input_tokens"].(float64); ok {
						metadata["input_tokens"] = int(it)
					}
					if ot, ok := usage["output_tokens"].(float64); ok {
						metadata["output_tokens"] = int(ot)
					}
					endOpts = append(endOpts, opik.WithSpanMetadata(metadata))
				}
			}
		}

		if respErr != nil {
			endOpts = append(endOpts, opik.WithSpanMetadata(map[string]any{
				"error": respErr.Error(),
			}))
		}

		_ = span.End(ctx, endOpts...)
	}

	return resp, respErr
}

func isAnthropicRequest(req *http.Request) bool {
	host := req.URL.Host
	return host == "api.anthropic.com"
}

func getOperationName(path string) string {
	switch {
	case bytes.Contains([]byte(path), []byte("/messages")):
		return "anthropic.messages"
	case bytes.Contains([]byte(path), []byte("/complete")):
		return "anthropic.complete"
	default:
		return "anthropic.api"
	}
}

// TracingHTTPClient creates an http.Client that traces Anthropic calls.
func TracingHTTPClient(opikClient *opik.Client, opts ...opik.SpanOption) *http.Client {
	return &http.Client{
		Transport: NewTracingTransport(nil, opikClient, opts...),
	}
}

// TracingProvider creates an Anthropic provider with automatic tracing.
func TracingProvider(opikClient *opik.Client, providerOpts ...Option) *Provider {
	httpClient := TracingHTTPClient(opikClient)
	allOpts := append([]Option{WithHTTPClient(httpClient)}, providerOpts...)
	return NewProvider(allOpts...)
}

// Wrap wraps an existing http.Client to add Anthropic tracing.
func Wrap(client *http.Client, opikClient *opik.Client, opts ...opik.SpanOption) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &http.Client{
		Transport:     NewTracingTransport(transport, opikClient, opts...),
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}
}
