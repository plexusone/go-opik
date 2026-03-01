# Release Notes v0.6.0

**Release Date:** March 1, 2026

## Highlights

- Organization rename from agentplexus to plexusone
- Repository rename from go-opik to opik-go

## Breaking Changes

### Module Path Changed

The Go module path has changed from `github.com/agentplexus/go-opik` to `github.com/plexusone/opik-go`.

**Before:**

```go
import opik "github.com/agentplexus/go-opik"
import _ "github.com/agentplexus/go-opik/llmops"
```

**After:**

```go
import opik "github.com/plexusone/opik-go"
import _ "github.com/plexusone/opik-go/llmops"
```

### Upgrade Guide

Update all import statements in your code:

```bash
# Using sed (macOS)
find . -name "*.go" -exec sed -i '' 's|github.com/agentplexus/go-opik|github.com/plexusone/opik-go|g' {} +

# Using sed (Linux)
find . -name "*.go" -exec sed -i 's|github.com/agentplexus/go-opik|github.com/plexusone/opik-go|g' {} +
```

Then update your dependencies:

```bash
go get github.com/plexusone/opik-go@v0.6.0
go mod tidy
```

## Dependencies

- Upgraded `github.com/plexusone/omnillm` from v0.12.0 to v0.13.0
- Upgraded `github.com/plexusone/omniobserve` from v0.5.1 to v0.7.0
