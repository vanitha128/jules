package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/go-moon/internal/todo" // Import your actual todo package
)

// MockTodoRepository is a mock type for the TodoRepository type
type MockTodoRepository struct {
	mock.Mock
}

// CreateTodo provides a mock function with given fields: ctx, t
func (m *MockTodoRepository) CreateTodo(ctx context.Context, t *todo.Todo) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

// UpdateTodo provides a mock function with given fields: ctx, t
func (m *MockTodoRepository) UpdateTodo(ctx context.Context, t *todo.Todo) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

// GetTodosByUserID provides a mock function with given fields: ctx, userID
func (m *MockTodoRepository) GetTodosByUserID(ctx context.Context, userID uuid.UUID) ([]todo.Todo, error) {
	args := m.Called(ctx, userID)
	var todos []todo.Todo
	if args.Get(0) != nil {
		todos = args.Get(0).([]todo.Todo)
	}
	return todos, args.Error(1)
}

// GetTodoByID provides a mock function with given fields: ctx, todoID
func (m *MockTodoRepository) GetTodoByID(ctx context.Context, todoID uuid.UUID) (*todo.Todo, error) {
	args := m.Called(ctx, todoID)
	var td *todo.Todo
	if args.Get(0) != nil {
		td = args.Get(0).(*todo.Todo)
	}
	return td, args.Error(1)
}

// DeleteTodo provides a mock function with given fields: ctx, todoID
func (m *MockTodoRepository) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	args := m.Called(ctx, todoID)
	return args.Error(0)
}
