# Experiments

Experiments track evaluation runs against datasets, allowing you to compare different models, prompts, or configurations.

## Creating Experiments

```go
// Create an experiment for a dataset
experiment, _ := client.CreateExperiment(ctx, "my-dataset",
    opik.WithExperimentName("gpt-4-evaluation-v1"),
    opik.WithExperimentMetadata(map[string]any{
        "model":       "gpt-4",
        "temperature": 0.7,
        "prompt_version": "v2",
    }),
)
```

## Logging Experiment Items

For each dataset item you evaluate, log the result:

```go
// Log an experiment item
experiment.LogItem(ctx, datasetItemID, traceID,
    opik.WithExperimentItemInput(map[string]any{"question": "What is 2+2?"}),
    opik.WithExperimentItemOutput(map[string]any{"answer": "4"}),
)
```

## Complete Evaluation Workflow

```go
func runExperiment(ctx context.Context, client *opik.Client, datasetName string) error {
    // Get dataset
    dataset, err := client.GetDatasetByName(ctx, datasetName)
    if err != nil {
        return err
    }

    // Create experiment
    experiment, err := client.CreateExperiment(ctx, datasetName,
        opik.WithExperimentName(fmt.Sprintf("eval-%s", time.Now().Format("20060102-150405"))),
        opik.WithExperimentMetadata(map[string]any{
            "model": "gpt-4",
        }),
    )
    if err != nil {
        return err
    }

    // Get dataset items
    items, err := dataset.GetItems(ctx, 1, 1000)
    if err != nil {
        return err
    }

    // Evaluate each item
    for _, item := range items {
        // Create trace for this evaluation
        trace, _ := client.Trace(ctx, "evaluate-item",
            opik.WithTraceInput(item.Data),
        )

        // Run your LLM
        input := item.Data["input"].(string)
        output, err := runLLM(ctx, input)

        // Log result
        experiment.LogItem(ctx, item.ID, trace.ID(),
            opik.WithExperimentItemInput(item.Data),
            opik.WithExperimentItemOutput(map[string]any{"response": output}),
        )

        // Add evaluation scores
        if expected, ok := item.Data["expected"].(string); ok {
            score := evaluateMatch(output, expected)
            trace.AddFeedbackScore(ctx, "accuracy", score, "")
        }

        trace.End(ctx)
    }

    // Mark experiment complete
    experiment.Complete(ctx)

    return nil
}
```

## Experiment States

```go
// Mark as complete when done
experiment.Complete(ctx)

// Cancel if something went wrong
experiment.Cancel(ctx)
```

## Listing Experiments

```go
// List experiments for a dataset
experiments, _ := client.ListExperiments(ctx, datasetID, 1, 100)

for _, exp := range experiments {
    fmt.Printf("Experiment: %s (ID: %s)\n", exp.Name, exp.ID)
}
```

## Deleting Experiments

```go
experiment.Delete(ctx)
```

## Experiment Metadata

Use metadata to track configuration and make experiments comparable:

```go
experiment, _ := client.CreateExperiment(ctx, "my-dataset",
    opik.WithExperimentName("comparison-test"),
    opik.WithExperimentMetadata(map[string]any{
        // Model configuration
        "model":       "gpt-4",
        "temperature": 0.7,
        "max_tokens":  1000,

        // Prompt information
        "prompt_version": "v2.1",
        "system_prompt":  "You are a helpful assistant...",

        // Environment
        "environment": "staging",
        "run_by":      "automated-pipeline",
    }),
)
```

## Best Practices

1. **Name experiments descriptively**: Include model, date, or version info
2. **Use consistent metadata**: Define standard fields for comparison
3. **Link traces to items**: Always associate traces with experiment items
4. **Add feedback scores**: Include evaluation metrics for analysis
5. **Complete or cancel**: Always finalize experiment state
