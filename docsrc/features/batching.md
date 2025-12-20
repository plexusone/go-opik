# Batching

Batch API calls for improved performance and efficiency.

## Batching Client

```go
// Create a batching client
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

## Local Recording (Testing)

For testing without sending data to the server:

```go
// Record traces locally
client := opik.RecordTracesLocally("my-project")

// Use normally
trace, _ := client.Trace(ctx, "test-trace")
span, _ := trace.Span(ctx, "test-span")
span.End(ctx)
trace.End(ctx)

// Access recorded data
recording := client.Recording()

traces := recording.Traces()
spans := recording.Spans()

// Inspect recorded data
for _, t := range traces {
    fmt.Printf("Trace: %s - %s\n", t.ID, t.Name)
}
```

## Attachments

Create and manage file attachments:

### From File

```go
attachment, err := opik.NewAttachmentFromFile("/path/to/image.png")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Name: %s\n", attachment.Name)
fmt.Printf("Type: %s\n", attachment.ContentType)
fmt.Printf("Size: %d bytes\n", len(attachment.Data))
```

### From Bytes

```go
jsonData := []byte(`{"key": "value"}`)
attachment := opik.NewAttachmentFromBytes("data.json", jsonData, "application/json")
```

### Text Attachment

```go
attachment := opik.NewTextAttachment("notes.txt", "Some text content")
```

### Using Attachments

```go
// Get data URL for embedding in HTML/markdown
dataURL := attachment.ToDataURL()
// Result: "data:image/png;base64,..."

// Include in span metadata
span.End(ctx, opik.WithSpanMetadata(map[string]any{
    "attachment": dataURL,
}))
```

## Batch Operations

### Async Feedback

```go
// Add feedback without waiting
client.AddFeedbackAsync("trace", traceID, "accuracy", 0.95, "High accuracy")
client.AddFeedbackAsync("span", spanID, "quality", 0.87, "Good quality")

// Flush when ready
client.Flush(5 * time.Second)
```

### Graceful Shutdown

```go
func main() {
    client, _ := opik.NewBatchingClient()

    // Ensure clean shutdown
    defer func() {
        if err := client.Close(10 * time.Second); err != nil {
            log.Printf("Error closing client: %v", err)
        }
    }()

    // Your application code...
}
```

## Configuration Options

### Batch Size

```go
client, _ := opik.NewBatchingClient(
    opik.WithBatchSize(100), // Send when 100 items accumulated
)
```

### Flush Interval

```go
client, _ := opik.NewBatchingClient(
    opik.WithFlushInterval(5 * time.Second), // Auto-flush every 5 seconds
)
```

## Best Practices

1. **Use batching for high-volume**: Reduce API calls in production
2. **Always close clients**: Ensure pending data is flushed
3. **Set appropriate timeouts**: Allow enough time for flushing
4. **Use local recording for tests**: Avoid external dependencies in tests
5. **Monitor batch sizes**: Tune for your workload
