package opik

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Client.
type Option func(*clientOptions)

// clientOptions holds the options for creating a Client.
type clientOptions struct {
	config     *Config
	httpClient *http.Client
	timeout    time.Duration
}

func defaultClientOptions() *clientOptions {
	return &clientOptions{
		config:  LoadConfig(),
		timeout: 60 * time.Second,
	}
}

// WithURL sets the API URL.
func WithURL(url string) Option {
	return func(o *clientOptions) {
		o.config.URL = url
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) Option {
	return func(o *clientOptions) {
		o.config.APIKey = apiKey
	}
}

// WithWorkspace sets the workspace name.
func WithWorkspace(workspace string) Option {
	return func(o *clientOptions) {
		o.config.Workspace = workspace
	}
}

// WithProjectName sets the default project name.
func WithProjectName(projectName string) Option {
	return func(o *clientOptions) {
		o.config.ProjectName = projectName
	}
}

// WithConfig sets the entire configuration.
func WithConfig(config *Config) Option {
	return func(o *clientOptions) {
		o.config = config
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *clientOptions) {
		o.httpClient = client
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithTracingDisabled disables tracing.
func WithTracingDisabled(disabled bool) Option {
	return func(o *clientOptions) {
		o.config.TracingDisabled = disabled
	}
}

// TraceOption is a functional option for configuring a Trace.
type TraceOption func(*traceOptions)

type traceOptions struct {
	projectName string
	input       any
	output      any
	metadata    map[string]any
	tags        []string
	threadID    string
}

func defaultTraceOptions() *traceOptions {
	return &traceOptions{
		metadata: make(map[string]any),
		tags:     []string{},
	}
}

// WithTraceProject sets the project name for the trace.
func WithTraceProject(projectName string) TraceOption {
	return func(o *traceOptions) {
		o.projectName = projectName
	}
}

// WithTraceInput sets the input for the trace.
func WithTraceInput(input any) TraceOption {
	return func(o *traceOptions) {
		o.input = input
	}
}

// WithTraceOutput sets the output for the trace.
func WithTraceOutput(output any) TraceOption {
	return func(o *traceOptions) {
		o.output = output
	}
}

// WithTraceMetadata sets the metadata for the trace.
func WithTraceMetadata(metadata map[string]any) TraceOption {
	return func(o *traceOptions) {
		o.metadata = metadata
	}
}

// WithTraceTags sets the tags for the trace.
func WithTraceTags(tags ...string) TraceOption {
	return func(o *traceOptions) {
		o.tags = tags
	}
}

// WithTraceThreadID sets the thread ID for the trace.
func WithTraceThreadID(threadID string) TraceOption {
	return func(o *traceOptions) {
		o.threadID = threadID
	}
}

// SpanOption is a functional option for configuring a Span.
type SpanOption func(*spanOptions)

type spanOptions struct {
	spanType string
	input    any
	output   any
	metadata map[string]any
	tags     []string
	model    string
	provider string
}

func defaultSpanOptions() *spanOptions {
	return &spanOptions{
		spanType: "general",
		metadata: make(map[string]any),
		tags:     []string{},
	}
}

// WithSpanType sets the type of the span (general, llm, tool, guardrail).
func WithSpanType(spanType string) SpanOption {
	return func(o *spanOptions) {
		o.spanType = spanType
	}
}

// WithSpanInput sets the input for the span.
func WithSpanInput(input any) SpanOption {
	return func(o *spanOptions) {
		o.input = input
	}
}

// WithSpanOutput sets the output for the span.
func WithSpanOutput(output any) SpanOption {
	return func(o *spanOptions) {
		o.output = output
	}
}

// WithSpanMetadata sets the metadata for the span.
func WithSpanMetadata(metadata map[string]any) SpanOption {
	return func(o *spanOptions) {
		o.metadata = metadata
	}
}

// WithSpanTags sets the tags for the span.
func WithSpanTags(tags ...string) SpanOption {
	return func(o *spanOptions) {
		o.tags = tags
	}
}

// WithSpanModel sets the model name for LLM spans.
func WithSpanModel(model string) SpanOption {
	return func(o *spanOptions) {
		o.model = model
	}
}

// WithSpanProvider sets the provider name for LLM spans.
func WithSpanProvider(provider string) SpanOption {
	return func(o *spanOptions) {
		o.provider = provider
	}
}

// SpanTypeLLM is the span type for LLM calls.
const SpanTypeLLM = "llm"

// SpanTypeTool is the span type for tool calls.
const SpanTypeTool = "tool"

// SpanTypeGeneral is the span type for general operations.
const SpanTypeGeneral = "general"

// SpanTypeGuardrail is the span type for guardrail checks.
const SpanTypeGuardrail = "guardrail"
