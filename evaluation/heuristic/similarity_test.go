package heuristic

import (
	"context"
	"math"
	"testing"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

const tolerance = 0.001

func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

func TestLevenshteinSimilarity(t *testing.T) {
	ctx := context.Background()
	metric := NewLevenshteinSimilarity(true)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"identical strings", "hello", "hello", 1.0},
		{"completely different", "abc", "xyz", 0.0},
		{"one char different", "hello", "hallo", 0.8},
		{"empty strings", "", "", 1.0},
		{"one empty", "hello", "", 0.0},
		{"case difference", "Hello", "hello", 0.8},
		{"substring", "hello", "hello world", 0.454},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if !approxEqual(result.Value, tt.want, tolerance) {
				t.Errorf("LevenshteinSimilarity() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"hello", "", 5},
		{"", "hello", 5},
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := levenshteinDistance(tt.a, tt.b); got != tt.want {
				t.Errorf("levenshteinDistance(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestJaccardSimilarity(t *testing.T) {
	ctx := context.Background()
	metric := NewJaccardSimilarity(true, true)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"identical", "hello world", "hello world", 1.0},
		{"no overlap", "hello", "world", 0.0},
		{"partial overlap", "hello world", "world peace", 0.333},
		{"subset", "a b", "a b c", 0.666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if !approxEqual(result.Value, tt.want, tolerance) {
				t.Errorf("JaccardSimilarity() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestCosineSimilarity(t *testing.T) {
	ctx := context.Background()
	metric := NewCosineSimilarity(true)

	tests := []struct {
		name     string
		output   string
		expected string
		want     float64
	}{
		{"identical", "hello world", "hello world", 1.0},
		{"no overlap", "aaa bbb", "ccc ddd", 0.0},
		{"partial overlap", "hello world", "hello there", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if !approxEqual(result.Value, tt.want, tolerance) {
				t.Errorf("CosineSimilarity() = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestBLEU(t *testing.T) {
	ctx := context.Background()
	metric := NewBLEU(4)

	tests := []struct {
		name     string
		output   string
		expected string
		wantMin  float64
		wantMax  float64
	}{
		{"identical", "the cat sat on the mat", "the cat sat on the mat", 0.9, 1.0},
		{"similar", "the cat sat on the mat", "the cat is on the mat", 0.1, 0.5},
		{"very different", "hello world", "goodbye universe", 0.0, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value < tt.wantMin || result.Value > tt.wantMax {
				t.Errorf("BLEU() = %v, want between %v and %v", result.Value, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestROUGE(t *testing.T) {
	ctx := context.Background()

	t.Run("ROUGE default", func(t *testing.T) {
		metric := NewROUGE(1.0)
		input := evaluation.NewMetricInput("", "the cat sat on the mat").WithExpected("the cat sat on the mat")
		result := metric.Score(ctx, input)
		if result.Value < 0.9 {
			t.Errorf("ROUGE for identical = %v, want >= 0.9", result.Value)
		}
	})

	t.Run("ROUGE different beta", func(t *testing.T) {
		metric := NewROUGE(0.5)
		input := evaluation.NewMetricInput("", "the quick brown fox").WithExpected("the quick brown dog")
		result := metric.Score(ctx, input)
		if result.Value < 0.3 || result.Value > 0.9 {
			t.Errorf("ROUGE = %v, want between 0.3 and 0.9", result.Value)
		}
	})
}

func TestFuzzyMatch(t *testing.T) {
	ctx := context.Background()
	metric := NewFuzzyMatch(0.8, true)

	tests := []struct {
		name     string
		output   string
		expected string
		wantMin  float64
		wantMax  float64
	}{
		{"identical", "hello world", "hello world", 0.99, 1.01},
		{"similar", "hello wrold", "hello world", 0.5, 0.95},
		{"different", "goodbye", "hello", 0.0, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value < tt.wantMin || result.Value > tt.wantMax {
				t.Errorf("FuzzyMatch() = %v, want between %v and %v", result.Value, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFuzzyMatchThreshold(t *testing.T) {
	ctx := context.Background()

	// FuzzyMatch returns the actual similarity score (average of levenshtein and jaccard)
	// The threshold only affects the reason message, not the score itself
	highThreshold := NewFuzzyMatch(0.95, true)
	lowThreshold := NewFuzzyMatch(0.5, true)

	output := "hello wrold"
	expected := "hello world"

	highInput := evaluation.NewMetricInput("", output).WithExpected(expected)
	lowInput := evaluation.NewMetricInput("", output).WithExpected(expected)

	highResult := highThreshold.Score(ctx, highInput)
	lowResult := lowThreshold.Score(ctx, lowInput)

	// Both should return the same similarity score
	if highResult.Value != lowResult.Value {
		t.Errorf("Same input should have same score: high=%v, low=%v", highResult.Value, lowResult.Value)
	}

	// Score should be between 0.5 and 0.9 for a typo
	if highResult.Value < 0.5 || highResult.Value > 0.9 {
		t.Errorf("FuzzyMatch score for typo = %v, want between 0.5 and 0.9", highResult.Value)
	}
}

func TestSemanticSimilarity(t *testing.T) {
	ctx := context.Background()
	metric := NewSemanticSimilarity()

	tests := []struct {
		name     string
		output   string
		expected string
		wantMin  float64
		wantMax  float64
	}{
		{"identical", "hello world", "hello world", 0.9, 1.0},
		{"similar meaning", "the dog runs fast", "the dog runs quickly", 0.5, 1.0},
		{"different", "hello", "goodbye", 0.0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output).WithExpected(tt.expected)
			result := metric.Score(ctx, input)
			if result.Value < tt.wantMin || result.Value > tt.wantMax {
				t.Errorf("SemanticSimilarity() = %v, want between %v and %v", result.Value, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestSimilarityMetricNames(t *testing.T) {
	tests := []struct {
		name   string
		metric evaluation.Metric
		want   string
	}{
		{"levenshtein", NewLevenshteinSimilarity(true), "levenshtein_similarity"},
		{"jaccard", NewJaccardSimilarity(true, true), "jaccard_similarity"},
		{"cosine", NewCosineSimilarity(true), "cosine_similarity"},
		{"bleu", NewBLEU(4), "bleu"},
		{"rouge", NewROUGE(1.0), "rouge_l"},
		{"fuzzy", NewFuzzyMatch(0.8, true), "fuzzy_match"},
		{"semantic", NewSemanticSimilarity(), "semantic_similarity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metric.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLCSLength(t *testing.T) {
	tests := []struct {
		a, b []string
		want int
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}, 3},
		{[]string{"a", "b", "c"}, []string{"a", "c"}, 2},
		{[]string{"a", "b", "c"}, []string{"d", "e", "f"}, 0},
		{[]string{}, []string{"a", "b"}, 0},
		{[]string{"a", "b"}, []string{}, 0},
	}

	for _, tt := range tests {
		got := lcsLength(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("lcsLength(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
