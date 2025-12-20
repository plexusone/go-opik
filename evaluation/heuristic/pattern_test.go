package heuristic

import (
	"context"
	"testing"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

func TestRegexMatch(t *testing.T) {
	ctx := context.Background()

	t.Run("valid pattern", func(t *testing.T) {
		metric, err := NewRegexMatch(`\d{3}-\d{4}`)
		if err != nil {
			t.Fatalf("NewRegexMatch error = %v", err)
		}

		if metric.Name() != "regex_match" {
			t.Errorf("Name() = %q, want %q", metric.Name(), "regex_match")
		}

		tests := []struct {
			name   string
			output string
			want   float64
		}{
			{"matches", "Call 123-4567", 1.0},
			{"exact match", "123-4567", 1.0},
			{"no match", "hello", 0.0},
			{"partial", "123-456", 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				input := evaluation.NewMetricInput("", tt.output)
				result := metric.Score(ctx, input)
				if result.Value != tt.want {
					t.Errorf("Score = %v, want %v", result.Value, tt.want)
				}
			})
		}
	})

	t.Run("invalid pattern", func(t *testing.T) {
		_, err := NewRegexMatch(`[invalid`)
		if err == nil {
			t.Error("Expected error for invalid pattern")
		}
	})
}

func TestMustRegexMatch(t *testing.T) {
	t.Run("valid pattern", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustRegexMatch panicked: %v", r)
			}
		}()
		metric := MustRegexMatch(`\w+`)
		if metric == nil {
			t.Error("MustRegexMatch returned nil")
		}
	})

	t.Run("invalid pattern panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustRegexMatch should panic for invalid pattern")
			}
		}()
		MustRegexMatch(`[invalid`)
	})
}

func TestRegexNotMatch(t *testing.T) {
	ctx := context.Background()

	metric, err := NewRegexNotMatch(`password|secret`)
	if err != nil {
		t.Fatalf("NewRegexNotMatch error = %v", err)
	}

	if metric.Name() != "regex_not_match" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "regex_not_match")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"no match good", "safe text", 1.0},
		{"contains password", "my password is 123", 0.0},
		{"contains secret", "a secret value", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Score = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestMustRegexNotMatch(t *testing.T) {
	t.Run("valid pattern", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustRegexNotMatch panicked: %v", r)
			}
		}()
		metric := MustRegexNotMatch(`\w+`)
		if metric == nil {
			t.Error("MustRegexNotMatch returned nil")
		}
	})

	t.Run("invalid pattern panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustRegexNotMatch should panic for invalid pattern")
			}
		}()
		MustRegexNotMatch(`[invalid`)
	})
}

func TestRegexFindAll(t *testing.T) {
	ctx := context.Background()

	t.Run("count in range", func(t *testing.T) {
		metric, err := NewRegexFindAll(`\d+`, 2, 4)
		if err != nil {
			t.Fatalf("NewRegexFindAll error = %v", err)
		}

		if metric.Name() != "regex_find_all" {
			t.Errorf("Name() = %q, want %q", metric.Name(), "regex_find_all")
		}

		tests := []struct {
			name   string
			output string
			want   float64
		}{
			{"in range", "1 2 3", 1.0},
			{"at min", "1 2", 1.0},
			{"at max", "1 2 3 4", 1.0},
			{"below min", "1", 0.0},
			{"above max", "1 2 3 4 5", 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				input := evaluation.NewMetricInput("", tt.output)
				result := metric.Score(ctx, input)
				if result.Value != tt.want {
					t.Errorf("Score = %v, want %v", result.Value, tt.want)
				}
			})
		}
	})

	t.Run("no max limit", func(t *testing.T) {
		metric, _ := NewRegexFindAll(`\d+`, 2, 0) // max=0 means no max
		input := evaluation.NewMetricInput("", "1 2 3 4 5 6 7 8 9 10")
		result := metric.Score(ctx, input)
		if result.Value != 1.0 {
			t.Errorf("Score = %v, want 1.0", result.Value)
		}
	})
}

func TestEmailFormat(t *testing.T) {
	ctx := context.Background()
	metric := NewEmailFormat()

	if metric.Name() != "email_format" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "email_format")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"valid simple", "user@example.com", 1.0},
		{"valid with subdomain", "user@mail.example.com", 1.0},
		{"valid with plus", "user+tag@example.com", 1.0},
		{"valid with dots", "first.last@example.com", 1.0},
		{"missing @", "userexample.com", 0.0},
		{"missing domain", "user@", 0.0},
		{"missing tld", "user@example", 0.0},
		{"spaces", "user @example.com", 0.0},
		{"empty", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Score = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestURLFormat(t *testing.T) {
	ctx := context.Background()
	metric := NewURLFormat()

	if metric.Name() != "url_format" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "url_format")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"http", "http://example.com", 1.0},
		{"https", "https://example.com", 1.0},
		{"with path", "https://example.com/path", 1.0},
		{"with query", "https://example.com?q=1", 1.0},
		{"with port", "https://example.com:8080/path", 1.0},
		{"no scheme", "example.com", 0.0},
		{"ftp", "ftp://example.com", 0.0},
		{"empty", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Score = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestPhoneFormat(t *testing.T) {
	ctx := context.Background()
	metric := NewPhoneFormat()

	if metric.Name() != "phone_format" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "phone_format")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"us format", "123-456-7890", 1.0},
		{"with country code", "+1-234-567-8901", 1.0},
		{"parentheses", "(234) 567-8901", 1.0},
		{"dots", "234.567.8901", 1.0},
		{"spaces", "234 567 8901", 1.0},
		{"too short", "123", 0.0},
		{"letters", "abc-def-ghij", 0.0},
		{"empty", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Score = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestDateFormat(t *testing.T) {
	ctx := context.Background()
	metric := NewDateFormat()

	if metric.Name() != "date_format" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "date_format")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"iso date", "2024-01-15", 1.0},
		{"iso datetime", "2024-01-15T10:30", 1.0},
		{"iso with seconds", "2024-01-15T10:30:00", 1.0},
		{"slash format", "2024/01/15", 1.0},
		{"us format", "01-15-2024", 0.0},
		{"text", "January 15, 2024", 0.0},
		{"empty", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Score = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestDateFormatWithPattern(t *testing.T) {
	ctx := context.Background()

	t.Run("custom pattern", func(t *testing.T) {
		// US date format: MM/DD/YYYY
		metric, err := NewDateFormatWithPattern(`^\d{2}/\d{2}/\d{4}$`)
		if err != nil {
			t.Fatalf("NewDateFormatWithPattern error = %v", err)
		}

		input := evaluation.NewMetricInput("", "01/15/2024")
		result := metric.Score(ctx, input)
		if result.Value != 1.0 {
			t.Errorf("Score = %v, want 1.0", result.Value)
		}
	})

	t.Run("invalid pattern", func(t *testing.T) {
		_, err := NewDateFormatWithPattern(`[invalid`)
		if err == nil {
			t.Error("Expected error for invalid pattern")
		}
	})
}

func TestUUIDFormat(t *testing.T) {
	ctx := context.Background()
	metric := NewUUIDFormat()

	if metric.Name() != "uuid_format" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "uuid_format")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", 1.0},
		{"uppercase", "550E8400-E29B-41D4-A716-446655440000", 1.0},
		{"mixed case", "550e8400-E29B-41d4-A716-446655440000", 1.0},
		{"with spaces", " 550e8400-e29b-41d4-a716-446655440000 ", 1.0},
		{"missing dashes", "550e8400e29b41d4a716446655440000", 0.0},
		{"too short", "550e8400-e29b-41d4-a716", 0.0},
		{"invalid chars", "550g8400-e29b-41d4-a716-446655440000", 0.0},
		{"empty", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := evaluation.NewMetricInput("", tt.output)
			result := metric.Score(ctx, input)
			if result.Value != tt.want {
				t.Errorf("Score = %v, want %v", result.Value, tt.want)
			}
		})
	}
}

func TestPatternMetricNames(t *testing.T) {
	tests := []struct {
		metric evaluation.Metric
		want   string
	}{
		{MustRegexMatch(`\w+`), "regex_match"},
		{MustRegexNotMatch(`\w+`), "regex_not_match"},
		{NewEmailFormat(), "email_format"},
		{NewURLFormat(), "url_format"},
		{NewPhoneFormat(), "phone_format"},
		{NewDateFormat(), "date_format"},
		{NewUUIDFormat(), "uuid_format"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.metric.Name(); got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}
