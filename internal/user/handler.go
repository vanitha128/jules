package user

import (
	"net/http"
	"errors" // For errors.Is
	"log"    // For placeholder logging
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // For uuid.Parse
	"go-moon/internal/middleware"
)

// UserService defines the interface for user related business logic.
// This should be implemented by a service in internal/user/service.go
type UserService interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Login(ctx context.Context, req LoginRequest) (accessToken string, refreshToken string, err error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword string, newPassword string) error
	UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*User, error)
}


// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=8"`
	FirstName string    `json:"firstName" binding:"required"`
	LastName  string    `json:"lastName" binding:"required"`
	DOB       time.Time `json:"dob" binding:"required"`
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterUser handles user registration.
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userRegistered, err := h.userService.Register(c.Request.Context(), req) // Renamed user to userRegistered
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) { // Use the error from the user package
			c.JSON(http.StatusConflict, gin.H{"error": ErrEmailAlreadyExists.Error()})
		} else {
			// Log the actual error for server-side diagnostics if desired
			// log.Printf("Failed to register user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "userID": userRegistered.ID})
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login.
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		// Check if the error is due to invalid credentials
		// Note: This requires userService.Login to return a specific error type or message
		// For now, we assume any error from Login service means unauthorized or bad request.
		// A more robust solution would be to check error types (e.g., errors.Is(err, userService.ErrInvalidCredentials))
		if err.Error() == "invalid email or password" { // This string matching is fragile; use typed errors.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

// ChangePasswordRequest defines the structure for the change password request.
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// ChangePassword handles changing the authenticated user's password.
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// userID, exists := c.Get("userID") // Assuming userID is set by AuthMiddleware
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
	// 	return
	// }
	// userIDStr, ok := userID.(string)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
	// 	return
	// }
	// parsedUserID, err := uuid.Parse(userIDStr)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse User ID from context"})
	// 	return
	// }

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// err = h.userService.ChangePassword(c.Request.Context(), parsedUserID, req.OldPassword, req.NewPassword)
	// if err != nil {
	//  // Handle specific errors like "old password mismatch" or "user not found"
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password: " + err.Error()})
	// 	return
	// }

	// For now, returning a placeholder as userService and userID extraction are not fully wired.
	userIDFromContextPlaceholder := "user-id-from-context-placeholder" // Simulate getting userID
	log.Printf("ChangePassword attempt for user %s with old: %s, new: %s", userIDFromContextPlaceholder, req.OldPassword, req.NewPassword)


	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully (placeholder)"})
}

// UpdateUserRequest defines the structure for updating user profile information.
// Email and DOB are not updatable as per current design.
type UpdateUserRequest struct {
	FirstName string `json:"firstName" binding:"omitempty,min=1"` // omitempty allows partial updates
	LastName  string `json:"lastName" binding:"omitempty,min=1"`
}

// UpdateProfile handles updating the authenticated user's profile information.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// userID, exists := c.Get("userID") // Assuming userID is set by AuthMiddleware
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
	// 	return
	// }
	// userIDStr, ok := userID.(string)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID in context is not a string"})
	// 	return
	// }
	// parsedUserID, err := uuid.Parse(userIDStr)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse User ID from context"})
	// 	return
	// }

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Basic validation: ensure at least one field is provided for update
	if req.FirstName == "" && req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields provided for update"})
		return
	}

	// updatedUser, err := h.userService.UpdateProfile(c.Request.Context(), parsedUserID, req)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
	// 	return
	// }

	// For now, returning a placeholder.
	userIDFromContextPlaceholder := "user-id-from-context-placeholder" // Simulate getting userID
	log.Printf("UpdateProfile attempt for user %s with data: %+v", userIDFromContextPlaceholder, req)

	// c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "user": updatedUser})
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully (placeholder)", "updated_fields": req})
}

// GetProfile handles retrieving the authenticated user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get(middleware.UserContextKey) // Using the constant from middleware
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	parsedUserID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID format in context"})
		return
	}

	userProfile, err := h.userService.GetProfile(c.Request.Context(), parsedUserID)
	if err != nil {
		// Handle specific errors, e.g., user not found from service
		if err.Error() == "user not found" { // This depends on error returned by service
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile: " + err.Error()})
		}
		return
	}

	// Return a DTO instead of the full user model to avoid exposing sensitive fields like Password hash
	// For now, returning placeholder or assuming User model is safe to return (e.g. password field is omitempty or cleared)
	// Example DTO:
	// type ProfileResponse struct {
	// 	ID uuid.UUID `json:"id"`
	// 	Email string `json:"email"`
	// 	FirstName string `json:"firstName"`
	//  LastName string `json:"lastName"`
	//  DOB time.Time `json:"dob"`
	// }
	// response := ProfileResponse{ ... }
	// c.JSON(http.StatusOK, response)

	// For now, returning the fetched userProfile directly, assuming it's safe or a placeholder.
	// Critical: Ensure password hash is NOT included in this response in a real app.
	// The User model has `json:"-"` for password, but GORM tags are for DB.
	// The service layer's GetProfile should ideally return a DTO or a User object with Password cleared.
	// Current User model does not have json tags for password, so it would be included if not cleared.
	// Let's assume service clears it or returns a DTO.
	// For placeholder:
	c.JSON(http.StatusOK, gin.H{
		"id": userProfile.ID,
		"email": userProfile.Email,
		"firstName": userProfile.FirstName,
		"lastName": userProfile.LastName,
		"dob": userProfile.DOB,
		// "created_at": userProfile.CreatedAt, // etc.
	})
}
