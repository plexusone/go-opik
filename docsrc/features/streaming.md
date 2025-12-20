# Streaming

Track streaming LLM responses with automatic chunk accumulation.

## Starting a Streaming Span

```go
// Start a streaming span
ctx, streamSpan, _ := opik.StartStreamingSpan(ctx, "stream-response",
    opik.WithSpanType(opik.SpanTypeLLM),
    opik.WithSpanModel("gpt-4"),
)
```

## Adding Chunks

As chunks arrive from your LLM, add them to the span:

```go
for chunk := range streamChannel {
    // Add chunk content
    streamSpan.AddChunk(chunk.Content)
}
```

### With Chunk Options

```go
// Add chunk with token count
streamSpan.AddChunk(chunk.Content,
    opik.WithChunkTokenCount(chunk.TokenCount),
)

// Mark final chunk with finish reason
streamSpan.AddChunk(lastChunk.Content,
    opik.WithChunkTokenCount(lastChunk.TokenCount),
    opik.WithChunkFinishReason("stop"),
)
```

## Ending the Streaming Span

```go
// End span - accumulated content is automatically captured
streamSpan.End(ctx)
```

## Complete Example

```go
func streamCompletion(ctx context.Context, prompt string) (string, error) {
    // Start streaming span
    ctx, streamSpan, err := opik.StartStreamingSpan(ctx, "openai-stream",
        opik.WithSpanType(opik.SpanTypeLLM),
        opik.WithSpanModel("gpt-4"),
        opik.WithSpanProvider("openai"),
        opik.WithSpanInput(map[string]any{"prompt": prompt}),
    )
    if err != nil {
        return "", err
    }

    // Create streaming request to OpenAI
    stream, err := openaiClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
        Model:    "gpt-4",
        Messages: []openai.ChatCompletionMessage{{Role: "user", Content: prompt}},
        Stream:   true,
    })
    if err != nil {
        streamSpan.End(ctx)
        return "", err
    }
    defer stream.Close()

    // Collect chunks
    var fullResponse strings.Builder

    for {
        response, err := stream.Recv()
        if err == io.EOF {
            // Mark final chunk
            streamSpan.AddChunk("", opik.WithChunkFinishReason("stop"))
            break
        }
        if err != nil {
            streamSpan.End(ctx)
            return "", err
        }

        content := response.Choices[0].Delta.Content
        fullResponse.WriteString(content)

        // Add chunk to span
        streamSpan.AddChunk(content)

        // Optionally output to user in real-time
        fmt.Print(content)
    }

    // End span with accumulated output
    streamSpan.End(ctx)

    return fullResponse.String(), nil
}
```

## Chunk Options

| Option | Description |
|--------|-------------|
| `WithChunkTokenCount(n)` | Number of tokens in this chunk |
| `WithChunkFinishReason(reason)` | Finish reason (e.g., "stop", "length") |

## Stream Accumulator

The streaming span automatically accumulates:

- **Content**: All chunk content concatenated
- **Token count**: Sum of all chunk token counts
- **Duration**: Time from first chunk to last
- **Finish reason**: From the final chunk

## Accessing Accumulated Data

```go
// Get the accumulated content before ending
accumulated := streamSpan.AccumulatedContent()

// Get chunk count
count := streamSpan.ChunkCount()

// Get total tokens
tokens := streamSpan.TotalTokens()
```

## Best Practices

1. **Start span before streaming**: Create the span before initiating the stream
2. **Add chunks as they arrive**: Don't buffer - add immediately for accurate timing
3. **Include token counts**: If available, include for cost tracking
4. **Mark finish reason**: Always mark the final chunk with a finish reason
5. **Handle errors**: End the span even if streaming fails
