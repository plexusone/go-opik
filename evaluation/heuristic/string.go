package heuristic

import (
	"context"
	"strings"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

// Equals checks if the output exactly matches the expected value.
type Equals struct {
	evaluation.BaseMetric
	caseSensitive bool
}

// NewEquals creates a new Equals metric.
func NewEquals(caseSensitive bool) *Equals {
	return &Equals{
		BaseMetric:    evaluation.NewBaseMetric("equals"),
		caseSensitive: caseSensitive,
	}
}

// Score evaluates if output equals expected.
func (m *Equals) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := input.Output
	expected := input.Expected

	if !m.caseSensitive {
		output = strings.ToLower(output)
		expected = strings.ToLower(expected)
	}

	if output == expected {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "exact match")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "no match")
}

// Contains checks if the output contains the expected value.
type Contains struct {
	evaluation.BaseMetric
	caseSensitive bool
}

// NewContains creates a new Contains metric.
func NewContains(caseSensitive bool) *Contains {
	return &Contains{
		BaseMetric:    evaluation.NewBaseMetric("contains"),
		caseSensitive: caseSensitive,
	}
}

// Score evaluates if output contains expected.
func (m *Contains) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := input.Output
	expected := input.Expected

	if !m.caseSensitive {
		output = strings.ToLower(output)
		expected = strings.ToLower(expected)
	}

	if strings.Contains(output, expected) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "contains expected value")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "does not contain expected value")
}

// StartsWith checks if the output starts with the expected value.
type StartsWith struct {
	evaluation.BaseMetric
	caseSensitive bool
}

// NewStartsWith creates a new StartsWith metric.
func NewStartsWith(caseSensitive bool) *StartsWith {
	return &StartsWith{
		BaseMetric:    evaluation.NewBaseMetric("starts_with"),
		caseSensitive: caseSensitive,
	}
}

// Score evaluates if output starts with expected.
func (m *StartsWith) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := input.Output
	expected := input.Expected

	if !m.caseSensitive {
		output = strings.ToLower(output)
		expected = strings.ToLower(expected)
	}

	if strings.HasPrefix(output, expected) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "starts with expected value")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "does not start with expected value")
}

// EndsWith checks if the output ends with the expected value.
type EndsWith struct {
	evaluation.BaseMetric
	caseSensitive bool
}

// NewEndsWith creates a new EndsWith metric.
func NewEndsWith(caseSensitive bool) *EndsWith {
	return &EndsWith{
		BaseMetric:    evaluation.NewBaseMetric("ends_with"),
		caseSensitive: caseSensitive,
	}
}

// Score evaluates if output ends with expected.
func (m *EndsWith) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := input.Output
	expected := input.Expected

	if !m.caseSensitive {
		output = strings.ToLower(output)
		expected = strings.ToLower(expected)
	}

	if strings.HasSuffix(output, expected) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "ends with expected value")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "does not end with expected value")
}

// ContainsAny checks if the output contains any of the specified values.
type ContainsAny struct {
	evaluation.BaseMetric
	values        []string
	caseSensitive bool
}

// NewContainsAny creates a new ContainsAny metric.
func NewContainsAny(values []string, caseSensitive bool) *ContainsAny {
	return &ContainsAny{
		BaseMetric:    evaluation.NewBaseMetric("contains_any"),
		values:        values,
		caseSensitive: caseSensitive,
	}
}

// Score evaluates if output contains any of the values.
func (m *ContainsAny) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := input.Output
	if !m.caseSensitive {
		output = strings.ToLower(output)
	}

	for _, v := range m.values {
		check := v
		if !m.caseSensitive {
			check = strings.ToLower(v)
		}
		if strings.Contains(output, check) {
			return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "contains: "+v)
		}
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "does not contain any expected value")
}

// ContainsAll checks if the output contains all of the specified values.
type ContainsAll struct {
	evaluation.BaseMetric
	values        []string
	caseSensitive bool
}

// NewContainsAll creates a new ContainsAll metric.
func NewContainsAll(values []string, caseSensitive bool) *ContainsAll {
	return &ContainsAll{
		BaseMetric:    evaluation.NewBaseMetric("contains_all"),
		values:        values,
		caseSensitive: caseSensitive,
	}
}

// Score evaluates if output contains all of the values.
func (m *ContainsAll) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := input.Output
	if !m.caseSensitive {
		output = strings.ToLower(output)
	}

	missing := []string{}
	for _, v := range m.values {
		check := v
		if !m.caseSensitive {
			check = strings.ToLower(v)
		}
		if !strings.Contains(output, check) {
			missing = append(missing, v)
		}
	}

	if len(missing) == 0 {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "contains all expected values")
	}

	// Return partial score based on how many were found
	found := len(m.values) - len(missing)
	score := float64(found) / float64(len(m.values))
	return evaluation.NewScoreResultWithReason(m.Name(), score,
		"missing: "+strings.Join(missing, ", "))
}

// NotEmpty checks if the output is not empty.
type NotEmpty struct {
	evaluation.BaseMetric
}

// NewNotEmpty creates a new NotEmpty metric.
func NewNotEmpty() *NotEmpty {
	return &NotEmpty{
		BaseMetric: evaluation.NewBaseMetric("not_empty"),
	}
}

// Score evaluates if output is not empty.
func (m *NotEmpty) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	if strings.TrimSpace(input.Output) != "" {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "output is not empty")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "output is empty")
}

// LengthBetween checks if the output length is within a range.
type LengthBetween struct {
	evaluation.BaseMetric
	min int
	max int
}

// NewLengthBetween creates a new LengthBetween metric.
func NewLengthBetween(min, max int) *LengthBetween {
	return &LengthBetween{
		BaseMetric: evaluation.NewBaseMetric("length_between"),
		min:        min,
		max:        max,
	}
}

// Score evaluates if output length is within range.
func (m *LengthBetween) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	length := len(input.Output)
	if length >= m.min && length <= m.max {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "length within range")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0,
		"length out of range: "+string(rune(length)))
}

// WordCount checks if the output word count is within a range.
type WordCount struct {
	evaluation.BaseMetric
	min int
	max int
}

// NewWordCount creates a new WordCount metric.
func NewWordCount(min, max int) *WordCount {
	return &WordCount{
		BaseMetric: evaluation.NewBaseMetric("word_count"),
		min:        min,
		max:        max,
	}
}

// Score evaluates if output word count is within range.
func (m *WordCount) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	words := strings.Fields(input.Output)
	count := len(words)
	if count >= m.min && count <= m.max {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "word count within range")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "word count out of range")
}

// NoOffensiveLanguage checks for offensive language patterns.
type NoOffensiveLanguage struct {
	evaluation.BaseMetric
	patterns []string
}

// NewNoOffensiveLanguage creates a new NoOffensiveLanguage metric with custom patterns.
func NewNoOffensiveLanguage(patterns []string) *NoOffensiveLanguage {
	return &NoOffensiveLanguage{
		BaseMetric: evaluation.NewBaseMetric("no_offensive_language"),
		patterns:   patterns,
	}
}

// Score evaluates if output contains offensive language.
func (m *NoOffensiveLanguage) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	lower := strings.ToLower(input.Output)
	for _, pattern := range m.patterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "contains offensive pattern")
		}
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "no offensive language detected")
}
