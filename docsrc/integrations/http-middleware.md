# HTTP Middleware

Add tracing to HTTP handlers and clients.

```go
import "github.com/grokify/go-comet-ml-opik/middleware"
```

## Server Middleware

### Wrap HTTP Handlers

Automatically create traces for incoming requests:

```go
opikClient, _ := opik.NewClient()

// Wrap a handler
handler := middleware.TracingMiddleware(opikClient, "api-request")(yourHandler)

// Use with http.ServeMux
mux := http.NewServeMux()
mux.Handle("/api/", middleware.TracingMiddleware(opikClient, "api")(apiHandler))
```

### What Gets Traced

Each request creates a trace with:

| Field | Description |
|-------|-------------|
| Name | Configured name (e.g., "api-request") |
| Input | Method, URL, headers |
| Output | Status code, response size |
| Metadata | Duration, client IP |

### Example Handler

```go
func main() {
    opikClient, _ := opik.NewClient()

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Access trace from context
        trace := opik.TraceFromContext(r.Context())

        // Create spans for sub-operations
        ctx, span, _ := opik.StartSpan(r.Context(), "process-request")

        // Do work...

        span.End(ctx)

        w.Write([]byte("OK"))
    })

    // Wrap with tracing
    tracedHandler := middleware.TracingMiddleware(opikClient, "my-api")(handler)

    http.ListenAndServe(":8080", tracedHandler)
}
```

## Client Middleware

### Tracing HTTP Client

Create a client that traces all outgoing requests:

```go
httpClient := middleware.TracingHTTPClient("external-api")

resp, _ := httpClient.Get("https://api.example.com/data")
```

### Tracing Round Tripper

Wrap an existing transport:

```go
transport := middleware.NewTracingRoundTripper(http.DefaultTransport, "api-call")

httpClient := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,
}
```

### What Gets Traced

Each outgoing request creates a span with:

| Field | Description |
|-------|-------------|
| Name | Configured name |
| Type | `general` |
| Input | Method, URL |
| Output | Status code |
| Metadata | Duration |

## Combining Server and Client

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Trace from middleware is in context
    trace := opik.TraceFromContext(ctx)

    // Create span for external call
    ctx, span, _ := opik.StartSpan(ctx, "fetch-data")

    // Use tracing client - span becomes parent
    client := middleware.TracingHTTPClient("external-api")
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
    resp, err := client.Do(req)

    span.End(ctx)

    // Process response...
}
```

## Framework Integration

### Chi

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Use(func(next http.Handler) http.Handler {
    return middleware.TracingMiddleware(opikClient, "api")(next)
})
```

### Gin

```go
import "github.com/gin-gonic/gin"

r := gin.Default()
r.Use(func(c *gin.Context) {
    handler := middleware.TracingMiddleware(opikClient, "api")(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Request = r.WithContext(r.Context())
            c.Next()
        }),
    )
    handler.ServeHTTP(c.Writer, c.Request)
})
```

### Echo

```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.Use(echo.WrapMiddleware(func(next http.Handler) http.Handler {
    return middleware.TracingMiddleware(opikClient, "api")(next)
}))
```

## Best Practices

1. **Name traces descriptively**: Use service/endpoint names
2. **Add to all entry points**: Trace all HTTP handlers
3. **Propagate context**: Pass request context to downstream calls
4. **Use with distributed tracing**: Combine with header propagation
