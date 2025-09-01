package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"
	"github.com/ilyulev/kart-challenge/backend-api/internal/services"
	"github.com/ilyulev/kart-challenge/backend-api/pkg/utils"

	"github.com/labstack/echo/v4"
)

// OrderHandler handles order-related requests
type OrderHandler struct {
	promoService   *services.PromoCodeService
	productHandler *ProductHandler
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(promoService *services.PromoCodeService) *OrderHandler {
	return &OrderHandler{
		promoService:   promoService,
		productHandler: NewProductHandler(),
	}
}

// PlaceOrder processes a new order
func (h *OrderHandler) PlaceOrder(c echo.Context) error {
	// Parse request body
	var orderReq models.OrderRequest
	if err := c.Bind(&orderReq); err != nil {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Type:    "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := h.validateOrderRequest(&orderReq); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, models.APIResponse{
			Code:    422,
			Type:    "error",
			Message: err.Error(),
		})
	}

	// Validate and collect products
	orderProducts, err := h.validateAndCollectProducts(orderReq.Items)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, models.APIResponse{
			Code:    422,
			Type:    "error",
			Message: err.Error(),
		})
	}

	// Validate promo code if provided
	if orderReq.CouponCode != "" {
		if !h.promoService.IsValidPromoCode(orderReq.CouponCode) {
			return c.JSON(http.StatusUnprocessableEntity, models.APIResponse{
				Code:    422,
				Type:    "error",
				Message: "Invalid promo code",
			})
		}
	}

	// Generate order ID
	orderID := h.generateOrderID()

	// Create order
	order := models.Order{
		ID:       orderID,
		Items:    orderReq.Items,
		Products: orderProducts,
	}

	// Return successful order
	return c.JSON(http.StatusOK, order)
}

// validateOrderRequest validates the order request
func (h *OrderHandler) validateOrderRequest(orderReq *models.OrderRequest) error {
	if len(orderReq.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	for _, item := range orderReq.Items {
		if item.ProductID == "" {
			return fmt.Errorf("productId is required for all items")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than 0")
		}
		if !utils.IsValidID(item.ProductID) {
			return fmt.Errorf("invalid productId format: %s", item.ProductID)
		}
	}

	return nil
}

// validateAndCollectProducts validates items and collects corresponding products
func (h *OrderHandler) validateAndCollectProducts(items []models.OrderItem) ([]models.Product, error) {
	var orderProducts []models.Product

	for _, item := range items {
		product, exists := h.productHandler.GetProductByID(item.ProductID)
		if !exists {
			return nil, fmt.Errorf("product with ID %s not found", item.ProductID)
		}
		orderProducts = append(orderProducts, *product)
	}

	return orderProducts, nil
}

// generateOrderID creates unique IDs without requiring mutex
func (h *OrderHandler) generateOrderID() string {
	// Option 1: Timestamp + Random (recommended for this use case)
	timestamp := time.Now().Format("060102-150405") // YYMMDD-HHMMSS

	// Generate 4 random bytes
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based if random fails (very unlikely)
		return fmt.Sprintf("ORD-%s-FALLBACK", timestamp)
	}

	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("ORD-%s-%s", timestamp, strings.ToUpper(randomHex))
}
