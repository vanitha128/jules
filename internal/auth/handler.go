package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"context" // Ensure context is imported

	"github.com/gin-gonic/gin"
)

// AuthService defines the interface for authentication-related business logic.
// This will be implemented by a service in internal/auth/service.go
type AuthService interface {
	Logout(ctx context.Context, tokenString string) error // Changed gin.Context to context.Context
	RefreshToken(ctx context.Context, refreshTokenString string) (newAccessToken string, newRefreshToken string, err error)
}

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Logout handles user logout by invalidating the provided JWT.
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
		return
	}
	tokenString := parts[1]

	err := h.authService.Logout(c, tokenString)
	if err != nil {
		// Specific error handling can be done here based on error types from service
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// RefreshTokenRequest defines the structure for the refresh token request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken handles the renewal of access and refresh tokens.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		// Specific error handling can be done here based on error types from service
		// For example, if refresh token is invalid or blacklisted
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}
