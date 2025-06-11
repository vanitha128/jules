package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go-moon/internal/cache" // Import your actual cache package
)

// MockCache is a mock type for the Cache type
type MockCache struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, key
func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	// Special handling for cache.ErrNotFound if it's the expected error
	if args.Get(1) != nil && args.Error(1).Error() == cache.ErrNotFound.Error() {
		return args.String(0), cache.ErrNotFound
	}
	return args.String(0), args.Error(1)
}

// Set provides a mock function with given fields: ctx, key, value, expiration
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

// Delete provides a mock function with given fields: ctx, key
func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Close provides a mock function with given fields:
func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}
