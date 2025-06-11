package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	userMocks "go-moon/internal/user/mocks" // Mock for UserRepository
	utilMocks "go-moon/pkg/utils/mocks"   // Mock for JWTUtil
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm" // For gorm.ErrRecordNotFound
)

func TestUserService_Register(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	// JWTUtil is not directly used by Register, so we can pass a nil or an unconfigured mock if NewUserService requires it.
	// The current NewUserService in service.go takes jwtUtil.
	mockJWTUtil := new(utilMocks.MockJWTUtil)
	userService := NewUserService(mockUserRepo, mockJWTUtil)

	ctx := context.Background()
	req := RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		DOB:       time.Now().AddDate(-20, 0, 0),
	}

	t.Run("Success", func(t *testing.T) {
		// For CreateUser, we expect any user object. We capture it to check details.
		mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*user.User")).Run(func(args mock.Arguments) {
			usrArg := args.Get(1).(*User)
			assert.Equal(t, req.Email, usrArg.Email)
			assert.Equal(t, req.FirstName, usrArg.FirstName)
			assert.Equal(t, req.LastName, usrArg.LastName)
			assert.WithinDuration(t, req.DOB, usrArg.DOB, time.Second)
			err := bcrypt.CompareHashAndPassword([]byte(usrArg.Password), []byte(req.Password))
			assert.NoError(t, err, "Password should be hashed")
		}).Return(nil).Once()

		createdUser, err := userService.Register(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, req.Email, createdUser.Email)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("CreateUser_Repository_Error", func(t *testing.T) {
		mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("*user.User")).Return(errors.New("db error")).Once()

		_, err := userService.Register(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		mockUserRepo.AssertExpectations(t)
	})

	// Note: Email existence check (GetUserByEmail) is not part of the current Register service logic.
	// If it were, we'd add a test case for "email already exists".
}

func TestUserService_Login(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	mockJWTUtil := new(utilMocks.MockJWTUtil)
	userService := NewUserService(mockUserRepo, mockJWTUtil)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userID := uuid.New()

	userInstance := &User{
		ID:        userID,
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: "Test",
		LastName:  "User",
	}

	loginReq := LoginRequest{Email: email, Password: password}
	expectedAccessToken := "expected.access.token"
	expectedRefreshToken := "expected.refresh.token"

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.On("GetUserByEmail", ctx, email).Return(userInstance, nil).Once()
		mockJWTUtil.On("GenerateAccessToken", userID).Return(expectedAccessToken, nil).Once()
		mockJWTUtil.On("GenerateRefreshToken", userID).Return(expectedRefreshToken, nil).Once()

		accessToken, refreshToken, err := userService.Login(ctx, loginReq)
		assert.NoError(t, err)
		assert.Equal(t, expectedAccessToken, accessToken)
		assert.Equal(t, expectedRefreshToken, refreshToken)

		mockUserRepo.AssertExpectations(t)
		mockJWTUtil.AssertExpectations(t)
	})

	t.Run("User_Not_Found", func(t *testing.T) {
		mockUserRepo.On("GetUserByEmail", ctx, email).Return(nil, gorm.ErrRecordNotFound).Once()

		_, _, err := userService.Login(ctx, loginReq)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err) // Service maps gorm.ErrRecordNotFound
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Incorrect_Password", func(t *testing.T) {
		mockUserRepo.On("GetUserByEmail", ctx, email).Return(userInstance, nil).Once()
		incorrectLoginReq := LoginRequest{Email: email, Password: "wrongpassword"}

		_, _, err := userService.Login(ctx, incorrectLoginReq)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err) // bcrypt.CompareHashAndPassword fails
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	mockJWTUtil := new(utilMocks.MockJWTUtil) // Not used by ChangePassword directly
	userService := NewUserService(mockUserRepo, mockJWTUtil)

	ctx := context.Background()
	userID := uuid.New()
	oldPassword := "oldPassword123"
	newPassword := "newPassword456"
	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)

	userInstance := &User{
		ID:       userID,
		Email:    "user@example.com",
		Password: string(hashedOldPassword),
	}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", ctx, userID).Return(userInstance, nil).Once()
		mockUserRepo.On("UpdateUser", ctx, mock.MatchedBy(func(u *User) bool {
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(newPassword))
			return u.ID == userID && err == nil
		})).Return(nil).Once()

		err := userService.ChangePassword(ctx, userID, oldPassword, newPassword)
		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("User_Not_Found", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound).Once()
		err := userService.ChangePassword(ctx, userID, oldPassword, newPassword)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Old_Password_Incorrect", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", ctx, userID).Return(userInstance, nil).Once()
		err := userService.ChangePassword(ctx, userID, "wrongOldPassword", newPassword)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	mockUserRepo := new(userMocks.MockUserRepository)
	mockJWTUtil := new(utilMocks.MockJWTUtil) // Not used by UpdateProfile
	userService := NewUserService(mockUserRepo, mockJWTUtil)

	ctx := context.Background()
	userID := uuid.New()
	originalUser := &User{
		ID:        userID,
		Email:     "original@example.com",
		FirstName: "OriginalFirst",
		LastName:  "OriginalLast",
	}
	updateReq := UpdateUserRequest{ // This is user.UpdateUserRequest from handler, service uses its fields
		FirstName: "UpdatedFirst",
		LastName:  "UpdatedLast",
	}

	t.Run("Success_Update_Both_Names", func(t *testing.T) {
		// Need to return a copy for GetUserByID as the service modifies the user object in place
		userToReturn := *originalUser
		mockUserRepo.On("GetUserByID", ctx, userID).Return(&userToReturn, nil).Once()
		mockUserRepo.On("UpdateUser", ctx, mock.MatchedBy(func(u *User) bool {
			return u.ID == userID && u.FirstName == updateReq.FirstName && u.LastName == updateReq.LastName
		})).Return(nil).Once()

		updatedUser, err := userService.UpdateProfile(ctx, userID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, updateReq.FirstName, updatedUser.FirstName)
		assert.Equal(t, updateReq.LastName, updatedUser.LastName)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Success_Update_FirstName_Only", func(t *testing.T) {
		userToReturn := *originalUser
		reqFirstOnly := UpdateUserRequest{FirstName: "NewFirstOnly"}
		mockUserRepo.On("GetUserByID", ctx, userID).Return(&userToReturn, nil).Once()
		mockUserRepo.On("UpdateUser", ctx, mock.MatchedBy(func(u *User) bool {
			return u.ID == userID && u.FirstName == reqFirstOnly.FirstName && u.LastName == originalUser.LastName
		})).Return(nil).Once()

		updatedUser, err := userService.UpdateProfile(ctx, userID, reqFirstOnly)
		assert.NoError(t, err)
		assert.Equal(t, reqFirstOnly.FirstName, updatedUser.FirstName)
		assert.Equal(t, originalUser.LastName, updatedUser.LastName) // Ensure last name is unchanged
		mockUserRepo.AssertExpectations(t)
	})


	t.Run("User_Not_Found", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound).Once()
		_, err := userService.UpdateProfile(ctx, userID, updateReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("No_Fields_To_Update", func(t *testing.T) {
		userToReturn := *originalUser
		reqEmpty := UpdateUserRequest{} // Empty request
		mockUserRepo.On("GetUserByID", ctx, userID).Return(&userToReturn, nil).Once()
		// UpdateUser should not be called if no fields are changed by the service logic

		updatedUser, err := userService.UpdateProfile(ctx, userID, reqEmpty)
		assert.NoError(t, err) // Service returns the original user if no update occurred
		assert.Equal(t, originalUser.FirstName, updatedUser.FirstName)
		assert.Equal(t, originalUser.LastName, updatedUser.LastName)
		mockUserRepo.AssertExpectations(t) // GetUserByID called
		mockUserRepo.AssertNotCalled(t, "UpdateUser", ctx, mock.Anything) // UpdateUser not called
	})
}
