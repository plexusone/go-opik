# OpenAI Integration

Auto-trace OpenAI API calls and use OpenAI as an evaluation judge.

```go
import "github.com/grokify/go-comet-ml-opik/integrations/openai"
```

## Auto-Tracing

### Tracing HTTP Client

Wrap HTTP calls to automatically create spans:

```go
opikClient, _ := opik.NewClient()

// Create tracing HTTP client
httpClient := openai.TracingHTTPClient(opikClient)

// Use with OpenAI SDK
config := openai.DefaultConfig("your-api-key")
config.HTTPClient = httpClient
client := openai.NewClientWithConfig(config)
```

### Tracing Provider

Create a complete tracing provider:

```go
tracingProvider := openai.TracingProvider(opikClient,
    openai.WithModel("gpt-4o"),
)
```

### Wrap Existing Client

Add tracing to an existing HTTP client:

```go
existingClient := &http.Client{Timeout: 30 * time.Second}
tracingClient := openai.Wrap(existingClient, opikClient)
```

## What Gets Traced

Each API call creates a span with:

| Field | Description |
|-------|-------------|
| Type | `LLM` |
| Provider | `openai` |
| Model | Model name from request |
| Input | Request body (messages, parameters) |
| Output | Response body (completions, choices) |
| Metadata | Token usage, duration |

## Evaluation Provider

Use OpenAI as an LLM judge:

```go
provider := openai.NewProvider(
    openai.WithAPIKey("your-api-key"),
    openai.WithModel("gpt-4o"),
    openai.WithTemperature(0.0), // Deterministic for evaluation
)

// Use with evaluation metrics
relevance := llm.NewAnswerRelevance(provider)
hallucination := llm.NewHallucination(provider)
```

### Provider Options

| Option | Description |
|--------|-------------|
| `WithAPIKey(key)` | Set API key (default: `OPENAI_API_KEY` env) |
| `WithModel(model)` | Set model (default: `gpt-4o`) |
| `WithBaseURL(url)` | Custom endpoint (for Azure, proxies) |
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
    httpClient := openai.TracingHTTPClient(opikClient)

    // Configure OpenAI with tracing
    config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
    config.HTTPClient = httpClient
    oaiClient := openai.NewClientWithConfig(config)

    // Start a trace
    ctx, trace, _ := opik.StartTrace(ctx, opikClient, "chat-request")
    defer trace.End(ctx)

    // Make OpenAI call - automatically traced!
    resp, err := oaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: "gpt-4o",
        Messages: []openai.ChatCompletionMessage{
            {Role: "user", Content: "Hello!"},
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Azure OpenAI

Use with Azure OpenAI Service:

```go
provider := openai.NewProvider(
    openai.WithAPIKey("your-azure-key"),
    openai.WithBaseURL("https://your-resource.openai.azure.com/openai/deployments/your-deployment"),
)
```
