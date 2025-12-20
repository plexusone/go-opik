# Traces and Spans

Traces and spans are the core building blocks for observability in Opik.

## Concepts

- **Trace**: Represents a complete request or workflow through your application
- **Span**: Represents a single operation within a trace (e.g., an LLM call, tool execution)

Spans can be nested to show parent-child relationships, creating a tree structure that represents your application's execution flow.

## Creating Traces

```go
client, _ := opik.NewClient()

// Create a simple trace
trace, _ := client.Trace(ctx, "my-trace")

// Create a trace with options
trace, _ := client.Trace(ctx, "my-trace",
    opik.WithTraceInput(map[string]any{"prompt": "Hello"}),
    opik.WithTraceMetadata(map[string]any{"user_id": "123"}),
    opik.WithTraceTags("production", "v2"),
)

// End the trace
trace.End(ctx)

// End with output
trace.End(ctx, opik.WithTraceOutput(map[string]any{"response": "Hi!"}))
```

## Creating Spans

Spans are created from traces or other spans:

```go
// Create a span from a trace
span, _ := trace.Span(ctx, "process-input")

// Create a nested span from another span
childSpan, _ := span.Span(ctx, "validate-input")

// End spans (in reverse order)
childSpan.End(ctx)
span.End(ctx)
```

## Span Types

Use span types to categorize operations:

```go
// LLM call span
span, _ := trace.Span(ctx, "llm-call",
    opik.WithSpanType(opik.SpanTypeLLM),
    opik.WithSpanModel("gpt-4"),
    opik.WithSpanProvider("openai"),
)

// Tool execution span
span, _ := trace.Span(ctx, "search-tool",
    opik.WithSpanType(opik.SpanTypeTool),
)

// General processing span (default)
span, _ := trace.Span(ctx, "process-data",
    opik.WithSpanType(opik.SpanTypeGeneral),
)
```

## Span Options

| Option | Description |
|--------|-------------|
| `WithSpanType(type)` | Set span type (LLM, Tool, General) |
| `WithSpanModel(model)` | Set the model name for LLM spans |
| `WithSpanProvider(provider)` | Set the provider name (openai, anthropic) |
| `WithSpanInput(data)` | Set input data |
| `WithSpanOutput(data)` | Set output data |
| `WithSpanMetadata(data)` | Set metadata |
| `WithSpanTags(tags...)` | Add tags |

## Complete Example

```go
func processQuery(ctx context.Context, client *opik.Client, query string) (string, error) {
    // Create trace for the entire request
    trace, _ := client.Trace(ctx, "process-query",
        opik.WithTraceInput(map[string]any{"query": query}),
    )
    defer trace.End(ctx)

    // Span for preprocessing
    preprocessSpan, _ := trace.Span(ctx, "preprocess")
    processedQuery := preprocess(query)
    preprocessSpan.End(ctx, opik.WithSpanOutput(map[string]any{
        "processed": processedQuery,
    }))

    // Span for LLM call
    llmSpan, _ := trace.Span(ctx, "llm-call",
        opik.WithSpanType(opik.SpanTypeLLM),
        opik.WithSpanModel("gpt-4"),
        opik.WithSpanInput(map[string]any{"prompt": processedQuery}),
    )

    response, err := callLLM(processedQuery)
    if err != nil {
        llmSpan.End(ctx, opik.WithSpanMetadata(map[string]any{"error": err.Error()}))
        return "", err
    }

    llmSpan.End(ctx, opik.WithSpanOutput(map[string]any{"response": response}))

    // Update trace with final output
    trace.End(ctx, opik.WithTraceOutput(map[string]any{"result": response}))

    return response, nil
}
```

## Updating Traces and Spans

Update metadata after creation:

```go
// Update trace
trace.Update(ctx,
    opik.WithTraceMetadata(map[string]any{"status": "completed"}),
    opik.WithTraceTags("success"),
)

// Update span
span.Update(ctx,
    opik.WithSpanMetadata(map[string]any{"tokens": 150}),
)
```
