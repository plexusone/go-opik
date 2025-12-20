package heuristic

import (
	"context"
	"testing"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

func TestEquals(t *testing.T) {
	ctx := context.Background()
	metric := NewEquals(true)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"exact match", "hello", "hello", 1.0},
		{"no match", "hello", "world", 0.0},
		{"case mismatch", "Hello", "hello", 0.0},
		{"empty strings", "", "", 1.0},
		{"whitespace difference", "hello ", "hello", 0.0},
		{"unicode match", "日本語", "日本語", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Equals() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestEqualsIgnoreCase(t *testing.T) {
	ctx := context.Background()
	metric := NewEquals(false)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"exact match", "hello", "hello", 1.0},
		{"case insensitive match", "Hello", "hello", 1.0},
		{"mixed case match", "HeLLo WoRLD", "hello world", 1.0},
		{"no match", "hello", "world", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("EqualsIgnoreCase() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	ctx := context.Background()
	metric := NewContains(true)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"contains substring", "hello world", "world", 1.0},
		{"exact match", "world", "world", 1.0},
		{"does not contain", "hello", "world", 0.0},
		{"case sensitive miss", "World", "world", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Contains() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	ctx := context.Background()
	metric := NewContains(false)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"contains lowercase", "hello world", "WORLD", 1.0},
		{"contains mixed case", "Hello World", "WORLD", 1.0},
		{"does not contain", "hello", "world", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("ContainsIgnoreCase() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestStartsWith(t *testing.T) {
	ctx := context.Background()
	metric := NewStartsWith(true)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"starts with prefix", "hello world", "hello", 1.0},
		{"exact prefix", "hello", "hello", 1.0},
		{"does not start with", "world hello", "hello", 0.0},
		{"case mismatch", "Hello world", "hello", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("StartsWith() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestEndsWith(t *testing.T) {
	ctx := context.Background()
	metric := NewEndsWith(true)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"ends with suffix", "hello world", "world", 1.0},
		{"exact suffix", "world", "world", 1.0},
		{"does not end with", "world hello", "world", 0.0},
		{"case mismatch", "hello World", "world", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("EndsWith() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	ctx := context.Background()
	values := []string{"apple", "banana", "cherry"}
	metric := NewContainsAny(values, true)

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"contains first", "I have an apple", 1.0},
		{"contains second", "banana split", 1.0},
		{"contains none", "I have a grape", 0.0},
		{"empty string", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("ContainsAny() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestContainsAll(t *testing.T) {
	ctx := context.Background()
	values := []string{"hello", "world"}
	metric := NewContainsAll(values, true)

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"contains all", "hello world", 1.0},
		{"contains all reversed", "world hello", 1.0},
		{"missing one", "hello there", 0.5}, // Partial score: 1/2 found
		{"missing all", "goodbye", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("ContainsAll() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestNotEmpty(t *testing.T) {
	ctx := context.Background()
	metric := NewNotEmpty()

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"non-empty", "hello", 1.0},
		{"empty", "", 0.0},
		{"whitespace only", "   ", 0.0},
		{"tabs and newlines", "\t\n", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("NotEmpty() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestLengthBetween(t *testing.T) {
	ctx := context.Background()
	metric := NewLengthBetween(5, 10)

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"within range", "hello", 1.0},
		{"at min", "12345", 1.0},
		{"at max", "1234567890", 1.0},
		{"too short", "hi", 0.0},
		{"too long", "hello world!", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("LengthBetween() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestWordCount(t *testing.T) {
	ctx := context.Background()
	metric := NewWordCount(2, 5)

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"within range", "hello world foo", 1.0},
		{"at min", "hello world", 1.0},
		{"at max", "one two three four five", 1.0},
		{"too few", "hello", 0.0},
		{"too many", "one two three four five six", 0.0},
		{"empty", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("WordCount() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestNoOffensiveLanguage(t *testing.T) {
	ctx := context.Background()
	metric := NewNoOffensiveLanguage(nil)

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"clean text", "hello world", 1.0},
		{"technical text", "This is a function call", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("NoOffensiveLanguage() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestMetricName(t *testing.T) {
	tests := []struct {
		name   string
		metric evaluation.Metric
		want   string
	}{
		{"equals", NewEquals(true), "equals"},
		{"contains", NewContains(true), "contains"},
		{"starts_with", NewStartsWith(true), "starts_with"},
		{"ends_with", NewEndsWith(true), "ends_with"},
		{"contains_any", NewContainsAny([]string{"x"}, true), "contains_any"},
		{"contains_all", NewContainsAll([]string{"x"}, true), "contains_all"},
		{"not_empty", NewNotEmpty(), "not_empty"},
		{"length_between", NewLengthBetween(1, 10), "length_between"},
		{"word_count", NewWordCount(1, 10), "word_count"},
		{"no_offensive_language", NewNoOffensiveLanguage(nil), "no_offensive_language"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metric.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
