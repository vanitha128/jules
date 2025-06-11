package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go-moon/internal/user" // Import your actual user package
)

// MockUserService is a mock type for the UserService type
type MockUserService struct {
	mock.Mock
}

// Register provides a mock function with given fields: ctx, req
func (m *MockUserService) Register(ctx context.Context, req user.RegisterRequest) (*user.User, error) {
	args := m.Called(ctx, req)
	var usr *user.User
	if args.Get(0) != nil {
		usr = args.Get(0).(*user.User)
	}
	return usr, args.Error(1)
}

// GetProfile provides a mock function with given fields: ctx, userID
func (m *MockUserService) GetProfile(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, userID)
	var usr *user.User
	if args.Get(0) != nil {
		usr = args.Get(0).(*user.User)
	}
	return usr, args.Error(1)
}

// Login provides a mock function with given fields: ctx, req
func (m *MockUserService) Login(ctx context.Context, req user.LoginRequest) (string, string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.String(1), args.Error(2)
}

// ChangePassword provides a mock function with given fields: ctx, userID, oldPassword, newPassword
func (m *MockUserService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

// UpdateProfile provides a mock function with given fields: ctx, userID, req
func (m *MockUserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req user.UpdateUserRequest) (*user.User, error) {
	args := m.Called(ctx, userID, req)
	var usr *user.User
	if args.Get(0) != nil {
		usr = args.Get(0).(*user.User)
	}
	return usr, args.Error(1)
}
