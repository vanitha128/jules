package todo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	todoMocks "github.com/your-username/go-moon/internal/todo/mocks" // Mock for TodoRepository
	"gorm.io/gorm"                                                   // For gorm.ErrRecordNotFound (though service maps this)
)

func TestTodoService_CreateTodo(t *testing.T) {
	mockTodoRepo := new(todoMocks.MockTodoRepository)
	todoService := NewTodoService(mockTodoRepo)

	ctx := context.Background()
	userID := uuid.New()
	req := CreateTodoRequest{
		Title:       "Test Todo",
		Description: "Test Description",
		DueDate:     time.Now().Add(24 * time.Hour),
	}

	t.Run("Success", func(t *testing.T) {
		mockTodoRepo.On("CreateTodo", ctx, mock.AnythingOfType("*todo.Todo")).Run(func(args mock.Arguments) {
			todoArg := args.Get(1).(*Todo)
			assert.Equal(t, req.Title, todoArg.Title)
			assert.Equal(t, req.Description, todoArg.Description)
			assert.Equal(t, userID, todoArg.UserID)
			assert.False(t, todoArg.IsDone) // Default
		}).Return(nil).Once()

		createdTodo, err := todoService.CreateTodo(ctx, userID, req)
		assert.NoError(t, err)
		assert.NotNil(t, createdTodo)
		assert.Equal(t, req.Title, createdTodo.Title)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Repository_Error", func(t *testing.T) {
		mockTodoRepo.On("CreateTodo", ctx, mock.AnythingOfType("*todo.Todo")).Return(errors.New("db error")).Once()
		_, err := todoService.CreateTodo(ctx, userID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create todo in repository")
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodoService_UpdateTodo(t *testing.T) {
	mockTodoRepo := new(todoMocks.MockTodoRepository)
	todoService := NewTodoService(mockTodoRepo)

	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	todoID := uuid.New()

	originalTodo := &Todo{
		ID:          todoID,
		UserID:      userID,
		Title:       "Original Title",
		Description: "Original Desc",
		IsDone:      false,
	}
	newTitle := "Updated Title"
	updateReq := UpdateTodoRequest{Title: &newTitle}

	t.Run("Success", func(t *testing.T) {
		// Important: return a copy for GetTodoByID because the service modifies it in place
		todoToReturn := *originalTodo
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(&todoToReturn, nil).Once()
		mockTodoRepo.On("UpdateTodo", ctx, mock.MatchedBy(func(td *Todo) bool {
			return td.ID == todoID && td.Title == newTitle
		})).Return(nil).Once()

		updatedTodo, err := todoService.UpdateTodo(ctx, userID, todoID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, updatedTodo)
		assert.Equal(t, newTitle, updatedTodo.Title)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Todo_Not_Found", func(t *testing.T) {
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(nil, gorm.ErrRecordNotFound).Once() // Simulate repo returning gorm error
		_, err := todoService.UpdateTodo(ctx, userID, todoID, updateReq)
		assert.Error(t, err)
		assert.Equal(t, ErrTodoNotFound, err) // Service should map to its own error
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Access_Denied", func(t *testing.T) {
		todoOwnedByOther := *originalTodo
		todoOwnedByOther.UserID = otherUserID
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(&todoOwnedByOther, nil).Once()

		_, err := todoService.UpdateTodo(ctx, userID, todoID, updateReq) // userID is different from todoOwnedByOther.UserID
		assert.Error(t, err)
		assert.Equal(t, ErrTodoAccessDenied, err)
		mockTodoRepo.AssertExpectations(t)
		mockTodoRepo.AssertNotCalled(t, "UpdateTodo", ctx, mock.Anything)
	})
}

func TestTodoService_ListTodosByUserID(t *testing.T) {
	mockTodoRepo := new(todoMocks.MockTodoRepository)
	todoService := NewTodoService(mockTodoRepo)
	ctx := context.Background()
	userID := uuid.New()
	expectedTodos := []Todo{{ID: uuid.New(), UserID: userID, Title: "Todo 1"}, {ID: uuid.New(), UserID: userID, Title: "Todo 2"}}

	t.Run("Success", func(t *testing.T) {
		mockTodoRepo.On("GetTodosByUserID", ctx, userID).Return(expectedTodos, nil).Once()
		todos, err := todoService.ListTodosByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedTodos, todos)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Repository_Error", func(t *testing.T) {
		mockTodoRepo.On("GetTodosByUserID", ctx, userID).Return(nil, errors.New("db error")).Once()
		_, err := todoService.ListTodosByUserID(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list todos from repository")
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodoService_GetTodoByID(t *testing.T) {
	mockTodoRepo := new(todoMocks.MockTodoRepository)
	todoService := NewTodoService(mockTodoRepo)
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	todoID := uuid.New()
	expectedTodo := &Todo{ID: todoID, UserID: userID, Title: "Test Todo"}

	t.Run("Success", func(t *testing.T) {
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(expectedTodo, nil).Once()
		todo, err := todoService.GetTodoByID(ctx, userID, todoID)
		assert.NoError(t, err)
		assert.Equal(t, expectedTodo, todo)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Not_Found", func(t *testing.T) {
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(nil, gorm.ErrRecordNotFound).Once()
		_, err := todoService.GetTodoByID(ctx, userID, todoID)
		assert.ErrorIs(t, err, ErrTodoNotFound)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Access_Denied", func(t *testing.T) {
		todoOwnedByOther := &Todo{ID: todoID, UserID: otherUserID, Title: "Other's Todo"}
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(todoOwnedByOther, nil).Once()
		_, err := todoService.GetTodoByID(ctx, userID, todoID)
		assert.ErrorIs(t, err, ErrTodoAccessDenied)
		mockTodoRepo.AssertExpectations(t)
	})
}

func TestTodoService_DeleteTodo(t *testing.T) {
	mockTodoRepo := new(todoMocks.MockTodoRepository)
	todoService := NewTodoService(mockTodoRepo)
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	todoID := uuid.New()
	todoToDelete := &Todo{ID: todoID, UserID: userID, Title: "To Delete"}

	t.Run("Success", func(t *testing.T) {
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(todoToDelete, nil).Once()
		mockTodoRepo.On("DeleteTodo", ctx, todoID).Return(nil).Once()
		err := todoService.DeleteTodo(ctx, userID, todoID)
		assert.NoError(t, err)
		mockTodoRepo.AssertExpectations(t)
	})

	t.Run("Not_Found", func(t *testing.T) {
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(nil, gorm.ErrRecordNotFound).Once()
		err := todoService.DeleteTodo(ctx, userID, todoID)
		assert.ErrorIs(t, err, ErrTodoNotFound)
		mockTodoRepo.AssertExpectations(t)
		mockTodoRepo.AssertNotCalled(t, "DeleteTodo", ctx, todoID)
	})

	t.Run("Access_Denied", func(t *testing.T) {
		todoOwnedByOther := &Todo{ID: todoID, UserID: otherUserID, Title: "Other's Todo"}
		mockTodoRepo.On("GetTodoByID", ctx, todoID).Return(todoOwnedByOther, nil).Once()
		err := todoService.DeleteTodo(ctx, userID, todoID)
		assert.ErrorIs(t, err, ErrTodoAccessDenied)
		mockTodoRepo.AssertExpectations(t)
		mockTodoRepo.AssertNotCalled(t, "DeleteTodo", ctx, todoID)
	})
}
