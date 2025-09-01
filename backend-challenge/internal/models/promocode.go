package models

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// PromoCodeService handles promo code validation
type PromoCodeService struct {
	validCodes map[string]bool
	once       sync.Once
}

func NewPromoCodeService() *PromoCodeService {
	return &PromoCodeService{
		validCodes: make(map[string]bool),
	}
}

// Init downloads and processes coupon files
func (p *PromoCodeService) Init() error {
	var initErr error
	p.once.Do(func() {
		// URLs for coupon files
		urls := []string{
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
			"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
		}

		// Store codes from each file
		fileCodes := make([]map[string]bool, len(urls))
		for i := range fileCodes {
			fileCodes[i] = make(map[string]bool)
		}

		// Download and process each file
		for i, url := range urls {
			codes, err := p.downloadAndProcessCouponFile(url)
			if err != nil {
				log.Printf("Error processing coupon file %s: %v", url, err)
				continue
			}
			fileCodes[i] = codes
		}

		// Find codes that appear in at least 2 files
		codeCount := make(map[string]int)
		for _, fileCodeMap := range fileCodes {
			for code := range fileCodeMap {
				codeCount[code]++
			}
		}

		// Store valid codes (appear in at least 2 files)
		for code, count := range codeCount {
			if count >= 2 {
				p.validCodes[code] = true
			}
		}

		log.Printf("Loaded %d valid promo codes", len(p.validCodes))
	})
	return initErr
}

// downloadAndProcessCouponFile downloads and extracts codes from a gzipped file
func (p *PromoCodeService) downloadAndProcessCouponFile(url string) (map[string]bool, error) {
	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	// Read content
	content, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}

	// Extract potential promo codes (8-10 character alphanumeric strings)
	codes := make(map[string]bool)
	text := string(content)
	words := strings.Fields(text)

	for _, word := range words {
		// Clean word (remove punctuation)
		cleaned := strings.ToUpper(strings.Trim(word, ".,!?;:\"'()[]{}"))

		// Check if it's a valid promo code format
		if len(cleaned) >= 8 && len(cleaned) <= 10 && isAlphanumeric(cleaned) {
			codes[cleaned] = true
		}
	}

	return codes, nil
}

// isAlphanumeric checks if string contains only letters and numbers
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

// IsValidPromoCode checks if a promo code is valid
func (p *PromoCodeService) IsValidPromoCode(code string) bool {
	if len(code) < 8 || len(code) > 10 {
		return false
	}
	return p.validCodes[strings.ToUpper(code)]
}
