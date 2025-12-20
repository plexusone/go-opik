package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	// Test default status code
	if w.statusCode != http.StatusOK {
		t.Errorf("default statusCode = %d, want %d", w.statusCode, http.StatusOK)
	}

	// Test WriteHeader
	w.WriteHeader(http.StatusNotFound)
	if w.statusCode != http.StatusNotFound {
		t.Errorf("statusCode after WriteHeader = %d, want %d", w.statusCode, http.StatusNotFound)
	}
}

func TestNewTracingRoundTripper(t *testing.T) {
	t.Run("nil transport", func(t *testing.T) {
		rt := NewTracingRoundTripper(nil, "")
		if rt.transport != http.DefaultTransport {
			t.Error("nil transport should use http.DefaultTransport")
		}
		if rt.spanName != "http-request" {
			t.Errorf("default spanName = %q, want %q", rt.spanName, "http-request")
		}
	})

	t.Run("custom transport and name", func(t *testing.T) {
		custom := &http.Transport{}
		rt := NewTracingRoundTripper(custom, "custom-span")
		if rt.transport != custom {
			t.Error("custom transport should be used")
		}
		if rt.spanName != "custom-span" {
			t.Errorf("spanName = %q, want %q", rt.spanName, "custom-span")
		}
	})

	t.Run("empty span name", func(t *testing.T) {
		rt := NewTracingRoundTripper(nil, "")
		if rt.spanName != "http-request" {
			t.Errorf("empty spanName should default to %q", "http-request")
		}
	})
}

func TestTracingHTTPClient(t *testing.T) {
	client := TracingHTTPClient("test-span")

	if client == nil {
		t.Fatal("TracingHTTPClient returned nil")
	}
	if client.Transport == nil {
		t.Error("Transport should not be nil")
	}

	rt, ok := client.Transport.(*TracingRoundTripper)
	if !ok {
		t.Fatal("Transport should be *TracingRoundTripper")
	}
	if rt.spanName != "test-span" {
		t.Errorf("spanName = %q, want %q", rt.spanName, "test-span")
	}
}

func TestInjectTraceHeaders(t *testing.T) {
	// Note: This requires context with trace/span, but without the opik package
	// we can only test that it doesn't panic with empty context
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	InjectTraceHeaders(req.Context(), req)

	// With empty context, no headers should be set
	if got := req.Header.Get("X-Opik-Trace-ID"); got != "" {
		t.Errorf("X-Opik-Trace-ID = %q, want empty", got)
	}
	if got := req.Header.Get("X-Opik-Span-ID"); got != "" {
		t.Errorf("X-Opik-Span-ID = %q, want empty", got)
	}
}

func TestExtractTraceContext(t *testing.T) {
	t.Run("with headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("X-Opik-Trace-ID", "trace-123")
		req.Header.Set("X-Opik-Span-ID", "span-456")

		traceID, spanID := ExtractTraceContext(req)

		if traceID != "trace-123" {
			t.Errorf("traceID = %q, want %q", traceID, "trace-123")
		}
		if spanID != "span-456" {
			t.Errorf("spanID = %q, want %q", spanID, "span-456")
		}
	})

	t.Run("without headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		traceID, spanID := ExtractTraceContext(req)

		if traceID != "" {
			t.Errorf("traceID = %q, want empty", traceID)
		}
		if spanID != "" {
			t.Errorf("spanID = %q, want empty", spanID)
		}
	})
}

func TestRequestMetadata(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com/api/v1/users?page=1", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "127.0.0.1:8080"

	metadata := RequestMetadata(req)

	if metadata["method"] != "POST" {
		t.Errorf("method = %v, want %q", metadata["method"], "POST")
	}
	if metadata["path"] != "/api/v1/users" {
		t.Errorf("path = %v, want %q", metadata["path"], "/api/v1/users")
	}
	if metadata["query"] != "page=1" {
		t.Errorf("query = %v, want %q", metadata["query"], "page=1")
	}
	if metadata["host"] != "example.com" {
		t.Errorf("host = %v, want %q", metadata["host"], "example.com")
	}
	if metadata["remote_addr"] != "127.0.0.1:8080" {
		t.Errorf("remote_addr = %v, want %q", metadata["remote_addr"], "127.0.0.1:8080")
	}
	if metadata["user_agent"] != "test-agent" {
		t.Errorf("user_agent = %v, want %q", metadata["user_agent"], "test-agent")
	}
	if metadata["protocol"] != "HTTP/1.1" {
		t.Errorf("protocol = %v, want %q", metadata["protocol"], "HTTP/1.1")
	}
}

func TestResponseMetadata(t *testing.T) {
	resp := &http.Response{
		StatusCode:    200,
		Status:        "200 OK",
		ContentLength: 1234,
	}
	duration := 150 * time.Millisecond

	metadata := ResponseMetadata(resp, duration)

	if metadata["status_code"] != 200 {
		t.Errorf("status_code = %v, want 200", metadata["status_code"])
	}
	if metadata["status"] != "200 OK" {
		t.Errorf("status = %v, want %q", metadata["status"], "200 OK")
	}
	if metadata["content_length"] != int64(1234) {
		t.Errorf("content_length = %v, want 1234", metadata["content_length"])
	}
	if metadata["duration_ms"] != int64(150) {
		t.Errorf("duration_ms = %v, want 150", metadata["duration_ms"])
	}
}

func TestStatusCodeCategory(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{200, "success"},
		{201, "success"},
		{299, "success"},
		{400, "client_error"},
		{404, "client_error"},
		{499, "client_error"},
		{500, "server_error"},
		{503, "server_error"},
		{100, "other"},
		{301, "other"},
		{399, "other"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.code), func(t *testing.T) {
			got := StatusCodeCategory(tt.code)
			if got != tt.want {
				t.Errorf("StatusCodeCategory(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{500 * time.Microsecond, "500µs"},
		{999 * time.Microsecond, "999µs"},
		{1 * time.Millisecond, "1ms"},
		{500 * time.Millisecond, "500ms"},
		{999 * time.Millisecond, "999ms"},
		{1 * time.Second, "1s"},
		{1500 * time.Millisecond, "1.5s"},
		{2 * time.Second, "2s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestTracingRoundTripperWithoutContext(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create request without trace context
	req, _ := http.NewRequest("GET", server.URL, nil)

	// Use tracing round tripper
	rt := NewTracingRoundTripper(nil, "test")
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip error = %v", err)
	}
	defer resp.Body.Close()

	// Should succeed without creating span (no context)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestMetadataContentLength(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com", nil)
	req.ContentLength = 100

	metadata := RequestMetadata(req)

	if metadata["content_length"] != int64(100) {
		t.Errorf("content_length = %v, want 100", metadata["content_length"])
	}
}
