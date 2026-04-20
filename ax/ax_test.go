package ax

import (
	"errors"
	"strconv"
	"testing"
)

func TestIsErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			code:     ErrTraceNotFound,
			expected: false,
		},
		{
			name:     "matching error code",
			err:      errors.New("API error: TRACE_NOT_FOUND - The trace was not found"),
			code:     ErrTraceNotFound,
			expected: true,
		},
		{
			name:     "non-matching error code",
			err:      errors.New("API error: TRACE_NOT_FOUND"),
			code:     ErrSpanNotFound,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("network timeout"),
			code:     ErrTraceNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsErrorCode(tt.err, tt.code)
			if result != tt.expected {
				t.Errorf("IsErrorCode(%v, %q) = %v, want %v", tt.err, tt.code, result, tt.expected)
			}
		})
	}
}

func TestContainsErrorCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
		expectedOk   bool
	}{
		{
			name:         "nil error",
			err:          nil,
			expectedCode: "",
			expectedOk:   false,
		},
		{
			name:         "contains TRACE_NOT_FOUND",
			err:          errors.New("Error: TRACE_NOT_FOUND"),
			expectedCode: ErrTraceNotFound,
			expectedOk:   true,
		},
		{
			name:         "contains DATASET_NOT_FOUND",
			err:          errors.New("Error: DATASET_NOT_FOUND - dataset does not exist"),
			expectedCode: ErrDatasetNotFound,
			expectedOk:   true,
		},
		{
			name:         "no known error code",
			err:          errors.New("unknown error"),
			expectedCode: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := ContainsErrorCode(tt.err)
			if code != tt.expectedCode || ok != tt.expectedOk {
				t.Errorf("ContainsErrorCode(%v) = (%q, %v), want (%q, %v)",
					tt.err, code, ok, tt.expectedCode, tt.expectedOk)
			}
		})
	}
}

func TestGetErrorInfo(t *testing.T) {
	// Test known error code
	info := GetErrorInfo(ErrTraceNotFound)
	if info == nil {
		t.Fatal("GetErrorInfo(ErrTraceNotFound) returned nil")
	}
	if info.Category != "not_found" {
		t.Errorf("expected category 'not_found', got %q", info.Category)
	}
	if info.HTTPStatus != 404 {
		t.Errorf("expected HTTP status 404, got %d", info.HTTPStatus)
	}

	// Test unknown error code
	info = GetErrorInfo("UNKNOWN_ERROR")
	if info != nil {
		t.Errorf("GetErrorInfo(UNKNOWN_ERROR) should return nil, got %+v", info)
	}
}

func TestErrorCategoryHelpers(t *testing.T) {
	tests := []struct {
		code         string
		isAuth       bool
		isNotFound   bool
		isValidation bool
		isConflict   bool
	}{
		{ErrUnauthorized, true, false, false, false},
		{ErrForbidden, true, false, false, false},
		{ErrTraceNotFound, false, true, false, false},
		{ErrDatasetNotFound, false, true, false, false},
		{ErrInvalidInput, false, false, true, false},
		{ErrConflict, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if IsAuthError(tt.code) != tt.isAuth {
				t.Errorf("IsAuthError(%q) = %v, want %v", tt.code, !tt.isAuth, tt.isAuth)
			}
			if IsNotFoundError(tt.code) != tt.isNotFound {
				t.Errorf("IsNotFoundError(%q) = %v, want %v", tt.code, !tt.isNotFound, tt.isNotFound)
			}
			if IsValidationError(tt.code) != tt.isValidation {
				t.Errorf("IsValidationError(%q) = %v, want %v", tt.code, !tt.isValidation, tt.isValidation)
			}
			if IsConflictError(tt.code) != tt.isConflict {
				t.Errorf("IsConflictError(%q) = %v, want %v", tt.code, !tt.isConflict, tt.isConflict)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		operationID string
		expected    bool
	}{
		{"findTraces", false}, // Not in map, defaults to false
		{"getTraceById", true},
		{"findDatasets", true},
		{"getProjectById", true},
		{"createTrace", false},
		{"createDataset", false},
		{"deleteTraceById", false},
		{"unknown_operation", false},
	}

	for _, tt := range tests {
		t.Run(tt.operationID, func(t *testing.T) {
			result := IsRetryable(tt.operationID)
			if result != tt.expected {
				t.Errorf("IsRetryable(%q) = %v, want %v", tt.operationID, result, tt.expected)
			}
		})
	}
}

func TestRetryableCount(t *testing.T) {
	retryable, nonRetryable := RetryableCount()

	// Verify we have a reasonable distribution
	total := retryable + nonRetryable
	if total != len(RetryPolicy) {
		t.Errorf("RetryableCount total %d != len(RetryPolicy) %d", total, len(RetryPolicy))
	}

	// Should have both categories
	if retryable == 0 {
		t.Error("Expected some retryable operations")
	}
	if nonRetryable == 0 {
		t.Error("Expected some non-retryable operations")
	}
}

func TestGetRequiredFields(t *testing.T) {
	// Test operation with required fields
	fields := GetRequiredFields("createTrace")
	if len(fields) == 0 {
		t.Error("createTrace should have required fields")
	}
	found := false
	for _, f := range fields {
		if f == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("createTrace should require 'name' field")
	}

	// Test operation without required fields (or unknown)
	fields = GetRequiredFields("findTraces")
	if len(fields) != 0 {
		t.Errorf("findTraces should have no required fields, got %v", fields)
	}
}

func TestMissingFields(t *testing.T) {
	present := map[string]bool{
		"name": true,
	}

	// All required fields present
	missing := MissingFields("createTrace", present)
	if len(missing) != 0 {
		t.Errorf("expected no missing fields, got %v", missing)
	}

	// Missing required field
	missing = MissingFields("createExperiment", present)
	if len(missing) == 0 {
		t.Error("expected missing fields for createExperiment")
	}
}

func TestValidateFields(t *testing.T) {
	present := map[string]bool{
		"name": true,
	}

	// Valid
	msg := ValidateFields("createTrace", present)
	if msg != "" {
		t.Errorf("expected empty validation message, got %q", msg)
	}

	// Invalid - missing fields
	msg = ValidateFields("createExperiment", present)
	if msg == "" {
		t.Error("expected validation error message for createExperiment")
	}
}

func TestCapabilities(t *testing.T) {
	// Test read operation
	caps := GetCapabilities("getTraceById")
	if len(caps) == 0 {
		t.Fatal("getTraceById should have capabilities")
	}
	if !HasCapability("getTraceById", CapRead) {
		t.Error("getTraceById should have read capability")
	}
	if HasCapability("getTraceById", CapWrite) {
		t.Error("getTraceById should not have write capability")
	}

	// Test write operation
	if !HasCapability("createTrace", CapWrite) {
		t.Error("createTrace should have write capability")
	}

	// Test delete operation
	if !HasCapability("deleteTraceById", CapDelete) {
		t.Error("deleteTraceById should have delete capability")
	}

	// Test evaluation operation
	if !IsEvaluation("evaluateTraces") {
		t.Error("evaluateTraces should be an evaluation operation")
	}

	// Test streaming operation
	if !SupportsStreaming("streamDatasetItems") {
		t.Error("streamDatasetItems should support streaming")
	}

	// Test admin operation
	if !RequiresAdmin("upsertWorkspaceConfiguration") {
		t.Error("upsertWorkspaceConfiguration should require admin")
	}
}

func TestIsReadOnly(t *testing.T) {
	if !IsReadOnly("getTraceById") {
		t.Error("getTraceById should be read-only")
	}
	if IsReadOnly("createTrace") {
		t.Error("createTrace should not be read-only")
	}
	if IsReadOnly("deleteTraceById") {
		t.Error("deleteTraceById should not be read-only")
	}
}

func TestErrorCodeForHTTPStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{400, ErrInvalidInput},
		{401, ErrUnauthorized},
		{403, ErrForbidden},
		{404, ""},
		{409, ErrConflict},
		{429, ErrRateLimited},
		{500, ErrInternalError},
		{503, ErrInternalError},
		{418, ""}, // I'm a teapot
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.status), func(t *testing.T) {
			result := ErrorCodeForHTTPStatus(tt.status)
			if result != tt.expected {
				t.Errorf("ErrorCodeForHTTPStatus(%d) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}
