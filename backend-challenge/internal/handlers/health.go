package handlers

import (
	"net/http"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"
	"github.com/ilyulev/kart-challenge/backend-api/internal/services"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	promoService *services.PromoCodeService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(promoService *services.PromoCodeService) *HealthHandler {
	return &HealthHandler{
		promoService: promoService,
	}
}

// Health returns comprehensive service health status
func (h *HealthHandler) Health(c echo.Context) error {
	// Get detailed promo service status
	promoStatus := h.promoService.GetServiceStatus()

	// Build response
	response := models.HealthResponse{
		Status:     "healthy",
		Service:    "github.com/ilyulev/kart-challenge/backend-api",
		PromoCodes: promoStatus.CodesLoaded,
		PromoStatus: models.PromoServiceStatus{
			Status:        promoStatus.Status,
			DataSource:    promoStatus.DataSource,
			CodesLoaded:   promoStatus.CodesLoaded,
			IsFullyLoaded: promoStatus.IsFullyLoaded,
			LastError:     promoStatus.LastError,
		},
	}

	// Determine overall health status
	httpStatus := http.StatusOK
	if promoStatus.CodesLoaded == 0 {
		response.Status = "starting"
		// Still return 200 OK for container health checks
	} else if !promoStatus.IsFullyLoaded && promoStatus.LastError != "" {
		response.Status = "degraded"
		// Still return 200 OK - service is functional with mock data
	}

	return c.JSON(httpStatus, response)
}

// LivenessProbe simple endpoint for container liveness checks
func (h *HealthHandler) LivenessProbe(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// ReadinessProbe endpoint for container readiness checks
func (h *HealthHandler) ReadinessProbe(c echo.Context) error {
	// Service is ready if it has any promo codes (mock or real)
	if h.promoService.GetValidCodesCount() > 0 {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ready",
		})
	}

	// Not ready yet
	return c.JSON(http.StatusServiceUnavailable, map[string]string{
		"status": "not_ready",
	})
}
