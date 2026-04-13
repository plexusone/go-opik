# AX Capabilities

The ax package defines 7 capability types for LLM observability operations.

## Capability Types

| Capability | Description | Example Operations |
|------------|-------------|-------------------|
| `CapRead` | Retrieves data without modification | getTraceById, findDatasets |
| `CapWrite` | Creates or modifies data | createTrace, updateSpan |
| `CapDelete` | Removes data | deleteTraceById, deleteDataset |
| `CapAdmin` | Administrative operations | upsertWorkspaceConfiguration |
| `CapStream` | Supports streaming responses | streamDatasetItems, streamExperiments |
| `CapEvaluate` | Performs LLM evaluation | evaluateTraces, evaluateSpans |
| `CapAnalytics` | Retrieves analytics/metrics data | getProjectMetrics, costsSummary |

## Checking Capabilities

```go
// Check if operation has a capability
if ax.HasCapability("evaluateTraces", ax.CapEvaluate) {
    // This is an evaluation operation
}

// Get all capabilities for an operation
caps := ax.GetCapabilities("getTraceById")
// [CapRead]

caps = ax.GetCapabilities("deleteTraceById")
// [CapWrite, CapDelete]
```

## Capability Helpers

```go
// Check if read-only
ax.IsReadOnly("getTraceById")  // true
ax.IsReadOnly("createTrace")   // false

// Check if requires admin
ax.RequiresAdmin("upsertWorkspaceConfiguration")  // true
ax.RequiresAdmin("createTrace")                   // false

// Check if evaluation operation
ax.IsEvaluation("evaluateTraces")  // true
ax.IsEvaluation("createTrace")     // false

// Check if supports streaming
ax.SupportsStreaming("streamExperimentItems")  // true
ax.SupportsStreaming("getExperimentById")      // false
```

## Finding Operations by Capability

```go
// Find all evaluation operations
evalOps := ax.GetOperationsByCapability(ax.CapEvaluate)
// ["evaluateSpans", "evaluateThreads", "evaluateTraces"]

// Find all streaming operations
streamOps := ax.GetOperationsByCapability(ax.CapStream)
// ["streamDatasetItems", "streamExperimentItems", "streamExperiments"]

// Find all analytics operations
analyticsOps := ax.GetOperationsByCapability(ax.CapAnalytics)
// ["getProjectMetrics", "getProjectStats", "costsSummary", ...]

// Find all admin operations
adminOps := ax.GetOperationsByCapability(ax.CapAdmin)
// ["upsertWorkspaceConfiguration", "storeLlmProviderApiKey", ...]
```

## Use Cases

### Discover Evaluation Options

```go
// Find available evaluation operations
evalOps := ax.GetOperationsByCapability(ax.CapEvaluate)

fmt.Println("Available evaluation methods:")
for _, op := range evalOps {
    fmt.Printf("  - %s\n", op)
}
// Available evaluation methods:
//   - evaluateSpans
//   - evaluateThreads
//   - evaluateTraces
```

### Choose Streaming vs Batch

```go
func getExperimentItems(ctx context.Context, experimentID string, large bool) {
    if large && ax.SupportsStreaming("streamExperimentItems") {
        // Use streaming for large result sets
        stream, _ := client.StreamExperimentItems(ctx, experimentID)
        for item := range stream {
            process(item)
        }
    } else {
        // Use batch for small result sets
        items, _ := client.GetExperimentItems(ctx, experimentID)
        for _, item := range items {
            process(item)
        }
    }
}
```

### Check Admin Requirements

```go
func updateWorkspaceConfig(ctx context.Context, config *Config) error {
    if ax.RequiresAdmin("upsertWorkspaceConfiguration") {
        // Verify admin permissions before calling
        if !hasAdminPermissions(ctx) {
            return ErrAdminRequired
        }
    }
    return client.UpsertWorkspaceConfiguration(ctx, config)
}
```

### Audit Read vs Write Operations

```go
func auditOperation(operationID string) {
    if ax.IsReadOnly(operationID) {
        log.Printf("READ: %s", operationID)
    } else if ax.HasCapability(operationID, ax.CapDelete) {
        log.Printf("DELETE: %s (requires confirmation)", operationID)
    } else if ax.HasCapability(operationID, ax.CapWrite) {
        log.Printf("WRITE: %s", operationID)
    }
}
```

## Complete Example

```go
func discoverCapabilities() {
    fmt.Println("=== Opik API Capabilities ===\n")

    // Evaluation
    fmt.Println("Evaluation Operations:")
    for _, op := range ax.GetOperationsByCapability(ax.CapEvaluate) {
        fmt.Printf("  - %s\n", op)
    }

    // Streaming
    fmt.Println("\nStreaming Operations:")
    for _, op := range ax.GetOperationsByCapability(ax.CapStream) {
        fmt.Printf("  - %s\n", op)
    }

    // Analytics
    fmt.Println("\nAnalytics Operations:")
    for _, op := range ax.GetOperationsByCapability(ax.CapAnalytics) {
        fmt.Printf("  - %s\n", op)
    }

    // Admin
    fmt.Println("\nAdmin Operations (require elevated permissions):")
    for _, op := range ax.GetOperationsByCapability(ax.CapAdmin) {
        fmt.Printf("  - %s\n", op)
    }
}
```
