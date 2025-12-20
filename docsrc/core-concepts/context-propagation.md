# Context Propagation

The SDK uses Go's `context.Context` for automatic trace and span propagation. This enables:

- Automatic parent-child span relationships
- Easy access to current trace/span from anywhere
- Clean API without passing trace objects everywhere

## Starting Traces with Context

```go
// Start a trace and get updated context
ctx, trace, _ := opik.StartTrace(ctx, client, "my-trace")

// The trace is now in the context
currentTrace := opik.TraceFromContext(ctx)
```

## Starting Spans with Context

```go
// Start a span - automatically nested under current span/trace
ctx, span, _ := opik.StartSpan(ctx, "my-span")

// Start another span - automatically nested under the previous span
ctx, childSpan, _ := opik.StartSpan(ctx, "child-span")
```

## Retrieving from Context

```go
// Get current trace
trace := opik.TraceFromContext(ctx)
if trace != nil {
    fmt.Printf("Current trace: %s\n", trace.ID())
}

// Get current span
span := opik.SpanFromContext(ctx)
if span != nil {
    fmt.Printf("Current span: %s\n", span.ID())
}
```

## Practical Example

Context propagation makes it easy to add tracing to existing code:

```go
func HandleRequest(ctx context.Context, req *Request) (*Response, error) {
    // Start trace at entry point
    ctx, trace, _ := opik.StartTrace(ctx, client, "handle-request",
        opik.WithTraceInput(req),
    )
    defer trace.End(ctx)

    // Call nested functions - they can access trace via context
    result, err := processRequest(ctx, req)
    if err != nil {
        return nil, err
    }

    trace.End(ctx, opik.WithTraceOutput(result))
    return result, nil
}

func processRequest(ctx context.Context, req *Request) (*Response, error) {
    // Create span - automatically nested under trace
    ctx, span, _ := opik.StartSpan(ctx, "process-request")
    defer span.End(ctx)

    // Call LLM
    return callLLM(ctx, req.Query)
}

func callLLM(ctx context.Context, query string) (*Response, error) {
    // Create LLM span - automatically nested under process-request span
    ctx, span, _ := opik.StartSpan(ctx, "llm-call",
        opik.WithSpanType(opik.SpanTypeLLM),
        opik.WithSpanModel("gpt-4"),
    )
    defer span.End(ctx)

    // Make the actual LLM call
    response, err := llmClient.Complete(ctx, query)

    span.End(ctx, opik.WithSpanOutput(map[string]any{"response": response}))
    return response, err
}
```

## Distributed Tracing

For microservices, propagate trace context across HTTP boundaries:

### Injecting Headers (Client Side)

```go
// Inject trace headers into outgoing request
req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
opik.InjectDistributedTraceHeaders(ctx, req)
```

### Extracting Headers (Server Side)

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Extract trace headers
    headers := opik.ExtractDistributedTraceHeaders(r)

    // Continue the distributed trace
    ctx, span, _ := client.ContinueTrace(r.Context(), headers, "handle-request")
    defer span.End(ctx)

    // Process request...
}
```

### Propagating HTTP Client

Use the built-in propagating client:

```go
// Create client that auto-injects trace headers
httpClient := opik.PropagatingHTTPClient()

// All requests will include trace headers
resp, _ := httpClient.Do(req.WithContext(ctx))
```

## Header Format

The SDK uses these headers for distributed tracing:

| Header | Description |
|--------|-------------|
| `X-Opik-Trace-ID` | The trace ID |
| `X-Opik-Parent-Span-ID` | The parent span ID |
