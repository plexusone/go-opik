package evaluation

import (
	"context"
	"fmt"
	"sync"
)

// EvaluationResult represents the result of evaluating a single item.
type EvaluationResult struct {
	// ItemID is the identifier for the evaluated item.
	ItemID string
	// Input is the input that was evaluated.
	Input MetricInput
	// Scores contains all metric scores for this item.
	Scores ScoreResults
	// Error is set if evaluation failed entirely.
	Error error
}

// IsSuccess returns true if evaluation completed without error.
func (r *EvaluationResult) IsSuccess() bool {
	return r.Error == nil
}

// AverageScore returns the average of all successful scores.
func (r *EvaluationResult) AverageScore() float64 {
	return r.Scores.Average()
}

// EvaluationResults is a collection of evaluation results.
type EvaluationResults []*EvaluationResult

// Successful returns only the successful evaluation results.
func (r EvaluationResults) Successful() EvaluationResults {
	results := make(EvaluationResults, 0, len(r))
	for _, res := range r {
		if res.IsSuccess() {
			results = append(results, res)
		}
	}
	return results
}

// Failed returns only the failed evaluation results.
func (r EvaluationResults) Failed() EvaluationResults {
	results := make(EvaluationResults, 0)
	for _, res := range r {
		if !res.IsSuccess() {
			results = append(results, res)
		}
	}
	return results
}

// AverageByMetric returns the average score for a specific metric across all items.
func (r EvaluationResults) AverageByMetric(metricName string) float64 {
	var sum float64
	var count int
	for _, res := range r {
		if score := res.Scores.ByName(metricName); score != nil && score.IsSuccess() {
			sum += score.Value
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// Summary returns a summary of scores by metric name.
func (r EvaluationResults) Summary() map[string]float64 {
	// Collect all metric names
	metricNames := make(map[string]bool)
	for _, res := range r {
		for _, score := range res.Scores {
			metricNames[score.Name] = true
		}
	}

	// Calculate averages
	summary := make(map[string]float64)
	for name := range metricNames {
		summary[name] = r.AverageByMetric(name)
	}
	return summary
}

// Engine runs evaluation metrics against data.
type Engine struct {
	metrics     []Metric
	concurrency int
	callbacks   []EvaluationCallback
}

// EvaluationCallback is called during evaluation for progress updates.
type EvaluationCallback func(completed, total int, result *EvaluationResult)

// EngineOption configures the evaluation engine.
type EngineOption func(*Engine)

// WithConcurrency sets the number of concurrent evaluations.
func WithConcurrency(n int) EngineOption {
	return func(e *Engine) {
		if n > 0 {
			e.concurrency = n
		}
	}
}

// WithCallback adds a callback for progress updates.
func WithCallback(cb EvaluationCallback) EngineOption {
	return func(e *Engine) {
		e.callbacks = append(e.callbacks, cb)
	}
}

// NewEngine creates a new evaluation engine.
func NewEngine(metrics []Metric, opts ...EngineOption) *Engine {
	e := &Engine{
		metrics:     metrics,
		concurrency: 1,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// EvaluateOne evaluates a single input against all metrics.
func (e *Engine) EvaluateOne(ctx context.Context, input MetricInput) *EvaluationResult {
	result := &EvaluationResult{
		Input:  input,
		Scores: make(ScoreResults, 0, len(e.metrics)),
	}

	for _, metric := range e.metrics {
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		default:
			score := metric.Score(ctx, input)
			result.Scores = append(result.Scores, score)
		}
	}

	return result
}

// EvaluateMany evaluates multiple inputs against all metrics.
func (e *Engine) EvaluateMany(ctx context.Context, inputs []MetricInput) EvaluationResults {
	results := make(EvaluationResults, len(inputs))

	if e.concurrency <= 1 {
		// Sequential evaluation
		for i, input := range inputs {
			results[i] = e.EvaluateOne(ctx, input)
			results[i].ItemID = fmt.Sprintf("item-%d", i)
			e.notifyCallbacks(i+1, len(inputs), results[i])
		}
		return results
	}

	// Concurrent evaluation
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, e.concurrency)
	completed := 0

	for i, input := range inputs {
		wg.Add(1)
		go func(idx int, inp MetricInput) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result := e.EvaluateOne(ctx, inp)
			result.ItemID = fmt.Sprintf("item-%d", idx)

			mu.Lock()
			results[idx] = result
			completed++
			e.notifyCallbacks(completed, len(inputs), result)
			mu.Unlock()
		}(i, input)
	}

	wg.Wait()
	return results
}

// EvaluateWithIDs evaluates inputs with explicit IDs.
func (e *Engine) EvaluateWithIDs(ctx context.Context, items map[string]MetricInput) EvaluationResults {
	results := make(EvaluationResults, 0, len(items))

	if e.concurrency <= 1 {
		i := 0
		for id, input := range items {
			result := e.EvaluateOne(ctx, input)
			result.ItemID = id
			results = append(results, result)
			i++
			e.notifyCallbacks(i, len(items), result)
		}
		return results
	}

	// Concurrent evaluation
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, e.concurrency)
	completed := 0

	for id, input := range items {
		wg.Add(1)
		go func(itemID string, inp MetricInput) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result := e.EvaluateOne(ctx, inp)
			result.ItemID = itemID

			mu.Lock()
			results = append(results, result)
			completed++
			e.notifyCallbacks(completed, len(items), result)
			mu.Unlock()
		}(id, input)
	}

	wg.Wait()
	return results
}

func (e *Engine) notifyCallbacks(completed, total int, result *EvaluationResult) {
	for _, cb := range e.callbacks {
		cb(completed, total, result)
	}
}

// Metrics returns the metrics used by this engine.
func (e *Engine) Metrics() []Metric {
	return e.metrics
}

// DatasetEvaluator evaluates metrics against a dataset.
type DatasetEvaluator struct {
	engine      *Engine
	inputMapper func(item map[string]any) MetricInput
}

// NewDatasetEvaluator creates a new dataset evaluator.
func NewDatasetEvaluator(engine *Engine, mapper func(item map[string]any) MetricInput) *DatasetEvaluator {
	return &DatasetEvaluator{
		engine:      engine,
		inputMapper: mapper,
	}
}

// Evaluate evaluates the metrics against dataset items.
func (d *DatasetEvaluator) Evaluate(ctx context.Context, items []map[string]any) EvaluationResults {
	inputs := make([]MetricInput, len(items))
	for i, item := range items {
		inputs[i] = d.inputMapper(item)
	}
	return d.engine.EvaluateMany(ctx, inputs)
}

// DefaultInputMapper creates a default input mapper for common dataset structures.
func DefaultInputMapper(inputKey, outputKey, expectedKey string) func(item map[string]any) MetricInput {
	return func(item map[string]any) MetricInput {
		input := MetricInput{
			Metadata: item,
		}
		if v, ok := item[inputKey].(string); ok {
			input.Input = v
		}
		if v, ok := item[outputKey].(string); ok {
			input.Output = v
		}
		if v, ok := item[expectedKey].(string); ok {
			input.Expected = v
		}
		return input
	}
}

// Evaluate is a convenience function to evaluate inputs with metrics.
func Evaluate(ctx context.Context, metrics []Metric, inputs []MetricInput, opts ...EngineOption) EvaluationResults {
	engine := NewEngine(metrics, opts...)
	return engine.EvaluateMany(ctx, inputs)
}

// EvaluateSingle is a convenience function to evaluate a single input.
func EvaluateSingle(ctx context.Context, metrics []Metric, input MetricInput) *EvaluationResult {
	engine := NewEngine(metrics)
	return engine.EvaluateOne(ctx, input)
}
