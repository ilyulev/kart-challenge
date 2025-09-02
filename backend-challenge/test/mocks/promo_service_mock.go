package mocks

import (
	"github.com/ilyulev/kart-challenge/backend-api/internal/services"
	"github.com/stretchr/testify/mock"
)

// MockPromoCodeService is a mock implementation of PromoCodeService
type MockPromoCodeService struct {
	mock.Mock
}

func (m *MockPromoCodeService) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPromoCodeService) IsValidPromoCode(code string) bool {
	args := m.Called(code)
	return args.Bool(0)
}

func (m *MockPromoCodeService) GetValidCodesCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPromoCodeService) GetServiceStatus() services.ServiceStatus {
	args := m.Called()
	return args.Get(0).(services.ServiceStatus)
}

func (m *MockPromoCodeService) ForceReload() {
	m.Called()
}
