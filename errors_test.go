package opik

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrMissingURL", ErrMissingURL, "opik: missing API URL"},
		{"ErrMissingAPIKey", ErrMissingAPIKey, "opik: missing API key for Opik Cloud"},
		{"ErrMissingWorkspace", ErrMissingWorkspace, "opik: missing workspace for Opik Cloud"},
		{"ErrTracingDisabled", ErrTracingDisabled, "opik: tracing is disabled"},
		{"ErrTraceNotFound", ErrTraceNotFound, "opik: trace not found"},
		{"ErrSpanNotFound", ErrSpanNotFound, "opik: span not found"},
		{"ErrDatasetNotFound", ErrDatasetNotFound, "opik: dataset not found"},
		{"ErrExperimentNotFound", ErrExperimentNotFound, "opik: experiment not found"},
		{"ErrPromptNotFound", ErrPromptNotFound, "opik: prompt not found"},
		{"ErrInvalidInput", ErrInvalidInput, "opik: invalid input"},
		{"ErrNoActiveTrace", ErrNoActiveTrace, "opik: no active trace in context"},
		{"ErrNoActiveSpan", ErrNoActiveSpan, "opik: no active span in context"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	t.Run("with details", func(t *testing.T) {
		err := &APIError{
			StatusCode: 400,
			Message:    "Bad Request",
			Details:    "invalid JSON",
		}
		want := "opik: API error (Bad Request): invalid JSON"
		if err.Error() != want {
			t.Errorf("Error() = %q, want %q", err.Error(), want)
		}
	})

	t.Run("without details", func(t *testing.T) {
		err := &APIError{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
		want := "opik: API error: Internal Server Error"
		if err.Error() != want {
			t.Errorf("Error() = %q, want %q", err.Error(), want)
		}
	})
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"ErrTraceNotFound", ErrTraceNotFound, true},
		{"ErrSpanNotFound", ErrSpanNotFound, true},
		{"ErrDatasetNotFound", ErrDatasetNotFound, true},
		{"ErrExperimentNotFound", ErrExperimentNotFound, true},
		{"ErrPromptNotFound", ErrPromptNotFound, true},
		{"APIError 404", &APIError{StatusCode: 404, Message: "Not Found"}, true},
		{"APIError 400", &APIError{StatusCode: 400, Message: "Bad Request"}, false},
		{"ErrMissingURL", ErrMissingURL, false},
		{"generic error", errors.New("some error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"APIError 401", &APIError{StatusCode: 401, Message: "Unauthorized"}, true},
		{"APIError 403", &APIError{StatusCode: 403, Message: "Forbidden"}, false},
		{"APIError 404", &APIError{StatusCode: 404, Message: "Not Found"}, false},
		{"sentinel error", ErrMissingAPIKey, false},
		{"generic error", errors.New("unauthorized"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.want {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"APIError 429", &APIError{StatusCode: 429, Message: "Too Many Requests"}, true},
		{"APIError 500", &APIError{StatusCode: 500, Message: "Internal Server Error"}, false},
		{"sentinel error", ErrMissingURL, false},
		{"generic error", errors.New("rate limited"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimited(tt.err); got != tt.want {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorsIs(t *testing.T) {
	// Test that errors.Is works with sentinel errors
	wrappedTrace := errors.New("wrapped: " + ErrTraceNotFound.Error())
	_ = wrappedTrace // We can't use errors.Is on non-wrapped errors, just verify sentinel works

	if !errors.Is(ErrTraceNotFound, ErrTraceNotFound) {
		t.Error("errors.Is(ErrTraceNotFound, ErrTraceNotFound) should be true")
	}

	if errors.Is(ErrTraceNotFound, ErrSpanNotFound) {
		t.Error("errors.Is(ErrTraceNotFound, ErrSpanNotFound) should be false")
	}
}

func TestAPIErrorStatusCodes(t *testing.T) {
	statusCodes := []int{200, 201, 400, 401, 403, 404, 429, 500, 502, 503}

	for _, code := range statusCodes {
		err := &APIError{StatusCode: code, Message: "test"}
		if err.StatusCode != code {
			t.Errorf("StatusCode = %d, want %d", err.StatusCode, code)
		}
	}
}
