package utils

import (
	"strconv"
	"strings"
)

// IsValidID validates if an ID is in correct format
func IsValidID(id string) bool {
	if id == "" {
		return false
	}

	// Check if it's a valid numeric ID
	if _, err := strconv.Atoi(id); err != nil {
		return false
	}

	return true
}

// IsAlphanumeric checks if string contains only letters and numbers
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

// SanitizeString removes common punctuation and converts to uppercase
func SanitizeString(s string) string {
	return strings.ToUpper(strings.Trim(s, ".,!?;:\"'()[]{}"))
}
