package utils

import (
	"strings"
	"unicode/utf8"
)

// TruncateUTF8 safely truncates a string to maxLen bytes
// without breaking UTF-8 characters
func TruncateUTF8(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Find the last valid UTF-8 character boundary before maxLen
	for i := maxLen; i > 0; i-- {
		if utf8.RuneStart(s[i]) {
			return s[:i]
		}
	}

	return ""
}

// SanitizeUTF8 removes invalid UTF-8 sequences from a string
func SanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	// Replace invalid UTF-8 with valid replacement character
	return strings.ToValidUTF8(s, "")
}

// TruncateUTF8WithEllipsis safely truncates and adds ellipsis
func TruncateUTF8WithEllipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Reserve space for ellipsis
	if maxLen < 3 {
		return "..."
	}

	truncated := TruncateUTF8(s, maxLen-3)
	return truncated + "..."
}