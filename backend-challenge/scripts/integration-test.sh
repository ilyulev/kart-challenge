#!/bin/bash
set -e

BASE_URL="http://localhost:8080"
API_KEY="apitest"

echo "Starting Oolio Food API integration tests..."

# Start the server in background
echo "Starting server..."
go run ./cmd/api/main.go &
SERVER_PID=$!

# Function to cleanup
cleanup() {
    echo "Cleaning up..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

# Wait for server to start
echo "Waiting for server to start..."
sleep 8

# Test health endpoint
echo "Testing health endpoint..."
curl -f "$BASE_URL/health" || { echo "Health check failed"; exit 1; }

# Test get all products
echo "Testing get all products..."
curl -f "$BASE_URL/api/product" || { echo "Get products failed"; exit 1; }

# Test get specific product
echo "Testing get specific product..."
curl -f "$BASE_URL/api/product/1" || { echo "Get product failed"; exit 1; }

# Test get non-existent product
echo "Testing get non-existent product..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/product/999")
if [ "$HTTP_CODE" != "404" ]; then
    echo "Expected 404 for non-existent product, got $HTTP_CODE"
    exit 1
fi

# Test place order without API key
echo "Testing place order without API key..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/api/order" \
    -H "Content-Type: application/json" \
    -d '{"items":[{"productId":"1","quantity":2}]}')
if [ "$HTTP_CODE" != "401" ]; then
    echo "Expected 401 for missing API key, got $HTTP_CODE"
    exit 1
fi

# Test place valid order
echo "Testing place valid order..."
curl -f -X POST "$BASE_URL/api/order" \
    -H "Content-Type: application/json" \
    -H "api_key: $API_KEY" \
    -d '{"items":[{"productId":"1","quantity":2},{"productId":"3","quantity":1}]}' \
    || { echo "Place order failed"; exit 1; }

# Test place order with invalid product
echo "Testing place order with invalid product..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/api/order" \
    -H "Content-Type: application/json" \
    -H "api_key: $API_KEY" \
    -d '{"items":[{"productId":"999","quantity":1}]}')
if [ "$HTTP_CODE" != "422" ]; then
    echo "Expected 422 for invalid product, got $HTTP_CODE"
    exit 1
fi

# Test place order with invalid promo code
echo "Testing place order with invalid promo code..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/api/order" \
    -H "Content-Type: application/json" \
    -H "api_key: $API_KEY" \
    -d '{"items":[{"productId":"1","quantity":1}],"couponCode":"INVALID"}')
if [ "$HTTP_CODE" != "422" ]; then
    echo "Expected 422 for invalid promo code, got $HTTP_CODE"
    exit 1
fi

echo "All integration tests passed successfully!"