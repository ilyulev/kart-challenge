package services

import (
	"testing"
)

func TestIsValidPromoCodeFormat(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		// Valid cases
		{"Valid 8 chars", "HAPPYHRS", true},
		{"Valid 9 chars", "FIFTYOFF1", true},
		{"Valid 10 chars", "DISCOUNT10", true},
		{"Valid with numbers", "SAVE20OFF", true},
		{"Valid all letters", "ABCDEFGH", true},
		{"Valid all caps", "NEWUSER", false}, // Too short

		// Invalid cases
		{"Too short", "SHORT", false},
		{"Too long", "VERYLONGCODE", false},
		{"Has lowercase", "happyhrs", false}, // Should be uppercase
		{"Has special chars", "SAVE@20%", false},
		{"Has space", "SAVE 20", false},
		{"Empty string", "", false},
		{"Numbers only", "12345678", true},
		{"Letters only", "ABCDEFGH", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPromoCodeFormat(tt.code)
			if result != tt.expected {
				t.Errorf("isValidPromoCodeFormat(%q) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

func TestPromoCodeService_IsValidPromoCode(t *testing.T) {
	service := NewPromoCodeService()

	// Manually add some test codes
	service.validCodes["HAPPYHRS"] = true
	service.validCodes["FIFTYOFF"] = true
	service.validCodes["TESTCODE"] = true

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"Valid existing code", "HAPPYHRS", true},
		{"Valid existing code lowercase - converted", "happyhrs", true},  // IsValidPromoCode converts to uppercase
		{"Valid existing code mixed case - converted", "FiftyOff", true}, // IsValidPromoCode converts to uppercase
		{"Invalid format too short", "SHORT", false},
		{"Invalid format too long", "VERYLONGCODE", false},
		{"Valid format but not in database", "NOTFOUND", false},
		{"Empty string", "", false},
		{"Valid code with numbers", "TESTCODE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsValidPromoCode(tt.code)
			if result != tt.expected {
				t.Errorf("IsValidPromoCode(%q) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

func TestPromoCodeService_GetServiceStatus(t *testing.T) {
	service := NewPromoCodeService()

	// Test initial state
	status := service.GetServiceStatus()
	if status.CodesLoaded != 0 {
		t.Errorf("Expected 0 codes initially, got %d", status.CodesLoaded)
	}
	if status.Status != "initializing" {
		t.Errorf("Expected 'initializing' status, got %s", status.Status)
	}

	// Add mock codes
	service.LoadMockPromoCodes()

	status = service.GetServiceStatus()
	if status.CodesLoaded == 0 {
		t.Error("Expected codes to be loaded after loadMockPromoCodes()")
	}
	if status.DataSource != "mock" {
		t.Errorf("Expected 'mock' data source, got %s", status.DataSource)
	}
}
