package models

// Product represents a food item available for order
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// OrderRequest represents the request body for placing an order
type OrderRequest struct {
	CouponCode string      `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items"`
}

// Order represents a completed order
type Order struct {
	ID       string      `json:"id"`
	Items    []OrderItem `json:"items"`
	Products []Product   `json:"products"`
}

// APIResponse represents a standard API error response
type APIResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string             `json:"status"`      // "healthy", "degraded", "starting"
	Service     string             `json:"service"`     // Service name
	PromoCodes  int                `json:"promoCodes"`  // Number of loaded codes
	PromoStatus PromoServiceStatus `json:"promoStatus"` // Detailed promo service status
}

// PromoServiceStatus represents detailed promo service status
type PromoServiceStatus struct {
	Status        string `json:"status"`              // "initializing", "loading", "ready"
	DataSource    string `json:"dataSource"`          // "none", "mock", "remote"
	CodesLoaded   int    `json:"codesLoaded"`         // Number of codes available
	IsFullyLoaded bool   `json:"isFullyLoaded"`       // True when real codes loaded
	LastError     string `json:"lastError,omitempty"` // Last error if any
}
