package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ilyulev/kart-challenge/backend-api/internal/handlers"
	"github.com/ilyulev/kart-challenge/backend-api/internal/middleware"
	"github.com/ilyulev/kart-challenge/backend-api/internal/models"
	"github.com/ilyulev/kart-challenge/backend-api/internal/services"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	echo           *echo.Echo
	promoService   *services.PromoCodeService
	productHandler *handlers.ProductHandler
	orderHandler   *handlers.OrderHandler
	healthHandler  *handlers.HealthHandler
}

func (suite *APITestSuite) SetupSuite() {
	// Initialize services
	suite.promoService = services.NewPromoCodeService()
	err := suite.promoService.Initialize()
	require.NoError(suite.T(), err)

	// Wait a moment for async initialization
	time.Sleep(100 * time.Millisecond)

	// Initialize handlers
	suite.productHandler = handlers.NewProductHandler()
	suite.orderHandler = handlers.NewOrderHandler(suite.promoService)
	suite.healthHandler = handlers.NewHealthHandler(suite.promoService)

	// Setup Echo
	suite.echo = echo.New()
	suite.echo.Use(echomiddleware.Logger())
	suite.echo.Use(echomiddleware.Recover())
	suite.setupRoutes()
}

func (suite *APITestSuite) setupRoutes() {
	// API routes
	api := suite.echo.Group("/api")
	api.GET("/product", suite.productHandler.ListProducts)
	api.GET("/product/:productId", suite.productHandler.GetProduct)
	api.POST("/order", suite.orderHandler.PlaceOrder, middleware.APIKeyAuth())

	// Health routes
	suite.echo.GET("/health", suite.healthHandler.Health)
	suite.echo.GET("/health/live", suite.healthHandler.LivenessProbe)
	suite.echo.GET("/health/ready", suite.healthHandler.ReadinessProbe)
}

func (suite *APITestSuite) TestHealthEndpoints() {
	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{"Health check", "/health", http.StatusOK},
		{"Liveness probe", "/health/live", http.StatusOK},
		{"Readiness probe", "/health/ready", http.StatusOK},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			rec := httptest.NewRecorder()

			suite.echo.ServeHTTP(rec, req)

			assert.Equal(suite.T(), tt.expectedStatus, rec.Code)
		})
	}
}

func (suite *APITestSuite) TestProductEndpoints() {
	// Test list products
	suite.Run("List all products", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/product", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		assert.Equal(suite.T(), http.StatusOK, rec.Code)

		var products []models.Product
		err := json.Unmarshal(rec.Body.Bytes(), &products)
		require.NoError(suite.T(), err)
		assert.Greater(suite.T(), len(products), 0)
	})

	// Test get specific product
	suite.Run("Get specific product", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/product/1", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		assert.Equal(suite.T(), http.StatusOK, rec.Code)

		var product models.Product
		err := json.Unmarshal(rec.Body.Bytes(), &product)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), "1", product.ID)
	})

	// Test non-existent product
	suite.Run("Get non-existent product", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/product/999", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		assert.Equal(suite.T(), http.StatusNotFound, rec.Code)
	})
}

func (suite *APITestSuite) TestOrderEndpoints() {
	// Test order without API key
	suite.Run("Order without API key should fail", func() {
		orderReq := models.OrderRequest{
			Items: []models.OrderItem{
				{ProductID: "1", Quantity: 2},
			},
		}

		body, err := json.Marshal(orderReq)
		require.NoError(suite.T(), err)

		req := httptest.NewRequest(http.MethodPost, "/api/order", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, rec.Code)
	})

	// Test valid order with API key
	suite.Run("Valid order with API key", func() {
		orderReq := models.OrderRequest{
			Items: []models.OrderItem{
				{ProductID: "1", Quantity: 2},
				{ProductID: "2", Quantity: 1},
			},
		}

		body, err := json.Marshal(orderReq)
		require.NoError(suite.T(), err)

		req := httptest.NewRequest(http.MethodPost, "/api/order", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api_key", "apitest")
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		assert.Equal(suite.T(), http.StatusOK, rec.Code)

		var order models.Order
		err = json.Unmarshal(rec.Body.Bytes(), &order)
		require.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), order.ID)
		assert.Len(suite.T(), order.Items, 2)
		assert.Len(suite.T(), order.Products, 2)
	})

	// Test order with valid promo code
	suite.Run("Order with valid promo code", func() {
		orderReq := models.OrderRequest{
			Items: []models.OrderItem{
				{ProductID: "1", Quantity: 1},
			},
			CouponCode: "HAPPYHRS", // This should be in mock data
		}

		body, err := json.Marshal(orderReq)
		require.NoError(suite.T(), err)

		req := httptest.NewRequest(http.MethodPost, "/api/order", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api_key", "apitest")
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	})

	// Test order with invalid promo code
	suite.Run("Order with invalid promo code", func() {
		orderReq := models.OrderRequest{
			Items: []models.OrderItem{
				{ProductID: "1", Quantity: 1},
			},
			CouponCode: "INVALIDCODE",
		}

		body, err := json.Marshal(orderReq)
		require.NoError(suite.T(), err)

		req := httptest.NewRequest(http.MethodPost, "/api/order", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api_key", "apitest")
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		assert.Equal(suite.T(), http.StatusUnprocessableEntity, rec.Code)
	})
}

func (suite *APITestSuite) TestFullOrderWorkflow() {
	suite.Run("Complete order workflow", func() {
		// Step 1: List products to see what's available
		req := httptest.NewRequest(http.MethodGet, "/api/product", nil)
		rec := httptest.NewRecorder()
		suite.echo.ServeHTTP(rec, req)

		require.Equal(suite.T(), http.StatusOK, rec.Code)

		var products []models.Product
		err := json.Unmarshal(rec.Body.Bytes(), &products)
		require.NoError(suite.T(), err)
		require.Greater(suite.T(), len(products), 1)

		// Step 2: Get details of first product
		req = httptest.NewRequest(http.MethodGet, "/api/product/"+products[0].ID, nil)
		rec = httptest.NewRecorder()
		suite.echo.ServeHTTP(rec, req)

		require.Equal(suite.T(), http.StatusOK, rec.Code)

		// Step 3: Place order with selected products
		orderReq := models.OrderRequest{
			Items: []models.OrderItem{
				{ProductID: products[0].ID, Quantity: 2},
				{ProductID: products[1].ID, Quantity: 1},
			},
			CouponCode: "HAPPYHRS", // Valid mock promo code
		}

		body, err := json.Marshal(orderReq)
		require.NoError(suite.T(), err)

		req = httptest.NewRequest(http.MethodPost, "/api/order", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api_key", "apitest")
		rec = httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		require.Equal(suite.T(), http.StatusOK, rec.Code)

		var order models.Order
		err = json.Unmarshal(rec.Body.Bytes(), &order)
		require.NoError(suite.T(), err)

		// Verify order structure
		assert.NotEmpty(suite.T(), order.ID)
		assert.Len(suite.T(), order.Items, 2)
		assert.Len(suite.T(), order.Products, 2)
		assert.Equal(suite.T(), orderReq.Items[0].ProductID, order.Items[0].ProductID)
		assert.Equal(suite.T(), orderReq.Items[0].Quantity, order.Items[0].Quantity)
	})
}

func TestAPIIntegration(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
