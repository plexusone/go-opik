// Package testutil provides testing utilities for the Opik SDK.
// This mirrors the Python SDK's testlib package.
package testutil

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

// Matcher is an interface for flexible value matching in tests.
type Matcher interface {
	Match(value any) bool
	String() string
}

// AnyMatcher matches any value including nil.
type AnyMatcher struct{}

func (m AnyMatcher) Match(value any) bool { return true }
func (m AnyMatcher) String() string       { return "<ANY>" }

// Any returns a matcher that matches any value.
func Any() Matcher { return AnyMatcher{} }

// AnyButNilMatcher matches any non-nil value.
type AnyButNilMatcher struct{}

func (m AnyButNilMatcher) Match(value any) bool {
	return value != nil && !reflect.ValueOf(value).IsZero()
}
func (m AnyButNilMatcher) String() string { return "<ANY_BUT_NIL>" }

// AnyButNil returns a matcher that matches any non-nil value.
func AnyButNil() Matcher { return AnyButNilMatcher{} }

// AnyStringMatcher matches any string, optionally with constraints.
type AnyStringMatcher struct {
	Prefix   string
	Suffix   string
	Contains string
	MinLen   int
	MaxLen   int
}

func (m AnyStringMatcher) Match(value any) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}
	if m.Prefix != "" && !strings.HasPrefix(s, m.Prefix) {
		return false
	}
	if m.Suffix != "" && !strings.HasSuffix(s, m.Suffix) {
		return false
	}
	if m.Contains != "" && !strings.Contains(s, m.Contains) {
		return false
	}
	if m.MinLen > 0 && len(s) < m.MinLen {
		return false
	}
	if m.MaxLen > 0 && len(s) > m.MaxLen {
		return false
	}
	return true
}
func (m AnyStringMatcher) String() string { return "<ANY_STRING>" }

// AnyString returns a matcher for strings.
func AnyString() *AnyStringMatcher { return &AnyStringMatcher{} }

// WithPrefix adds a prefix requirement.
func (m *AnyStringMatcher) WithPrefix(prefix string) *AnyStringMatcher {
	m.Prefix = prefix
	return m
}

// WithSuffix adds a suffix requirement.
func (m *AnyStringMatcher) WithSuffix(suffix string) *AnyStringMatcher {
	m.Suffix = suffix
	return m
}

// Containing adds a contains requirement.
func (m *AnyStringMatcher) Containing(substr string) *AnyStringMatcher {
	m.Contains = substr
	return m
}

// AnyMapMatcher matches any map, optionally requiring specific keys.
type AnyMapMatcher struct {
	RequiredKeys []string
}

func (m AnyMapMatcher) Match(value any) bool {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Map {
		return false
	}
	if len(m.RequiredKeys) > 0 {
		for _, key := range m.RequiredKeys {
			found := false
			for _, k := range v.MapKeys() {
				if k.String() == key {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}
func (m AnyMapMatcher) String() string { return "<ANY_MAP>" }

// AnyMap returns a matcher for maps.
func AnyMap(requiredKeys ...string) AnyMapMatcher {
	return AnyMapMatcher{RequiredKeys: requiredKeys}
}

// AnySliceMatcher matches any slice.
type AnySliceMatcher struct {
	MinLen int
	MaxLen int
}

func (m AnySliceMatcher) Match(value any) bool {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return false
	}
	if m.MinLen > 0 && v.Len() < m.MinLen {
		return false
	}
	if m.MaxLen > 0 && v.Len() > m.MaxLen {
		return false
	}
	return true
}
func (m AnySliceMatcher) String() string { return "<ANY_SLICE>" }

// AnySlice returns a matcher for slices.
func AnySlice() AnySliceMatcher { return AnySliceMatcher{} }

// AnyFloatMatcher matches floats within a range or tolerance.
type AnyFloatMatcher struct {
	Min       float64
	Max       float64
	Tolerance float64
	Expected  float64
}

func (m AnyFloatMatcher) Match(value any) bool {
	var f float64
	switch v := value.(type) {
	case float64:
		f = v
	case float32:
		f = float64(v)
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	default:
		return false
	}
	if m.Tolerance > 0 {
		return math.Abs(f-m.Expected) <= m.Tolerance
	}
	if m.Min != 0 || m.Max != 0 {
		return f >= m.Min && f <= m.Max
	}
	return true
}
func (m AnyFloatMatcher) String() string { return "<ANY_FLOAT>" }

// AnyFloat returns a matcher for floats.
func AnyFloat() *AnyFloatMatcher { return &AnyFloatMatcher{} }

// Between sets min/max range.
func (m *AnyFloatMatcher) Between(min, max float64) *AnyFloatMatcher {
	m.Min = min
	m.Max = max
	return m
}

// Near sets expected value with tolerance.
func (m *AnyFloatMatcher) Near(expected, tolerance float64) *AnyFloatMatcher {
	m.Expected = expected
	m.Tolerance = tolerance
	return m
}

// AssertMatch checks if a value matches a matcher.
func AssertMatch(t *testing.T, matcher Matcher, value any) {
	t.Helper()
	if !matcher.Match(value) {
		t.Errorf("value %v does not match %s", value, matcher.String())
	}
}

// AssertMapHasKeys checks if a map has all required keys.
func AssertMapHasKeys(t *testing.T, m map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if _, ok := m[key]; !ok {
			t.Errorf("map missing key %q", key)
		}
	}
}

// AssertJSONEqual compares two values as JSON.
func AssertJSONEqual(t *testing.T, expected, actual any) {
	t.Helper()
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal expected: %v", err)
	}
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("failed to marshal actual: %v", err)
	}
	if string(expectedJSON) != string(actualJSON) {
		t.Errorf("JSON mismatch:\nexpected: %s\nactual: %s", expectedJSON, actualJSON)
	}
}

// AssertFloatNear checks if two floats are within tolerance.
func AssertFloatNear(t *testing.T, expected, actual, tolerance float64) {
	t.Helper()
	if math.Abs(expected-actual) > tolerance {
		t.Errorf("floats not near: expected %v, got %v (tolerance %v)", expected, actual, tolerance)
	}
}

// MockEnv temporarily sets environment variables for a test.
type MockEnv struct {
	t       *testing.T
	origEnv map[string]string
}

// NewMockEnv creates a new environment mocker.
func NewMockEnv(t *testing.T) *MockEnv {
	return &MockEnv{
		t:       t,
		origEnv: make(map[string]string),
	}
}

// Set sets an environment variable and records it for cleanup.
func (m *MockEnv) Set(key, value string) {
	m.t.Setenv(key, value)
}

// Unset unsets an environment variable.
func (m *MockEnv) Unset(key string) {
	m.t.Setenv(key, "")
}
