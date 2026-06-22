package agentobs

import (
	"regexp"
	"strings"
)

// SanitizeConfig configures content sanitization behavior.
type SanitizeConfig struct {
	// RemoveInternalMarkers removes framework-specific internal markers.
	RemoveInternalMarkers bool

	// RemoveMetadataBlocks removes metadata blocks like <system-reminder>.
	RemoveMetadataBlocks bool

	// RemoveSecrets removes patterns that look like secrets or API keys.
	RemoveSecrets bool

	// RemoveAnsiCodes removes ANSI escape codes from terminal output.
	RemoveAnsiCodes bool

	// MaxLength truncates content exceeding this length (0 = no limit).
	MaxLength int

	// CustomPatterns are additional regex patterns to remove.
	CustomPatterns []*regexp.Regexp
}

// DefaultSanitizeConfig returns a SanitizeConfig with sensible defaults.
func DefaultSanitizeConfig() SanitizeConfig {
	return SanitizeConfig{
		RemoveInternalMarkers: true,
		RemoveMetadataBlocks:  true,
		RemoveSecrets:         true,
		RemoveAnsiCodes:       true,
		MaxLength:             0, // No limit by default
	}
}

// Pre-compiled regular expressions for sanitization.
var (
	// Matches XML-style tags like <system-reminder>...</system-reminder>
	// Note: Go's regexp doesn't support backreferences, so we use separate patterns
	metadataBlockPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?s)<system-reminder[^>]*>.*?</system-reminder>`),
		regexp.MustCompile(`(?s)<internal[^>]*>.*?</internal>`),
		regexp.MustCompile(`(?s)<metadata[^>]*>.*?</metadata>`),
		regexp.MustCompile(`(?s)<debug[^>]*>.*?</debug>`),
		regexp.MustCompile(`(?s)<private[^>]*>.*?</private>`),
	}

	// Matches common internal markers
	internalMarkerPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?m)^\[INTERNAL\].*$`),
		regexp.MustCompile(`(?m)^\[DEBUG\].*$`),
		regexp.MustCompile(`(?m)^\[PRIVATE\].*$`),
		regexp.MustCompile(`(?s)---\s*(internal|debug|private)\s*---.*?---\s*end\s*---`),
	}

	// Matches patterns that look like API keys or secrets
	secretPatterns = []*regexp.Regexp{
		// API keys (common formats)
		regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret[_-]?key|secretkey|access[_-]?token|auth[_-]?token)\s*[:=]\s*["']?[a-zA-Z0-9_\-]{16,}["']?`),
		// Bearer tokens
		regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_\-\.]+`),
		// AWS-style keys
		regexp.MustCompile(`(?i)(AKIA|ABIA|ACCA|ASIA)[A-Z0-9]{16}`),
		// Generic long alphanumeric tokens that look like secrets
		regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["']?[^\s"']{8,}["']?`),
	}

	// ANSI escape codes
	ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

	// Multiple consecutive newlines (more than 2)
	excessiveNewlines = regexp.MustCompile(`\n{3,}`)

	// Multiple consecutive spaces (more than 2)
	excessiveSpaces = regexp.MustCompile(`  {2,}`)
)

// Sanitize removes internal markers, metadata blocks, and sensitive content
// from the input string according to the provided configuration.
func Sanitize(content string, cfg SanitizeConfig) string {
	if content == "" {
		return content
	}

	result := content

	// Remove ANSI codes first as they can interfere with other patterns
	if cfg.RemoveAnsiCodes {
		result = ansiPattern.ReplaceAllString(result, "")
	}

	// Remove metadata blocks
	if cfg.RemoveMetadataBlocks {
		for _, pattern := range metadataBlockPatterns {
			result = pattern.ReplaceAllString(result, "")
		}
	}

	// Remove internal markers
	if cfg.RemoveInternalMarkers {
		for _, pattern := range internalMarkerPatterns {
			result = pattern.ReplaceAllString(result, "")
		}
	}

	// Remove secrets
	if cfg.RemoveSecrets {
		for _, pattern := range secretPatterns {
			result = pattern.ReplaceAllStringFunc(result, func(match string) string {
				// Find the key name and redact the value
				parts := strings.SplitN(match, ":", 2)
				if len(parts) == 2 {
					return parts[0] + ": [REDACTED]"
				}
				parts = strings.SplitN(match, "=", 2)
				if len(parts) == 2 {
					return parts[0] + "=[REDACTED]"
				}
				return "[REDACTED]"
			})
		}
	}

	// Apply custom patterns
	for _, pattern := range cfg.CustomPatterns {
		result = pattern.ReplaceAllString(result, "")
	}

	// Clean up excessive whitespace
	result = excessiveNewlines.ReplaceAllString(result, "\n\n")
	result = excessiveSpaces.ReplaceAllString(result, "  ")

	// Trim whitespace
	result = strings.TrimSpace(result)

	// Apply max length
	if cfg.MaxLength > 0 && len(result) > cfg.MaxLength {
		result = result[:cfg.MaxLength-3] + "..."
	}

	return result
}

// SanitizeMap recursively sanitizes all string values in a map.
func SanitizeMap(data map[string]any, cfg SanitizeConfig) map[string]any {
	if data == nil {
		return nil
	}

	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = sanitizeValue(v, cfg)
	}
	return result
}

// sanitizeValue sanitizes a single value, recursing into maps and slices.
func sanitizeValue(v any, cfg SanitizeConfig) any {
	switch val := v.(type) {
	case string:
		return Sanitize(val, cfg)
	case map[string]any:
		return SanitizeMap(val, cfg)
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = sanitizeValue(item, cfg)
		}
		return result
	case []string:
		result := make([]string, len(val))
		for i, item := range val {
			result[i] = Sanitize(item, cfg)
		}
		return result
	default:
		return v
	}
}

// SanitizeKeys is a list of keys that should have their values redacted.
var SanitizeKeys = []string{
	"password",
	"passwd",
	"secret",
	"api_key",
	"apikey",
	"access_token",
	"auth_token",
	"bearer",
	"authorization",
	"credentials",
	"private_key",
}

// SanitizeMapKeys redacts values for sensitive keys in a map.
func SanitizeMapKeys(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	result := make(map[string]any, len(data))
	for k, v := range data {
		lowerKey := strings.ToLower(k)
		shouldRedact := false
		for _, sensitiveKey := range SanitizeKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				shouldRedact = true
				break
			}
		}

		if shouldRedact {
			result[k] = "[REDACTED]"
		} else if m, ok := v.(map[string]any); ok {
			result[k] = SanitizeMapKeys(m)
		} else {
			result[k] = v
		}
	}
	return result
}

// RemoveEmptyStrings removes empty string values from a map.
func RemoveEmptyStrings(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	result := make(map[string]any)
	for k, v := range data {
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		if m, ok := v.(map[string]any); ok {
			cleaned := RemoveEmptyStrings(m)
			if len(cleaned) > 0 {
				result[k] = cleaned
			}
		} else {
			result[k] = v
		}
	}
	return result
}
