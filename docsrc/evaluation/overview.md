# Evaluation Framework

The evaluation framework provides tools for measuring LLM output quality using both rule-based heuristics and LLM-as-judge approaches.

## Architecture

```
evaluation/
├── Metric           # Interface for all metrics
├── MetricInput      # Input data for evaluation
├── ScoreResult      # Result of a metric evaluation
├── Engine           # Runs multiple metrics concurrently
├── heuristic/       # Rule-based metrics
│   ├── string.go    # String matching (equals, contains)
│   ├── parsing.go   # Format validation (JSON, XML)
│   ├── pattern.go   # Regex and format patterns
│   └── similarity.go # Text similarity (BLEU, ROUGE)
└── llm/             # LLM-based judge metrics
    ├── provider.go  # LLM provider interface
    └── metrics.go   # Judge metrics (relevance, hallucination)
```

## Quick Example

```go
import (
    "github.com/grokify/go-comet-ml-opik/evaluation"
    "github.com/grokify/go-comet-ml-opik/evaluation/heuristic"
)

// Create metrics
metrics := []evaluation.Metric{
    heuristic.NewEquals(false),           // Case-insensitive equality
    heuristic.NewContains(false),         // Substring check
    heuristic.NewIsJSON(),                // JSON validation
}

// Create engine
engine := evaluation.NewEngine(metrics,
    evaluation.WithConcurrency(4),
)

// Create input
input := evaluation.NewMetricInput("What is 2+2?", "The answer is 4.")
input = input.WithExpected("4")

// Evaluate
result := engine.EvaluateOne(ctx, input)

fmt.Printf("Average score: %.2f\n", result.AverageScore())
for name, score := range result.Scores {
    fmt.Printf("  %s: %.2f\n", name, score.Value)
}
```

## Metric Interface

All metrics implement this interface:

```go
type Metric interface {
    Name() string
    Score(ctx context.Context, input MetricInput) *ScoreResult
}
```

## MetricInput

Contains all data needed for evaluation:

```go
type MetricInput struct {
    Input    string         // The original input/prompt
    Output   string         // The LLM's output to evaluate
    Expected string         // Expected/ground truth output
    Context  string         // Additional context
    Metadata map[string]any // Any extra data
}

// Create input
input := evaluation.NewMetricInput(prompt, llmOutput)
input = input.WithExpected(expectedOutput)
input = input.WithContext(additionalContext)
```

## ScoreResult

Contains the evaluation result:

```go
type ScoreResult struct {
    Name     string         // Metric name
    Value    float64        // Score (typically 0.0 to 1.0)
    Reason   string         // Explanation for the score
    Metadata map[string]any // Additional data
    Error    error          // Error if evaluation failed
}

// Helper constructors
score := evaluation.NewScoreResult("accuracy", 0.95)
score := evaluation.NewScoreResultWithReason("accuracy", 0.95, "Exact match found")
score := evaluation.BooleanScore("is_valid", true) // 1.0 for true, 0.0 for false
```

## Evaluation Engine

Run multiple metrics concurrently:

```go
// Create engine with options
engine := evaluation.NewEngine(metrics,
    evaluation.WithConcurrency(4),  // Run 4 metrics in parallel
)

// Evaluate single input
result := engine.EvaluateOne(ctx, input)

// Evaluate multiple inputs
inputs := []evaluation.MetricInput{input1, input2, input3}
results := engine.EvaluateMany(ctx, inputs)

// Evaluate with item IDs (for datasets)
itemResults := engine.EvaluateWithIDs(ctx, map[string]evaluation.MetricInput{
    "item-1": input1,
    "item-2": input2,
})
```

## Dataset Evaluator

Evaluate entire datasets:

```go
evaluator := evaluation.NewDatasetEvaluator(engine, client)

results, err := evaluator.Evaluate(ctx, dataset,
    func(item map[string]any) string {
        // Generate output for each dataset item
        return llmClient.Complete(item["input"].(string))
    },
)
```

## Metric Categories

| Category | Description | Examples |
|----------|-------------|----------|
| [Heuristic](heuristic-metrics.md) | Rule-based, deterministic | Equals, Contains, IsJSON, BLEU |
| [LLM Judge](llm-judges.md) | Uses LLM to evaluate | Relevance, Hallucination, Factuality |

## Best Practices

1. **Combine metrics**: Use multiple metrics for comprehensive evaluation
2. **Use heuristics first**: They're faster and cheaper than LLM judges
3. **Set appropriate concurrency**: Balance speed vs. rate limits
4. **Handle errors**: Check `ScoreResult.Error` for failed evaluations
5. **Log to traces**: Add scores as feedback to traces for tracking
