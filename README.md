# Comet ML Opik SDK for Go

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

Go SDK for [Opik](https://github.com/comet-ml/opik) - an open-source LLM observability platform by Comet ML.

## Installation

```bash
go get github.com/grokify/go-comet-ml-opik
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/grokify/go-comet-ml-opik"
)

func main() {
    // Create client (uses OPIK_API_KEY and OPIK_WORKSPACE env vars for Opik Cloud)
    client, err := opik.NewClient(
        opik.WithProjectName("My Project"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Create a trace
    trace, _ := client.Trace(ctx, "my-trace",
        opik.WithTraceInput(map[string]any{"prompt": "Hello"}),
    )

    // Create a span for an LLM call
    span, _ := trace.Span(ctx, "llm-call",
        opik.WithSpanType(opik.SpanTypeLLM),
        opik.WithSpanModel("gpt-4"),
    )

    // Do your LLM call here...

    // End span with output
    span.End(ctx, opik.WithSpanOutput(map[string]any{"response": "Hi!"}))

    // End trace
    trace.End(ctx)
}
```

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPIK_URL_OVERRIDE` | API endpoint URL |
| `OPIK_API_KEY` | API key for Opik Cloud |
| `OPIK_WORKSPACE` | Workspace name for Opik Cloud |
| `OPIK_PROJECT_NAME` | Default project name |
| `OPIK_TRACK_DISABLE` | Set to "true" to disable tracing |

### Config File

Create `~/.opik.config`:

```ini
[opik]
url_override = https://www.comet.com/opik/api
api_key = your-api-key
workspace = your-workspace
project_name = My Project
```

### Programmatic Configuration

```go
client, err := opik.NewClient(
    opik.WithURL("https://www.comet.com/opik/api"),
    opik.WithAPIKey("your-api-key"),
    opik.WithWorkspace("your-workspace"),
    opik.WithProjectName("My Project"),
)
```

## Features

### Traces and Spans

```go
// Create a trace
trace, _ := client.Trace(ctx, "my-trace",
    opik.WithTraceInput(input),
    opik.WithTraceMetadata(map[string]any{"key": "value"}),
    opik.WithTraceTags("tag1", "tag2"),
)

// Create spans (supports nesting)
span1, _ := trace.Span(ctx, "outer-span")
span2, _ := span1.Span(ctx, "inner-span")

// End spans and traces
span2.End(ctx, opik.WithSpanOutput(output))
span1.End(ctx)
trace.End(ctx)
```

### Span Types

```go
// LLM spans
span, _ := trace.Span(ctx, "llm-call",
    opik.WithSpanType(opik.SpanTypeLLM),
    opik.WithSpanModel("gpt-4"),
    opik.WithSpanProvider("openai"),
)

// Tool spans
span, _ := trace.Span(ctx, "tool-call",
    opik.WithSpanType(opik.SpanTypeTool),
)

// General spans (default)
span, _ := trace.Span(ctx, "processing",
    opik.WithSpanType(opik.SpanTypeGeneral),
)
```

### Feedback Scores

```go
// Add feedback to traces
trace.AddFeedbackScore(ctx, "accuracy", 0.95, "High accuracy")

// Add feedback to spans
span.AddFeedbackScore(ctx, "relevance", 0.87, "Mostly relevant")
```

### Context Propagation

```go
// Start trace with context
ctx, trace, _ := opik.StartTrace(ctx, client, "my-trace")

// Start nested spans
ctx, span1, _ := opik.StartSpan(ctx, "span-1")
ctx, span2, _ := opik.StartSpan(ctx, "span-2") // Automatically nested under span1

// Get current trace/span from context
currentTrace := opik.TraceFromContext(ctx)
currentSpan := opik.SpanFromContext(ctx)
```

### Distributed Tracing

```go
// Inject trace headers into outgoing requests
opik.InjectDistributedTraceHeaders(ctx, req)

// Extract trace headers from incoming requests
headers := opik.ExtractDistributedTraceHeaders(req)

// Continue a distributed trace
ctx, span, _ := client.ContinueTrace(ctx, headers, "handle-request")

// Use propagating HTTP client
httpClient := opik.PropagatingHTTPClient()
```

### Streaming Support

```go
// Start a streaming span
ctx, streamSpan, _ := opik.StartStreamingSpan(ctx, "stream-response",
    opik.WithSpanType(opik.SpanTypeLLM),
)

// Add chunks as they arrive
for chunk := range chunks {
    streamSpan.AddChunk(chunk.Content,
        opik.WithChunkTokenCount(chunk.Tokens),
    )
}

// Mark final chunk
streamSpan.AddChunk(lastChunk, opik.WithChunkFinishReason("stop"))

// End with accumulated data
streamSpan.End(ctx)
```

### Datasets

```go
// Create a dataset
dataset, _ := client.CreateDataset(ctx, "my-dataset",
    opik.WithDatasetDescription("Test data for evaluation"),
    opik.WithDatasetTags("test", "evaluation"),
)

// Insert items
dataset.InsertItem(ctx, map[string]any{
    "input":    "What is the capital of France?",
    "expected": "Paris",
})

// Insert multiple items
items := []map[string]any{
    {"input": "2+2", "expected": "4"},
    {"input": "3+3", "expected": "6"},
}
dataset.InsertItems(ctx, items)

// Retrieve items
items, _ := dataset.GetItems(ctx, 1, 100)

// List datasets
datasets, _ := client.ListDatasets(ctx, 1, 100)

// Get dataset by name
dataset, _ := client.GetDatasetByName(ctx, "my-dataset")

// Delete dataset
dataset.Delete(ctx)
```

### Experiments

```go
// Create an experiment
experiment, _ := client.CreateExperiment(ctx, "my-dataset",
    opik.WithExperimentName("gpt-4-evaluation"),
    opik.WithExperimentMetadata(map[string]any{"model": "gpt-4"}),
)

// Log experiment items
experiment.LogItem(ctx, datasetItemID, traceID,
    opik.WithExperimentItemInput(input),
    opik.WithExperimentItemOutput(output),
)

// Complete or cancel experiments
experiment.Complete(ctx)
experiment.Cancel(ctx)

// List experiments
experiments, _ := client.ListExperiments(ctx, datasetID, 1, 100)

// Delete experiment
experiment.Delete(ctx)
```

### Prompts

```go
// Create a prompt
prompt, _ := client.CreatePrompt(ctx, "greeting-prompt",
    opik.WithPromptDescription("Greeting template"),
    opik.WithPromptTemplate("Hello, {{name}}! Welcome to {{place}}."),
    opik.WithPromptTags("greeting", "template"),
)

// Get prompt by name
version, _ := client.GetPromptByName(ctx, "greeting-prompt", "")

// Render template with variables
rendered := version.Render(map[string]string{
    "name":  "Alice",
    "place": "Wonderland",
})
// Result: "Hello, Alice! Welcome to Wonderland."

// Extract variables from template
vars := version.ExtractVariables()
// Result: ["name", "place"]

// Create new version
newVersion, _ := prompt.CreateVersion(ctx, "Hi, {{name}}!",
    opik.WithVersionChangeDescription("Simplified greeting"),
)

// List all versions
versions, _ := prompt.GetVersions(ctx, 1, 100)

// List all prompts
prompts, _ := client.ListPrompts(ctx, 1, 100)
```

### HTTP Middleware

```go
import "github.com/grokify/go-comet-ml-opik/middleware"

// Wrap HTTP handlers with automatic tracing
handler := middleware.TracingMiddleware(client, "api-request")(yourHandler)

// Use tracing HTTP client for outgoing requests
httpClient := middleware.TracingHTTPClient("external-call")
resp, _ := httpClient.Get("https://api.example.com/data")

// Or wrap an existing transport
transport := middleware.NewTracingRoundTripper(http.DefaultTransport, "api-call")
httpClient := &http.Client{Transport: transport}
```

### Local Recording (Testing)

```go
// Record traces locally without sending to server
client := opik.RecordTracesLocally("my-project")
trace, _ := client.Trace(ctx, "test-trace")
span, _ := trace.Span(ctx, "test-span")
span.End(ctx)
trace.End(ctx)

// Access recorded data
traces := client.Recording().Traces()
spans := client.Recording().Spans()
```

### Batching

```go
// Create a batching client for efficient API calls
client, _ := opik.NewBatchingClient(
    opik.WithProjectName("My Project"),
)

// Operations are batched automatically
client.AddFeedbackAsync("trace", traceID, "score", 0.95, "reason")

// Flush pending operations
client.Flush(5 * time.Second)

// Close and flush on shutdown
defer client.Close(10 * time.Second)
```

### Attachments

```go
// Create attachment from file
attachment, _ := opik.NewAttachmentFromFile("/path/to/image.png")

// Create attachment from bytes
attachment := opik.NewAttachmentFromBytes("data.json", jsonBytes, "application/json")

// Create text attachment
attachment := opik.NewTextAttachment("notes.txt", "Some text content")

// Get data URL for embedding
dataURL := attachment.ToDataURL()
```

## Evaluation Framework

### Heuristic Metrics

```go
import (
    "github.com/grokify/go-comet-ml-opik/evaluation"
    "github.com/grokify/go-comet-ml-opik/evaluation/heuristic"
)

// Create metrics
metrics := []evaluation.Metric{
    heuristic.NewEquals(false),           // Case-insensitive equality
    heuristic.NewContains(false),         // Substring check
    heuristic.NewIsJSON(),                // JSON validation
    heuristic.NewLevenshteinSimilarity(false), // Edit distance
    heuristic.NewBLEU(4),                 // BLEU score
    heuristic.NewROUGE(1.0),              // ROUGE-L score
    heuristic.MustRegexMatch(`\d+`),      // Regex matching
}

// Evaluate
engine := evaluation.NewEngine(metrics, evaluation.WithConcurrency(4))
input := evaluation.NewMetricInput("What is 2+2?", "The answer is 4.")
input = input.WithExpected("4")

result := engine.EvaluateOne(ctx, input)
fmt.Printf("Average score: %.2f\n", result.AverageScore())
```

### LLM Judge Metrics

```go
import (
    "github.com/grokify/go-comet-ml-opik/evaluation/llm"
    "github.com/grokify/go-comet-ml-opik/integrations/openai"
)

// Create LLM provider
provider := openai.NewProvider(openai.WithModel("gpt-4o"))

// Create LLM-based metrics
metrics := []evaluation.Metric{
    llm.NewAnswerRelevance(provider),
    llm.NewHallucination(provider),
    llm.NewFactuality(provider),
    llm.NewCoherence(provider),
    llm.NewHelpfulness(provider),
}

// Custom judge with custom prompt
customJudge := llm.NewCustomJudge("tone_check", `
Evaluate whether the response maintains a professional tone.

User message: {{input}}
AI response: {{output}}

Return JSON: {"score": <0.0-1.0>, "reason": "<explanation>"}
`, provider)
```

### G-EVAL

```go
geval := llm.NewGEval(provider, "fluency and coherence").
    WithEvaluationSteps([]string{
        "Check if the response is grammatically correct",
        "Evaluate logical flow of ideas",
        "Assess clarity of expression",
    })

score := geval.Score(ctx, input)
```

## LLM Provider Integrations

### OpenAI

```go
import "github.com/grokify/go-comet-ml-opik/integrations/openai"

// Create provider for evaluation
provider := openai.NewProvider(
    openai.WithAPIKey("your-api-key"),
    openai.WithModel("gpt-4o"),
)

// Create tracing provider (auto-traces all calls)
tracingProvider := openai.TracingProvider(opikClient,
    openai.WithModel("gpt-4o"),
)

// Use tracing HTTP client with existing code
httpClient := openai.TracingHTTPClient(opikClient)
```

### Anthropic

```go
import "github.com/grokify/go-comet-ml-opik/integrations/anthropic"

// Create provider for evaluation
provider := anthropic.NewProvider(
    anthropic.WithAPIKey("your-api-key"),
    anthropic.WithModel("claude-sonnet-4-20250514"),
)

// Create tracing provider (auto-traces all calls)
tracingProvider := anthropic.TracingProvider(opikClient)

// Use tracing HTTP client with existing code
httpClient := anthropic.TracingHTTPClient(opikClient)
```

## CLI Tool

```bash
# Install CLI
go install github.com/grokify/go-comet-ml-opik/cmd/opik@latest

# Configure
opik configure -api-key=your-key -workspace=your-workspace

# List projects
opik projects -list

# Create project
opik projects -create="New Project"

# List traces
opik traces -list -project="My Project" -limit=20

# List datasets
opik datasets -list

# Create dataset
opik datasets -create="evaluation-data"

# List experiments
opik experiments -list -dataset="my-dataset"
```

## API Client Access

For advanced usage, access the underlying ogen-generated API client:

```go
api := client.API()
// Use api.* methods for full API access
```

## Error Handling

```go
// Check for specific errors
if opik.IsNotFound(err) {
    // Handle not found
}

if opik.IsUnauthorized(err) {
    // Handle auth failure
}

if opik.IsRateLimited(err) {
    // Handle rate limiting
}
```

## Tutorials

### Agentic Observability

For integrating Opik with agent frameworks like Google ADK and Eino, see the [Agentic Observability Tutorial](docsrc/tutorials/agentic-observability.md). This tutorial covers:

- Tracing Google ADK agents with tools
- Tracing Eino workflow graphs
- Multi-agent orchestration observability
- Best practices for agent debugging and monitoring

## Development

### Running Tests

```bash
go test ./...
```

### Running Linter

```bash
golangci-lint run
```

### Regenerating API Client

The SDK uses [ogen](https://github.com/ogen-go/ogen) to generate a type-safe API client from the Opik OpenAPI specification. When the upstream API changes, regenerate the client:

**Prerequisites:**

```bash
# Install ogen
go install github.com/ogen-go/ogen/cmd/ogen@latest
```

**Update the OpenAPI spec:**

```bash
# Download latest spec from Opik repository
curl -o openapi/openapi.yaml \
  https://raw.githubusercontent.com/comet-ml/opik/main/sdks/code_generation/fern/openapi/openapi.yaml
```

**Generate the client:**

```bash
./generate.sh
```

This script runs ogen, applies necessary fixes, and verifies the build. For detailed documentation on the generation process and troubleshooting, see the [Development Guide](docsrc/getting-started/development.md).

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related

- [Opik](https://github.com/comet-ml/opik) - Open-source LLM observability platform
- [Opik Python SDK](https://github.com/comet-ml/opik/tree/main/sdks/python) - Official Python SDK
- [Opik Documentation](https://www.comet.com/docs/opik/) - Official documentation

 [build-status-svg]: https://github.com/grokify/go-comet-ml-opik/actions/workflows/ci.yaml/badge.svg?branch=main
 [build-status-url]: https://github.com/grokify/go-comet-ml-opik/actions/workflows/ci.yaml
 [lint-status-svg]: https://github.com/grokify/go-comet-ml-opik/actions/workflows/lint.yaml/badge.svg?branch=main
 [lint-status-url]: https://github.com/grokify/go-comet-ml-opik/actions/workflows/lint.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/go-comet-ml-opik
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/go-comet-ml-opik
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/go-comet-ml-opik
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/go-comet-ml-opik
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/go-comet-ml-opik/blob/master/LICENSE
 [used-by-svg]: https://sourcegraph.com/github.com/grokify/go-comet-ml-opik/-/badge.svg
 [used-by-url]: https://sourcegraph.com/github.com/grokify/go-comet-ml-opik?badge