// Package evaluation provides a framework for evaluating LLM outputs.
//
// The evaluation framework consists of:
//   - Metrics: Interfaces and base types for implementing evaluation metrics
//   - Scores: Result types for metric evaluations
//   - Engine: Runs metrics against inputs concurrently
//
// # Sub-packages
//
//   - heuristic: Rule-based metrics (string matching, JSON validation, text similarity)
//   - llm: LLM-based judge metrics (relevance, hallucination, factuality)
//
// # Basic Usage
//
//	import (
//	    "github.com/grokify/go-comet-ml-opik/evaluation"
//	    "github.com/grokify/go-comet-ml-opik/evaluation/heuristic"
//	)
//
//	// Create metrics
//	metrics := []evaluation.Metric{
//	    heuristic.NewEquals(false),
//	    heuristic.NewContains(false),
//	    heuristic.NewIsJSON(),
//	}
//
//	// Create engine
//	engine := evaluation.NewEngine(metrics,
//	    evaluation.WithConcurrency(4),
//	)
//
//	// Create input
//	input := evaluation.NewMetricInput("What is 2+2?", "4")
//	input = input.WithExpected("4")
//
//	// Evaluate
//	result := engine.EvaluateOne(ctx, input)
//	fmt.Printf("Score: %.2f\n", result.AverageScore())
//
// # Custom Metrics
//
//	type MyMetric struct {
//	    evaluation.BaseMetric
//	}
//
//	func NewMyMetric() *MyMetric {
//	    return &MyMetric{
//	        BaseMetric: evaluation.NewBaseMetric("my_metric"),
//	    }
//	}
//
//	func (m *MyMetric) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
//	    // Custom evaluation logic
//	    return evaluation.NewScoreResult(m.Name(), 0.95)
//	}
package evaluation
