package todo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// postgresTodoRepository implements the TodoRepository interface using GORM.
type postgresTodoRepository struct {
	db *gorm.DB
}

// NewPostgresTodoRepository creates a new instance of postgresTodoRepository.
func NewPostgresTodoRepository(db *gorm.DB) TodoRepository {
	return &postgresTodoRepository{db: db}
}

// CreateTodo creates a new todo record in the database.
func (r *postgresTodoRepository) CreateTodo(ctx context.Context, todo *Todo) error {
	// Todo.ID is set to uuid.New() in the service layer.
	// CreatedAt and UpdatedAt are handled by GORM via autoCreateTime/autoUpdateTime.
	if err := r.db.WithContext(ctx).Create(todo).Error; err != nil {
		return err
	}
	return nil
}

// UpdateTodo updates an existing todo item in the database.
// The service layer is responsible for fetching the todo, checking ownership,
// making changes, and then passing the full Todo object to this method.
func (r *postgresTodoRepository) UpdateTodo(ctx context.Context, todo *Todo) error {
	// GORM's Save method updates all fields of the record if the primary key is provided.
	// It also automatically updates the UpdatedAt timestamp if `gorm:"autoUpdateTime"` is used.
	result := r.db.WithContext(ctx).Save(todo)
	if result.Error != nil {
		return result.Error
	}
	// Optionally, check result.RowsAffected if you need to confirm the record was actually found and updated.
	// If RowsAffected is 0, it might mean the record with that ID didn't exist.
	// GORM itself doesn't error on 0 rows affected for Save if the operation itself was valid.
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Or a custom domain error like ErrTodoNotFound
	}
	return nil
}

// GetTodosByUserID retrieves all todo items for a specific user ID.
func (r *postgresTodoRepository) GetTodosByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error) {
	var todos []Todo
	// Order by CreatedAt or DueDate, for example
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&todos).Error; err != nil {
		// If no records found, GORM Find doesn't return ErrRecordNotFound, it returns an empty slice and nil error.
		return nil, err
	}
	return todos, nil
}

// GetTodoByID retrieves a specific todo item by its ID.
// The service layer will handle the ownership check (UserID match).
func (r *postgresTodoRepository) GetTodoByID(ctx context.Context, todoID uuid.UUID) (*Todo, error) {
	var todo Todo
	if err := r.db.WithContext(ctx).Where("id = ?", todoID).First(&todo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Propagate gorm.ErrRecordNotFound, service layer will map to domain specific ErrTodoNotFound
			return nil, err
		}
		return nil, err
	}
	return &todo, nil
}

// DeleteTodo deletes a todo item by its ID.
// The service layer is responsible for ownership check before calling this.
func (r *postgresTodoRepository) DeleteTodo(ctx context.Context, todoID uuid.UUID) error {
	// GORM's Delete method requires a model instance or a primary key.
	// If you pass a primary key, you must specify the model.
	result := r.db.WithContext(ctx).Delete(&Todo{}, todoID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// If no rows were affected, it means the record was not found.
		return gorm.ErrRecordNotFound // Service layer will map to ErrTodoNotFound
	}
	return nil
}
