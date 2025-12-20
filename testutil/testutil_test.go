package testutil

import (
	"testing"
)

func TestAnyMatcher(t *testing.T) {
	m := Any()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, true},
		{"string", "hello", true},
		{"int", 42, true},
		{"map", map[string]any{}, true},
		{"slice", []int{1, 2, 3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Match(tt.value); got != tt.want {
				t.Errorf("Match(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestAnyButNilMatcher(t *testing.T) {
	m := AnyButNil()

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"empty string", "", false},
		{"zero int", 0, false},
		{"string", "hello", true},
		{"int", 42, true},
		// Note: empty map is not the zero value (nil is), so it matches
		{"empty map", map[string]any{}, true},
		{"non-empty map", map[string]any{"a": 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Match(tt.value); got != tt.want {
				t.Errorf("Match(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestAnyStringMatcher(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		m := AnyString()
		if !m.Match("hello") {
			t.Error("should match any string")
		}
		if m.Match(123) {
			t.Error("should not match non-string")
		}
	})

	t.Run("with prefix", func(t *testing.T) {
		m := AnyString().WithPrefix("hello")
		if !m.Match("hello world") {
			t.Error("should match string with prefix")
		}
		if m.Match("world hello") {
			t.Error("should not match string without prefix")
		}
	})

	t.Run("with suffix", func(t *testing.T) {
		m := AnyString().WithSuffix("world")
		if !m.Match("hello world") {
			t.Error("should match string with suffix")
		}
		if m.Match("world hello") {
			t.Error("should not match string without suffix")
		}
	})

	t.Run("containing", func(t *testing.T) {
		m := AnyString().Containing("middle")
		if !m.Match("start middle end") {
			t.Error("should match string containing substring")
		}
		if m.Match("no match here") {
			t.Error("should not match string without substring")
		}
	})

	t.Run("min length", func(t *testing.T) {
		m := &AnyStringMatcher{MinLen: 5}
		if !m.Match("hello") {
			t.Error("should match string at min length")
		}
		if m.Match("hi") {
			t.Error("should not match short string")
		}
	})

	t.Run("max length", func(t *testing.T) {
		m := &AnyStringMatcher{MaxLen: 5}
		if !m.Match("hello") {
			t.Error("should match string at max length")
		}
		if m.Match("hello world") {
			t.Error("should not match long string")
		}
	})
}

func TestAnyMapMatcher(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		m := AnyMap()
		if !m.Match(map[string]any{"a": 1}) {
			t.Error("should match any map")
		}
		if m.Match("not a map") {
			t.Error("should not match non-map")
		}
	})

	t.Run("required keys", func(t *testing.T) {
		m := AnyMap("a", "b")
		if !m.Match(map[string]any{"a": 1, "b": 2, "c": 3}) {
			t.Error("should match map with required keys")
		}
		if m.Match(map[string]any{"a": 1}) {
			t.Error("should not match map missing required key")
		}
	})
}

func TestAnySliceMatcher(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		m := AnySlice()
		if !m.Match([]int{1, 2, 3}) {
			t.Error("should match any slice")
		}
		if m.Match("not a slice") {
			t.Error("should not match non-slice")
		}
	})

	t.Run("min length", func(t *testing.T) {
		m := AnySliceMatcher{MinLen: 3}
		if !m.Match([]int{1, 2, 3}) {
			t.Error("should match slice at min length")
		}
		if m.Match([]int{1, 2}) {
			t.Error("should not match short slice")
		}
	})

	t.Run("max length", func(t *testing.T) {
		m := AnySliceMatcher{MaxLen: 3}
		if !m.Match([]int{1, 2, 3}) {
			t.Error("should match slice at max length")
		}
		if m.Match([]int{1, 2, 3, 4}) {
			t.Error("should not match long slice")
		}
	})
}

func TestAnyFloatMatcher(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		m := AnyFloat()
		if !m.Match(3.14) {
			t.Error("should match float64")
		}
		if !m.Match(float32(3.14)) {
			t.Error("should match float32")
		}
		if !m.Match(42) {
			t.Error("should match int as float")
		}
		if m.Match("not a number") {
			t.Error("should not match non-number")
		}
	})

	t.Run("between", func(t *testing.T) {
		m := AnyFloat().Between(0.0, 1.0)
		if !m.Match(0.5) {
			t.Error("should match value in range")
		}
		if m.Match(1.5) {
			t.Error("should not match value out of range")
		}
	})

	t.Run("near", func(t *testing.T) {
		m := AnyFloat().Near(0.7, 0.01)
		if !m.Match(0.699) {
			t.Error("should match value within tolerance")
		}
		if !m.Match(0.701) {
			t.Error("should match value within tolerance")
		}
		if m.Match(0.72) {
			t.Error("should not match value outside tolerance")
		}
	})
}

func TestAssertMapHasKeys(t *testing.T) {
	m := map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	// This should pass
	mockT := &testing.T{}
	AssertMapHasKeys(mockT, m, "a", "b")
	// Can't easily check if it failed without more infrastructure
}

func TestAssertFloatNear(t *testing.T) {
	// This should pass
	mockT := &testing.T{}
	AssertFloatNear(mockT, 0.7, 0.699, 0.01)
}

func TestMatcherStrings(t *testing.T) {
	tests := []struct {
		matcher Matcher
		want    string
	}{
		{Any(), "<ANY>"},
		{AnyButNil(), "<ANY_BUT_NIL>"},
		{AnyString(), "<ANY_STRING>"},
		{AnyMap(), "<ANY_MAP>"},
		{AnySlice(), "<ANY_SLICE>"},
		{AnyFloat(), "<ANY_FLOAT>"},
	}

	for _, tt := range tests {
		if got := tt.matcher.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}
