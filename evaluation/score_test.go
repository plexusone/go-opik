package evaluation

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
)

// floatEquals compares two floats with tolerance for floating-point precision
func floatEquals(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestScoreResult(t *testing.T) {
	t.Run("successful score", func(t *testing.T) {
		result := NewScoreResult("test", 0.85)

		if result.Name != "test" {
			t.Errorf("Name = %q, want %q", result.Name, "test")
		}
		if result.Value != 0.85 {
			t.Errorf("Value = %v, want 0.85", result.Value)
		}
		if !result.IsSuccess() {
			t.Error("IsSuccess() should be true")
		}
	})

	t.Run("score with reason", func(t *testing.T) {
		result := NewScoreResultWithReason("test", 0.5, "partial match")

		if result.Reason != "partial match" {
			t.Errorf("Reason = %q, want %q", result.Reason, "partial match")
		}
	})

	t.Run("failed score", func(t *testing.T) {
		err := errors.New("evaluation failed")
		result := NewFailedScoreResult("test", err)

		if result.IsSuccess() {
			t.Error("IsSuccess() should be false for failed score")
		}
		if result.Error != err {
			t.Errorf("Error = %v, want %v", result.Error, err)
		}
	})
}

func TestScoreResultString(t *testing.T) {
	tests := []struct {
		name   string
		result *ScoreResult
		want   string
	}{
		{
			"simple",
			NewScoreResult("test", 0.85),
			"test: 0.8500",
		},
		{
			"with reason",
			NewScoreResultWithReason("test", 0.5, "partial"),
			"test: 0.5000 (partial)",
		},
		{
			"with error",
			NewFailedScoreResult("test", errors.New("failed")),
			"test: error - failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScoreResultToJSON(t *testing.T) {
	result := NewScoreResultWithReason("test_metric", 0.75, "good match")

	jsonBytes, err := result.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if decoded["name"] != "test_metric" {
		t.Errorf("JSON name = %v, want %q", decoded["name"], "test_metric")
	}
	if decoded["value"] != 0.75 {
		t.Errorf("JSON value = %v, want 0.75", decoded["value"])
	}
	if decoded["reason"] != "good match" {
		t.Errorf("JSON reason = %v, want %q", decoded["reason"], "good match")
	}
}

func TestBooleanScore(t *testing.T) {
	t.Run("true value", func(t *testing.T) {
		result := BooleanScore("check", true)
		if result.Value != 1.0 {
			t.Errorf("Value = %v, want 1.0", result.Value)
		}
	})

	t.Run("false value", func(t *testing.T) {
		result := BooleanScore("check", false)
		if result.Value != 0.0 {
			t.Errorf("Value = %v, want 0.0", result.Value)
		}
	})
}

func TestBooleanScoreWithReason(t *testing.T) {
	t.Run("true with reason", func(t *testing.T) {
		result := BooleanScoreWithReason("check", true, "passed")
		if result.Value != 1.0 {
			t.Errorf("Value = %v, want 1.0", result.Value)
		}
		if result.Reason != "passed" {
			t.Errorf("Reason = %q, want %q", result.Reason, "passed")
		}
	})

	t.Run("false with reason", func(t *testing.T) {
		result := BooleanScoreWithReason("check", false, "failed")
		if result.Value != 0.0 {
			t.Errorf("Value = %v, want 0.0", result.Value)
		}
		if result.Reason != "failed" {
			t.Errorf("Reason = %q, want %q", result.Reason, "failed")
		}
	})
}

func TestScoreResults(t *testing.T) {
	results := ScoreResults{
		NewScoreResult("m1", 0.8),
		NewScoreResult("m2", 0.6),
		NewScoreResult("m1", 0.9),
		NewFailedScoreResult("m3", errors.New("error")),
	}

	t.Run("ByName", func(t *testing.T) {
		result := results.ByName("m1")
		if result == nil {
			t.Fatal("ByName should return first match")
		}
		if result.Value != 0.8 {
			t.Errorf("Value = %v, want 0.8", result.Value)
		}

		missing := results.ByName("missing")
		if missing != nil {
			t.Error("ByName should return nil for missing name")
		}
	})

	t.Run("AllByName", func(t *testing.T) {
		all := results.AllByName("m1")
		if len(all) != 2 {
			t.Errorf("AllByName length = %d, want 2", len(all))
		}
	})

	t.Run("Successful", func(t *testing.T) {
		successful := results.Successful()
		if len(successful) != 3 {
			t.Errorf("Successful length = %d, want 3", len(successful))
		}
		for _, s := range successful {
			if !s.IsSuccess() {
				t.Error("Successful should only return successful scores")
			}
		}
	})

	t.Run("Failed", func(t *testing.T) {
		failed := results.Failed()
		if len(failed) != 1 {
			t.Errorf("Failed length = %d, want 1", len(failed))
		}
		if failed[0].Name != "m3" {
			t.Errorf("Failed[0].Name = %q, want %q", failed[0].Name, "m3")
		}
	})
}

func TestScoreResultsAverage(t *testing.T) {
	t.Run("normal average", func(t *testing.T) {
		results := ScoreResults{
			NewScoreResult("m1", 0.8),
			NewScoreResult("m2", 0.6),
			NewScoreResult("m3", 0.7),
		}

		avg := results.Average()
		expected := (0.8 + 0.6 + 0.7) / 3.0
		if !floatEquals(avg, expected, 0.0001) {
			t.Errorf("Average() = %v, want %v", avg, expected)
		}
	})

	t.Run("with failed scores excluded", func(t *testing.T) {
		results := ScoreResults{
			NewScoreResult("m1", 0.8),
			NewFailedScoreResult("m2", errors.New("error")),
			NewScoreResult("m3", 0.6),
		}

		avg := results.Average()
		expected := (0.8 + 0.6) / 2.0
		if avg != expected {
			t.Errorf("Average() = %v, want %v", avg, expected)
		}
	})

	t.Run("empty results", func(t *testing.T) {
		results := ScoreResults{}
		if avg := results.Average(); avg != 0 {
			t.Errorf("Average() = %v, want 0", avg)
		}
	})

	t.Run("all failed", func(t *testing.T) {
		results := ScoreResults{
			NewFailedScoreResult("m1", errors.New("e1")),
			NewFailedScoreResult("m2", errors.New("e2")),
		}
		if avg := results.Average(); avg != 0 {
			t.Errorf("Average() = %v, want 0", avg)
		}
	})
}

func TestScoreResultsAverageByName(t *testing.T) {
	results := ScoreResults{
		NewScoreResult("accuracy", 0.9),
		NewScoreResult("accuracy", 0.8),
		NewScoreResult("precision", 0.7),
		NewFailedScoreResult("accuracy", errors.New("error")),
	}

	t.Run("existing name", func(t *testing.T) {
		avg := results.AverageByName("accuracy")
		expected := (0.9 + 0.8) / 2.0
		if !floatEquals(avg, expected, 0.0001) {
			t.Errorf("AverageByName(accuracy) = %v, want %v", avg, expected)
		}
	})

	t.Run("single entry", func(t *testing.T) {
		avg := results.AverageByName("precision")
		if avg != 0.7 {
			t.Errorf("AverageByName(precision) = %v, want 0.7", avg)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		avg := results.AverageByName("missing")
		if avg != 0 {
			t.Errorf("AverageByName(missing) = %v, want 0", avg)
		}
	})
}

func TestScoreResultMetadata(t *testing.T) {
	result := &ScoreResult{
		Name:  "test",
		Value: 0.5,
		Metadata: map[string]any{
			"tokens": 100,
			"model":  "gpt-4",
		},
	}

	if result.Metadata["tokens"] != 100 {
		t.Errorf("Metadata[tokens] = %v, want 100", result.Metadata["tokens"])
	}
	if result.Metadata["model"] != "gpt-4" {
		t.Errorf("Metadata[model] = %v, want %q", result.Metadata["model"], "gpt-4")
	}
}

func TestScoreResultEdgeCases(t *testing.T) {
	t.Run("zero score", func(t *testing.T) {
		result := NewScoreResult("test", 0.0)
		if result.Value != 0.0 {
			t.Errorf("Value = %v, want 0.0", result.Value)
		}
		if !result.IsSuccess() {
			t.Error("Zero score should still be successful")
		}
	})

	t.Run("negative score", func(t *testing.T) {
		result := NewScoreResult("test", -0.5)
		if result.Value != -0.5 {
			t.Errorf("Value = %v, want -0.5", result.Value)
		}
	})

	t.Run("score greater than 1", func(t *testing.T) {
		result := NewScoreResult("test", 1.5)
		if result.Value != 1.5 {
			t.Errorf("Value = %v, want 1.5", result.Value)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		result := NewScoreResult("", 0.5)
		if result.Name != "" {
			t.Errorf("Name = %q, want empty", result.Name)
		}
	})
}
