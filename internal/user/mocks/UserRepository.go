package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go-moon/internal/user" // Import your actual user package
)

// MockUserRepository is a mock type for the UserRepository type
type MockUserRepository struct {
	mock.Mock
}

// CreateUser provides a mock function with given fields: ctx, u
func (m *MockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

// GetUserByEmail provides a mock function with given fields: ctx, email
func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	var usr *user.User
	if args.Get(0) != nil {
		usr = args.Get(0).(*user.User)
	}
	return usr, args.Error(1)
}

// GetUserByID provides a mock function with given fields: ctx, userID
func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, userID)
	var usr *user.User
	if args.Get(0) != nil {
		usr = args.Get(0).(*user.User)
	}
	return usr, args.Error(1)
}

// UpdateUser provides a mock function with given fields: ctx, u
func (m *MockUserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
