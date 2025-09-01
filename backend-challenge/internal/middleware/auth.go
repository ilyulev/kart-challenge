package middleware

import (
	"net/http"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"

	"github.com/labstack/echo/v4"
)

// APIKeyAuth middleware validates API key for protected endpoints
func APIKeyAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check API key
			apiKey := c.Request().Header.Get("api_key")
			if apiKey != "apitest" {
				return c.JSON(http.StatusUnauthorized, models.APIResponse{
					Code:    401,
					Type:    "error",
					Message: "Invalid or missing API key",
				})
			}

			return next(c)
		}
	}
}
