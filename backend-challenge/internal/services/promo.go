package services

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cavaliergopher/grab/v3"
)

// PromoCodeService handles promo code validation with async download
type PromoCodeService struct {
	validCodes map[string]bool
	codesMutex sync.RWMutex // Protects validCodes map
	isLoaded   int32        // Atomic flag: 0 = loading, 1 = loaded
	loadError  error        // Last load error
	errorMutex sync.RWMutex // Protects loadError
	codesCount int32        // Atomic counter for loaded codes
}

// NewPromoCodeService creates a new promo code service
func NewPromoCodeService() *PromoCodeService {
	return &PromoCodeService{
		validCodes: make(map[string]bool),
		isLoaded:   0,
	}
}

// Initialize sets up the service and starts async download
func (p *PromoCodeService) Initialize() error {
	log.Println("Promo code service initializing...")

	// Load mock data immediately for quick startup
	p.LoadMockPromoCodes()

	// Start async download in background
	go p.downloadCodesAsync()

	log.Println("Promo code service initialized with mock data, downloading real codes in background")
	return nil
}

// downloadCodesAsync downloads coupon files in background
func (p *PromoCodeService) downloadCodesAsync() {
	log.Println("Starting background download of coupon files...")

	urls := []string{
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz",
		"https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz",
	}

	// Set loading state
	atomic.StoreInt32(&p.isLoaded, 0)

	// Clear any previous error
	p.errorMutex.Lock()
	p.loadError = nil
	p.errorMutex.Unlock()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "oolio-coupons-")
	if err != nil {
		p.setLoadError(fmt.Errorf("failed to create temp directory: %w", err))
		return
	}
	defer os.RemoveAll(tempDir)

	// Create grab client
	client := grab.NewClient()
	client.UserAgent = "OolioFoodAPI/1.0"

	// Download and process files
	fileCodes := make([]map[string]bool, len(urls))
	successCount := 0

	for i, url := range urls {
		log.Printf("Background download: processing file %d of %d", i+1, len(urls))

		codes, err := p.downloadAndProcessWithGrab(client, url, tempDir, i+1)
		if err != nil {
			log.Printf("Background download warning: failed to process %s: %v", url, err)
			continue
		}

		fileCodes[i] = codes
		successCount++
		log.Printf("Background download: successfully processed file %d (%d potential codes)", i+1, len(codes))
	}

	// Process results
	if successCount == 0 {
		p.setLoadError(fmt.Errorf("background download failed: no coupon files processed"))
		log.Println("Background download failed, continuing with mock data")
		return
	}

	// Replace mock data with real codes
	if err := p.replaceWithRealCodes(fileCodes, successCount); err != nil {
		p.setLoadError(fmt.Errorf("background processing failed: %w", err))
		log.Printf("Background processing failed: %v, continuing with mock data", err)
		return
	}

	// Mark as successfully loaded
	atomic.StoreInt32(&p.isLoaded, 1)
	log.Printf("Background download completed successfully: %d valid promo codes loaded",
		atomic.LoadInt32(&p.codesCount))
}

// downloadAndProcessWithGrab downloads a single file
func (p *PromoCodeService) downloadAndProcessWithGrab(client *grab.Client, url, tempDir string, fileNum int) (map[string]bool, error) {
	filename := filepath.Join(tempDir, fmt.Sprintf("coupon%d.gz", fileNum))
	req, err := grab.NewRequest(filename, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	// Set timeout for background download
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	req = req.WithContext(ctx)

	log.Printf("Background: starting download of file %d", fileNum)

	resp := client.Do(req)

	// Simple progress logging (less frequent for background)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if !resp.IsComplete() {
					if resp.Size() > 0 {
						log.Printf("Background file %d: %.0f%% complete",
							fileNum, 100*resp.Progress())
					} else {
						log.Printf("Background file %d: %.1f MB downloaded",
							fileNum, float64(resp.BytesComplete())/(1024*1024))
					}
				}
			case <-resp.Done:
				return
			}
		}
	}()

	if err := resp.Err(); err != nil {
		return nil, fmt.Errorf("background download failed: %w", err)
	}

	log.Printf("Background: file %d downloaded (%.1f MB)",
		fileNum, float64(resp.Size())/(1024*1024))

	return p.processGzipFile(filename)
}

// processGzipFile extracts promo codes from gzipped file
func (p *PromoCodeService) processGzipFile(filename string) (map[string]bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	codes := make(map[string]bool)
	buffer := make([]byte, 64*1024)
	wordCount := 0

	for {
		n, err := gzReader.Read(buffer)
		if n > 0 {
			text := string(buffer[:n])
			words := strings.Fields(text)

			for _, word := range words {
				cleaned := strings.ToUpper(strings.Trim(word, ".,!?;:\"'()[]{}"))
				if isValidPromoCodeFormat(cleaned) {
					codes[cleaned] = true
				}
				wordCount++
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}
	}

	return codes, nil
}

// replaceWithRealCodes replaces mock data with real coupon codes
func (p *PromoCodeService) replaceWithRealCodes(fileCodes []map[string]bool, successCount int) error {
	// Count code occurrences
	codeCount := make(map[string]int)
	for _, fileCodeMap := range fileCodes {
		for code := range fileCodeMap {
			codeCount[code]++
		}
	}

	// Determine minimum occurrences
	minOccurrences := 2
	if successCount == 1 {
		minOccurrences = 1
	}

	// Build new valid codes map
	newValidCodes := make(map[string]bool)
	for code, count := range codeCount {
		if count >= minOccurrences {
			newValidCodes[code] = true
		}
	}

	if len(newValidCodes) == 0 {
		return fmt.Errorf("no valid codes found")
	}

	// Atomically replace the codes map
	p.codesMutex.Lock()
	p.validCodes = newValidCodes
	p.codesMutex.Unlock()

	// Update counter
	atomic.StoreInt32(&p.codesCount, int32(len(newValidCodes)))

	return nil
}

// setLoadError sets the load error in a thread-safe way
func (p *PromoCodeService) setLoadError(err error) {
	p.errorMutex.Lock()
	p.loadError = err
	p.errorMutex.Unlock()
}

// loadMockPromoCodes loads initial mock data for immediate availability
func (p *PromoCodeService) LoadMockPromoCodes() {
	mockCodes := []string{
		"HAPPYHRS", "FIFTYOFF", "WELCOME1", "NEWUSER2", "DISCOUNT",
		"SAVE20PC", "FREESHIP", "SUMMER25", "AUTUMN30", "WINTER15",
		"STUDENT", "BIRTHDAY", "LOYALTY5", "REFERRAL", "COMEBACK",
	}

	p.codesMutex.Lock()
	for _, code := range mockCodes {
		p.validCodes[code] = true
	}
	p.codesMutex.Unlock()

	atomic.StoreInt32(&p.codesCount, int32(len(mockCodes)))
	log.Printf("Loaded %d mock promo codes for immediate availability", len(mockCodes))
}

// IsValidPromoCode checks if a promo code is valid (thread-safe)
func (p *PromoCodeService) IsValidPromoCode(code string) bool {
	upperCode := strings.ToUpper(code)
	if !isValidPromoCodeFormat(upperCode) {
		return false
	}

	p.codesMutex.RLock() // ✅ Keep this
	valid := p.validCodes[upperCode]
	p.codesMutex.RUnlock()

	return valid
}

// GetValidCodesCount returns the number of loaded valid codes
func (p *PromoCodeService) GetValidCodesCount() int {
	return int(atomic.LoadInt32(&p.codesCount))
}

// GetServiceStatus returns the current service status
func (p *PromoCodeService) GetServiceStatus() ServiceStatus {
	status := ServiceStatus{
		CodesLoaded:   p.GetValidCodesCount(),
		IsFullyLoaded: atomic.LoadInt32(&p.isLoaded) == 1,
	}

	p.errorMutex.RLock()
	if p.loadError != nil {
		status.LastError = p.loadError.Error()
	}
	p.errorMutex.RUnlock()

	if status.IsFullyLoaded {
		status.Status = "ready"
		status.DataSource = "remote"
	} else if status.CodesLoaded > 0 {
		status.Status = "loading"
		status.DataSource = "mock"
	} else {
		status.Status = "initializing"
		status.DataSource = "none"
	}

	return status
}

// ServiceStatus represents the current state of the promo service
type ServiceStatus struct {
	Status        string `json:"status"`              // "initializing", "loading", "ready"
	DataSource    string `json:"dataSource"`          // "none", "mock", "remote"
	CodesLoaded   int    `json:"codesLoaded"`         // Number of codes currently available
	IsFullyLoaded bool   `json:"isFullyLoaded"`       // True when real codes are loaded
	LastError     string `json:"lastError,omitempty"` // Last error if any
}

// isValidPromoCodeFormat validates promo code format
func isValidPromoCodeFormat(code string) bool {
	n := len(code)
	if n < 8 || n > 10 {
		return false
	}

	for i := 0; i < n; i++ {
		c := code[i]
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// ForceReload manually triggers a reload of coupon codes
func (p *PromoCodeService) ForceReload() {
	log.Println("Manual reload of promo codes requested")
	go p.downloadCodesAsync()
}
