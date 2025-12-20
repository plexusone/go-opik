# Development Guide

This guide covers development setup, code generation, and contribution guidelines for the Go Opik SDK.

## Prerequisites

Before contributing to the SDK, ensure you have the following installed:

- **Go 1.21+** - [Download Go](https://go.dev/dl/)
- **golangci-lint** - For linting: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **ogen** - For API client generation: `go install github.com/ogen-go/ogen/cmd/ogen@latest`

## Project Structure

```
go-comet-ml-opik/
├── internal/api/           # ogen-generated API client (DO NOT EDIT)
├── integrations/           # LLM provider integrations
│   ├── openai/
│   ├── anthropic/
│   └── gollm/
├── evaluation/             # Evaluation framework
│   ├── heuristic/
│   └── llm/
├── middleware/             # HTTP tracing middleware
├── testutil/               # Test utilities
├── cmd/opik/               # CLI tool
├── examples/               # Usage examples
├── openapi/                # OpenAPI specification
│   └── openapi.yaml
└── docsrc/                 # Documentation source
```

## Regenerating the API Client

The SDK uses [ogen](https://github.com/ogen-go/ogen) to generate a type-safe API client from the Opik OpenAPI specification. When the upstream API changes, you'll need to regenerate the client code.

### Step 1: Update the OpenAPI Specification

The OpenAPI spec is located at `openapi/openapi.yaml`. To update it:

```bash
# Option A: Copy from your local Opik clone
cp /path/to/opik/sdks/code_generation/fern/openapi/openapi.yaml openapi/

# Option B: Download from the Opik repository
curl -o openapi/openapi.yaml \
  https://raw.githubusercontent.com/comet-ml/opik/main/sdks/code_generation/fern/openapi/openapi.yaml
```

### Step 2: Install ogen (if not already installed)

```bash
go install github.com/ogen-go/ogen/cmd/ogen@latest
```

Verify installation:

```bash
ogen --version
```

### Step 3: Run the Generation Script

```bash
./generate.sh
```

This script:

1. Runs ogen to generate Go code from the OpenAPI spec
2. Applies fixes for known issues (e.g., `jx.Raw` comparison)
3. Runs `go mod tidy` to update dependencies
4. Verifies the build compiles

### Manual Generation

If you need to run the generation manually:

```bash
# Generate the API client
ogen --package api --target internal/api --clean openapi/openapi.yaml

# Update dependencies
go mod tidy

# Verify build
go build ./...
```

### Known Issues and Fixes

The generated code may have issues that require post-processing:

#### jx.Raw Comparison

The ogen generator creates equality functions that compare `jx.Raw` values directly, which doesn't work in Go. The `generate.sh` script automatically fixes this by replacing direct comparisons with `bytes.Equal()`.

**Before (generated):**
```go
if a.Input != b.Input {
    return false
}
```

**After (fixed):**
```go
if !bytes.Equal([]byte(a.Input), []byte(b.Input)) {
    return false
}
```

### Troubleshooting

#### "ogen: command not found"

Ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Build Errors After Generation

If you encounter build errors after regenerating:

1. Check that the OpenAPI spec is valid
2. Ensure you're using a compatible ogen version
3. Review the generated code for any new fields that need handling
4. Run `go mod tidy` to resolve dependency issues

#### API Changes Breaking Existing Code

When the API changes, you may need to update the SDK wrapper code (files outside `internal/api/`):

1. Check for new required fields in request/response types
2. Update any `OptXxx` type handling for new optional fields
3. Add support for new endpoints in the appropriate files

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Running Linter

```bash
# Run all linters
golangci-lint run

# Run with auto-fix where possible
golangci-lint run --fix
```

## Building the CLI

```bash
# Build for current platform
make build-cli

# Or manually
go build -o bin/opik ./cmd/opik
```

## Building Documentation

```bash
# Generate presentation HTML
make presentation

# Build MkDocs site
make docs

# Serve locally for development
make docs-serve
```

## Code Style Guidelines

- Follow standard Go conventions and idioms
- Use the functional options pattern for configuration
- Handle all errors explicitly (checked by `errcheck` linter)
- Add `//nolint` directives only with explanatory comments
- Keep the `internal/api/` directory untouched (auto-generated)

## Submitting Changes

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linter: `go test ./... && golangci-lint run`
5. Submit a pull request

## Getting Help

- [GitHub Issues](https://github.com/grokify/go-comet-ml-opik/issues)
- [Opik Documentation](https://www.comet.com/docs/opik/)
- [ogen Documentation](https://ogen.dev/)
