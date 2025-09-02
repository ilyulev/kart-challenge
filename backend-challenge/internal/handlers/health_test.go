package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"
	"github.com/ilyulev/kart-challenge/backend-api/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Health(t *testing.T) {
	// Setup
	promoService := services.NewPromoCodeService()
	promoService.LoadMockPromoCodes() // Load some test data

	handler := NewHealthHandler(promoService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Health(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.HealthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "github.com/ilyulev/kart-challenge/backend-api", response.Service)
	assert.Contains(t, []string{"healthy", "degraded", "starting"}, response.Status)
	assert.Greater(t, response.PromoCodes, 0, "Should have some promo codes loaded")

	// Check promo status structure
	assert.NotEmpty(t, response.PromoStatus.Status)
	assert.NotEmpty(t, response.PromoStatus.DataSource)
	assert.Equal(t, response.PromoCodes, response.PromoStatus.CodesLoaded)
}

func TestHealthHandler_LivenessProbe(t *testing.T) {
	promoService := services.NewPromoCodeService()
	handler := NewHealthHandler(promoService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.LivenessProbe(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}

func TestHealthHandler_ReadinessProbe(t *testing.T) {
	tests := []struct {
		name           string
		setupService   func(*services.PromoCodeService)
		expectedStatus int
		expectedReady  string
	}{
		{
			name: "Ready when codes loaded",
			setupService: func(s *services.PromoCodeService) {
				s.LoadMockPromoCodes()
			},
			expectedStatus: http.StatusOK,
			expectedReady:  "ready",
		},
		{
			name: "Not ready when no codes",
			setupService: func(s *services.PromoCodeService) {
				// Don't load any codes
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedReady:  "not_ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promoService := services.NewPromoCodeService()
			tt.setupService(promoService)

			handler := NewHealthHandler(promoService)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err := handler.ReadinessProbe(c)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedReady, response["status"])
		})
	}
}
