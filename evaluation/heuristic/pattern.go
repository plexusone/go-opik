package heuristic

import (
	"context"
	"regexp"
	"strings"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

// RegexMatch checks if the output matches a regular expression.
type RegexMatch struct {
	evaluation.BaseMetric
	pattern *regexp.Regexp
}

// NewRegexMatch creates a new RegexMatch metric.
func NewRegexMatch(pattern string) (*RegexMatch, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexMatch{
		BaseMetric: evaluation.NewBaseMetric("regex_match"),
		pattern:    re,
	}, nil
}

// MustRegexMatch creates a new RegexMatch metric, panics on invalid pattern.
func MustRegexMatch(pattern string) *RegexMatch {
	m, err := NewRegexMatch(pattern)
	if err != nil {
		panic(err)
	}
	return m
}

// Score evaluates if output matches the regex pattern.
func (m *RegexMatch) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	if m.pattern.MatchString(input.Output) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "matches pattern")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "does not match pattern")
}

// RegexNotMatch checks if the output does NOT match a regular expression.
type RegexNotMatch struct {
	evaluation.BaseMetric
	pattern *regexp.Regexp
}

// NewRegexNotMatch creates a new RegexNotMatch metric.
func NewRegexNotMatch(pattern string) (*RegexNotMatch, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexNotMatch{
		BaseMetric: evaluation.NewBaseMetric("regex_not_match"),
		pattern:    re,
	}, nil
}

// MustRegexNotMatch creates a new RegexNotMatch metric, panics on invalid pattern.
func MustRegexNotMatch(pattern string) *RegexNotMatch {
	m, err := NewRegexNotMatch(pattern)
	if err != nil {
		panic(err)
	}
	return m
}

// Score evaluates if output does not match the regex pattern.
func (m *RegexNotMatch) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	if !m.pattern.MatchString(input.Output) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "does not match pattern")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "matches pattern (unexpected)")
}

// RegexFindAll counts how many times a pattern matches.
type RegexFindAll struct {
	evaluation.BaseMetric
	pattern     *regexp.Regexp
	minMatches  int
	maxMatches  int
	normalizeBy int // if > 0, score = min(matches/normalizeBy, 1.0)
}

// NewRegexFindAll creates a new RegexFindAll metric.
func NewRegexFindAll(pattern string, minMatches, maxMatches int) (*RegexFindAll, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexFindAll{
		BaseMetric: evaluation.NewBaseMetric("regex_find_all"),
		pattern:    re,
		minMatches: minMatches,
		maxMatches: maxMatches,
	}, nil
}

// Score evaluates the number of pattern matches.
func (m *RegexFindAll) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	matches := m.pattern.FindAllString(input.Output, -1)
	count := len(matches)

	if m.normalizeBy > 0 {
		score := float64(count) / float64(m.normalizeBy)
		if score > 1.0 {
			score = 1.0
		}
		return evaluation.NewScoreResult(m.Name(), score)
	}

	if count >= m.minMatches && (m.maxMatches <= 0 || count <= m.maxMatches) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "match count within range")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "match count out of range")
}

// EmailFormat checks if the output contains a valid email format.
type EmailFormat struct {
	evaluation.BaseMetric
	pattern *regexp.Regexp
}

// NewEmailFormat creates a new EmailFormat metric.
func NewEmailFormat() *EmailFormat {
	return &EmailFormat{
		BaseMetric: evaluation.NewBaseMetric("email_format"),
		pattern:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	}
}

// Score evaluates if output is a valid email format.
func (m *EmailFormat) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := strings.TrimSpace(input.Output)
	if m.pattern.MatchString(output) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid email format")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "invalid email format")
}

// URLFormat checks if the output contains a valid URL format.
type URLFormat struct {
	evaluation.BaseMetric
	pattern *regexp.Regexp
}

// NewURLFormat creates a new URLFormat metric.
func NewURLFormat() *URLFormat {
	return &URLFormat{
		BaseMetric: evaluation.NewBaseMetric("url_format"),
		pattern:    regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`),
	}
}

// Score evaluates if output is a valid URL format.
func (m *URLFormat) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := strings.TrimSpace(input.Output)
	if m.pattern.MatchString(output) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid URL format")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "invalid URL format")
}

// PhoneFormat checks if the output contains a phone number format.
type PhoneFormat struct {
	evaluation.BaseMetric
	pattern *regexp.Regexp
}

// NewPhoneFormat creates a new PhoneFormat metric.
func NewPhoneFormat() *PhoneFormat {
	return &PhoneFormat{
		BaseMetric: evaluation.NewBaseMetric("phone_format"),
		// Matches common phone formats like +1-234-567-8901, (234) 567-8901, etc.
		pattern: regexp.MustCompile(`^[\+]?[(]?[0-9]{1,4}[)]?[-\s\./0-9]*$`),
	}
}

// Score evaluates if output is a valid phone format.
func (m *PhoneFormat) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := strings.TrimSpace(input.Output)
	if len(output) >= 7 && m.pattern.MatchString(output) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid phone format")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "invalid phone format")
}

// DateFormat checks if the output matches a date format.
type DateFormat struct {
	evaluation.BaseMetric
	pattern *regexp.Regexp
}

// NewDateFormat creates a new DateFormat metric with ISO format by default.
func NewDateFormat() *DateFormat {
	return &DateFormat{
		BaseMetric: evaluation.NewBaseMetric("date_format"),
		// Matches ISO 8601 dates and common formats
		pattern: regexp.MustCompile(`^\d{4}[-/]\d{2}[-/]\d{2}(T\d{2}:\d{2}(:\d{2})?)?`),
	}
}

// NewDateFormatWithPattern creates a new DateFormat metric with custom pattern.
func NewDateFormatWithPattern(pattern string) (*DateFormat, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &DateFormat{
		BaseMetric: evaluation.NewBaseMetric("date_format"),
		pattern:    re,
	}, nil
}

// Score evaluates if output matches a date format.
func (m *DateFormat) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := strings.TrimSpace(input.Output)
	if m.pattern.MatchString(output) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid date format")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "invalid date format")
}

// UUIDFormat checks if the output is a valid UUID format.
type UUIDFormat struct {
	evaluation.BaseMetric
	pattern *regexp.Regexp
}

// NewUUIDFormat creates a new UUIDFormat metric.
func NewUUIDFormat() *UUIDFormat {
	return &UUIDFormat{
		BaseMetric: evaluation.NewBaseMetric("uuid_format"),
		pattern:    regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
	}
}

// Score evaluates if output is a valid UUID format.
func (m *UUIDFormat) Score(ctx context.Context, input evaluation.MetricInput) *evaluation.ScoreResult {
	output := strings.TrimSpace(input.Output)
	if m.pattern.MatchString(output) {
		return evaluation.NewScoreResultWithReason(m.Name(), 1.0, "valid UUID format")
	}
	return evaluation.NewScoreResultWithReason(m.Name(), 0.0, "invalid UUID format")
}
