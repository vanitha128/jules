package todo

import (
	"context"

	"github.com/google/uuid"
)

// TodoRepository defines the interface for database operations related to todos.
// The actual implementation (e.g., for PostgreSQL using GORM) will be in a different package,
// likely under internal/database.
type TodoRepository interface {
	CreateTodo(ctx context.Context, todo *Todo) error
	UpdateTodo(ctx context.Context, todo *Todo) error
	GetTodosByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error)
	GetTodoByID(ctx context.Context, todoID uuid.UUID) (*Todo, error) // Should this also take userID for pre-filtering?
	                                                                // For now, service layer handles ownership check after fetch.
	DeleteTodo(ctx context.Context, todoID uuid.UUID) error
}

// Note: The Todo model is defined in todo.go within this package, so it's directly accessible.
// If TodoRepository were in a different package (e.g., internal/database),
// it would need to import the todo model (e.g., "github.com/your-username/go-moon/internal/todo").
