package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// APITestHelper provides utilities for API testing
type APITestHelper struct {
	Echo *echo.Echo
	T    *testing.T
}

// NewAPITestHelper creates a new API test helper
func NewAPITestHelper(t *testing.T, e *echo.Echo) *APITestHelper {
	return &APITestHelper{
		Echo: e,
		T:    t,
	}
}

// GET performs a GET request and returns the response
func (h *APITestHelper) GET(path string, headers ...map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)

	// Add headers if provided
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	rec := httptest.NewRecorder()
	h.Echo.ServeHTTP(rec, req)
	return rec
}

// POST performs a POST request and returns the response
func (h *APITestHelper) POST(path string, body interface{}, headers ...map[string]string) *httptest.ResponseRecorder {
	var reqBody string
	if body != nil {
		if str, ok := body.(string); ok {
			reqBody = str
		} else {
			bodyBytes, err := json.Marshal(body)
			require.NoError(h.T, err)
			reqBody = string(bodyBytes)
		}
	}

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Add headers if provided
	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	rec := httptest.NewRecorder()
	h.Echo.ServeHTTP(rec, req)
	return rec
}

// AssertJSON unmarshals response body to the provided interface
func (h *APITestHelper) AssertJSON(rec *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(rec.Body.Bytes(), v)
	require.NoError(h.T, err)
}

// AssertHealthy checks if health endpoint returns healthy status
func (h *APITestHelper) AssertHealthy() {
	rec := h.GET("/health")
	require.Equal(h.T, http.StatusOK, rec.Code)

	var health models.HealthResponse
	h.AssertJSON(rec, &health)
	require.Contains(h.T, []string{"healthy", "degraded"}, health.Status)
}

// AssertAPIKeyRequired checks if endpoint requires API key
func (h *APITestHelper) AssertAPIKeyRequired(method, path string, body interface{}) {
	var rec *httptest.ResponseRecorder

	if method == "GET" {
		rec = h.GET(path)
	} else if method == "POST" {
		rec = h.POST(path, body)
	}

	require.Equal(h.T, http.StatusUnauthorized, rec.Code)

	var apiResp models.APIResponse
	h.AssertJSON(rec, &apiResp)
	require.Equal(h.T, 401, apiResp.Code)
}
