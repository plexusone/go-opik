# Anthropic Integration

Auto-trace Anthropic API calls and use Claude as an evaluation judge.

```go
import "github.com/grokify/go-comet-ml-opik/integrations/anthropic"
```

## Auto-Tracing

### Tracing HTTP Client

Wrap HTTP calls to automatically create spans:

```go
opikClient, _ := opik.NewClient()

// Create tracing HTTP client
httpClient := anthropic.TracingHTTPClient(opikClient)

// Use with your Anthropic client
```

### Tracing Provider

Create a complete tracing provider:

```go
tracingProvider := anthropic.TracingProvider(opikClient,
    anthropic.WithModel("claude-sonnet-4-20250514"),
)
```

### Wrap Existing Client

Add tracing to an existing HTTP client:

```go
existingClient := &http.Client{Timeout: 60 * time.Second}
tracingClient := anthropic.Wrap(existingClient, opikClient)
```

## What Gets Traced

Each API call creates a span with:

| Field | Description |
|-------|-------------|
| Type | `LLM` |
| Provider | `anthropic` |
| Model | Model name from request |
| Input | Request body (messages, system prompt) |
| Output | Response body (content, stop reason) |
| Metadata | Token usage (input_tokens, output_tokens), duration |

## Evaluation Provider

Use Claude as an LLM judge:

```go
provider := anthropic.NewProvider(
    anthropic.WithAPIKey("your-api-key"),
    anthropic.WithModel("claude-sonnet-4-20250514"),
    anthropic.WithTemperature(0.0),
)

// Use with evaluation metrics
relevance := llm.NewAnswerRelevance(provider)
coherence := llm.NewCoherence(provider)
```

### Provider Options

| Option | Description |
|--------|-------------|
| `WithAPIKey(key)` | Set API key (default: `ANTHROPIC_API_KEY` env) |
| `WithModel(model)` | Set model (default: `claude-sonnet-4-20250514`) |
| `WithBaseURL(url)` | Custom endpoint |
| `WithHTTPClient(client)` | Custom HTTP client |
| `WithTemperature(temp)` | Generation temperature |
| `WithMaxTokens(max)` | Maximum tokens |

## Complete Example

```go
func main() {
    ctx := context.Background()

    // Create Opik client
    opikClient, _ := opik.NewClient()

    // Create tracing HTTP client
    httpClient := anthropic.TracingHTTPClient(opikClient)

    // Start a trace
    ctx, trace, _ := opik.StartTrace(ctx, opikClient, "claude-request")
    defer trace.End(ctx)

    // Make Anthropic API call
    req, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.anthropic.com/v1/messages",
        bytes.NewReader(requestBody))

    req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("content-type", "application/json")

    // Call is automatically traced!
    resp, err := httpClient.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // Process response...
}
```

## Supported Operations

The integration traces these Anthropic API endpoints:

| Endpoint | Span Name |
|----------|-----------|
| `/v1/messages` | `anthropic.messages` |
| `/v1/complete` | `anthropic.complete` |
| Other | `anthropic.api` |
