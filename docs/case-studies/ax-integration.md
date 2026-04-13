# Agent Experience (AX) Integration Case Study

## Executive Summary

This case study documents the integration of Agent Experience (AX) principles into opik-go, an LLM observability SDK. The integration enables AI agents to reliably trace their own behavior with machine-readable error handling, explicit retry policies, and pre-flight validation.

**Key Results:**

- 19 domain-specific error codes defined
- 201 operations mapped with retry policies
- 67 operations have required field definitions
- 7 capability types including streaming and evaluation

## Background

### The Observability Challenge

AI agents need to observe their own behavior to learn and improve. This creates a unique challenge: the observability system itself must be agent-friendly.

| Aspect | Human Developer | AI Agent |
|--------|-----------------|----------|
| Error handling | Read logs, debug | Programmatic recovery |
| Missing traces | Manually investigate | Auto-create resources |
| Evaluation | Review dashboards | Run programmatically |
| Retries | Intuitive judgment | Explicit policies needed |

### The Project

opik-go is a Go SDK for Comet ML's Opik observability platform:

- **201 API endpoints** covering traces, spans, datasets, experiments, and evaluations
- **14,820 line OpenAPI specification**
- **LLM evaluation framework** with heuristic and model-based scorers
- **Streaming support** for large result sets
- Used by AI agents to observe and improve their own behavior

## The Problem

When tracing fails, agents face ambiguous errors:

```json
{
  "status": 404,
  "message": "Not found"
}
```

Questions the agent cannot answer:

1. **What wasn't found?** — Trace? Span? Dataset? Project?
2. **Should it retry?** — Will it help or create duplicates?
3. **How to recover?** — Create the resource? Use a different one?
4. **Is it permanent?** — Or a transient failure?

### Before AX Integration

```go
trace, err := client.GetTrace(ctx, traceID)
if err != nil {
    if apiErr, ok := err.(*APIError); ok {
        if apiErr.StatusCode == 404 {
            // Which resource is missing?
            // The trace? The project? The workspace?
            // No way to determine programmatically
        }
    }
    return nil, err
}
```

## Solution: AX Integration

### Error Code Design

19 domain-specific error codes organized by category:

| Category | Error Codes | HTTP Status |
|----------|-------------|-------------|
| **not_found** | TRACE_NOT_FOUND, SPAN_NOT_FOUND, DATASET_NOT_FOUND, EXPERIMENT_NOT_FOUND, PROMPT_NOT_FOUND, PROJECT_NOT_FOUND, FEEDBACK_NOT_FOUND, ATTACHMENT_NOT_FOUND, WORKSPACE_NOT_FOUND, EVALUATOR_NOT_FOUND, ALERT_NOT_FOUND, QUEUE_NOT_FOUND, DASHBOARD_NOT_FOUND | 404 |
| **auth** | UNAUTHORIZED, FORBIDDEN | 401, 403 |
| **validation** | INVALID_INPUT | 400 |
| **conflict** | CONFLICT | 409 |
| **rate_limit** | RATE_LIMITED | 429 |
| **server** | INTERNAL_ERROR | 500 |

### Retry Policy Mapping

All 201 operations mapped with retry safety:

```go
var RetryPolicy = map[string]bool{
    // Safe to retry (GET operations)
    "getTraceById":      true,
    "findTraces":        true,
    "getSpanById":       true,
    "findDatasets":      true,
    "streamExperiments": true,

    // Not safe (mutations)
    "createTrace":       false,
    "createSpan":        false,
    "createExperiment":  false,
    "deleteTraceById":   false,
    "evaluateTraces":    false,
}
```

**Distribution:**

| Category | Count | Retryable |
|----------|-------|-----------|
| GET (read) | 78 | Yes |
| POST (create) | 62 | No |
| PUT/PATCH (update) | 31 | No |
| DELETE | 30 | No |

### Required Fields Extraction

67 operations have required field definitions:

```go
var RequiredFields = map[string][]string{
    "createTrace":       {"name"},
    "createSpan":        {"trace_id", "name"},
    "createExperiment":  {"dataset_name", "name"},
    "createDataset":     {"name"},
    "createPrompt":      {"name", "template"},
    "evaluateTraces":    {"trace_ids", "evaluator_ids"},
    "evaluateSpans":     {"span_ids", "evaluator_ids"},
    // ... 60 more
}
```

### Capability Mapping

7 capability types for observability operations:

```go
const (
    CapRead      Capability = "read"      // Data retrieval
    CapWrite     Capability = "write"     // Data creation/modification
    CapDelete    Capability = "delete"    // Data removal
    CapAdmin     Capability = "admin"     // Administrative operations
    CapStream    Capability = "stream"    // Streaming responses
    CapEvaluate  Capability = "evaluate"  // LLM evaluation
    CapAnalytics Capability = "analytics" // Metrics and BI
)
```

## Results

### Error Handling Improvement

**Before:**

```go
trace, err := client.GetTrace(ctx, traceID)
if err != nil {
    // Generic error handling
    log.Printf("Error: %v", err)
    return nil, err
}
```

**After:**

```go
trace, err := client.GetTrace(ctx, traceID)
if err != nil {
    code, ok := opik.GetAXErrorCode(err)
    if !ok {
        return nil, err
    }

    switch code {
    case ax.ErrTraceNotFound:
        // Create the trace first
        return client.CreateTrace(ctx, &Trace{ID: traceID, Name: "auto-created"})

    case ax.ErrProjectNotFound:
        // Create the project, then retry
        client.CreateProject(ctx, projectName)
        return client.GetTrace(ctx, traceID)

    case ax.ErrUnauthorized:
        // Re-authenticate
        return nil, ErrNeedsAuth

    case ax.ErrRateLimited:
        // Back off and retry
        time.Sleep(time.Second)
        return client.GetTrace(ctx, traceID)
    }

    return nil, err
}
```

### Self-Healing Tracing Pattern

```go
func (a *Agent) recordAction(ctx context.Context, action Action) error {
    trace := &Trace{
        Name:  action.Name,
        Input: action.Input,
    }

    err := a.client.CreateTrace(ctx, trace)
    if err == nil {
        return nil
    }

    // Self-healing based on AX metadata
    info := opik.GetAXErrorInfo(err)
    if info == nil {
        return err
    }

    switch info.Category {
    case "not_found":
        // Create missing resource
        if code, _ := opik.GetAXErrorCode(err); code == ax.ErrProjectNotFound {
            a.client.CreateProject(ctx, a.projectName)
            return a.client.CreateTrace(ctx, trace)
        }

    case "rate_limit":
        // Exponential backoff (retryable)
        if info.Retryable {
            time.Sleep(time.Second * 2)
            return a.recordAction(ctx, action)
        }

    case "conflict":
        // Resource exists, fetch it instead
        existing, _ := a.client.GetTrace(ctx, trace.ID)
        return a.updateTrace(ctx, existing, action)
    }

    return err
}
```

### Pre-flight Validation

```go
func validateRequest(operationID string, req interface{}) error {
    // Extract present fields via reflection or manual mapping
    present := extractPresentFields(req)

    if msg := ax.ValidateFields(operationID, present); msg != "" {
        return fmt.Errorf("validation failed: %s", msg)
    }
    return nil
}

// Usage
func (c *Client) CreateExperiment(ctx context.Context, req *ExperimentRequest) error {
    if err := validateRequest("createExperiment", req); err != nil {
        return err // Fail fast without API call
    }
    return c.api.CreateExperiment(ctx, req)
}
```

### Capability-Based Discovery

```go
// Find operations that support streaming
streamOps := ax.GetOperationsByCapability(ax.CapStream)
// ["streamDatasetItems", "streamExperimentItems", "streamExperiments"]

// Check if evaluation is available
if ax.HasCapability("evaluateTraces", ax.CapEvaluate) {
    // Run automatic quality evaluation
    scores, _ := client.EvaluateTraces(ctx, traceIDs, evaluatorIDs)
}

// Find analytics operations for dashboards
analyticsOps := ax.GetOperationsByCapability(ax.CapAnalytics)
// ["getProjectMetrics", "getProjectStats", "costsSummary", ...]
```

## Metrics

### Code Changes

| Component | Files | Lines |
|-----------|-------|-------|
| ax package | 6 new files | ~950 lines |
| errors.go | 1 modified | ~80 lines |
| **Total** | **7 files** | **~1,030 lines** |

### Coverage

| Metadata Type | Count | Coverage |
|---------------|-------|----------|
| Error codes | 19 | Domain-complete |
| Retry policies | 201 | 100% of operations |
| Required fields | 67 | 33% (mutation operations) |
| Capabilities | ~100 | Key operations |

### Test Results

```bash
$ go test -v ./ax/...
=== RUN   TestIsErrorCode
--- PASS: TestIsErrorCode (0.00s)
=== RUN   TestContainsErrorCode
--- PASS: TestContainsErrorCode (0.00s)
=== RUN   TestGetErrorInfo
--- PASS: TestGetErrorInfo (0.00s)
=== RUN   TestErrorCategoryHelpers
--- PASS: TestErrorCategoryHelpers (0.00s)
=== RUN   TestIsRetryable
--- PASS: TestIsRetryable (0.00s)
=== RUN   TestRetryableCount
--- PASS: TestRetryableCount (0.00s)
=== RUN   TestGetRequiredFields
--- PASS: TestGetRequiredFields (0.00s)
=== RUN   TestCapabilities
--- PASS: TestCapabilities (0.00s)
...
PASS
ok      github.com/plexusone/opik-go/ax
```

## Key Learnings

### 1. Observability Needs Precision

Generic "not found" errors are insufficient when agents trace themselves. Each resource type needs its own error code:

- `TRACE_NOT_FOUND` — The trace doesn't exist
- `PROJECT_NOT_FOUND` — The project doesn't exist (create it first)
- `SPAN_NOT_FOUND` — The span doesn't exist (but trace might)

### 2. Self-Healing Patterns are Essential

Agents that observe themselves must handle their own tracing failures:

```go
// Bad: Agent stops observing itself on first error
// Good: Agent recovers and continues observing

if code == ax.ErrProjectNotFound {
    createProject()
    retryTrace()
}
```

### 3. Domain-Specific Capabilities Matter

Observability has unique capabilities that generic CRUD doesn't capture:

- **CapEvaluate** — LLM quality evaluation
- **CapStream** — Large result streaming
- **CapAnalytics** — Metrics and BI operations

### 4. Retry Policies Prevent Data Corruption

Observability data is append-only. Retrying creates risks:

```
createTrace (not retryable) — Would create duplicate traces
evaluateTraces (not retryable) — Would run evaluation twice
getTraceById (retryable) — Safe to retry reads
```

### 5. HTTP Status is Not Enough

Status 404 could mean:

- Trace not found
- Span not found
- Project not found
- Workspace not found
- Any of 13 other "not found" conditions

AX error codes disambiguate completely.

## Comparison with elevenlabs-go

| Aspect | elevenlabs-go | opik-go |
|--------|---------------|---------|
| Domain | Voice generation | LLM observability |
| Endpoints | 204 | 201 |
| Error codes | 9 (API discovered) | 19 (domain defined) |
| Retry policies | 236 | 201 |
| Required fields | 72 | 67 |
| Special capabilities | - | Stream, Evaluate, Analytics |
| Self-healing | Media errors | Tracing errors |

Both SDKs benefit from AX, but the specific error codes and capabilities differ by domain.

## Future Work

### API Discovery

Run ax-spec discovery against the real Opik API to find additional error codes not documented in the spec.

### Idempotency Support

Add `x-ax-idempotent` extension support for safe retry of create operations with idempotency keys.

### Batch Operation Handling

Define patterns for partial success in batch operations (some items succeed, some fail).

### Evaluation Metadata

Expose evaluator capabilities (what metrics they produce, what inputs they need).

## Conclusion

The AX integration transforms opik-go from a basic observability SDK to an agent-friendly one:

| Aspect | Before | After |
|--------|--------|-------|
| Error handling | HTTP status parsing | 19 typed error codes |
| Retry decisions | Hardcoded or missing | 201 operations mapped |
| Validation | Runtime API errors | Pre-flight validation |
| Capabilities | Unknown | Domain-specific discovery |
| Agent behavior | Fragile tracing | Self-healing observability |

For AI agents that need to observe themselves, reliable observability is foundational. AX makes that reliability achievable.

## References

- [DIRECT Principles](https://github.com/grokify/direct-principles)
- [AX Spec](https://github.com/grokify/ax-spec)
- [opik-go ax package](https://github.com/plexusone/opik-go/tree/main/ax)
- [elevenlabs-go AX Case Study](https://github.com/plexusone/elevenlabs-go/tree/main/docs/case-studies)
