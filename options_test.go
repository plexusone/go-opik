package opik

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultClientOptions(t *testing.T) {
	opts := defaultClientOptions()

	if opts.config == nil {
		t.Error("config should not be nil")
	}
	if opts.timeout != 60*time.Second {
		t.Errorf("timeout = %v, want %v", opts.timeout, 60*time.Second)
	}
}

func TestWithURL(t *testing.T) {
	opts := defaultClientOptions()
	WithURL("https://custom.example.com/api")(opts)

	if opts.config.URL != "https://custom.example.com/api" {
		t.Errorf("URL = %q, want %q", opts.config.URL, "https://custom.example.com/api")
	}
}

func TestWithAPIKey(t *testing.T) {
	opts := defaultClientOptions()
	WithAPIKey("test-api-key")(opts)

	if opts.config.APIKey != "test-api-key" {
		t.Errorf("APIKey = %q, want %q", opts.config.APIKey, "test-api-key")
	}
}

func TestWithWorkspace(t *testing.T) {
	opts := defaultClientOptions()
	WithWorkspace("my-workspace")(opts)

	if opts.config.Workspace != "my-workspace" {
		t.Errorf("Workspace = %q, want %q", opts.config.Workspace, "my-workspace")
	}
}

func TestWithProjectName(t *testing.T) {
	opts := defaultClientOptions()
	WithProjectName("my-project")(opts)

	if opts.config.ProjectName != "my-project" {
		t.Errorf("ProjectName = %q, want %q", opts.config.ProjectName, "my-project")
	}
}

func TestWithConfig(t *testing.T) {
	customConfig := &Config{
		URL:       "https://custom.url",
		APIKey:    "custom-key",
		Workspace: "custom-workspace",
	}

	opts := defaultClientOptions()
	WithConfig(customConfig)(opts)

	if opts.config != customConfig {
		t.Error("config should be the custom config")
	}
	if opts.config.URL != "https://custom.url" {
		t.Errorf("URL = %q, want %q", opts.config.URL, "https://custom.url")
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 30 * time.Second}

	opts := defaultClientOptions()
	WithHTTPClient(customClient)(opts)

	if opts.httpClient != customClient {
		t.Error("httpClient should be the custom client")
	}
}

func TestWithTimeout(t *testing.T) {
	opts := defaultClientOptions()
	WithTimeout(120 * time.Second)(opts)

	if opts.timeout != 120*time.Second {
		t.Errorf("timeout = %v, want %v", opts.timeout, 120*time.Second)
	}
}

func TestWithTracingDisabled(t *testing.T) {
	opts := defaultClientOptions()

	WithTracingDisabled(true)(opts)
	if !opts.config.TracingDisabled {
		t.Error("TracingDisabled should be true")
	}

	WithTracingDisabled(false)(opts)
	if opts.config.TracingDisabled {
		t.Error("TracingDisabled should be false")
	}
}

func TestMultipleClientOptions(t *testing.T) {
	opts := defaultClientOptions()

	// Apply multiple options
	WithURL("https://example.com")(opts)
	WithAPIKey("key123")(opts)
	WithWorkspace("ws")(opts)
	WithProjectName("proj")(opts)
	WithTimeout(90 * time.Second)(opts)

	if opts.config.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", opts.config.URL, "https://example.com")
	}
	if opts.config.APIKey != "key123" {
		t.Errorf("APIKey = %q, want %q", opts.config.APIKey, "key123")
	}
	if opts.config.Workspace != "ws" {
		t.Errorf("Workspace = %q, want %q", opts.config.Workspace, "ws")
	}
	if opts.config.ProjectName != "proj" {
		t.Errorf("ProjectName = %q, want %q", opts.config.ProjectName, "proj")
	}
	if opts.timeout != 90*time.Second {
		t.Errorf("timeout = %v, want %v", opts.timeout, 90*time.Second)
	}
}

// Trace Options Tests

func TestDefaultTraceOptions(t *testing.T) {
	opts := defaultTraceOptions()

	if opts.metadata == nil {
		t.Error("metadata should not be nil")
	}
	if opts.tags == nil {
		t.Error("tags should not be nil")
	}
	if len(opts.tags) != 0 {
		t.Errorf("tags should be empty, got %d", len(opts.tags))
	}
}

func TestWithTraceProject(t *testing.T) {
	opts := defaultTraceOptions()
	WithTraceProject("my-project")(opts)

	if opts.projectName != "my-project" {
		t.Errorf("projectName = %q, want %q", opts.projectName, "my-project")
	}
}

func TestWithTraceInput(t *testing.T) {
	opts := defaultTraceOptions()
	input := map[string]any{"prompt": "hello"}
	WithTraceInput(input)(opts)

	if opts.input == nil {
		t.Error("input should not be nil")
	}
}

func TestWithTraceOutput(t *testing.T) {
	opts := defaultTraceOptions()
	output := map[string]any{"response": "world"}
	WithTraceOutput(output)(opts)

	if opts.output == nil {
		t.Error("output should not be nil")
	}
}

func TestWithTraceMetadata(t *testing.T) {
	opts := defaultTraceOptions()
	metadata := map[string]any{"user": "test"}
	WithTraceMetadata(metadata)(opts)

	if opts.metadata == nil {
		t.Error("metadata should not be nil")
	}
	if opts.metadata["user"] != "test" {
		t.Errorf("metadata[user] = %v, want %q", opts.metadata["user"], "test")
	}
}

func TestWithTraceTags(t *testing.T) {
	opts := defaultTraceOptions()
	WithTraceTags("tag1", "tag2", "tag3")(opts)

	if len(opts.tags) != 3 {
		t.Errorf("tags length = %d, want 3", len(opts.tags))
	}
	if opts.tags[0] != "tag1" {
		t.Errorf("tags[0] = %q, want %q", opts.tags[0], "tag1")
	}
}

func TestWithTraceThreadID(t *testing.T) {
	opts := defaultTraceOptions()
	WithTraceThreadID("thread-123")(opts)

	if opts.threadID != "thread-123" {
		t.Errorf("threadID = %q, want %q", opts.threadID, "thread-123")
	}
}

// Span Options Tests

func TestDefaultSpanOptions(t *testing.T) {
	opts := defaultSpanOptions()

	if opts.spanType != "general" {
		t.Errorf("spanType = %q, want %q", opts.spanType, "general")
	}
	if opts.metadata == nil {
		t.Error("metadata should not be nil")
	}
	if opts.tags == nil {
		t.Error("tags should not be nil")
	}
}

func TestWithSpanType(t *testing.T) {
	tests := []struct {
		spanType string
	}{
		{SpanTypeLLM},
		{SpanTypeTool},
		{SpanTypeGeneral},
		{SpanTypeGuardrail},
	}

	for _, tt := range tests {
		t.Run(tt.spanType, func(t *testing.T) {
			opts := defaultSpanOptions()
			WithSpanType(tt.spanType)(opts)

			if opts.spanType != tt.spanType {
				t.Errorf("spanType = %q, want %q", opts.spanType, tt.spanType)
			}
		})
	}
}

func TestWithSpanInput(t *testing.T) {
	opts := defaultSpanOptions()
	input := map[string]any{"messages": []string{"hello"}}
	WithSpanInput(input)(opts)

	if opts.input == nil {
		t.Error("input should not be nil")
	}
}

func TestWithSpanOutput(t *testing.T) {
	opts := defaultSpanOptions()
	output := map[string]any{"response": "hi"}
	WithSpanOutput(output)(opts)

	if opts.output == nil {
		t.Error("output should not be nil")
	}
}

func TestWithSpanMetadata(t *testing.T) {
	opts := defaultSpanOptions()
	metadata := map[string]any{"temperature": 0.7}
	WithSpanMetadata(metadata)(opts)

	if opts.metadata == nil {
		t.Error("metadata should not be nil")
	}
}

func TestWithSpanTags(t *testing.T) {
	opts := defaultSpanOptions()
	WithSpanTags("production", "v1")(opts)

	if len(opts.tags) != 2 {
		t.Errorf("tags length = %d, want 2", len(opts.tags))
	}
}

func TestWithSpanModel(t *testing.T) {
	opts := defaultSpanOptions()
	WithSpanModel("gpt-4")(opts)

	if opts.model != "gpt-4" {
		t.Errorf("model = %q, want %q", opts.model, "gpt-4")
	}
}

func TestWithSpanProvider(t *testing.T) {
	opts := defaultSpanOptions()
	WithSpanProvider("openai")(opts)

	if opts.provider != "openai" {
		t.Errorf("provider = %q, want %q", opts.provider, "openai")
	}
}

func TestSpanTypeConstants(t *testing.T) {
	if SpanTypeLLM != "llm" {
		t.Errorf("SpanTypeLLM = %q, want %q", SpanTypeLLM, "llm")
	}
	if SpanTypeTool != "tool" {
		t.Errorf("SpanTypeTool = %q, want %q", SpanTypeTool, "tool")
	}
	if SpanTypeGeneral != "general" {
		t.Errorf("SpanTypeGeneral = %q, want %q", SpanTypeGeneral, "general")
	}
	if SpanTypeGuardrail != "guardrail" {
		t.Errorf("SpanTypeGuardrail = %q, want %q", SpanTypeGuardrail, "guardrail")
	}
}

func TestMultipleSpanOptions(t *testing.T) {
	opts := defaultSpanOptions()

	WithSpanType(SpanTypeLLM)(opts)
	WithSpanModel("gpt-4-turbo")(opts)
	WithSpanProvider("openai")(opts)
	WithSpanTags("chat", "production")(opts)
	WithSpanInput(map[string]any{"prompt": "test"})(opts)
	WithSpanOutput(map[string]any{"response": "result"})(opts)

	if opts.spanType != SpanTypeLLM {
		t.Errorf("spanType = %q, want %q", opts.spanType, SpanTypeLLM)
	}
	if opts.model != "gpt-4-turbo" {
		t.Errorf("model = %q, want %q", opts.model, "gpt-4-turbo")
	}
	if opts.provider != "openai" {
		t.Errorf("provider = %q, want %q", opts.provider, "openai")
	}
	if len(opts.tags) != 2 {
		t.Errorf("tags length = %d, want 2", len(opts.tags))
	}
	if opts.input == nil {
		t.Error("input should not be nil")
	}
	if opts.output == nil {
		t.Error("output should not be nil")
	}
}
