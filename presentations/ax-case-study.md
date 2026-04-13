---
marp: true
theme: default
paginate: true
backgroundColor: #fff
---

# Agent Experience (AX) Case Study

## Integrating AX Spec into opik-go

Building Agent-Friendly LLM Observability SDKs

---

# The Challenge

AI agents need to observe their own behavior:

- **Trace LLM calls** — Track inputs, outputs, latency
- **Record spans** — Measure individual operations
- **Run evaluations** — Score quality automatically
- **Handle errors** — Recover from failures gracefully

Observability SDKs must be agent-friendly.

---

# The Problem

When an agent's tracing fails:

```json
{
  "status": 404,
  "message": "Not found"
}
```

**What can the agent do?**

- Is it a trace? Span? Dataset? Experiment?
- Should it retry? Create the resource first?
- How does it recover automatically?

---

# The Vision: Agent Experience (AX)

Design interfaces that enable agents to:

> **understand → call → recover → iterate**

For observability: Agents must trace reliably to learn from their actions.

---

# DIRECT Principles for Observability

| Principle | Observability Application |
|-----------|---------------------------|
| **D**eterministic | Same trace data → same recorded trace |
| **I**ntrospectable | Discover available metrics, evaluators |
| **R**ecoverable | Know if trace failed, why, how to fix |
| **E**xplicit | Required fields for spans, traces clear |
| **C**onsistent | Uniform patterns across all resources |
| **T**estable | Mock tracing for unit tests |

---

# The Project: opik-go

A Go SDK for Comet ML's Opik observability platform:

- **201 endpoints** (traces, spans, datasets, experiments)
- **14,820 line OpenAPI spec**
- **LLM evaluation** framework built-in
- Used by AI agents to observe themselves

---

# Why Observability Needs AX

Observability is critical for agent reliability:

```
Agent executes task
    ↓
Records trace → Error: TRACE_NOT_FOUND
    ↓
Without AX: Agent stalls, loses observability
With AX: Agent creates project first, retries
```

Self-observing agents need self-healing tracing.

---

# Implementation Approach

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  OpenAPI Spec   │────▶│   ax-spec CLI   │────▶│  Generated Go   │
│                 │     │                 │     │     Code        │
│ opik-openapi.yaml│    │ enrich + gen    │     │ ax/errors.go    │
│  (14,820 lines) │     │                 │     │ ax/retry.go     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

---

# Step 1: Define Error Codes

19 domain-specific error codes for observability:

| Category | Codes |
|----------|-------|
| **not_found** | TRACE_NOT_FOUND, SPAN_NOT_FOUND, DATASET_NOT_FOUND, EXPERIMENT_NOT_FOUND, PROMPT_NOT_FOUND, PROJECT_NOT_FOUND, ... |
| **auth** | UNAUTHORIZED, FORBIDDEN |
| **validation** | INVALID_INPUT |
| **conflict** | CONFLICT |
| **server** | RATE_LIMITED, INTERNAL_ERROR |

---

# Step 2: Map Retry Policies

201 operations mapped:

| Category | Count | Retryable |
|----------|-------|-----------|
| GET (read) | 78 | Yes |
| POST (create) | 62 | No |
| PUT/PATCH (update) | 31 | No |
| DELETE | 30 | No |

**Key insight:** Observability writes are not retryable.

---

# Step 3: Extract Required Fields

67 operations have required field definitions:

```go
var RequiredFields = map[string][]string{
    "createTrace":       {"name"},
    "createSpan":        {"trace_id", "name"},
    "createExperiment":  {"dataset_name", "name"},
    "createDataset":     {"name"},
    "evaluateTraces":    {"trace_ids", "evaluator_ids"},
}
```

---

# Step 4: Define Capabilities

Domain-specific capabilities for observability:

```go
const (
    CapRead      = "read"
    CapWrite     = "write"
    CapDelete    = "delete"
    CapAdmin     = "admin"
    CapStream    = "stream"     // Streaming responses
    CapEvaluate  = "evaluate"   // LLM evaluation
    CapAnalytics = "analytics"  // Metrics/BI
)
```

---

# Before: HTTP Status Handling

```go
trace, err := client.GetTrace(ctx, traceID)
if err != nil {
    if apiErr, ok := err.(*APIError); ok {
        if apiErr.StatusCode == 404 {
            // Is it the trace? The project?
            // No way to know...
        }
    }
}
```

---

# After: AX Error Handling

```go
trace, err := client.GetTrace(ctx, traceID)
if err != nil {
    if opik.IsAXError(err, ax.ErrTraceNotFound) {
        // Specific: the trace doesn't exist
        return createTraceFirst(ctx, traceID)
    }
    if opik.IsAXError(err, ax.ErrProjectNotFound) {
        // Specific: the project doesn't exist
        return createProjectFirst(ctx, projectName)
    }
}
```

---

# Error Recovery Patterns

```go
code, ok := opik.GetAXErrorCode(err)
if !ok {
    return err
}

info := ax.GetErrorInfo(code)
switch info.Category {
case "not_found":
    // Create the missing resource
case "auth":
    // Re-authenticate
case "rate_limit":
    // Back off and retry (info.Retryable = true)
case "conflict":
    // Resource already exists, fetch it
}
```

---

# Retry Policy for Tracing

```go
func recordSpan(ctx context.Context, span *Span) error {
    // Check if safe to retry
    if !ax.IsRetryable("createSpan") {
        // Not retryable - would create duplicates
        return client.CreateSpan(ctx, span)
    }

    // Safe operations can use retry logic
    return retry.Do(func() error {
        return client.CreateSpan(ctx, span)
    })
}
```

---

# Pre-flight Validation

```go
func validateExperimentRequest(req *ExperimentRequest) error {
    present := map[string]bool{
        "name":         req.Name != "",
        "dataset_name": req.DatasetName != "",
    }

    if msg := ax.ValidateFields("createExperiment", present); msg != "" {
        return fmt.Errorf("invalid request: %s", msg)
    }
    return nil
}
```

---

# Capability-Based Discovery

```go
// Find all evaluation operations
evalOps := ax.GetOperationsByCapability(ax.CapEvaluate)
// ["evaluateSpans", "evaluateThreads", "evaluateTraces"]

// Check if operation supports streaming
if ax.SupportsStreaming("streamExperimentItems") {
    // Use streaming API for efficiency
}

// Check admin requirements
if ax.RequiresAdmin("upsertWorkspaceConfiguration") {
    // Verify admin permissions first
}
```

---

# Results Summary

| Metric | Before | After |
|--------|--------|-------|
| Error handling | HTTP status only | 19 typed error codes |
| Retry decisions | Hardcoded | 201 operations mapped |
| Required fields | Runtime errors | Pre-validation for 67 ops |
| Error categories | None | 6 categories with metadata |
| Agent behavior | Guesswork | Deterministic recovery |

---

# Observability-Specific Benefits

| Feature | Agent Benefit |
|---------|---------------|
| Trace error codes | Know exactly what failed |
| Span validation | Pre-check before recording |
| Evaluation detection | Discover available evaluators |
| Streaming support | Efficient large result handling |
| Analytics capability | Find metrics operations |

---

# Code Impact

**New files:**

- `ax/doc.go` - Package documentation
- `ax/errors.go` - 19 error constants
- `ax/retry.go` - 201 retry policies
- `ax/validation.go` - 67 required field maps
- `ax/capabilities.go` - 7 capability types
- `ax/ax_test.go` - Comprehensive tests

**Modified:** `errors.go` - AX integration methods

---

# Example: Self-Healing Tracing

```go
func (a *Agent) recordAction(ctx context.Context, action Action) error {
    trace := &Trace{Name: action.Name, Input: action.Input}

    err := a.client.CreateTrace(ctx, trace)
    if err == nil {
        return nil
    }

    // Self-healing based on AX error code
    switch code, _ := opik.GetAXErrorCode(err); code {
    case ax.ErrProjectNotFound:
        a.client.CreateProject(ctx, a.projectName)
        return a.client.CreateTrace(ctx, trace)
    case ax.ErrRateLimited:
        time.Sleep(time.Second)
        return a.recordAction(ctx, action)
    default:
        return err
    }
}
```

---

# Evaluation Integration

```go
// Discover if evaluation is possible
if ax.IsEvaluation("evaluateTraces") {
    // Run automatic evaluation
    result, err := client.EvaluateTraces(ctx, &EvaluateRequest{
        TraceIDs:     traceIDs,
        EvaluatorIDs: evaluatorIDs,
    })

    if err != nil {
        if info := opik.GetAXErrorInfo(err); info != nil {
            log.Printf("Evaluation failed: %s (retryable=%v)",
                info.Description, info.Retryable)
        }
    }
}
```

---

# Key Learnings

1. **Observability needs precision** — "Not found" isn't enough
2. **Tracing must be reliable** — Self-healing patterns essential
3. **201 operations is a lot** — Code generation scales
4. **Categories enable strategies** — not_found vs auth vs rate_limit
5. **Capabilities aid discovery** — Streaming, evaluation, analytics

---

# Comparison: elevenlabs-go vs opik-go

| Aspect | elevenlabs-go | opik-go |
|--------|---------------|---------|
| Endpoints | 204 | 201 |
| Error codes | 9 (discovered) | 19 (defined) |
| Retry policies | 236 | 201 |
| Required fields | 72 | 67 |
| Domain | Voice generation | LLM observability |
| Special caps | - | Stream, Evaluate, Analytics |

---

# The AX Workflow

```bash
# 1. Lint spec for AX compliance
ax-spec lint opik-openapi.yaml

# 2. Enrich with x-ax-* extensions
ax-spec enrich opik-openapi.yaml --output opik-openapi-ax.yaml

# 3. Generate SDK code
ax-spec gen opik-openapi-ax.yaml --output ax/

# 4. Integrate with existing SDK
# Add IsAXError(), GetAXErrorCode() helpers
```

---

# What's Next

- **API discovery** — Probe real API for actual error codes
- **Idempotency keys** — Safe retry for create operations
- **Batch operation handling** — Partial success patterns
- **Evaluation metadata** — Evaluator capabilities

---

# Resources

- **DIRECT Principles:** github.com/grokify/direct-principles
- **AX Spec:** github.com/grokify/ax-spec
- **opik-go:** github.com/plexusone/opik-go
- **elevenlabs-go:** github.com/plexusone/elevenlabs-go

---

# Summary

Agent Experience (AX) for observability enables:

| Traditional SDK | AX-Enhanced SDK |
|-----------------|-----------------|
| HTTP status codes | Domain error codes |
| Implicit retry rules | Explicit retry policies |
| Runtime validation | Pre-flight validation |
| Generic capabilities | Domain-specific (evaluate, stream) |

**Reliable observability = Reliable agents**

---

# Thank You

Questions?

---

# Appendix: Error Code Reference

```go
const (
    ErrTraceNotFound      = "TRACE_NOT_FOUND"
    ErrSpanNotFound       = "SPAN_NOT_FOUND"
    ErrDatasetNotFound    = "DATASET_NOT_FOUND"
    ErrExperimentNotFound = "EXPERIMENT_NOT_FOUND"
    ErrPromptNotFound     = "PROMPT_NOT_FOUND"
    ErrProjectNotFound    = "PROJECT_NOT_FOUND"
    ErrUnauthorized       = "UNAUTHORIZED"
    ErrForbidden          = "FORBIDDEN"
    ErrInvalidInput       = "INVALID_INPUT"
    ErrConflict           = "CONFLICT"
    ErrRateLimited        = "RATE_LIMITED"
    ErrInternalError      = "INTERNAL_ERROR"
    // ... and 7 more
)
```
