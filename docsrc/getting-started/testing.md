# Testing

The Go Opik SDK includes a comprehensive test suite to ensure reliability and correctness.

## Running Tests

### Quick Start

Run all tests without any API key or external dependencies:

```bash
go test ./...
```

### Verbose Output

See detailed test output:

```bash
go test ./... -v
```

### Run Specific Test Packages

```bash
# Core SDK tests (config, context)
go test github.com/grokify/go-comet-ml-opik -v

# Evaluation heuristic metrics
go test github.com/grokify/go-comet-ml-opik/evaluation/heuristic -v
```

## API Key Requirements

!!! note "No API Key Required for Unit Tests"
    **All unit tests run locally without requiring an Opik API key or server connection.** The test suite is designed to test the SDK's logic, data structures, and algorithms without making external API calls.

### What Tests Don't Require API Keys

| Test Category | Description |
|--------------|-------------|
| **Config Tests** | Configuration loading from environment variables and files |
| **Context Tests** | Context propagation for traces and spans |
| **String Metrics** | Equals, Contains, StartsWith, EndsWith, etc. |
| **Similarity Metrics** | Levenshtein, Jaccard, Cosine, BLEU, ROUGE, etc. |

### When You Need an API Key

API keys are only required for:

- **Integration tests** (if added) that verify end-to-end functionality with an Opik server
- **Running the CLI** to interact with a live Opik instance
- **Production usage** of the SDK to send traces to Opik Cloud

## Test Coverage

### Core SDK (`config_test.go`, `context_test.go`)

Tests for configuration management and context propagation:

```bash
go test github.com/grokify/go-comet-ml-opik -v -run TestConfig
go test github.com/grokify/go-comet-ml-opik -v -run TestContext
```

**Config tests cover:**

- `TestNewConfig` - Default configuration values
- `TestLoadConfigFromEnv` - Environment variable loading
- `TestLoadConfigTracingDisableVariants` - Tracing disable flag variants
- `TestLoadConfigDefaultURL` - URL defaults based on API key presence
- `TestConfigIsCloud` - Cloud vs local detection
- `TestConfigValidate` - Configuration validation
- `TestSaveConfig` - Configuration file saving

**Context tests cover:**

- `TestContextWithTrace` / `TestTraceFromContext` - Trace context storage
- `TestContextWithSpan` / `TestSpanFromContext` - Span context storage
- `TestContextWithClient` / `TestClientFromContext` - Client context storage
- `TestContextChaining` - Multiple values in context
- `TestCurrentTraceID` / `TestCurrentSpanID` - ID retrieval from context
- `TestStartSpanNoActiveTrace` - Error handling

### Heuristic Metrics (`evaluation/heuristic/`)

Tests for string and similarity metrics:

```bash
go test github.com/grokify/go-comet-ml-opik/evaluation/heuristic -v
```

**String metric tests (`string_test.go`):**

- `TestEquals` / `TestEqualsIgnoreCase` - Exact string matching
- `TestContains` / `TestContainsIgnoreCase` - Substring matching
- `TestStartsWith` / `TestEndsWith` - Prefix/suffix matching
- `TestContainsAny` / `TestContainsAll` - Multiple value matching
- `TestNotEmpty` - Empty string detection
- `TestLengthBetween` / `TestWordCount` - Length constraints
- `TestNoOffensiveLanguage` - Content filtering
- `TestMetricName` - Metric naming consistency

**Similarity metric tests (`similarity_test.go`):**

- `TestLevenshteinSimilarity` - Edit distance similarity
- `TestLevenshteinDistance` - Raw edit distance calculation
- `TestJaccardSimilarity` - Set-based similarity
- `TestCosineSimilarity` - Vector-based similarity
- `TestBLEU` - Machine translation metric
- `TestROUGE` - Summarization metric
- `TestFuzzyMatch` - Combined fuzzy matching
- `TestSemanticSimilarity` - Word-based semantic similarity
- `TestLCSLength` - Longest common subsequence

## Writing New Tests

### Test File Naming

Follow Go conventions:

- `foo.go` → `foo_test.go`
- Package remains the same (not `_test` suffix for internal tests)

### Test Structure Example

```go
package opik

import (
    "context"
    "testing"
)

func TestMyFunction(t *testing.T) {
    ctx := context.Background()

    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"basic case", "input", "expected"},
        {"edge case", "", ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(ctx, tt.input)
            if result != tt.expected {
                t.Errorf("MyFunction() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Testing Metrics

Use the `evaluation.MetricInput` type:

```go
func TestMyMetric(t *testing.T) {
    ctx := context.Background()
    metric := NewMyMetric()

    input := evaluation.NewMetricInput("", "output text").
        WithExpected("expected text")

    result := metric.Score(ctx, input)

    if result.Value < 0.9 {
        t.Errorf("Score = %v, want >= 0.9", result.Value)
    }
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run Tests
        run: go test ./... -v -race

      - name: Run Tests with Coverage
        run: go test ./... -coverprofile=coverage.out

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: coverage.out
```

## Test Coverage Report

Generate a coverage report:

```bash
# Generate coverage profile
go test ./... -coverprofile=coverage.out

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

## Troubleshooting

### Tests Fail Due to Missing Dependencies

```bash
go mod download
go mod tidy
```

### Timeout Issues

Increase test timeout for slow machines:

```bash
go test ./... -timeout 5m
```

### Race Condition Detection

Run tests with race detector:

```bash
go test ./... -race
```
