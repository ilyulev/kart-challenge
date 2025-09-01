package main

import (
	"log"

	"github.com/ilyulev/kart-challenge/backend-api/internal/config"
	"github.com/ilyulev/kart-challenge/backend-api/internal/handlers"
	"github.com/ilyulev/kart-challenge/backend-api/internal/middleware"
	"github.com/ilyulev/kart-challenge/backend-api/internal/services"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize promo code service - fail fast on errors
	promoService := services.NewPromoCodeService()
	log.Println("Initializing promo code service...")

	if err := promoService.Initialize(); err != nil {
		log.Fatalf("Failed to initialize promo codes: %v", err)
		// Fail fast - better than running with broken promo validation
	}

	log.Printf("Promo code service ready with %d valid codes",
		promoService.GetValidCodesCount())

	// Create Echo instance
	e := echo.New()

	// Apply middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())
	e.Use(echomiddleware.LoggerWithConfig(echomiddleware.LoggerConfig{
		Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
	}))

	// Initialize handlers
	productHandler := handlers.NewProductHandler()
	orderHandler := handlers.NewOrderHandler(promoService)
	healthHandler := handlers.NewHealthHandler(promoService)

	// Register all routes
	registerRoutes(e, productHandler, orderHandler, healthHandler)

	// Start server
	log.Printf("Starting Echo server on port %s", cfg.Port)
	log.Printf("Access your API at: http://localhost:%s", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}

// registerRoutes centralizes all route registration
func registerRoutes(e *echo.Echo, productHandler *handlers.ProductHandler, orderHandler *handlers.OrderHandler, healthHandler *handlers.HealthHandler) {
	// API routes group - DON'T apply middleware to the entire group
	api := e.Group("/api")

	// Product routes (no auth required for GET)
	api.GET("/product", productHandler.ListProducts)
	api.GET("/product/:productId", productHandler.GetProduct)

	// Order routes (auth required) - apply auth middleware only to these routes
	api.POST("/order", orderHandler.PlaceOrder, middleware.APIKeyAuth())

	// Health check endpoints (no auth required)
	e.GET("/health", healthHandler.Health)
	e.GET("/health/live", healthHandler.LivenessProbe)
	e.GET("/health/ready", healthHandler.ReadinessProbe)
}
