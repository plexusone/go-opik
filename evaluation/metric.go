package evaluation

import (
	"context"
)

// Metric is the interface for all evaluation metrics.
type Metric interface {
	// Name returns the name of the metric.
	Name() string
	// Score evaluates the metric and returns a score result.
	Score(ctx context.Context, input MetricInput) *ScoreResult
}

// MetricInput contains the inputs for metric evaluation.
type MetricInput struct {
	// Input is the input to the LLM/model.
	Input string
	// Output is the output from the LLM/model.
	Output string
	// Expected is the expected/reference output (for comparison metrics).
	Expected string
	// Context is additional context provided to the model.
	Context string
	// Metadata contains additional key-value pairs.
	Metadata map[string]any
}

// NewMetricInput creates a new MetricInput with the given input and output.
func NewMetricInput(input, output string) MetricInput {
	return MetricInput{
		Input:  input,
		Output: output,
	}
}

// WithExpected returns a copy of the input with the expected value set.
func (m MetricInput) WithExpected(expected string) MetricInput {
	m.Expected = expected
	return m
}

// WithContext returns a copy of the input with the context value set.
func (m MetricInput) WithContext(ctx string) MetricInput {
	m.Context = ctx
	return m
}

// WithMetadata returns a copy of the input with additional metadata.
func (m MetricInput) WithMetadata(key string, value any) MetricInput {
	if m.Metadata == nil {
		m.Metadata = make(map[string]any)
	}
	m.Metadata[key] = value
	return m
}

// Get retrieves a value from metadata.
func (m MetricInput) Get(key string) (any, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	v, ok := m.Metadata[key]
	return v, ok
}

// GetString retrieves a string value from metadata.
func (m MetricInput) GetString(key string) string {
	v, ok := m.Get(key)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetStringSlice retrieves a string slice from metadata.
func (m MetricInput) GetStringSlice(key string) []string {
	v, ok := m.Get(key)
	if !ok {
		return nil
	}
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		result := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

// AsyncMetric is a metric that can be evaluated asynchronously.
type AsyncMetric interface {
	Metric
	// ScoreAsync evaluates the metric asynchronously.
	ScoreAsync(ctx context.Context, input MetricInput) <-chan *ScoreResult
}

// BatchMetric is a metric that can evaluate multiple inputs at once.
type BatchMetric interface {
	Metric
	// ScoreBatch evaluates multiple inputs and returns results for each.
	ScoreBatch(ctx context.Context, inputs []MetricInput) ScoreResults
}

// BaseMetric provides common functionality for metrics.
type BaseMetric struct {
	name string
}

// NewBaseMetric creates a new base metric with the given name.
func NewBaseMetric(name string) BaseMetric {
	return BaseMetric{name: name}
}

// Name returns the name of the metric.
func (m BaseMetric) Name() string {
	return m.name
}

// MetricFunc is a function-based metric implementation.
type MetricFunc struct {
	BaseMetric
	fn func(ctx context.Context, input MetricInput) *ScoreResult
}

// NewMetricFunc creates a new metric from a function.
func NewMetricFunc(name string, fn func(ctx context.Context, input MetricInput) *ScoreResult) *MetricFunc {
	return &MetricFunc{
		BaseMetric: NewBaseMetric(name),
		fn:         fn,
	}
}

// Score evaluates the metric.
func (m *MetricFunc) Score(ctx context.Context, input MetricInput) *ScoreResult {
	return m.fn(ctx, input)
}

// CompositeMetric combines multiple metrics into one.
type CompositeMetric struct {
	BaseMetric
	metrics []Metric
}

// NewCompositeMetric creates a new composite metric.
func NewCompositeMetric(name string, metrics ...Metric) *CompositeMetric {
	return &CompositeMetric{
		BaseMetric: NewBaseMetric(name),
		metrics:    metrics,
	}
}

// Score evaluates all contained metrics and returns the average score.
func (m *CompositeMetric) Score(ctx context.Context, input MetricInput) *ScoreResult {
	var scores ScoreResults
	for _, metric := range m.metrics {
		scores = append(scores, metric.Score(ctx, input))
	}
	return NewScoreResult(m.name, scores.Average())
}

// Metrics returns the contained metrics.
func (m *CompositeMetric) Metrics() []Metric {
	return m.metrics
}

// ScoreAll evaluates all contained metrics and returns all results.
func (m *CompositeMetric) ScoreAll(ctx context.Context, input MetricInput) ScoreResults {
	var scores ScoreResults
	for _, metric := range m.metrics {
		scores = append(scores, metric.Score(ctx, input))
	}
	return scores
}

// ConditionalMetric evaluates a metric only if a condition is met.
type ConditionalMetric struct {
	BaseMetric
	condition func(input MetricInput) bool
	metric    Metric
}

// NewConditionalMetric creates a new conditional metric.
func NewConditionalMetric(name string, condition func(input MetricInput) bool, metric Metric) *ConditionalMetric {
	return &ConditionalMetric{
		BaseMetric: NewBaseMetric(name),
		condition:  condition,
		metric:     metric,
	}
}

// Score evaluates the metric if the condition is met.
func (m *ConditionalMetric) Score(ctx context.Context, input MetricInput) *ScoreResult {
	if !m.condition(input) {
		return NewScoreResultWithReason(m.name, 0, "condition not met")
	}
	return m.metric.Score(ctx, input)
}

// WeightedMetric applies a weight to a metric's score.
type WeightedMetric struct {
	metric Metric
	weight float64
}

// NewWeightedMetric creates a new weighted metric.
func NewWeightedMetric(metric Metric, weight float64) *WeightedMetric {
	return &WeightedMetric{
		metric: metric,
		weight: weight,
	}
}

// Name returns the name of the underlying metric.
func (m *WeightedMetric) Name() string {
	return m.metric.Name()
}

// Score evaluates the metric and applies the weight.
func (m *WeightedMetric) Score(ctx context.Context, input MetricInput) *ScoreResult {
	result := m.metric.Score(ctx, input)
	if result.Error != nil {
		return result
	}
	return &ScoreResult{
		Name:     result.Name,
		Value:    result.Value * m.weight,
		Reason:   result.Reason,
		Metadata: result.Metadata,
	}
}

// Weight returns the weight factor.
func (m *WeightedMetric) Weight() float64 {
	return m.weight
}
