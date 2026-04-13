# AX Package Overview

The `ax` package provides Agent Experience (AX) metadata for the Opik API. It enables AI agents to handle errors, make retry decisions, and validate requests programmatically.

## Installation

The ax package is included in opik-go:

```go
import "github.com/plexusone/opik-go/ax"
```

## Features

### Error Codes

19 domain-specific error codes for LLM observability:

```go
ax.ErrTraceNotFound      // Trace doesn't exist
ax.ErrSpanNotFound       // Span doesn't exist
ax.ErrDatasetNotFound    // Dataset doesn't exist
ax.ErrProjectNotFound    // Project doesn't exist
ax.ErrUnauthorized       // Authentication failed
ax.ErrRateLimited        // Rate limited (retryable)
```

### Retry Policies

201 operations mapped with retry safety:

```go
ax.IsRetryable("getTraceById")  // true - safe to retry
ax.IsRetryable("createTrace")   // false - would create duplicates
```

### Required Fields

67 operations have required field definitions:

```go
ax.GetRequiredFields("createExperiment")
// ["dataset_name", "name"]
```

### Capabilities

7 capability types for operation introspection:

```go
ax.HasCapability("evaluateTraces", ax.CapEvaluate)  // true
ax.SupportsStreaming("streamExperimentItems")        // true
```

## Quick Example

```go
import (
    "github.com/plexusone/opik-go"
    "github.com/plexusone/opik-go/ax"
)

trace, err := client.GetTrace(ctx, traceID)
if err != nil {
    code, ok := opik.GetAXErrorCode(err)
    if !ok {
        return nil, err
    }

    switch code {
    case ax.ErrTraceNotFound:
        // Create the trace
        return client.CreateTrace(ctx, &Trace{ID: traceID})
    case ax.ErrProjectNotFound:
        // Create project first
        client.CreateProject(ctx, projectName)
        return client.GetTrace(ctx, traceID)
    case ax.ErrRateLimited:
        // Back off and retry
        time.Sleep(time.Second)
        return client.GetTrace(ctx, traceID)
    }
}
```

## Why AX for Observability?

AI agents that observe themselves need reliable tracing. When tracing fails, the agent loses visibility into its own behavior.

AX enables self-healing observability:

1. **Precise error identification** — Know exactly what failed
2. **Automatic recovery** — Create missing resources, retry safely
3. **Pre-flight validation** — Catch errors before API calls
4. **Capability discovery** — Find streaming and evaluation operations

## Next Steps

- [Error Codes](errors.md) — Complete error code reference
- [Retry Policies](retry.md) — Operation retry safety
- [Capabilities](capabilities.md) — Operation capability mapping
- [Case Study](../case-studies/ax-integration.md) — Full integration details
