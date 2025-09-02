package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyulev/kart-challenge/backend-api/internal/middleware"
	"github.com/ilyulev/kart-challenge/backend-api/internal/models"
	"github.com/ilyulev/kart-challenge/backend-api/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderHandler_PlaceOrder(t *testing.T) {
	// Setup mock promo service
	promoService := services.NewPromoCodeService()
	promoService.LoadMockPromoCodes()

	handler := NewOrderHandler(promoService)

	tests := []struct {
		name           string
		requestBody    models.OrderRequest
		apiKey         string
		expectedStatus int
		shouldHaveID   bool
	}{
		{
			name: "Valid order with valid promo code",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 2},
					{ProductID: "2", Quantity: 1},
				},
				CouponCode: "HAPPYHRS",
			},
			apiKey:         "apitest",
			expectedStatus: http.StatusOK,
			shouldHaveID:   true,
		},
		{
			name: "Valid order without promo code",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 1},
				},
			},
			apiKey:         "apitest",
			expectedStatus: http.StatusOK,
			shouldHaveID:   true,
		},
		{
			name: "Invalid API key",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 1},
				},
			},
			apiKey:         "wrongkey",
			expectedStatus: http.StatusUnauthorized,
			shouldHaveID:   false,
		},
		{
			name: "Missing API key",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 1},
				},
			},
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
			shouldHaveID:   false,
		},
		{
			name: "Empty items list",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{},
			},
			apiKey:         "apitest",
			expectedStatus: http.StatusUnprocessableEntity,
			shouldHaveID:   false,
		},
		{
			name: "Invalid product ID",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "999", Quantity: 1},
				},
			},
			apiKey:         "apitest",
			expectedStatus: http.StatusUnprocessableEntity,
			shouldHaveID:   false,
		},
		{
			name: "Invalid quantity",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 0},
				},
			},
			apiKey:         "apitest",
			expectedStatus: http.StatusUnprocessableEntity,
			shouldHaveID:   false,
		},
		{
			name: "Invalid promo code",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{ProductID: "1", Quantity: 1},
				},
				CouponCode: "INVALID",
			},
			apiKey:         "apitest",
			expectedStatus: http.StatusUnprocessableEntity,
			shouldHaveID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup request
			bodyBytes, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/order", bytes.NewBuffer(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.apiKey != "" {
				req.Header.Set("api_key", tt.apiKey)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			handlerChain := middleware.APIKeyAuth()(handler.PlaceOrder)
			err = handlerChain(c) // ✅ Middleware runs first!

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK && tt.shouldHaveID {
				var order models.Order
				err = json.Unmarshal(rec.Body.Bytes(), &order)
				require.NoError(t, err)
				assert.NotEmpty(t, order.ID, "Order ID should not be empty")
				assert.Equal(t, len(tt.requestBody.Items), len(order.Items), "Order items should match request items")
			}
		})
	}
}

func TestOrderHandler_generateOrderID(t *testing.T) {
	promoService := services.NewPromoCodeService()
	handler := NewOrderHandler(promoService)

	// Generate multiple IDs to test uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := handler.generateOrderID()

		// Check format
		assert.Regexp(t, `^ORD-\d{6}-\d{6}-[A-F0-9]{8}$`, id, "Order ID should match expected format")

		// Check uniqueness
		assert.False(t, ids[id], "Order ID should be unique")
		ids[id] = true
	}
}
