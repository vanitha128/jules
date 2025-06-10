package todo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	// "github.com/your-username/go-moon/internal/user" // Might be needed for user ownership checks later if complex
)

// Errors for Todo service
var (
	ErrTodoNotFound     = errors.New("todo not found")
	ErrTodoAccessDenied = errors.New("user does not have permission to access this todo")
)

// TodoService defines the interface for todo-related business logic.
// This should match the interface in handler.go or be a single canonical definition.
type TodoService interface {
	CreateTodo(ctx context.Context, userID uuid.UUID, req CreateTodoRequest) (*Todo, error)
	UpdateTodo(ctx context.Context, userID uuid.UUID, todoID uuid.UUID, req UpdateTodoRequest) (*Todo, error)
	ListTodosByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error)
	GetTodoByID(ctx context.Context, userID uuid.UUID, todoID uuid.UUID) (*Todo, error)
	DeleteTodo(ctx context.Context, userID uuid.UUID, todoID uuid.UUID) error
}

type todoService struct {
	todoRepo TodoRepository // Dependency on TodoRepository
}

// NewTodoService creates a new TodoService.
func NewTodoService(todoRepo TodoRepository) TodoService {
	return &todoService{todoRepo: todoRepo}
}

// CreateTodo creates a new todo item.
func (s *todoService) CreateTodo(ctx context.Context, userID uuid.UUID, req CreateTodoRequest) (*Todo, error) {
	newTodo := &Todo{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		IsDone:      false, // Default to not done
		UserID:      userID,
		// CreatedAt and UpdatedAt will be set by GORM or database
	}

	err := s.todoRepo.CreateTodo(ctx, newTodo)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo in repository: %w", err)
	}
	return newTodo, nil
}

// UpdateTodo updates an existing todo item.
// It also ensures that the user attempting the update owns the todo.
func (s *todoService) UpdateTodo(ctx context.Context, userID uuid.UUID, todoID uuid.UUID, req UpdateTodoRequest) (*Todo, error) {
	todo, err := s.todoRepo.GetTodoByID(ctx, todoID)
	if err != nil {
		// Could be a generic DB error or a "not found" error from the repo.
		// The repo should ideally return a specific ErrNotFound.
		return nil, ErrTodoNotFound // Assuming repo might return an error that leads to this.
	}

	if todo.UserID != userID {
		return nil, ErrTodoAccessDenied
	}

	updated := false
	if req.Title != nil {
		todo.Title = *req.Title
		updated = true
	}
	if req.Description != nil {
		todo.Description = *req.Description
		updated = true
	}
	if req.DueDate != nil {
		todo.DueDate = *req.DueDate
		updated = true
	}
	if req.IsDone != nil {
		todo.IsDone = *req.IsDone
		updated = true
	}

	if !updated {
		return todo, nil // No fields were actually changed
	}
	todo.UpdatedAt = time.Now() // GORM might handle this automatically if configured

	err = s.todoRepo.UpdateTodo(ctx, todo)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo in repository: %w", err)
	}
	return todo, nil
}

// ListTodosByUserID retrieves all todo items for a specific user.
func (s *todoService) ListTodosByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error) {
	todos, err := s.todoRepo.GetTodosByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list todos from repository: %w", err)
	}
	return todos, nil
}

// GetTodoByID retrieves a specific todo item.
// It also ensures that the user attempting to get the todo owns it.
func (s *todoService) GetTodoByID(ctx context.Context, userID uuid.UUID, todoID uuid.UUID) (*Todo, error) {
	todo, err := s.todoRepo.GetTodoByID(ctx, todoID)
	if err != nil {
		// Handle potential ErrNotFound from repo layer
		return nil, ErrTodoNotFound
	}

	if todo.UserID != userID {
		return nil, ErrTodoAccessDenied
	}
	return todo, nil
}

// DeleteTodo deletes a specific todo item.
// It also ensures that the user attempting the delete owns the todo.
func (s *todoService) DeleteTodo(ctx context.Context, userID uuid.UUID, todoID uuid.UUID) error {
	todo, err := s.todoRepo.GetTodoByID(ctx, todoID)
	if err != nil {
		// Handle potential ErrNotFound from repo layer
		return ErrTodoNotFound
	}

	if todo.UserID != userID {
		return ErrTodoAccessDenied
	}

	err = s.todoRepo.DeleteTodo(ctx, todoID)
	if err != nil {
		return fmt.Errorf("failed to delete todo from repository: %w", err)
	}
	return nil
}


// CreateTodoRequest and UpdateTodoRequest types are needed by the service methods.
// They are defined in handler.go (package todo).
// If they were in a different package, they would need to be imported or redefined here.
// Since both handler and service are in package `todo`, they are accessible.
// For example, `req CreateTodoRequest` refers to `todo.CreateTodoRequest`.
// This is fine for smaller services, but for larger applications,
// request/response DTOs might live in a shared `types` or `dto` sub-package of `todo`.
// Or even a global DTO package if shared across different domains (e.g. user.UserDTO, todo.TodoDTO).
// For this task, keeping them within the `todo` package (accessible from both handler and service) is acceptable.
// The `CreateTodoRequest` and `UpdateTodoRequest` structs are already defined in `internal/todo/handler.go`
// and since `service.go` is in the same package `todo`, they are directly usable.
// No redefinition is needed here.
// The interface definition for TodoService in handler.go takes *gin.Context for ctx,
// while here it's context.Context. This should be consistent.
// Standard practice is context.Context for services.
// I will ensure the interface definition is consistent (using context.Context).
// The handler should pass c.Request.Context() to the service.
// I need to correct the TodoService interface in handler.go to use context.Context.
// And update the handler calls to pass c.Request.Context().
// Or, more simply, ensure this file's TodoService interface is the canonical one and handler.go uses it.
// The current `handler.go` definition of `TodoService` uses `*gin.Context`. I will correct this later.
// For now, this service file defines its `TodoService` interface using `context.Context`.
// The definition of CreateTodoRequest and UpdateTodoRequest are in handler.go,
// and since they are in the same package `todo`, they are accessible.
// The parameters like `req CreateTodoRequest` will correctly refer to `todo.CreateTodoRequest`.
// No redefinition is needed here.
