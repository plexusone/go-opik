package heuristic

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

// IsJSON checks if the output is valid JSON.
type IsJSON struct {
	evaluation.BaseMetric
}

// NewIsJSON creates a new IsJSON metric.
func NewIsJSON() *IsJSON {
	return &IsJSON{
		BaseMetric: evaluation.NewBaseMetric("is_json"),
	}
}

// Score evaluates if output is valid JSON.
func (m *IsJSON) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(input.Output), &js); err != nil {
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "invalid JSON: "+err.Error())
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid JSON")
}

// IsJSONObject checks if the output is a valid JSON object.
type IsJSONObject struct {
	evaluation.BaseMetric
}

// NewIsJSONObject creates a new IsJSONObject metric.
func NewIsJSONObject() *IsJSONObject {
	return &IsJSONObject{
		BaseMetric: evaluation.NewBaseMetric("is_json_object"),
	}
}

// Score evaluates if output is a valid JSON object.
func (m *IsJSONObject) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	var obj map[string]any
	if err := json.Unmarshal([]byte(input.Output), &obj); err != nil {
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "not a valid JSON object")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid JSON object")
}

// IsJSONArray checks if the output is a valid JSON array.
type IsJSONArray struct {
	evaluation.BaseMetric
}

// NewIsJSONArray creates a new IsJSONArray metric.
func NewIsJSONArray() *IsJSONArray {
	return &IsJSONArray{
		BaseMetric: evaluation.NewBaseMetric("is_json_array"),
	}
}

// Score evaluates if output is a valid JSON array.
func (m *IsJSONArray) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	var arr []any
	if err := json.Unmarshal([]byte(input.Output), &arr); err != nil {
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "not a valid JSON array")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid JSON array")
}

// JSONHasKeys checks if the JSON output has the specified keys.
type JSONHasKeys struct {
	evaluation.BaseMetric
	keys []string
}

// NewJSONHasKeys creates a new JSONHasKeys metric.
func NewJSONHasKeys(keys []string) *JSONHasKeys {
	return &JSONHasKeys{
		BaseMetric: evaluation.NewBaseMetric("json_has_keys"),
		keys:       keys,
	}
}

// Score evaluates if JSON output has all required keys.
func (m *JSONHasKeys) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	var obj map[string]any
	if err := json.Unmarshal([]byte(input.Output), &obj); err != nil {
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "not a valid JSON object")
	}

	missing := []string{}
	for _, key := range m.keys {
		if _, ok := obj[key]; !ok {
			missing = append(missing, key)
		}
	}

	if len(missing) == 0 {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "has all required keys")
	}

	found := len(m.keys) - len(missing)
	score := float64(found) / float64(len(m.keys))
	return evaluation.NewScoreResultWithReason(m.Name(), score,
		"missing keys: "+strings.Join(missing, ", "))
}

// JSONSchemaValid checks if the JSON output matches a simple schema.
type JSONSchemaValid struct {
	evaluation.BaseMetric
	required map[string]string // key -> expected type (string, number, boolean, array, object)
}

// NewJSONSchemaValid creates a new JSONSchemaValid metric.
func NewJSONSchemaValid(required map[string]string) *JSONSchemaValid {
	return &JSONSchemaValid{
		BaseMetric: evaluation.NewBaseMetric("json_schema_valid"),
		required:   required,
	}
}

// Score evaluates if JSON output matches the schema.
func (m *JSONSchemaValid) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	var obj map[string]any
	if err := json.Unmarshal([]byte(input.Output), &obj); err != nil {
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "not a valid JSON object")
	}

	errors := []string{}
	for key, expectedType := range m.required {
		val, ok := obj[key]
		if !ok {
			errors = append(errors, key+": missing")
			continue
		}

		actualType := getJSONType(val)
		if actualType != expectedType {
			errors = append(errors, key+": expected "+expectedType+", got "+actualType)
		}
	}

	if len(errors) == 0 {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid schema")
	}

	valid := len(m.required) - len(errors)
	score := float64(valid) / float64(len(m.required))
	return evaluation.NewScoreResultWithReason(m.Name(), score, strings.Join(errors, "; "))
}

func getJSONType(v any) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// IsXML checks if the output is valid XML.
type IsXML struct {
	evaluation.BaseMetric
}

// NewIsXML creates a new IsXML metric.
func NewIsXML() *IsXML {
	return &IsXML{
		BaseMetric: evaluation.NewBaseMetric("is_xml"),
	}
}

// Score evaluates if output is valid XML.
func (m *IsXML) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	decoder := xml.NewDecoder(strings.NewReader(input.Output))
	for {
		_, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "invalid XML: "+err.Error())
		}
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid XML")
}

// ExtractJSON extracts JSON from markdown code blocks or raw text.
type ExtractJSON struct {
	evaluation.BaseMetric
	inner evaluation.Metric
}

// NewExtractJSON creates a new ExtractJSON metric that extracts JSON before evaluation.
func NewExtractJSON(inner evaluation.Metric) *ExtractJSON {
	return &ExtractJSON{
		BaseMetric: evaluation.NewBaseMetric("extract_json"),
		inner:      inner,
	}
}

// Score extracts JSON and passes to inner metric.
func (m *ExtractJSON) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	extracted := extractJSONFromText(input.Output)
	if extracted == "" {
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "no JSON found in output")
	}

	newInput := input
	newInput.Output = extracted
	return m.inner.Score(ctx, newInput)
}

func extractJSONFromText(text string) string {
	// Try to extract from markdown code blocks
	codeBlockPattern := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	if matches := codeBlockPattern.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find raw JSON object
	text = strings.TrimSpace(text)
	if (strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}")) ||
		(strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]")) {
		return text
	}

	// Try to find JSON in the middle of text
	objectPattern := regexp.MustCompile(`\{[^{}]*\}`)
	if match := objectPattern.FindString(text); match != "" {
		return match
	}

	arrayPattern := regexp.MustCompile(`\[[^\[\]]*\]`)
	if match := arrayPattern.FindString(text); match != "" {
		return match
	}

	return ""
}

// IsNumber checks if the output is a valid number.
type IsNumber struct {
	evaluation.BaseMetric
}

// NewIsNumber creates a new IsNumber metric.
func NewIsNumber() *IsNumber {
	return &IsNumber{
		BaseMetric: evaluation.NewBaseMetric("is_number"),
	}
}

// Score evaluates if output is a valid number.
func (m *IsNumber) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	var num float64
	if err := json.Unmarshal([]byte(input.Output), &num); err != nil {
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "not a valid number")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid number")
}

// IsBoolean checks if the output is a valid boolean (true/false/yes/no).
type IsBoolean struct {
	evaluation.BaseMetric
}

// NewIsBoolean creates a new IsBoolean metric.
func NewIsBoolean() *IsBoolean {
	return &IsBoolean{
		BaseMetric: evaluation.NewBaseMetric("is_boolean"),
	}
}

// Score evaluates if output is a valid boolean.
func (m *IsBoolean) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	lower := strings.ToLower(strings.TrimSpace(input.Output))
	switch lower {
	case "true", "false", "yes", "no", "1", "0":
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid boolean: "+lower)
	default:
		return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "not a valid boolean")
	}
}
