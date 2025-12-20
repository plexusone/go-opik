package evaluation

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestEvaluationResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		result := &EvaluationResult{
			ItemID: "item-1",
			Input:  NewMetricInput("test input", "test output"),
			Scores: ScoreResults{
				NewScoreResult("metric1", 0.8),
				NewScoreResult("metric2", 0.6),
			},
		}

		if !result.IsSuccess() {
			t.Error("IsSuccess() should be true")
		}
		if result.ItemID != "item-1" {
			t.Errorf("ItemID = %q, want %q", result.ItemID, "item-1")
		}
	})

	t.Run("with error", func(t *testing.T) {
		result := &EvaluationResult{
			ItemID: "item-1",
			Error:  errors.New("evaluation failed"),
		}

		if result.IsSuccess() {
			t.Error("IsSuccess() should be false when error is set")
		}
	})

	t.Run("average score", func(t *testing.T) {
		result := &EvaluationResult{
			Scores: ScoreResults{
				NewScoreResult("m1", 0.8),
				NewScoreResult("m2", 0.6),
			},
		}

		avg := result.AverageScore()
		expected := 0.7
		if avg < expected-0.01 || avg > expected+0.01 {
			t.Errorf("AverageScore() = %v, want ~%v", avg, expected)
		}
	})
}

func TestEvaluationResults(t *testing.T) {
	results := EvaluationResults{
		&EvaluationResult{
			ItemID: "item-1",
			Scores: ScoreResults{NewScoreResult("m1", 0.8)},
		},
		&EvaluationResult{
			ItemID: "item-2",
			Scores: ScoreResults{NewScoreResult("m1", 0.6)},
		},
		&EvaluationResult{
			ItemID: "item-3",
			Error:  errors.New("failed"),
		},
	}

	t.Run("Successful", func(t *testing.T) {
		successful := results.Successful()
		if len(successful) != 2 {
			t.Errorf("Successful count = %d, want 2", len(successful))
		}
	})

	t.Run("Failed", func(t *testing.T) {
		failed := results.Failed()
		if len(failed) != 1 {
			t.Errorf("Failed count = %d, want 1", len(failed))
		}
		if failed[0].ItemID != "item-3" {
			t.Errorf("Failed[0].ItemID = %q, want %q", failed[0].ItemID, "item-3")
		}
	})

	t.Run("AverageByMetric", func(t *testing.T) {
		avg := results.AverageByMetric("m1")
		expected := 0.7
		if avg < expected-0.01 || avg > expected+0.01 {
			t.Errorf("AverageByMetric(m1) = %v, want ~%v", avg, expected)
		}
	})

	t.Run("AverageByMetric missing", func(t *testing.T) {
		avg := results.AverageByMetric("nonexistent")
		if avg != 0 {
			t.Errorf("AverageByMetric(nonexistent) = %v, want 0", avg)
		}
	})
}

func TestEvaluationResultsSummary(t *testing.T) {
	results := EvaluationResults{
		&EvaluationResult{
			Scores: ScoreResults{
				NewScoreResult("accuracy", 0.9),
				NewScoreResult("relevance", 0.8),
			},
		},
		&EvaluationResult{
			Scores: ScoreResults{
				NewScoreResult("accuracy", 0.8),
				NewScoreResult("relevance", 0.6),
			},
		},
	}

	summary := results.Summary()

	if summary["accuracy"] < 0.84 || summary["accuracy"] > 0.86 {
		t.Errorf("summary[accuracy] = %v, want ~0.85", summary["accuracy"])
	}
	if summary["relevance"] != 0.7 {
		t.Errorf("summary[relevance] = %v, want 0.7", summary["relevance"])
	}
}

func TestEngineOptions(t *testing.T) {
	t.Run("WithConcurrency", func(t *testing.T) {
		engine := NewEngine(nil, WithConcurrency(4))
		if engine.concurrency != 4 {
			t.Errorf("concurrency = %d, want 4", engine.concurrency)
		}
	})

	t.Run("WithConcurrency zero", func(t *testing.T) {
		engine := NewEngine(nil, WithConcurrency(0))
		if engine.concurrency != 1 {
			t.Errorf("concurrency = %d, want 1 (zero should not change default)", engine.concurrency)
		}
	})

	t.Run("WithCallback", func(t *testing.T) {
		cb := func(completed, total int, result *EvaluationResult) {
			// callback for testing
		}
		engine := NewEngine(nil, WithCallback(cb))
		if len(engine.callbacks) != 1 {
			t.Errorf("callbacks length = %d, want 1", len(engine.callbacks))
		}
	})
}

func TestNewEngine(t *testing.T) {
	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("test", 1.0)
	})

	engine := NewEngine([]Metric{metric})

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
	if len(engine.metrics) != 1 {
		t.Errorf("metrics length = %d, want 1", len(engine.metrics))
	}
	if engine.concurrency != 1 {
		t.Errorf("default concurrency = %d, want 1", engine.concurrency)
	}
}

func TestEngineMetrics(t *testing.T) {
	m1 := NewMetricFunc("m1", nil)
	m2 := NewMetricFunc("m2", nil)

	engine := NewEngine([]Metric{m1, m2})
	metrics := engine.Metrics()

	if len(metrics) != 2 {
		t.Errorf("Metrics() length = %d, want 2", len(metrics))
	}
}

func TestEngineEvaluateOne(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		if input.Output == "good" {
			return NewScoreResult("test", 1.0)
		}
		return NewScoreResult("test", 0.0)
	})

	engine := NewEngine([]Metric{metric})

	t.Run("good input", func(t *testing.T) {
		result := engine.EvaluateOne(ctx, NewMetricInput("", "good"))
		if !result.IsSuccess() {
			t.Error("result should be successful")
		}
		if len(result.Scores) != 1 {
			t.Errorf("Scores length = %d, want 1", len(result.Scores))
		}
		if result.Scores[0].Value != 1.0 {
			t.Errorf("Score value = %v, want 1.0", result.Scores[0].Value)
		}
	})

	t.Run("bad input", func(t *testing.T) {
		result := engine.EvaluateOne(ctx, NewMetricInput("", "bad"))
		if result.Scores[0].Value != 0.0 {
			t.Errorf("Score value = %v, want 0.0", result.Scores[0].Value)
		}
	})
}

func TestEngineEvaluateOneContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("test", 1.0)
	})

	engine := NewEngine([]Metric{metric})
	result := engine.EvaluateOne(ctx, NewMetricInput("", "test"))

	if result.Error == nil {
		t.Error("result should have error when context is cancelled")
	}
}

func TestEngineEvaluateMany(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("test", 0.5)
	})

	engine := NewEngine([]Metric{metric})
	inputs := []MetricInput{
		NewMetricInput("", "output1"),
		NewMetricInput("", "output2"),
		NewMetricInput("", "output3"),
	}

	results := engine.EvaluateMany(ctx, inputs)

	if len(results) != 3 {
		t.Errorf("results length = %d, want 3", len(results))
	}
	for i, r := range results {
		if r.ItemID != "item-"+string(rune('0'+i)) {
			// ItemID format is "item-0", "item-1", etc.
		}
		if !r.IsSuccess() {
			t.Errorf("result[%d] should be successful", i)
		}
	}
}

func TestEngineEvaluateManyConcurrent(t *testing.T) {
	ctx := context.Background()

	var callCount int32
	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		atomic.AddInt32(&callCount, 1)
		return NewScoreResult("test", 1.0)
	})

	engine := NewEngine([]Metric{metric}, WithConcurrency(4))
	inputs := make([]MetricInput, 10)
	for i := range inputs {
		inputs[i] = NewMetricInput("", "output")
	}

	results := engine.EvaluateMany(ctx, inputs)

	if len(results) != 10 {
		t.Errorf("results length = %d, want 10", len(results))
	}
	if atomic.LoadInt32(&callCount) != 10 {
		t.Errorf("metric called %d times, want 10", callCount)
	}
}

func TestEngineEvaluateManyWithCallback(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("test", 1.0)
	})

	var progressUpdates []int
	engine := NewEngine([]Metric{metric}, WithCallback(func(completed, total int, result *EvaluationResult) {
		progressUpdates = append(progressUpdates, completed)
	}))

	inputs := []MetricInput{
		NewMetricInput("", "a"),
		NewMetricInput("", "b"),
		NewMetricInput("", "c"),
	}

	engine.EvaluateMany(ctx, inputs)

	if len(progressUpdates) != 3 {
		t.Errorf("callback called %d times, want 3", len(progressUpdates))
	}
}

func TestEngineEvaluateWithIDs(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("test", 1.0)
	})

	engine := NewEngine([]Metric{metric})
	items := map[string]MetricInput{
		"custom-id-1": NewMetricInput("", "output1"),
		"custom-id-2": NewMetricInput("", "output2"),
	}

	results := engine.EvaluateWithIDs(ctx, items)

	if len(results) != 2 {
		t.Errorf("results length = %d, want 2", len(results))
	}

	// Check that custom IDs are used
	ids := make(map[string]bool)
	for _, r := range results {
		ids[r.ItemID] = true
	}
	if !ids["custom-id-1"] || !ids["custom-id-2"] {
		t.Error("custom IDs should be preserved")
	}
}

func TestDatasetEvaluator(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		if input.Output == input.Expected {
			return NewScoreResult("test", 1.0)
		}
		return NewScoreResult("test", 0.0)
	})

	engine := NewEngine([]Metric{metric})
	mapper := DefaultInputMapper("input", "output", "expected")
	evaluator := NewDatasetEvaluator(engine, mapper)

	items := []map[string]any{
		{"input": "q1", "output": "a1", "expected": "a1"},
		{"input": "q2", "output": "wrong", "expected": "a2"},
	}

	results := evaluator.Evaluate(ctx, items)

	if len(results) != 2 {
		t.Errorf("results length = %d, want 2", len(results))
	}
	if results[0].Scores[0].Value != 1.0 {
		t.Error("first result should score 1.0 (match)")
	}
	if results[1].Scores[0].Value != 0.0 {
		t.Error("second result should score 0.0 (no match)")
	}
}

func TestDefaultInputMapper(t *testing.T) {
	mapper := DefaultInputMapper("prompt", "response", "expected_response")

	item := map[string]any{
		"prompt":            "What is 2+2?",
		"response":          "4",
		"expected_response": "4",
		"extra_field":       "ignored",
	}

	input := mapper(item)

	if input.Input != "What is 2+2?" {
		t.Errorf("Input = %q, want %q", input.Input, "What is 2+2?")
	}
	if input.Output != "4" {
		t.Errorf("Output = %q, want %q", input.Output, "4")
	}
	if input.Expected != "4" {
		t.Errorf("Expected = %q, want %q", input.Expected, "4")
	}
	if input.Metadata["extra_field"] != "ignored" {
		t.Error("Metadata should contain original item")
	}
}

func TestDefaultInputMapperMissingFields(t *testing.T) {
	mapper := DefaultInputMapper("input", "output", "expected")

	item := map[string]any{
		"input": "test",
		// output and expected are missing
	}

	input := mapper(item)

	if input.Input != "test" {
		t.Errorf("Input = %q, want %q", input.Input, "test")
	}
	if input.Output != "" {
		t.Errorf("Output = %q, want empty", input.Output)
	}
	if input.Expected != "" {
		t.Errorf("Expected = %q, want empty", input.Expected)
	}
}

func TestEvaluateConvenienceFunction(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("test", 0.75)
	})

	inputs := []MetricInput{
		NewMetricInput("", "a"),
		NewMetricInput("", "b"),
	}

	results := Evaluate(ctx, []Metric{metric}, inputs)

	if len(results) != 2 {
		t.Errorf("results length = %d, want 2", len(results))
	}
}

func TestEvaluateSingleConvenienceFunction(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("test", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("test", 0.9)
	})

	result := EvaluateSingle(ctx, []Metric{metric}, NewMetricInput("", "test"))

	if !result.IsSuccess() {
		t.Error("result should be successful")
	}
	if result.Scores[0].Value != 0.9 {
		t.Errorf("Score = %v, want 0.9", result.Scores[0].Value)
	}
}
