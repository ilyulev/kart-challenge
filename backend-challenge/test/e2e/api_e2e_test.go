package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ilyulev/kart-challenge/backend-api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	baseURL string
	client  *http.Client
}

func (suite *E2ETestSuite) SetupSuite() {
	// Check if we should run E2E tests
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		suite.T().Skip("E2E tests disabled. Set RUN_E2E_TESTS=true to enable")
	}

	// Get base URL from environment or use default
	suite.baseURL = os.Getenv("API_BASE_URL")
	if suite.baseURL == "" {
		suite.baseURL = "http://localhost:8080"
	}

	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Wait for service to be ready
	suite.waitForService()
}

func (suite *E2ETestSuite) waitForService() {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := suite.client.Get(suite.baseURL + "/health/ready")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	suite.T().Fatal("Service did not become ready within timeout")
}

func (suite *E2ETestSuite) TestHealthEndpoints() {
	endpoints := []string{"/health", "/health/live", "/health/ready"}

	for _, endpoint := range endpoints {
		suite.Run(fmt.Sprintf("GET %s", endpoint), func() {
			resp, err := suite.client.Get(suite.baseURL + endpoint)
			require.NoError(suite.T(), err)
			defer resp.Body.Close()

			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		})
	}
}

func (suite *E2ETestSuite) TestProductsAPI() {
	// Test list products
	suite.Run("List products", func() {
		resp, err := suite.client.Get(suite.baseURL + "/api/product")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var products []models.Product
		err = json.NewDecoder(resp.Body).Decode(&products)
		require.NoError(suite.T(), err)
		assert.Greater(suite.T(), len(products), 0)
	})

	// Test get specific product
	suite.Run("Get product by ID", func() {
		resp, err := suite.client.Get(suite.baseURL + "/api/product/1")
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var product models.Product
		err = json.NewDecoder(resp.Body).Decode(&product)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), "1", product.ID)
	})
}

func (suite *E2ETestSuite) TestOrdersAPI() {
	// Test order without API key
	suite.Run("Order without API key should fail", func() {
		orderReq := models.OrderRequest{
			Items: []models.OrderItem{
				{ProductID: "1", Quantity: 1},
			},
		}

		body, err := json.Marshal(orderReq)
		require.NoError(suite.T(), err)

		resp, err := suite.client.Post(
			suite.baseURL+"/api/order",
			"application/json",
			bytes.NewBuffer(body),
		)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	})

	// Test order with API key
	suite.Run("Order with valid API key", func() {
		orderReq := models.OrderRequest{
			Items: []models.OrderItem{
				{ProductID: "1", Quantity: 2},
				{ProductID: "2", Quantity: 1},
			},
		}

		body, err := json.Marshal(orderReq)
		require.NoError(suite.T(), err)

		req, err := http.NewRequest("POST", suite.baseURL+"/api/order", bytes.NewBuffer(body))
		require.NoError(suite.T(), err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api_key", "apitest")

		resp, err := suite.client.Do(req)
		require.NoError(suite.T(), err)
		defer resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var order models.Order
		err = json.NewDecoder(resp.Body).Decode(&order)
		require.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), order.ID)
	})
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
