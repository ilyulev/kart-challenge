package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductHandler_ListProducts(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/product", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := NewProductHandler()

	// Execute
	err := handler.ListProducts(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var products []models.Product
	err = json.Unmarshal(rec.Body.Bytes(), &products)
	require.NoError(t, err)

	assert.Greater(t, len(products), 0, "Should return at least one product")

	// Check first product structure
	if len(products) > 0 {
		product := products[0]
		assert.NotEmpty(t, product.ID, "Product ID should not be empty")
		assert.NotEmpty(t, product.Name, "Product name should not be empty")
		assert.Greater(t, product.Price, 0.0, "Product price should be greater than 0")
		assert.NotEmpty(t, product.Category, "Product category should not be empty")
	}
}

func TestProductHandler_GetProduct(t *testing.T) {
	tests := []struct {
		name           string
		productID      string
		expectedStatus int
		shouldHaveBody bool
	}{
		{"Valid product ID", "1", http.StatusOK, true},
		{"Another valid product ID", "2", http.StatusOK, true},
		{"Invalid product ID", "999", http.StatusNotFound, true},
		{"Non-numeric product ID", "abc", http.StatusBadRequest, true},
		{"Empty product ID", "", http.StatusBadRequest, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/product/"+tt.productID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("productId")
			c.SetParamValues(tt.productID)

			handler := NewProductHandler()

			// Execute
			err := handler.GetProduct(c)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.shouldHaveBody {
				assert.Greater(t, len(rec.Body.Bytes()), 0, "Response should have a body")
			}

			if tt.expectedStatus == http.StatusOK {
				var product models.Product
				err = json.Unmarshal(rec.Body.Bytes(), &product)
				require.NoError(t, err)
				assert.Equal(t, tt.productID, product.ID)
			}
		})
	}
}

func TestProductHandler_GetProductByID(t *testing.T) {
	handler := NewProductHandler()

	// Test existing product
	product, exists := handler.GetProductByID("1")
	assert.True(t, exists, "Product with ID '1' should exist")
	assert.NotNil(t, product, "Product should not be nil")
	assert.Equal(t, "1", product.ID)

	// Test non-existing product
	product, exists = handler.GetProductByID("999")
	assert.False(t, exists, "Product with ID '999' should not exist")
	assert.Nil(t, product, "Product should be nil for non-existing ID")
}
