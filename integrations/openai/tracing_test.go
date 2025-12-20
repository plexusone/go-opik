package openai

import (
	"net/http"
	"net/url"
	"testing"
)

func TestIsOpenAIRequest(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		{"OpenAI", "api.openai.com", true},
		{"Azure OpenAI", "openai.azure.com", true},
		{"Other API", "api.anthropic.com", false},
		{"Local", "localhost:8080", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{Host: tt.host},
			}
			if got := isOpenAIRequest(req); got != tt.want {
				t.Errorf("isOpenAIRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOperationName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/v1/chat/completions", "openai.chat.completion"},
		{"/chat/completions", "openai.chat.completion"},
		{"/v1/completions", "openai.completion"},
		{"/v1/embeddings", "openai.embeddings"},
		{"/v1/images/generations", "openai.images"},
		{"/v1/audio/transcriptions", "openai.audio"},
		{"/v1/moderations", "openai.moderations"},
		{"/v1/models", "openai.api"},
		{"/unknown", "openai.api"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := getOperationName(tt.path); got != tt.want {
				t.Errorf("getOperationName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"/v1/chat/completions", "chat", true},
		{"", "foo", false},
		{"foo", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"/"+tt.substr, func(t *testing.T) {
			if got := contains(tt.s, tt.substr); got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
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
	p := TracingProvider(nil, WithAPIKey("test-key"), WithModel("gpt-4"))

	if p == nil {
		t.Fatal("provider should not be nil")
	}
	if p.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want %q", p.apiKey, "test-key")
	}
	if p.model != "gpt-4" {
		t.Errorf("model = %q, want %q", p.model, "gpt-4")
	}

	// HTTP client should have tracing transport
	if p.client == nil {
		t.Fatal("client should not be nil")
	}
	if _, ok := p.client.Transport.(*TracingTransport); !ok {
		t.Error("provider client should have TracingTransport")
	}
}

func TestTracingTransportNonOpenAIRequest(t *testing.T) {
	// Create a tracing transport
	tr := NewTracingTransport(http.DefaultTransport, nil)

	// Create a request to a non-OpenAI URL
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)

	// Should not trace, just pass through
	// This test verifies that non-OpenAI requests pass through without issues
	if isOpenAIRequest(req) {
		t.Error("example.com should not be detected as OpenAI request")
	}

	// The RoundTrip will use inner transport directly
	// We can't easily test the actual HTTP call without a server,
	// but we verify the detection logic
	_ = tr
}
