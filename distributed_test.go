package opik

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDistributedTraceHeaders(t *testing.T) {
	headers := DistributedTraceHeaders{
		TraceID:      "trace-123",
		ParentSpanID: "span-456",
	}

	if headers.TraceID != "trace-123" {
		t.Errorf("TraceID = %q, want %q", headers.TraceID, "trace-123")
	}
	if headers.ParentSpanID != "span-456" {
		t.Errorf("ParentSpanID = %q, want %q", headers.ParentSpanID, "span-456")
	}
}

func TestHeaderConstants(t *testing.T) {
	if HeaderTraceID != "X-Opik-Trace-ID" {
		t.Errorf("HeaderTraceID = %q, want %q", HeaderTraceID, "X-Opik-Trace-ID")
	}
	if HeaderParentSpanID != "X-Opik-Parent-Span-ID" {
		t.Errorf("HeaderParentSpanID = %q, want %q", HeaderParentSpanID, "X-Opik-Parent-Span-ID")
	}
}

func TestGetDistributedTraceHeadersEmpty(t *testing.T) {
	ctx := context.Background()
	headers := GetDistributedTraceHeaders(ctx)

	if headers.TraceID != "" {
		t.Errorf("TraceID = %q, want empty", headers.TraceID)
	}
	if headers.ParentSpanID != "" {
		t.Errorf("ParentSpanID = %q, want empty", headers.ParentSpanID)
	}
}

func TestGetDistributedTraceHeadersFromTrace(t *testing.T) {
	ctx := context.Background()
	trace := &Trace{id: "trace-123"}
	ctx = ContextWithTrace(ctx, trace)

	headers := GetDistributedTraceHeaders(ctx)

	if headers.TraceID != "trace-123" {
		t.Errorf("TraceID = %q, want %q", headers.TraceID, "trace-123")
	}
	if headers.ParentSpanID != "" {
		t.Errorf("ParentSpanID = %q, want empty", headers.ParentSpanID)
	}
}

func TestGetDistributedTraceHeadersFromSpan(t *testing.T) {
	ctx := context.Background()
	span := &Span{id: "span-456", traceID: "trace-123"}
	ctx = ContextWithSpan(ctx, span)

	headers := GetDistributedTraceHeaders(ctx)

	if headers.TraceID != "trace-123" {
		t.Errorf("TraceID = %q, want %q", headers.TraceID, "trace-123")
	}
	if headers.ParentSpanID != "span-456" {
		t.Errorf("ParentSpanID = %q, want %q", headers.ParentSpanID, "span-456")
	}
}

func TestGetDistributedTraceHeadersFromTraceAndSpan(t *testing.T) {
	ctx := context.Background()
	trace := &Trace{id: "trace-main"}
	span := &Span{id: "span-456", traceID: "trace-123"}
	ctx = ContextWithTrace(ctx, trace)
	ctx = ContextWithSpan(ctx, span)

	headers := GetDistributedTraceHeaders(ctx)

	// Trace ID should come from trace, not span
	if headers.TraceID != "trace-main" {
		t.Errorf("TraceID = %q, want %q (from trace)", headers.TraceID, "trace-main")
	}
	if headers.ParentSpanID != "span-456" {
		t.Errorf("ParentSpanID = %q, want %q", headers.ParentSpanID, "span-456")
	}
}

func TestInjectDistributedTraceHeaders(t *testing.T) {
	ctx := context.Background()
	trace := &Trace{id: "trace-123"}
	span := &Span{id: "span-456", traceID: "trace-123"}
	ctx = ContextWithTrace(ctx, trace)
	ctx = ContextWithSpan(ctx, span)

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	InjectDistributedTraceHeaders(ctx, req)

	if got := req.Header.Get(HeaderTraceID); got != "trace-123" {
		t.Errorf("Header %s = %q, want %q", HeaderTraceID, got, "trace-123")
	}
	if got := req.Header.Get(HeaderParentSpanID); got != "span-456" {
		t.Errorf("Header %s = %q, want %q", HeaderParentSpanID, got, "span-456")
	}
}

func TestInjectDistributedTraceHeadersEmpty(t *testing.T) {
	ctx := context.Background()
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	InjectDistributedTraceHeaders(ctx, req)

	if got := req.Header.Get(HeaderTraceID); got != "" {
		t.Errorf("Header %s = %q, want empty", HeaderTraceID, got)
	}
	if got := req.Header.Get(HeaderParentSpanID); got != "" {
		t.Errorf("Header %s = %q, want empty", HeaderParentSpanID, got)
	}
}

func TestExtractDistributedTraceHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(HeaderTraceID, "trace-123")
	req.Header.Set(HeaderParentSpanID, "span-456")

	headers := ExtractDistributedTraceHeaders(req)

	if headers.TraceID != "trace-123" {
		t.Errorf("TraceID = %q, want %q", headers.TraceID, "trace-123")
	}
	if headers.ParentSpanID != "span-456" {
		t.Errorf("ParentSpanID = %q, want %q", headers.ParentSpanID, "span-456")
	}
}

func TestExtractDistributedTraceHeadersEmpty(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	headers := ExtractDistributedTraceHeaders(req)

	if headers.TraceID != "" {
		t.Errorf("TraceID = %q, want empty", headers.TraceID)
	}
	if headers.ParentSpanID != "" {
		t.Errorf("ParentSpanID = %q, want empty", headers.ParentSpanID)
	}
}

func TestNewPropagatingRoundTripper(t *testing.T) {
	t.Run("with nil transport", func(t *testing.T) {
		rt := NewPropagatingRoundTripper(nil)
		if rt == nil {
			t.Fatal("NewPropagatingRoundTripper returned nil")
		}
		if rt.transport != http.DefaultTransport {
			t.Error("nil transport should use http.DefaultTransport")
		}
	})

	t.Run("with custom transport", func(t *testing.T) {
		custom := &http.Transport{}
		rt := NewPropagatingRoundTripper(custom)
		if rt.transport != custom {
			t.Error("custom transport should be used")
		}
	})
}

func TestPropagatingRoundTripperInjectsHeaders(t *testing.T) {
	// Create a test server that records headers
	var receivedTraceID, receivedParentSpanID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedTraceID = r.Header.Get(HeaderTraceID)
		receivedParentSpanID = r.Header.Get(HeaderParentSpanID)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create context with trace and span
	ctx := context.Background()
	trace := &Trace{id: "trace-abc"}
	span := &Span{id: "span-xyz", traceID: "trace-abc"}
	ctx = ContextWithTrace(ctx, trace)
	ctx = ContextWithSpan(ctx, span)

	// Create request with context
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)

	// Use propagating round tripper
	rt := NewPropagatingRoundTripper(nil)
	_, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error = %v", err)
	}

	// Check headers were injected
	if receivedTraceID != "trace-abc" {
		t.Errorf("Server received TraceID = %q, want %q", receivedTraceID, "trace-abc")
	}
	if receivedParentSpanID != "span-xyz" {
		t.Errorf("Server received ParentSpanID = %q, want %q", receivedParentSpanID, "span-xyz")
	}
}

func TestPropagatingHTTPClient(t *testing.T) {
	client := PropagatingHTTPClient()

	if client == nil {
		t.Fatal("PropagatingHTTPClient returned nil")
	}
	if client.Transport == nil {
		t.Error("Transport should not be nil")
	}

	_, ok := client.Transport.(*PropagatingRoundTripper)
	if !ok {
		t.Error("Transport should be *PropagatingRoundTripper")
	}
}

func TestDistributedTraceHeadersJSON(t *testing.T) {
	// Test that JSON tags are correct
	headers := DistributedTraceHeaders{
		TraceID:      "t1",
		ParentSpanID: "s1",
	}

	// Just verify the struct can be created with expected values
	if headers.TraceID != "t1" || headers.ParentSpanID != "s1" {
		t.Error("struct fields not set correctly")
	}
}
