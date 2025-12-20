package anthropic

import (
	"net/http"
	"net/url"
	"testing"
)

func TestIsAnthropicRequest(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		{"Anthropic", "api.anthropic.com", true},
		{"OpenAI", "api.openai.com", false},
		{"Local", "localhost:8080", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{Host: tt.host},
			}
			if got := isAnthropicRequest(req); got != tt.want {
				t.Errorf("isAnthropicRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOperationName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/v1/messages", "anthropic.messages"},
		{"/messages", "anthropic.messages"},
		{"/v1/complete", "anthropic.complete"},
		{"/complete", "anthropic.complete"},
		{"/v1/models", "anthropic.api"},
		{"/unknown", "anthropic.api"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := getOperationName(tt.path); got != tt.want {
				t.Errorf("getOperationName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestNewTracingTransport(t *testing.T) {
	t.Run("with nil inner", func(t *testing.T) {
		tr := NewTracingTransport(nil, nil)
		if tr.inner != http.DefaultTransport {
			t.Error("inner should default to http.DefaultTransport")
		}
	})

	t.Run("with custom inner", func(t *testing.T) {
		custom := &http.Transport{}
		tr := NewTracingTransport(custom, nil)
		if tr.inner != custom {
			t.Error("inner should be custom transport")
		}
	})
}

func TestTracingHTTPClient(t *testing.T) {
	client := TracingHTTPClient(nil)
	if client == nil {
		t.Fatal("client should not be nil")
	}

	tr, ok := client.Transport.(*TracingTransport)
	if !ok {
		t.Fatal("transport should be *TracingTransport")
	}
	if tr.inner != http.DefaultTransport {
		t.Error("inner should default to http.DefaultTransport")
	}
}

func TestWrap(t *testing.T) {
	original := &http.Client{
		Timeout: 30,
	}

	wrapped := Wrap(original, nil)

	if wrapped == nil {
		t.Fatal("wrapped should not be nil")
	}

	// Timeout should be preserved
	if wrapped.Timeout != original.Timeout {
		t.Errorf("Timeout = %v, want %v", wrapped.Timeout, original.Timeout)
	}

	// Transport should be TracingTransport
	tr, ok := wrapped.Transport.(*TracingTransport)
	if !ok {
		t.Fatal("transport should be *TracingTransport")
	}

	// Inner should be DefaultTransport (since original.Transport was nil)
	if tr.inner != http.DefaultTransport {
		t.Error("inner should be http.DefaultTransport")
	}
}

func TestWrapWithCustomTransport(t *testing.T) {
	customTransport := &http.Transport{
		MaxIdleConns: 100,
	}
	original := &http.Client{
		Transport: customTransport,
	}

	wrapped := Wrap(original, nil)

	tr, ok := wrapped.Transport.(*TracingTransport)
	if !ok {
		t.Fatal("transport should be *TracingTransport")
	}

	if tr.inner != customTransport {
		t.Error("inner should be custom transport")
	}
}

func TestTracingProvider(t *testing.T) {
	p := TracingProvider(nil, WithAPIKey("test-key"), WithModel("claude-3-opus-20240229"))

	if p == nil {
		t.Fatal("provider should not be nil")
	}
	if p.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want %q", p.apiKey, "test-key")
	}
	if p.model != "claude-3-opus-20240229" {
		t.Errorf("model = %q, want %q", p.model, "claude-3-opus-20240229")
	}

	// HTTP client should have tracing transport
	if p.client == nil {
		t.Fatal("client should not be nil")
	}
	if _, ok := p.client.Transport.(*TracingTransport); !ok {
		t.Error("provider client should have TracingTransport")
	}
}

func TestTracingTransportNonAnthropicRequest(t *testing.T) {
	// Create a tracing transport
	tr := NewTracingTransport(http.DefaultTransport, nil)

	// Create a request to a non-Anthropic URL
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)

	// Should not trace, just pass through
	if isAnthropicRequest(req) {
		t.Error("example.com should not be detected as Anthropic request")
	}

	_ = tr
}
