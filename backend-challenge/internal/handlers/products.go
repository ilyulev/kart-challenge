package handlers

import (
	"net/http"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"
	"github.com/ilyulev/kart-challenge/backend-api/pkg/utils"

	"github.com/labstack/echo/v4"
)

// ProductHandler handles product-related requests
type ProductHandler struct {
	products []models.Product
}

// NewProductHandler creates a new product handler
func NewProductHandler() *ProductHandler {
	// In production, this would come from a database
	products := []models.Product{
		{ID: "1", Name: "Chicken Waffle", Price: 12.99, Category: "Waffle"},
		{ID: "2", Name: "Belgian Waffle", Price: 10.99, Category: "Waffle"},
		{ID: "3", Name: "Pancake Stack", Price: 8.99, Category: "Pancake"},
		{ID: "4", Name: "Avocado Toast", Price: 9.99, Category: "Toast"},
		{ID: "5", Name: "Caesar Salad", Price: 11.99, Category: "Salad"},
		{ID: "6", Name: "Burger Deluxe", Price: 14.99, Category: "Burger"},
		{ID: "7", Name: "Fish & Chips", Price: 13.99, Category: "Main"},
		{ID: "8", Name: "Chocolate Cake", Price: 6.99, Category: "Dessert"},
	}

	return &ProductHandler{
		products: products,
	}
}

// ListProducts returns all available products
func (h *ProductHandler) ListProducts(c echo.Context) error {
	return c.JSON(http.StatusOK, h.products)
}

// GetProduct returns a specific product by ID
func (h *ProductHandler) GetProduct(c echo.Context) error {
	productID := c.Param("productId")

	// Validate product ID
	if !utils.IsValidID(productID) {
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Code:    400,
			Type:    "error",
			Message: "Invalid product ID format",
		})
	}

	// Find product
	for _, product := range h.products {
		if product.ID == productID {
			return c.JSON(http.StatusOK, product)
		}
	}

	// Product not found
	return c.JSON(http.StatusNotFound, models.APIResponse{
		Code:    404,
		Type:    "error",
		Message: "Product not found",
	})
}

// GetProductByID helper method for internal use
func (h *ProductHandler) GetProductByID(id string) (*models.Product, bool) {
	for _, product := range h.products {
		if product.ID == id {
			return &product, true
		}
	}
	return nil, false
}
