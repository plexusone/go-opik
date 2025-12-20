package evaluation

import (
	"context"
	"errors"
	"testing"
)

func TestNewMetricInput(t *testing.T) {
	input := NewMetricInput("user query", "model response")

	if input.Input != "user query" {
		t.Errorf("Input = %q, want %q", input.Input, "user query")
	}
	if input.Output != "model response" {
		t.Errorf("Output = %q, want %q", input.Output, "model response")
	}
	if input.Expected != "" {
		t.Errorf("Expected = %q, want empty", input.Expected)
	}
	if input.Context != "" {
		t.Errorf("Context = %q, want empty", input.Context)
	}
}

func TestMetricInputWithExpected(t *testing.T) {
	input := NewMetricInput("", "output").WithExpected("expected")

	if input.Expected != "expected" {
		t.Errorf("Expected = %q, want %q", input.Expected, "expected")
	}
}

func TestMetricInputWithContext(t *testing.T) {
	input := NewMetricInput("", "output").WithContext("some context")

	if input.Context != "some context" {
		t.Errorf("Context = %q, want %q", input.Context, "some context")
	}
}

func TestMetricInputWithMetadata(t *testing.T) {
	input := NewMetricInput("", "output").
		WithMetadata("key1", "value1").
		WithMetadata("key2", 123)

	if input.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}
	if input.Metadata["key1"] != "value1" {
		t.Errorf("Metadata[key1] = %v, want %q", input.Metadata["key1"], "value1")
	}
	if input.Metadata["key2"] != 123 {
		t.Errorf("Metadata[key2] = %v, want %d", input.Metadata["key2"], 123)
	}
}

func TestMetricInputGet(t *testing.T) {
	input := NewMetricInput("", "").WithMetadata("key", "value")

	t.Run("existing key", func(t *testing.T) {
		val, ok := input.Get("key")
		if !ok {
			t.Error("Get should return true for existing key")
		}
		if val != "value" {
			t.Errorf("Get = %v, want %q", val, "value")
		}
	})

	t.Run("missing key", func(t *testing.T) {
		_, ok := input.Get("missing")
		if ok {
			t.Error("Get should return false for missing key")
		}
	})

	t.Run("nil metadata", func(t *testing.T) {
		emptyInput := NewMetricInput("", "")
		_, ok := emptyInput.Get("any")
		if ok {
			t.Error("Get should return false for nil metadata")
		}
	})
}

func TestMetricInputGetString(t *testing.T) {
	input := NewMetricInput("", "").
		WithMetadata("string", "value").
		WithMetadata("number", 123)

	if got := input.GetString("string"); got != "value" {
		t.Errorf("GetString(string) = %q, want %q", got, "value")
	}
	if got := input.GetString("number"); got != "" {
		t.Errorf("GetString(number) = %q, want empty", got)
	}
	if got := input.GetString("missing"); got != "" {
		t.Errorf("GetString(missing) = %q, want empty", got)
	}
}

func TestMetricInputGetStringSlice(t *testing.T) {
	input := NewMetricInput("", "").
		WithMetadata("strings", []string{"a", "b", "c"}).
		WithMetadata("anys", []any{"x", "y"}).
		WithMetadata("notSlice", "single")

	t.Run("string slice", func(t *testing.T) {
		got := input.GetStringSlice("strings")
		if len(got) != 3 || got[0] != "a" {
			t.Errorf("GetStringSlice(strings) = %v, want [a b c]", got)
		}
	})

	t.Run("any slice", func(t *testing.T) {
		got := input.GetStringSlice("anys")
		if len(got) != 2 || got[0] != "x" {
			t.Errorf("GetStringSlice(anys) = %v, want [x y]", got)
		}
	})

	t.Run("not a slice", func(t *testing.T) {
		got := input.GetStringSlice("notSlice")
		if got != nil {
			t.Errorf("GetStringSlice(notSlice) = %v, want nil", got)
		}
	})
}

func TestBaseMetric(t *testing.T) {
	metric := NewBaseMetric("test_metric")

	if metric.Name() != "test_metric" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "test_metric")
	}
}

func TestMetricFunc(t *testing.T) {
	ctx := context.Background()

	metric := NewMetricFunc("custom", func(ctx context.Context, input MetricInput) *ScoreResult {
		if input.Output == "good" {
			return NewScoreResult("custom", 1.0)
		}
		return NewScoreResult("custom", 0.0)
	})

	if metric.Name() != "custom" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "custom")
	}

	t.Run("good output", func(t *testing.T) {
		result := metric.Score(ctx, NewMetricInput("", "good"))
		if result.Value != 1.0 {
			t.Errorf("Score = %v, want 1.0", result.Value)
		}
	})

	t.Run("bad output", func(t *testing.T) {
		result := metric.Score(ctx, NewMetricInput("", "bad"))
		if result.Value != 0.0 {
			t.Errorf("Score = %v, want 0.0", result.Value)
		}
	})
}

func TestCompositeMetric(t *testing.T) {
	ctx := context.Background()

	// Create simple metrics that return fixed scores
	metric1 := NewMetricFunc("m1", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("m1", 0.8)
	})
	metric2 := NewMetricFunc("m2", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("m2", 0.6)
	})

	composite := NewCompositeMetric("composite", metric1, metric2)

	if composite.Name() != "composite" {
		t.Errorf("Name() = %q, want %q", composite.Name(), "composite")
	}

	t.Run("Score returns average", func(t *testing.T) {
		result := composite.Score(ctx, NewMetricInput("", "test"))
		// Average of 0.8 and 0.6 = 0.7
		if result.Value != 0.7 {
			t.Errorf("Score = %v, want 0.7", result.Value)
		}
	})

	t.Run("Metrics returns contained metrics", func(t *testing.T) {
		metrics := composite.Metrics()
		if len(metrics) != 2 {
			t.Errorf("Metrics() length = %d, want 2", len(metrics))
		}
	})

	t.Run("ScoreAll returns all results", func(t *testing.T) {
		results := composite.ScoreAll(ctx, NewMetricInput("", "test"))
		if len(results) != 2 {
			t.Errorf("ScoreAll() length = %d, want 2", len(results))
		}
		if results[0].Name != "m1" || results[1].Name != "m2" {
			t.Error("ScoreAll should return results from both metrics")
		}
	})
}

func TestConditionalMetric(t *testing.T) {
	ctx := context.Background()

	innerMetric := NewMetricFunc("inner", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("inner", 1.0)
	})

	condition := func(input MetricInput) bool {
		return input.Output != ""
	}

	conditional := NewConditionalMetric("conditional", condition, innerMetric)

	if conditional.Name() != "conditional" {
		t.Errorf("Name() = %q, want %q", conditional.Name(), "conditional")
	}

	t.Run("condition met", func(t *testing.T) {
		result := conditional.Score(ctx, NewMetricInput("", "has output"))
		if result.Value != 1.0 {
			t.Errorf("Score = %v, want 1.0", result.Value)
		}
	})

	t.Run("condition not met", func(t *testing.T) {
		result := conditional.Score(ctx, NewMetricInput("", ""))
		if result.Value != 0.0 {
			t.Errorf("Score = %v, want 0.0", result.Value)
		}
		if result.Reason != "condition not met" {
			t.Errorf("Reason = %q, want %q", result.Reason, "condition not met")
		}
	})
}

func TestWeightedMetric(t *testing.T) {
	ctx := context.Background()

	innerMetric := NewMetricFunc("inner", func(ctx context.Context, input MetricInput) *ScoreResult {
		return NewScoreResult("inner", 0.8)
	})

	weighted := NewWeightedMetric(innerMetric, 0.5)

	if weighted.Name() != "inner" {
		t.Errorf("Name() = %q, want %q", weighted.Name(), "inner")
	}

	if weighted.Weight() != 0.5 {
		t.Errorf("Weight() = %v, want 0.5", weighted.Weight())
	}

	t.Run("score is weighted", func(t *testing.T) {
		result := weighted.Score(ctx, NewMetricInput("", "test"))
		// 0.8 * 0.5 = 0.4
		if result.Value != 0.4 {
			t.Errorf("Score = %v, want 0.4", result.Value)
		}
	})

	t.Run("error is not weighted", func(t *testing.T) {
		errorMetric := NewMetricFunc("error", func(ctx context.Context, input MetricInput) *ScoreResult {
			return NewFailedScoreResult("error", errors.New("failed"))
		})
		errorWeighted := NewWeightedMetric(errorMetric, 0.5)
		result := errorWeighted.Score(ctx, NewMetricInput("", "test"))
		if result.Error == nil {
			t.Error("Error should be preserved")
		}
	})
}

func TestMetricInputChaining(t *testing.T) {
	input := NewMetricInput("query", "response").
		WithExpected("expected").
		WithContext("context").
		WithMetadata("key", "value")

	if input.Input != "query" {
		t.Errorf("Input = %q, want %q", input.Input, "query")
	}
	if input.Output != "response" {
		t.Errorf("Output = %q, want %q", input.Output, "response")
	}
	if input.Expected != "expected" {
		t.Errorf("Expected = %q, want %q", input.Expected, "expected")
	}
	if input.Context != "context" {
		t.Errorf("Context = %q, want %q", input.Context, "context")
	}
	if input.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want %q", input.Metadata["key"], "value")
	}
}
