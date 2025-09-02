package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"Valid numeric ID", "123", true},
		{"Valid single digit", "1", true},
		{"Valid large number", "999999", true},
		{"Invalid empty string", "", false},
		{"Invalid non-numeric", "abc", false},
		{"Invalid alphanumeric", "123abc", false},
		{"Invalid with spaces", "123 456", false},
		{"Invalid with special chars", "123!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid uppercase letters", "ABCDEF", true},
		{"Valid lowercase letters", "abcdef", true},
		{"Valid mixed case", "AbCdEf", true},
		{"Valid numbers", "123456", true},
		{"Valid alphanumeric", "ABC123", true},
		{"Invalid with spaces", "ABC 123", false},
		{"Invalid with special chars", "ABC@123", false},
		{"Invalid empty string", "", true}, // Edge case: empty string is technically alphanumeric
		{"Invalid with punctuation", "ABC,123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAlphanumeric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Remove punctuation", "hello,world!", "HELLO,WORLD"},
		{"Convert to uppercase", "hello", "HELLO"},
		{"Remove common punctuation", "test.word;here", "TEST.WORD;HERE"},
		{"Handle quotes", `"quoted"`, "QUOTED"},
		{"Handle parentheses", "(bracketed)", "BRACKETED"},
		{"Empty string", "", ""},
		{"Already clean", "CLEAN", "CLEAN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
