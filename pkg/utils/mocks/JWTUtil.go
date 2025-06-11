package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/google/uuid"
	"go-moon/pkg/utils" // Import your actual utils package
)

// MockJWTUtil is a mock type for the JWTUtil type
type MockJWTUtil struct {
	mock.Mock
}

// GenerateAccessToken provides a mock function.
func (m *MockJWTUtil) GenerateAccessToken(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// GenerateRefreshToken provides a mock function.
func (m *MockJWTUtil) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// ValidateToken provides a mock function.
func (m *MockJWTUtil) ValidateToken(tokenString string) (*utils.Claims, error) {
	args := m.Called(tokenString)
	var claims *utils.Claims
	if args.Get(0) != nil {
		claims = args.Get(0).(*utils.Claims)
	}
	return claims, args.Error(1)
}

// NewJWTUtil is not typically mocked this way as it's a constructor.
// Tests would use this mock by instantiating it directly: `mockJWT := new(mocks.MockJWTUtil)`
// and then setting up expectations on its methods.
// The actual NewJWTUtil constructor from the utils package would be used to create real instances if needed.
