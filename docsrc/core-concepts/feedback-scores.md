# Feedback Scores

Feedback scores allow you to attach quality metrics to traces and spans. Use them to:

- Record user feedback (thumbs up/down, ratings)
- Store evaluation results from automated metrics
- Track quality metrics over time

## Adding Feedback to Traces

```go
// Add a numeric score
trace.AddFeedbackScore(ctx, "accuracy", 0.95, "High accuracy response")

// Add multiple scores
trace.AddFeedbackScore(ctx, "relevance", 0.87, "Mostly relevant")
trace.AddFeedbackScore(ctx, "helpfulness", 0.92, "Very helpful")
```

## Adding Feedback to Spans

```go
// Add feedback to specific spans
span.AddFeedbackScore(ctx, "latency_score", 0.75, "Response time acceptable")
span.AddFeedbackScore(ctx, "quality", 0.90, "Good quality output")
```

## Score Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Name of the score (e.g., "accuracy", "relevance") |
| `value` | float64 | Score value (typically 0.0 to 1.0) |
| `reason` | string | Optional explanation for the score |

## Use Cases

### User Feedback

```go
func handleFeedback(ctx context.Context, traceID string, rating int) {
    // Convert 1-5 rating to 0-1 score
    score := float64(rating-1) / 4.0

    trace := getTrace(traceID)
    trace.AddFeedbackScore(ctx, "user_rating", score,
        fmt.Sprintf("User rated %d/5", rating))
}
```

### Automated Evaluation

```go
func evaluateResponse(ctx context.Context, trace *opik.Trace, response string) {
    // Run evaluation metrics
    relevanceScore := evaluateRelevance(response)
    factualScore := evaluateFactuality(response)

    // Record as feedback scores
    trace.AddFeedbackScore(ctx, "relevance", relevanceScore, "Automated relevance check")
    trace.AddFeedbackScore(ctx, "factuality", factualScore, "Automated fact check")
}
```

### A/B Testing

```go
func recordABResult(ctx context.Context, trace *opik.Trace, variant string, converted bool) {
    score := 0.0
    if converted {
        score = 1.0
    }

    trace.AddFeedbackScore(ctx, "conversion", score,
        fmt.Sprintf("Variant %s, converted: %v", variant, converted))
}
```

## Viewing Feedback Scores

Feedback scores are visible in the Opik UI:

- On the trace detail page
- In trace list summaries
- In experiment comparisons
- In analytics dashboards

## Best Practices

1. **Use consistent names**: Standardize score names across your application
2. **Normalize values**: Use 0.0-1.0 range for easy comparison
3. **Include reasons**: Add explanations for debugging and analysis
4. **Score at appropriate level**: Use trace-level for overall quality, span-level for specific operations
