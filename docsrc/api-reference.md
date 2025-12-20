# API Reference

## Client

### Creating a Client

```go
// Default client (uses env vars and config file)
client, err := opik.NewClient()

// With options
client, err := opik.NewClient(
    opik.WithURL("https://www.comet.com/opik/api"),
    opik.WithAPIKey("your-api-key"),
    opik.WithWorkspace("your-workspace"),
    opik.WithProjectName("My Project"),
    opik.WithHTTPClient(customHTTPClient),
)
```

### Client Options

| Option | Description |
|--------|-------------|
| `WithURL(url)` | API endpoint URL |
| `WithAPIKey(key)` | API key |
| `WithWorkspace(name)` | Workspace name |
| `WithProjectName(name)` | Default project |
| `WithHTTPClient(client)` | Custom HTTP client |

## Accessing the Generated API

For advanced usage, access the underlying ogen-generated API client:

```go
api := client.API()

// Use generated methods directly
resp, err := api.GetProjects(ctx, api.GetProjectsParams{
    Page: api.NewOptInt32(1),
    Size: api.NewOptInt32(100),
})
```

## Error Handling

### Error Types

```go
// Check for specific errors
if opik.IsNotFound(err) {
    // Resource doesn't exist
}

if opik.IsUnauthorized(err) {
    // Invalid credentials
}

if opik.IsRateLimited(err) {
    // Too many requests
}

if opik.IsBadRequest(err) {
    // Invalid request parameters
}
```

### Error Details

```go
if apiErr, ok := err.(*opik.APIError); ok {
    fmt.Printf("Status: %d\n", apiErr.StatusCode)
    fmt.Printf("Message: %s\n", apiErr.Message)
}
```

## Types

### Trace

```go
type Trace struct {
    // Methods
    ID() string
    Name() string
    End(ctx context.Context, opts ...TraceOption) error
    Update(ctx context.Context, opts ...TraceOption) error
    Span(ctx context.Context, name string, opts ...SpanOption) (*Span, error)
    AddFeedbackScore(ctx context.Context, name string, value float64, reason string) error
}
```

### Span

```go
type Span struct {
    // Methods
    ID() string
    TraceID() string
    Name() string
    End(ctx context.Context, opts ...SpanOption) error
    Update(ctx context.Context, opts ...SpanOption) error
    Span(ctx context.Context, name string, opts ...SpanOption) (*Span, error)
    AddFeedbackScore(ctx context.Context, name string, value float64, reason string) error
}
```

### Dataset

```go
type Dataset struct {
    // Fields
    Name        string
    Description string
    ID          string

    // Methods
    ID() string
    InsertItem(ctx context.Context, data map[string]any) error
    InsertItems(ctx context.Context, items []map[string]any) error
    GetItems(ctx context.Context, page, size int) ([]*DatasetItem, error)
    Delete(ctx context.Context) error
}
```

### Experiment

```go
type Experiment struct {
    // Fields
    Name string
    ID   string

    // Methods
    LogItem(ctx context.Context, itemID, traceID string, opts ...ExperimentItemOption) error
    Complete(ctx context.Context) error
    Cancel(ctx context.Context) error
    Delete(ctx context.Context) error
}
```

### Prompt

```go
type Prompt struct {
    // Fields
    Name        string
    Description string

    // Methods
    CreateVersion(ctx context.Context, template string, opts ...VersionOption) (*PromptVersion, error)
    GetVersions(ctx context.Context, page, size int) ([]*PromptVersion, error)
}

type PromptVersion struct {
    // Fields
    Commit   string
    Template string

    // Methods
    Render(vars map[string]string) string
    ExtractVariables() []string
}
```

## Span Types

```go
const (
    SpanTypeLLM     = "llm"
    SpanTypeTool    = "tool"
    SpanTypeGeneral = "general"
)
```

## Context Functions

```go
// Start trace with context
ctx, trace, err := opik.StartTrace(ctx, client, "name", opts...)

// Start span with context
ctx, span, err := opik.StartSpan(ctx, "name", opts...)

// Get from context
trace := opik.TraceFromContext(ctx)
span := opik.SpanFromContext(ctx)

// Add to context
ctx = opik.ContextWithTrace(ctx, trace)
ctx = opik.ContextWithSpan(ctx, span)
```

## Distributed Tracing

```go
// Get headers for propagation
headers := opik.GetDistributedTraceHeaders(ctx)

// Inject into HTTP request
opik.InjectDistributedTraceHeaders(ctx, req)

// Extract from HTTP request
headers := opik.ExtractDistributedTraceHeaders(req)

// Continue a distributed trace
ctx, span, err := client.ContinueTrace(ctx, headers, "name", opts...)
```

## Configuration

```go
// Load configuration
cfg := opik.LoadConfig()

// Save configuration
err := opik.SaveConfig(cfg)

// Config struct
type Config struct {
    URL         string
    APIKey      string
    Workspace   string
    ProjectName string
}
```

## Testing Utilities

```go
// Record traces locally (no server)
client := opik.RecordTracesLocally("project-name")

// Access recorded data
recording := client.Recording()
traces := recording.Traces()
spans := recording.Spans()
```
