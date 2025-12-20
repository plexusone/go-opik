package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/grokify/go-comet-ml-opik/evaluation"
)

// BaseJudge provides common functionality for LLM-based evaluation metrics.
type BaseJudge struct {
	evaluation.BaseMetric
	provider    Provider
	model       string
	temperature float64
}

// NewBaseJudge creates a new base judge.
func NewBaseJudge(name string, provider Provider, opts ...JudgeOption) *BaseJudge {
	j := &BaseJudge{
		BaseMetric:  evaluation.NewBaseMetric(name),
		provider:    provider,
		model:       provider.DefaultModel(),
		temperature: 0.0, // Deterministic by default for evaluation
	}

	for _, opt := range opts {
		opt(j)
	}

	return j
}

// JudgeOption configures a judge metric.
type JudgeOption func(*BaseJudge)

// WithJudgeModel sets the model for the judge.
func WithJudgeModel(model string) JudgeOption {
	return func(j *BaseJudge) {
		j.model = model
	}
}

// WithJudgeTemperature sets the temperature for the judge.
func WithJudgeTemperature(temp float64) JudgeOption {
	return func(j *BaseJudge) {
		j.temperature = temp
	}
}

// Provider returns the LLM provider.
func (j *BaseJudge) Provider() Provider {
	return j.provider
}

// Model returns the model name.
func (j *BaseJudge) Model() string {
	return j.model
}

// Complete sends a completion request to the provider.
func (j *BaseJudge) Complete(ctx context.Context, messages []Message) (*CompletionResponse, error) {
	return j.provider.Complete(ctx, CompletionRequest{
		Messages:    messages,
		Model:       j.model,
		Temperature: j.temperature,
	})
}

// ParseScoreFromResponse extracts a numeric score from an LLM response.
func ParseScoreFromResponse(response string) (float64, error) {
	// Try to find JSON with score
	jsonPattern := regexp.MustCompile(`\{[^}]*"score"\s*:\s*([\d.]+)[^}]*\}`)
	if matches := jsonPattern.FindStringSubmatch(response); len(matches) > 1 {
		return strconv.ParseFloat(matches[1], 64)
	}

	// Try to find a standalone number
	numberPattern := regexp.MustCompile(`(?:^|[^\d])((?:0|1)(?:\.\d+)?|(?:0?\.\d+))(?:[^\d]|$)`)
	if matches := numberPattern.FindStringSubmatch(response); len(matches) > 1 {
		return strconv.ParseFloat(matches[1], 64)
	}

	// Try to find score out of 10 or 100 and normalize
	scorePattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:/|out of)\s*(\d+)`)
	if matches := scorePattern.FindStringSubmatch(response); len(matches) > 2 {
		score, err1 := strconv.ParseFloat(matches[1], 64)
		max, err2 := strconv.ParseFloat(matches[2], 64)
		if err1 == nil && err2 == nil && max > 0 {
			return score / max, nil
		}
	}

	// Try to parse yes/no as 1.0/0.0
	lower := strings.ToLower(strings.TrimSpace(response))
	if strings.HasPrefix(lower, "yes") || strings.HasPrefix(lower, "true") {
		return 1.0, nil
	}
	if strings.HasPrefix(lower, "no") || strings.HasPrefix(lower, "false") {
		return 0.0, nil
	}

	return 0, fmt.Errorf("could not parse score from response: %s", truncate(response, 100))
}

// ParseJSONResponse extracts JSON from an LLM response.
func ParseJSONResponse(response string, v any) error {
	// Try to find JSON in code blocks
	codeBlockPattern := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	if matches := codeBlockPattern.FindStringSubmatch(response); len(matches) > 1 {
		return json.Unmarshal([]byte(matches[1]), v)
	}

	// Try direct parsing
	response = strings.TrimSpace(response)
	return json.Unmarshal([]byte(response), v)
}

// ParseReasonFromResponse extracts a reason/explanation from an LLM response.
func ParseReasonFromResponse(response string) string {
	// Look for common patterns
	patterns := []string{
		`(?i)reason(?:ing)?:\s*(.+?)(?:\n|$)`,
		`(?i)explanation:\s*(.+?)(?:\n|$)`,
		`(?i)because\s+(.+?)(?:\n|$)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(response); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// If no pattern found, use first sentence
	sentences := strings.SplitN(response, ".", 2)
	if len(sentences) > 0 {
		return strings.TrimSpace(sentences[0])
	}

	return truncate(response, 200)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ScoreResponse represents a structured scoring response.
type ScoreResponse struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason,omitempty"`
}

// ParseScoreResponse parses a JSON score response.
func ParseScoreResponse(response string) (*ScoreResponse, error) {
	var sr ScoreResponse
	if err := ParseJSONResponse(response, &sr); err != nil {
		// Try extracting score directly
		score, err := ParseScoreFromResponse(response)
		if err != nil {
			return nil, err
		}
		return &ScoreResponse{
			Score:  score,
			Reason: ParseReasonFromResponse(response),
		}, nil
	}
	return &sr, nil
}

// FormatPromptTemplate formats a prompt template with variable substitution.
func FormatPromptTemplate(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
		result = strings.ReplaceAll(result, "{{ "+k+" }}", v)
	}
	return result
}

// ScoreWithRetry attempts to score with retries on failure.
func ScoreWithRetry(ctx context.Context, j *BaseJudge, messages []Message, maxRetries int) (*ScoreResponse, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		resp, err := j.Complete(ctx, messages)
		if err != nil {
			lastErr = err
			continue
		}

		sr, err := ParseScoreResponse(resp.Content)
		if err != nil {
			lastErr = err
			continue
		}

		return sr, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
