package opik

import "errors"

// Sentinel errors for the Opik SDK.
var (
	// ErrMissingURL is returned when the API URL is not configured.
	ErrMissingURL = errors.New("opik: missing API URL")

	// ErrMissingAPIKey is returned when the API key is required but not provided.
	ErrMissingAPIKey = errors.New("opik: missing API key for Opik Cloud")

	// ErrMissingWorkspace is returned when the workspace is required but not provided.
	ErrMissingWorkspace = errors.New("opik: missing workspace for Opik Cloud")

	// ErrTracingDisabled is returned when tracing is disabled.
	ErrTracingDisabled = errors.New("opik: tracing is disabled")

	// ErrTraceNotFound is returned when a trace cannot be found.
	ErrTraceNotFound = errors.New("opik: trace not found")

	// ErrSpanNotFound is returned when a span cannot be found.
	ErrSpanNotFound = errors.New("opik: span not found")

	// ErrDatasetNotFound is returned when a dataset cannot be found.
	ErrDatasetNotFound = errors.New("opik: dataset not found")

	// ErrExperimentNotFound is returned when an experiment cannot be found.
	ErrExperimentNotFound = errors.New("opik: experiment not found")

	// ErrPromptNotFound is returned when a prompt cannot be found.
	ErrPromptNotFound = errors.New("opik: prompt not found")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("opik: invalid input")

	// ErrNoActiveTrace is returned when there is no active trace in context.
	ErrNoActiveTrace = errors.New("opik: no active trace in context")

	// ErrNoActiveSpan is returned when there is no active span in context.
	ErrNoActiveSpan = errors.New("opik: no active span in context")
)

// APIError represents an error returned by the Opik API.
type APIError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return "opik: API error (" + e.Message + "): " + e.Details
	}
	return "opik: API error: " + e.Message
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return errors.Is(err, ErrTraceNotFound) ||
		errors.Is(err, ErrSpanNotFound) ||
		errors.Is(err, ErrDatasetNotFound) ||
		errors.Is(err, ErrExperimentNotFound) ||
		errors.Is(err, ErrPromptNotFound)
}

// IsUnauthorized returns true if the error indicates an authentication failure.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsRateLimited returns true if the error indicates rate limiting.
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 429
	}
	return false
}
