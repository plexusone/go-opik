# Datasets

Datasets store evaluation data for testing and benchmarking your LLM applications.

## Creating Datasets

```go
// Create a simple dataset
dataset, _ := client.CreateDataset(ctx, "my-dataset")

// Create with options
dataset, _ := client.CreateDataset(ctx, "evaluation-data",
    opik.WithDatasetDescription("Test data for Q&A evaluation"),
    opik.WithDatasetTags("qa", "evaluation", "v1"),
)
```

## Inserting Items

### Single Item

```go
dataset.InsertItem(ctx, map[string]any{
    "input":    "What is the capital of France?",
    "expected": "Paris",
    "category": "geography",
})
```

### Multiple Items

```go
items := []map[string]any{
    {
        "input":    "What is 2+2?",
        "expected": "4",
        "category": "math",
    },
    {
        "input":    "What is 3+3?",
        "expected": "6",
        "category": "math",
    },
    {
        "input":    "What color is the sky?",
        "expected": "Blue",
        "category": "general",
    },
}

dataset.InsertItems(ctx, items)
```

## Retrieving Items

```go
// Get items with pagination
items, _ := dataset.GetItems(ctx, 1, 100) // page 1, 100 items per page

for _, item := range items {
    fmt.Printf("Input: %v\n", item.Data["input"])
    fmt.Printf("Expected: %v\n", item.Data["expected"])
}
```

## Listing Datasets

```go
// List all datasets
datasets, _ := client.ListDatasets(ctx, 1, 100)

for _, ds := range datasets {
    fmt.Printf("Dataset: %s (ID: %s)\n", ds.Name, ds.ID)
}
```

## Getting a Dataset by Name

```go
dataset, _ := client.GetDatasetByName(ctx, "my-dataset")

fmt.Printf("Name: %s\n", dataset.Name)
fmt.Printf("Description: %s\n", dataset.Description)
```

## Deleting Datasets

```go
// Delete by dataset object
dataset.Delete(ctx)
```

## Dataset Item Structure

Each dataset item can contain any fields you need:

| Common Field | Description |
|-------------|-------------|
| `input` | The input/prompt to evaluate |
| `expected` | Expected/ground truth output |
| `context` | Additional context for evaluation |
| `metadata` | Any additional metadata |

## Example: Building an Evaluation Dataset

```go
func buildEvaluationDataset(ctx context.Context, client *opik.Client) (*opik.Dataset, error) {
    // Create dataset
    dataset, err := client.CreateDataset(ctx, "qa-evaluation-v1",
        opik.WithDatasetDescription("Q&A evaluation dataset"),
        opik.WithDatasetTags("qa", "production"),
    )
    if err != nil {
        return nil, err
    }

    // Load test cases from file or database
    testCases := loadTestCases()

    // Convert to dataset items
    items := make([]map[string]any, len(testCases))
    for i, tc := range testCases {
        items[i] = map[string]any{
            "input":    tc.Question,
            "expected": tc.Answer,
            "context":  tc.Context,
            "category": tc.Category,
        }
    }

    // Bulk insert
    err = dataset.InsertItems(ctx, items)
    if err != nil {
        return nil, err
    }

    return dataset, nil
}
```

## Using Datasets with Experiments

See [Experiments](experiments.md) for how to run evaluations against datasets.
