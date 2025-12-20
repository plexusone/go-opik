package heuristic

import (
	"context"
	"testing"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

func TestIsJSON(t *testing.T) {
	ctx := context.Background()
	metric := NewIsJSON()

	if metric.Name() != "is_json" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "is_json")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"valid object", `{"key": "value"}`, 1.0},
		{"valid array", `[1, 2, 3]`, 1.0},
		{"valid string", `"hello"`, 1.0},
		{"valid number", `123.45`, 1.0},
		{"valid boolean", `true`, 1.0},
		{"valid null", `null`, 1.0},
		{"nested object", `{"outer": {"inner": "value"}}`, 1.0},
		{"invalid json", `{key: value}`, 0.0},
		{"plain text", `hello world`, 0.0},
		{"empty string", ``, 0.0},
		{"unclosed brace", `{"key": "value"`, 0.0},
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

func TestIsJSONObject(t *testing.T) {
	ctx := context.Background()
	metric := NewIsJSONObject()

	if metric.Name() != "is_json_object" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "is_json_object")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"valid object", `{"key": "value"}`, 1.0},
		{"empty object", `{}`, 1.0},
		{"nested object", `{"a": {"b": 1}}`, 1.0},
		{"array", `[1, 2, 3]`, 0.0},
		{"string", `"hello"`, 0.0},
		{"number", `123`, 0.0},
		{"invalid json", `not json`, 0.0},
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

func TestIsJSONArray(t *testing.T) {
	ctx := context.Background()
	metric := NewIsJSONArray()

	if metric.Name() != "is_json_array" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "is_json_array")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"valid array", `[1, 2, 3]`, 1.0},
		{"empty array", `[]`, 1.0},
		{"string array", `["a", "b"]`, 1.0},
		{"object array", `[{"a": 1}]`, 1.0},
		{"object", `{"key": "value"}`, 0.0},
		{"string", `"hello"`, 0.0},
		{"invalid json", `not json`, 0.0},
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

func TestJSONHasKeys(t *testing.T) {
	ctx := context.Background()

	t.Run("all keys present", func(t *testing.T) {
		metric := NewJSONHasKeys([]string{"name", "age"})
		input := evaluation.NewMetricInput("", `{"name": "John", "age": 30}`)
		result := metric.Score(ctx, input)
		if result.Value != 1.0 {
			t.Errorf("Score = %v, want 1.0", result.Value)
		}
	})

	t.Run("partial keys", func(t *testing.T) {
		metric := NewJSONHasKeys([]string{"name", "age", "email"})
		input := evaluation.NewMetricInput("", `{"name": "John", "age": 30}`)
		result := metric.Score(ctx, input)
		expected := 2.0 / 3.0
		if result.Value < expected-0.01 || result.Value > expected+0.01 {
			t.Errorf("Score = %v, want ~%v", result.Value, expected)
		}
	})

	t.Run("no keys present", func(t *testing.T) {
		metric := NewJSONHasKeys([]string{"x", "y"})
		input := evaluation.NewMetricInput("", `{"a": 1}`)
		result := metric.Score(ctx, input)
		if result.Value != 0.0 {
			t.Errorf("Score = %v, want 0.0", result.Value)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		metric := NewJSONHasKeys([]string{"key"})
		input := evaluation.NewMetricInput("", `not json`)
		result := metric.Score(ctx, input)
		if result.Value != 0.0 {
			t.Errorf("Score = %v, want 0.0", result.Value)
		}
	})

	if NewJSONHasKeys(nil).Name() != "json_has_keys" {
		t.Error("Name() should be json_has_keys")
	}
}

func TestJSONSchemaValid(t *testing.T) {
	ctx := context.Background()

	t.Run("valid schema", func(t *testing.T) {
		metric := NewJSONSchemaValid(map[string]string{
			"name": "string",
			"age":  "number",
		})
		input := evaluation.NewMetricInput("", `{"name": "John", "age": 30}`)
		result := metric.Score(ctx, input)
		if result.Value != 1.0 {
			t.Errorf("Score = %v, want 1.0", result.Value)
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		metric := NewJSONSchemaValid(map[string]string{
			"age": "string",
		})
		input := evaluation.NewMetricInput("", `{"age": 30}`)
		result := metric.Score(ctx, input)
		if result.Value != 0.0 {
			t.Errorf("Score = %v, want 0.0", result.Value)
		}
	})

	t.Run("missing field", func(t *testing.T) {
		metric := NewJSONSchemaValid(map[string]string{
			"name":  "string",
			"email": "string",
		})
		input := evaluation.NewMetricInput("", `{"name": "John"}`)
		result := metric.Score(ctx, input)
		if result.Value != 0.5 {
			t.Errorf("Score = %v, want 0.5", result.Value)
		}
	})

	t.Run("all types", func(t *testing.T) {
		metric := NewJSONSchemaValid(map[string]string{
			"s": "string",
			"n": "number",
			"b": "boolean",
			"a": "array",
			"o": "object",
		})
		input := evaluation.NewMetricInput("", `{"s": "str", "n": 1.5, "b": true, "a": [], "o": {}}`)
		result := metric.Score(ctx, input)
		if result.Value != 1.0 {
			t.Errorf("Score = %v, want 1.0", result.Value)
		}
	})

	if NewJSONSchemaValid(nil).Name() != "json_schema_valid" {
		t.Error("Name() should be json_schema_valid")
	}
}

func TestIsXML(t *testing.T) {
	ctx := context.Background()
	metric := NewIsXML()

	if metric.Name() != "is_xml" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "is_xml")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"valid xml", `<root><child>text</child></root>`, 1.0},
		{"self-closing", `<item/>`, 1.0},
		{"with attributes", `<div class="test">content</div>`, 1.0},
		{"invalid xml", `<unclosed>`, 0.0},
		// Note: Go's XML decoder treats plain text as valid character data
		{"plain text", `not xml`, 1.0},
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

func TestExtractJSON(t *testing.T) {
	ctx := context.Background()

	// Use IsJSONObject as inner metric
	inner := NewIsJSONObject()
	metric := NewExtractJSON(inner)

	if metric.Name() != "extract_json" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "extract_json")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"raw json", `{"key": "value"}`, 1.0},
		{"markdown code block", "```json\n{\"key\": \"value\"}\n```", 1.0},
		{"code block no lang", "```\n{\"key\": \"value\"}\n```", 1.0},
		{"embedded json", `Here is the result: {"data": true}`, 1.0},
		{"no json", `no json here`, 0.0},
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

func TestIsNumber(t *testing.T) {
	ctx := context.Background()
	metric := NewIsNumber()

	if metric.Name() != "is_number" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "is_number")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"integer", `42`, 1.0},
		{"float", `3.14`, 1.0},
		{"negative", `-100`, 1.0},
		{"scientific", `1.5e10`, 1.0},
		{"zero", `0`, 1.0},
		{"string number", `"42"`, 0.0},
		{"text", `hello`, 0.0},
		{"empty", ``, 0.0},
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

func TestIsBoolean(t *testing.T) {
	ctx := context.Background()
	metric := NewIsBoolean()

	if metric.Name() != "is_boolean" {
		t.Errorf("Name() = %q, want %q", metric.Name(), "is_boolean")
	}

	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{"true", `true`, 1.0},
		{"false", `false`, 1.0},
		{"yes", `yes`, 1.0},
		{"no", `no`, 1.0},
		{"1", `1`, 1.0},
		{"0", `0`, 1.0},
		{"TRUE", `TRUE`, 1.0},
		{"FALSE", `FALSE`, 1.0},
		{"with spaces", `  true  `, 1.0},
		{"maybe", `maybe`, 0.0},
		{"number", `123`, 0.0},
		{"empty", ``, 0.0},
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

func TestGetJSONType(t *testing.T) {
	tests := []struct {
		value any
		want  string
	}{
		{"hello", "string"},
		{float64(123), "number"},
		{true, "boolean"},
		{[]any{1, 2}, "array"},
		{map[string]any{"a": 1}, "object"},
		{nil, "null"},
	}

	for _, tt := range tests {
		got := getJSONType(tt.value)
		if got != tt.want {
			t.Errorf("getJSONType(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}
