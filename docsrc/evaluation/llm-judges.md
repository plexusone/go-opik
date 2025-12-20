# LLM Judge Metrics

Use an LLM to evaluate outputs that can't be measured with simple rules.

```go
import (
    "github.com/grokify/go-comet-ml-opik/evaluation/llm"
    "github.com/grokify/go-comet-ml-opik/integrations/openai"
)
```

## Setting Up a Provider

First, create an LLM provider:

```go
// OpenAI
provider := openai.NewProvider(
    openai.WithAPIKey("your-api-key"),
    openai.WithModel("gpt-4o"),
)

// Anthropic
provider := anthropic.NewProvider(
    anthropic.WithAPIKey("your-api-key"),
    anthropic.WithModel("claude-sonnet-4-20250514"),
)

// gollm (any provider)
provider := gollm.NewProvider(gollmClient,
    gollm.WithModel("gpt-4o"),
)
```

## Built-in Judge Metrics

### Answer Relevance

Evaluates how relevant the answer is to the question.

```go
metric := llm.NewAnswerRelevance(provider)
```

### Hallucination

Detects factual claims not supported by the context.

```go
metric := llm.NewHallucination(provider)
```

Requires context in the input:

```go
input := evaluation.NewMetricInput(question, answer).
    WithContext(relevantDocuments)
```

### Factuality

Checks if the response is factually accurate.

```go
metric := llm.NewFactuality(provider)
```

### Context Recall

Measures how much of the expected information is captured.

```go
metric := llm.NewContextRecall(provider)
```

### Context Precision

Measures precision of information retrieval.

```go
metric := llm.NewContextPrecision(provider)
```

### Moderation

Checks for harmful, inappropriate, or policy-violating content.

```go
metric := llm.NewModeration(provider)
```

### Coherence

Evaluates logical flow and consistency.

```go
metric := llm.NewCoherence(provider)
```

### Helpfulness

Measures how helpful the response is to the user.

```go
metric := llm.NewHelpfulness(provider)
```

## G-EVAL

Flexible evaluation with custom criteria and evaluation steps.

```go
geval := llm.NewGEval(provider, "fluency and coherence")

// With custom evaluation steps
geval = geval.WithEvaluationSteps([]string{
    "Check if the response is grammatically correct",
    "Evaluate the logical flow of ideas",
    "Assess clarity and readability",
    "Check for appropriate vocabulary usage",
})

score := geval.Score(ctx, input)
```

## Custom Judge

Create a judge with a custom prompt template:

```go
prompt := `
Evaluate whether the response maintains a professional tone.

User message: {{input}}
AI response: {{output}}

Provide a score from 0.0 to 1.0 where:
- 1.0: Completely professional
- 0.5: Somewhat professional with minor issues
- 0.0: Unprofessional

Return JSON: {"score": <float>, "reason": "<explanation>"}
`

judge := llm.NewCustomJudge("tone_check", prompt, provider)
```

### Template Variables

| Variable | Description |
|----------|-------------|
| `{{input}}` | The original input/prompt |
| `{{output}}` | The LLM's response |
| `{{expected}}` | Expected/ground truth output |
| `{{context}}` | Additional context |

## Using Multiple Judges

```go
metrics := []evaluation.Metric{
    llm.NewAnswerRelevance(provider),
    llm.NewHallucination(provider),
    llm.NewCoherence(provider),
    llm.NewHelpfulness(provider),
}

engine := evaluation.NewEngine(metrics,
    evaluation.WithConcurrency(2), // Limit concurrent LLM calls
)

input := evaluation.NewMetricInput(question, answer).
    WithExpected(expectedAnswer).
    WithContext(documents)

result := engine.EvaluateOne(ctx, input)
```

## Caching Responses

Reduce costs by caching identical evaluations:

```go
// Wrap provider with caching
cachedProvider := llm.NewCachingProvider(provider)

// Use cached provider for metrics
metric := llm.NewAnswerRelevance(cachedProvider)
```

## Best Practices

1. **Choose appropriate models**: GPT-4 or Claude 3 for nuanced evaluation
2. **Limit concurrency**: Respect rate limits
3. **Use caching**: For repeated evaluations
4. **Combine with heuristics**: Use LLM judges only when needed
5. **Monitor costs**: LLM evaluations add up

## Cost Considerations

Each LLM judge metric makes an API call. For large datasets:

1. Pre-filter with heuristic metrics
2. Use caching for duplicate inputs
3. Batch evaluations during off-peak hours
4. Consider smaller models for simple judgments
