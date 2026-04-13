# AX Retry Policies

The ax package maps all 201 Opik API operations with retry safety information.

## Checking Retry Safety

```go
// Check if operation is safe to retry
if ax.IsRetryable("getTraceById") {
    // Safe to retry with exponential backoff
}

if !ax.IsRetryable("createTrace") {
    // Not safe - would create duplicate traces
}
```

## Retry Policy Distribution

| Category | Count | Retryable | Reason |
|----------|-------|-----------|--------|
| GET (read) | 78 | Yes | Idempotent reads |
| POST (create) | 62 | No | Would create duplicates |
| PUT/PATCH (update) | 31 | No | May cause conflicts |
| DELETE | 30 | No | Data loss risk |

## Retryable Operations

All GET operations are safe to retry:

```go
// Safe to retry
"getTraceById"
"findTraces"
"getSpanById"
"getSpansByProject"
"findDatasets"
"getDatasetById"
"findExperiments"
"getExperimentById"
"getPrompts"
"getPromptById"
// ... and 68 more
```

## Non-Retryable Operations

Mutation operations are not safe to retry:

```go
// Not safe - creates resources
"createTrace"
"createSpan"
"createDataset"
"createExperiment"
"createPrompt"

// Not safe - updates resources
"updateTrace"
"updateSpan"
"updateDataset"

// Not safe - deletes resources
"deleteTraceById"
"deleteSpanById"
"deleteDataset"

// Not safe - evaluation (consumes resources)
"evaluateTraces"
"evaluateSpans"
```

## Helper Functions

```go
// Get all retryable operation IDs
retryable := ax.GetRetryableOperations()

// Get all non-retryable operation IDs
nonRetryable := ax.GetNonRetryableOperations()

// Get counts
safe, unsafe := ax.RetryableCount()
```

## Using with Retry Libraries

```go
import "github.com/cenkalti/backoff/v4"

func withRetry(operationID string, fn func() error) error {
    if !ax.IsRetryable(operationID) {
        // Not safe to retry, call once
        return fn()
    }

    // Safe to retry with backoff
    return backoff.Retry(fn, backoff.NewExponentialBackOff())
}
```

## Complete Example

```go
func recordTrace(ctx context.Context, trace *Trace) error {
    operationID := "createTrace"

    // Check retry safety
    if ax.IsRetryable(operationID) {
        // This shouldn't happen for createTrace, but check anyway
        return retryWithBackoff(func() error {
            return client.CreateTrace(ctx, trace)
        })
    }

    // Not retryable - call once and handle errors
    err := client.CreateTrace(ctx, trace)
    if err != nil {
        // Check if the error itself is retryable
        if opik.IsAXError(err, ax.ErrRateLimited) {
            // Rate limit is retryable even for non-retryable operations
            time.Sleep(time.Second)
            return client.CreateTrace(ctx, trace)
        }
        return err
    }
    return nil
}
```

## Error-Based Retry Decisions

Some errors are retryable even for non-retryable operations:

```go
code, _ := opik.GetAXErrorCode(err)
info := ax.GetErrorInfo(code)

if info != nil && info.Retryable {
    // Error itself is retryable (rate limit, server error)
    // Safe to retry even for mutation operations
    time.Sleep(time.Second)
    return retry()
}
```
