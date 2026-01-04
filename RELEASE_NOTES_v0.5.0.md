# Release Notes v0.5.0

## Overview

This release includes two major changes:

1. **Module Rename**: The module has been renamed from `go-comet-ml-opik` to `go-opik` for a cleaner, shorter import path
2. **Built-in OmniObserve Provider**: The SDK now includes a `llmops` subpackage that implements the OmniObserve `llmops.Provider` interface

## Breaking Changes

### Module Renamed to go-opik

The Go module has been renamed from `github.com/agentplexus/go-comet-ml-opik` to `github.com/agentplexus/go-opik`.

**All users must update their imports:**

```go
// Before (v0.4.x)
import opik "github.com/agentplexus/go-comet-ml-opik"

// After (v0.5.0+)
import opik "github.com/agentplexus/go-opik"
```

This affects all subpackages as well:

| Before | After |
|--------|-------|
| `github.com/agentplexus/go-comet-ml-opik` | `github.com/agentplexus/go-opik` |
| `github.com/agentplexus/go-comet-ml-opik/llmops` | `github.com/agentplexus/go-opik/llmops` |
| `github.com/agentplexus/go-comet-ml-opik/middleware` | `github.com/agentplexus/go-opik/middleware` |
| `github.com/agentplexus/go-comet-ml-opik/evaluation` | `github.com/agentplexus/go-opik/evaluation` |
| `github.com/agentplexus/go-comet-ml-opik/integrations/openai` | `github.com/agentplexus/go-opik/integrations/openai` |
| `github.com/agentplexus/go-comet-ml-opik/integrations/anthropic` | `github.com/agentplexus/go-opik/integrations/anthropic` |

## New Features

### Built-in OmniObserve llmops Provider

The SDK now includes a `llmops/` subpackage that registers as an OmniObserve provider, enabling use through a unified observability abstraction consistent with other providers like go-phoenix:

```go
import (
    "github.com/agentplexus/omniobserve/llmops"
    _ "github.com/agentplexus/go-opik/llmops" // Register Opik provider
)

// Open the Opik provider through OmniObserve
provider, err := llmops.Open("opik",
    llmops.WithAPIKey("your-api-key"),
    llmops.WithWorkspace("your-workspace"),
    llmops.WithProjectName("my-project"),
)
```

### Provider Capabilities

The Opik provider registers with the following capabilities:

- `CapabilityTracing` - Full trace and span support
- `CapabilityEvaluation` - Metric evaluation
- `CapabilityPrompts` - Prompt management
- `CapabilityDatasets` - Dataset operations
- `CapabilityExperiments` - Experiment tracking
- `CapabilityStreaming` - Streaming span support
- `CapabilityDistributed` - Distributed tracing
- `CapabilityCostTracking` - Token usage and cost tracking

### Implemented Interfaces

The following OmniObserve interfaces are fully implemented:

- `llmops.Provider` - Core provider interface
- `llmops.Trace` - Trace operations (ID, Name, StartSpan, SetInput, SetOutput, SetMetadata, AddTag, AddFeedbackScore, End, Duration)
- `llmops.Span` - Span operations (all Trace methods plus TraceID, ParentSpanID, Type, SetModel, SetProvider, SetUsage)

## Architectural Change

### Previous Pattern (v0.4.x and earlier)

```
omniobserve/llmops/opik/ → imports → go-opik
```

The adapter code lived in the OmniObserve repository, requiring updates to OmniObserve whenever the Opik SDK changed.

### New Pattern (v0.5.0+)

```
go-opik/llmops/ → imports → omniobserve/llmops (interfaces only)
```

The adapter code now lives in the go-opik repository itself. This pattern:

- Makes the SDK self-contained
- Allows SDK and adapter to be updated together
- Follows the same pattern as go-phoenix
- Simplifies dependency management

## Files Added

- `llmops/provider.go` - Provider implementation with all llmops.Provider methods
- `llmops/trace.go` - Trace and span adapter implementations

## Dependencies

Added new dependency:

- `github.com/agentplexus/omniobserve v0.5.0` - For llmops interfaces

## New in v0.5.0

The llmops provider now implements additional interfaces from omniobserve v0.5.0:

- `GetDatasetByID` - Get dataset by ID
- `DeleteDataset` - Delete a dataset
- `CreateAnnotation` - Stub (returns not implemented, Opik uses feedback scores)
- `ListAnnotations` - Stub (returns not implemented, Opik uses feedback scores)

## Migration Guide

### For All Users (Module Rename)

Update all imports from `go-comet-ml-opik` to `go-opik`:

```bash
# Using sed (macOS)
find . -name "*.go" -exec sed -i '' 's|github.com/agentplexus/go-comet-ml-opik|github.com/agentplexus/go-opik|g' {} \;

# Using sed (Linux)
find . -name "*.go" -exec sed -i 's|github.com/agentplexus/go-comet-ml-opik|github.com/agentplexus/go-opik|g' {} \;
```

Then run:
```bash
go mod tidy
```

### For OmniObserve Users

If you were using the external adapter in omniobserve, switch to the built-in adapter:

**Before (v0.4.x):**
```go
import _ "github.com/agentplexus/omniobserve/llmops/opik"
```

**After (v0.5.0+):**
```go
import _ "github.com/agentplexus/go-opik/llmops"
```

The API remains identical - only the import path changes.

## Testing

All existing tests pass. The new llmops package has been verified to:

- Build successfully with `go build ./...`
- Pass linting with `golangci-lint run`
- Register correctly with OmniObserve's provider registry
