package evaluation

import (
	"encoding/json"
	"fmt"
)

// ScoreResult represents the result of a metric evaluation.
type ScoreResult struct {
	// Name is the name of the metric.
	Name string `json:"name"`
	// Value is the numeric score value (typically 0.0 to 1.0).
	Value float64 `json:"value"`
	// Reason is an optional explanation for the score.
	Reason string `json:"reason,omitempty"`
	// Metadata contains additional information about the score.
	Metadata map[string]any `json:"metadata,omitempty"`
	// Error is set if the metric evaluation failed.
	Error error `json:"error,omitempty"`
}

// IsSuccess returns true if the score was computed successfully.
func (s *ScoreResult) IsSuccess() bool {
	return s.Error == nil
}

// String returns a human-readable representation of the score.
func (s *ScoreResult) String() string {
	if s.Error != nil {
		return fmt.Sprintf("%s: error - %v", s.Name, s.Error)
	}
	if s.Reason != "" {
		return fmt.Sprintf("%s: %.4f (%s)", s.Name, s.Value, s.Reason)
	}
	return fmt.Sprintf("%s: %.4f", s.Name, s.Value)
}

// ToJSON returns the score as JSON bytes.
func (s *ScoreResult) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// ScoreResults is a collection of score results.
type ScoreResults []*ScoreResult

// ByName returns the first score result with the given name.
func (r ScoreResults) ByName(name string) *ScoreResult {
	for _, s := range r {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// AllByName returns all score results with the given name.
func (r ScoreResults) AllByName(name string) ScoreResults {
	results := make(ScoreResults, 0)
	for _, s := range r {
		if s.Name == name {
			results = append(results, s)
		}
	}
	return results
}

// Successful returns only the successful score results.
func (r ScoreResults) Successful() ScoreResults {
	results := make(ScoreResults, 0, len(r))
	for _, s := range r {
		if s.IsSuccess() {
			results = append(results, s)
		}
	}
	return results
}

// Failed returns only the failed score results.
func (r ScoreResults) Failed() ScoreResults {
	results := make(ScoreResults, 0)
	for _, s := range r {
		if !s.IsSuccess() {
			results = append(results, s)
		}
	}
	return results
}

// Average returns the average value of successful scores.
func (r ScoreResults) Average() float64 {
	successful := r.Successful()
	if len(successful) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range successful {
		sum += s.Value
	}
	return sum / float64(len(successful))
}

// AverageByName returns the average value of scores with the given name.
func (r ScoreResults) AverageByName(name string) float64 {
	named := r.AllByName(name).Successful()
	if len(named) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range named {
		sum += s.Value
	}
	return sum / float64(len(named))
}

// NewScoreResult creates a new successful score result.
func NewScoreResult(name string, value float64) *ScoreResult {
	return &ScoreResult{
		Name:  name,
		Value: value,
	}
}

// NewScoreResultWithReason creates a new score result with a reason.
func NewScoreResultWithReason(name string, value float64, reason string) *ScoreResult {
	return &ScoreResult{
		Name:   name,
		Value:  value,
		Reason: reason,
	}
}

// NewFailedScoreResult creates a new failed score result.
func NewFailedScoreResult(name string, err error) *ScoreResult {
	return &ScoreResult{
		Name:  name,
		Error: err,
	}
}

// BooleanScore converts a boolean to a score (1.0 for true, 0.0 for false).
func BooleanScore(name string, value bool) *ScoreResult {
	if value {
		return NewScoreResult(name, 1.0)
	}
	return NewScoreResult(name, 0.0)
}

// BooleanScoreWithReason converts a boolean to a score with a reason.
func BooleanScoreWithReason(name string, value bool, reason string) *ScoreResult {
	if value {
		return NewScoreResultWithReason(name, 1.0, reason)
	}
	return NewScoreResultWithReason(name, 0.0, reason)
}
