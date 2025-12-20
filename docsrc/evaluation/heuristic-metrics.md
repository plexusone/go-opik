# Heuristic Metrics

Rule-based metrics that don't require an LLM. Fast, deterministic, and free.

```go
import "github.com/grokify/go-comet-ml-opik/evaluation/heuristic"
```

## String Matching

### Equals

Check if output matches expected exactly.

```go
// Case-insensitive
metric := heuristic.NewEquals(false)

// Case-sensitive
metric := heuristic.NewEquals(true)
```

### Contains

Check if output contains expected as substring.

```go
metric := heuristic.NewContains(false) // case-insensitive
```

### StartsWith / EndsWith

```go
metric := heuristic.NewStartsWith(false)
metric := heuristic.NewEndsWith(false)
```

### ContainsAny / ContainsAll

Check for multiple substrings.

```go
// Match if ANY substring is found
metric := heuristic.NewContainsAny([]string{"yes", "correct", "right"}, false)

// Match if ALL substrings are found
metric := heuristic.NewContainsAll([]string{"name", "email", "phone"}, false)
```

### NotEmpty

Check that output is not empty.

```go
metric := heuristic.NewNotEmpty()
```

### LengthBetween

Check output length is within range.

```go
metric := heuristic.NewLengthBetween(10, 1000) // 10-1000 characters
```

### WordCount

Check word count is within range.

```go
metric := heuristic.NewWordCount(5, 100) // 5-100 words
```

## Parsing/Format Validation

### JSON Validation

```go
// Check if valid JSON
metric := heuristic.NewIsJSON()

// Check if JSON object (not array)
metric := heuristic.NewIsJSONObject()

// Check if JSON array
metric := heuristic.NewIsJSONArray()

// Check if JSON has specific keys
metric := heuristic.NewJSONHasKeys([]string{"name", "email", "status"})

// Validate against JSON schema
schema := `{
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "age": {"type": "number"}
    },
    "required": ["name"]
}`
metric := heuristic.NewJSONSchemaValid(schema)
```

### XML Validation

```go
metric := heuristic.NewIsXML()
```

### Type Validation

```go
metric := heuristic.NewIsNumber()   // Valid number
metric := heuristic.NewIsBoolean()  // "true"/"false"
```

## Pattern Matching

### Regex Match

```go
// Must match pattern
metric := heuristic.MustRegexMatch(`^\d{3}-\d{4}$`) // phone format

// Must NOT match pattern
metric := heuristic.MustRegexNotMatch(`(?i)error|fail`)
```

### Common Formats

```go
// Email format
metric := heuristic.NewEmailFormat()

// URL format
metric := heuristic.NewURLFormat()

// Phone format (flexible)
metric := heuristic.NewPhoneFormat()

// Date format (ISO 8601)
metric := heuristic.NewDateFormat()

// UUID format
metric := heuristic.NewUUIDFormat()
```

## Text Similarity

### Levenshtein Similarity

Edit distance based similarity (0-1 scale).

```go
metric := heuristic.NewLevenshteinSimilarity(false) // case-insensitive
```

### Jaccard Similarity

Word overlap similarity.

```go
metric := heuristic.NewJaccardSimilarity(false)
```

### Cosine Similarity

Word vector cosine similarity.

```go
metric := heuristic.NewCosineSimilarity(false)
```

### BLEU Score

Machine translation evaluation metric.

```go
metric := heuristic.NewBLEU(4) // n-gram size up to 4
```

### ROUGE Score

Recall-oriented similarity metric.

```go
metric := heuristic.NewROUGE(1.0) // beta parameter
```

### Fuzzy Match

Flexible string matching with threshold.

```go
metric := heuristic.NewFuzzyMatch(0.8, false) // 80% threshold
```

## Using Multiple Metrics

```go
metrics := []evaluation.Metric{
    heuristic.NewEquals(false),
    heuristic.NewContains(false),
    heuristic.NewIsJSON(),
    heuristic.NewLevenshteinSimilarity(false),
    heuristic.NewBLEU(4),
}

engine := evaluation.NewEngine(metrics)

input := evaluation.NewMetricInput("prompt", "output").WithExpected("expected")
result := engine.EvaluateOne(ctx, input)

for name, score := range result.Scores {
    fmt.Printf("%s: %.2f\n", name, score.Value)
}
```

## Creating Custom Heuristics

Implement the `Metric` interface:

```go
type MyMetric struct {
    evaluation.BaseMetric
}

func NewMyMetric() *MyMetric {
    return &MyMetric{
        BaseMetric: evaluation.NewBaseMetric("my_metric"),
    }
}

func (m *MyMetric) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
    // Your custom logic
    score := calculateScore(input.Output, input.Expected)
    return evaluation.NewScoreResult(m.Name(), score)
}
```
