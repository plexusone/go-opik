package opik

import (
	"errors"
	"strings"

	"github.com/plexusone/opik-go/ax"
)

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

// AXErrorCode extracts the AX error code from the API error, if present.
// Returns the error code constant (e.g., ax.ErrTraceNotFound) and true if found.
// Use this for machine-readable error handling:
//
//	if apiErr, ok := err.(*APIError); ok {
//	    if code, ok := apiErr.AXErrorCode(); ok {
//	        switch code {
//	        case ax.ErrTraceNotFound:
//	            // Handle trace not found
//	        case ax.ErrUnauthorized:
//	            // Handle auth required
//	        }
//	    }
//	}
func (e *APIError) AXErrorCode() (string, bool) {
	// Check Message and Details for known error codes
	for _, code := range ax.AllErrorCodes {
		if strings.Contains(e.Message, code) || strings.Contains(e.Details, code) {
			return code, true
		}
	}
	// Fall back to HTTP status code mapping
	if code := ax.ErrorCodeForHTTPStatus(e.StatusCode); code != "" {
		return code, true
	}
	return "", false
}

// HasAXCode checks if the API error contains a specific AX error code.
func (e *APIError) HasAXCode(code string) bool {
	axCode, ok := e.AXErrorCode()
	return ok && axCode == code
}

// IsAXError checks if an error contains a specific AX error code.
// This works with any error type by first trying to extract as an APIError.
//
// Usage:
//
//	if IsAXError(err, ax.ErrTraceNotFound) {
//	    // Handle trace not found
//	}
func IsAXError(err error, code string) bool {
	if err == nil {
		return false
	}

	// Try to get as APIError first for structured checking
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.HasAXCode(code)
	}

	// Fall back to string matching
	return ax.IsErrorCode(err, code)
}

// GetAXErrorCode extracts the AX error code from any error.
// Returns the error code and true if found, empty string and false otherwise.
func GetAXErrorCode(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	// Try structured extraction first
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.AXErrorCode()
	}

	// Fall back to string matching
	return ax.ContainsErrorCode(err)
}

// GetAXErrorInfo returns metadata about an error's AX code, if present.
// Returns nil if no AX code is found or the code is not recognized.
func GetAXErrorInfo(err error) *ax.ErrorCodeInfo {
	code, ok := GetAXErrorCode(err)
	if !ok {
		return nil
	}
	return ax.GetErrorInfo(code)
}
