package agentobs

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		cfg      SanitizeConfig
		contains []string
		excludes []string
	}{
		{
			name:  "removes metadata blocks",
			input: "Hello <system-reminder>secret stuff</system-reminder> world",
			cfg: SanitizeConfig{
				RemoveMetadataBlocks: true,
			},
			contains: []string{"Hello", "world"},
			excludes: []string{"secret stuff", "system-reminder"},
		},
		{
			name:  "removes internal markers",
			input: "Normal text\n[INTERNAL] secret\n[DEBUG] debug info\nMore text",
			cfg: SanitizeConfig{
				RemoveInternalMarkers: true,
			},
			contains: []string{"Normal text", "More text"},
			excludes: []string{"[INTERNAL]", "secret", "[DEBUG]"},
		},
		{
			name:  "redacts API keys",
			input: "api_key: sk-1234567890abcdef1234567890",
			cfg: SanitizeConfig{
				RemoveSecrets: true,
			},
			contains: []string{"api_key:", "[REDACTED]"},
			excludes: []string{"sk-1234567890abcdef1234567890"},
		},
		{
			name:  "redacts bearer tokens",
			input: "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.payload.signature",
			cfg: SanitizeConfig{
				RemoveSecrets: true,
			},
			contains: []string{"[REDACTED]"},
			excludes: []string{"eyJhbGciOiJIUzI1NiJ9"},
		},
		{
			name:  "removes ANSI codes",
			input: "\x1b[31mRed text\x1b[0m normal",
			cfg: SanitizeConfig{
				RemoveAnsiCodes: true,
			},
			contains: []string{"Red text", "normal"},
			excludes: []string{"\x1b[31m", "\x1b[0m"},
		},
		{
			name:  "truncates long content",
			input: strings.Repeat("a", 100),
			cfg: SanitizeConfig{
				MaxLength: 50,
			},
			contains: []string{"..."},
		},
		{
			name:     "collapses excessive newlines",
			input:    "line1\n\n\n\n\nline2",
			cfg:      SanitizeConfig{},
			contains: []string{"line1\n\nline2"},
			excludes: []string{"\n\n\n"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Sanitize(tc.input, tc.cfg)

			for _, s := range tc.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected result to contain %q, got %q", s, result)
				}
			}

			for _, s := range tc.excludes {
				if strings.Contains(result, s) {
					t.Errorf("expected result to NOT contain %q, got %q", s, result)
				}
			}
		})
	}
}

func TestSanitizeEmpty(t *testing.T) {
	result := Sanitize("", DefaultSanitizeConfig())
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestSanitizeMap(t *testing.T) {
	cfg := SanitizeConfig{
		RemoveSecrets: true,
	}

	input := map[string]any{
		"message": "api_key: sk-secret123456789012",
		"nested": map[string]any{
			"password": "password: mypassword123",
		},
		"list": []any{
			"bearer token123456789012345",
		},
		"number": 42,
	}

	result := SanitizeMap(input, cfg)

	// Check that secrets are redacted
	if msg, ok := result["message"].(string); ok {
		if strings.Contains(msg, "sk-secret") {
			t.Error("api key should be redacted")
		}
	}

	// Check nested maps
	if nested, ok := result["nested"].(map[string]any); ok {
		if pwd, ok := nested["password"].(string); ok {
			if strings.Contains(pwd, "mypassword123") {
				t.Error("password should be redacted")
			}
		}
	}

	// Check number is preserved
	if num, ok := result["number"].(int); !ok || num != 42 {
		t.Error("number should be preserved")
	}
}

func TestSanitizeMapNil(t *testing.T) {
	result := SanitizeMap(nil, DefaultSanitizeConfig())
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

func TestSanitizeMapKeys(t *testing.T) {
	input := map[string]any{
		"username": "john",
		"password": "secret123",
		"api_key":  "sk-12345",
		"data": map[string]any{
			"authorization": "Bearer token",
			"normal_field":  "value",
		},
	}

	result := SanitizeMapKeys(input)

	if result["username"] != "john" {
		t.Error("username should be preserved")
	}
	if result["password"] != "[REDACTED]" {
		t.Error("password should be redacted")
	}
	if result["api_key"] != "[REDACTED]" {
		t.Error("api_key should be redacted")
	}

	nested := result["data"].(map[string]any)
	if nested["authorization"] != "[REDACTED]" {
		t.Error("authorization should be redacted")
	}
	if nested["normal_field"] != "value" {
		t.Error("normal_field should be preserved")
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	input := map[string]any{
		"filled": "value",
		"empty":  "",
		"number": 42,
		"nested": map[string]any{
			"also_empty": "",
			"has_value":  "yes",
		},
		"all_empty": map[string]any{
			"empty1": "",
		},
	}

	result := RemoveEmptyStrings(input)

	if result["filled"] != "value" {
		t.Error("filled should be preserved")
	}
	if _, exists := result["empty"]; exists {
		t.Error("empty should be removed")
	}
	if result["number"] != 42 {
		t.Error("number should be preserved")
	}

	nested := result["nested"].(map[string]any)
	if _, exists := nested["also_empty"]; exists {
		t.Error("also_empty should be removed from nested")
	}
	if nested["has_value"] != "yes" {
		t.Error("has_value should be preserved in nested")
	}

	// all_empty should be removed since it becomes empty
	if _, exists := result["all_empty"]; exists {
		t.Error("all_empty map should be removed when empty")
	}
}

func TestDefaultSanitizeConfig(t *testing.T) {
	cfg := DefaultSanitizeConfig()

	if !cfg.RemoveInternalMarkers {
		t.Error("default should remove internal markers")
	}
	if !cfg.RemoveMetadataBlocks {
		t.Error("default should remove metadata blocks")
	}
	if !cfg.RemoveSecrets {
		t.Error("default should remove secrets")
	}
	if !cfg.RemoveAnsiCodes {
		t.Error("default should remove ANSI codes")
	}
	if cfg.MaxLength != 0 {
		t.Error("default should have no max length")
	}
}
