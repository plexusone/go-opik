# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-01-03

### Added

- Built-in OmniObserve `llmops` provider adapter
  - `llmops/provider.go` - Full `llmops.Provider` implementation
  - `llmops/trace.go` - Trace and span adapters
- New dataset methods in llmops provider
  - `GetDatasetByID` for retrieving datasets by ID
  - `DeleteDataset` for removing datasets
- Annotation interface stubs (returns not implemented, Opik uses feedback scores)
  - `CreateAnnotation`
  - `ListAnnotations`
- Integration tests for SDK and llmops provider
  - `integration_test.go` - SDK integration tests
  - `llmops/provider_test.go` - 27 provider tests
- Feature comparison matrix in README

### Changed

- Module renamed from `github.com/agentplexus/go-comet-ml-opik` to `github.com/agentplexus/go-opik`
- Updated dependency: `omniobserve v0.4.0` → `v0.5.0`

### Migration

Update all imports:

```go
// Before
import opik "github.com/agentplexus/go-comet-ml-opik"

// After
import opik "github.com/agentplexus/go-opik"
```

For OmniObserve users, switch to the built-in adapter:

```go
// Before (omniobserve v0.4.x)
import _ "github.com/agentplexus/omniobserve/llmops/opik"

// After (v0.5.0+)
import _ "github.com/agentplexus/go-opik/llmops"
```

## [0.4.0] - 2025-12-XX

### Fixed

- Span project association
- JSON encoding issues

See git history for earlier changes.
