package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go-moon/internal/cache"
	"go-moon/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9" // For redis.Nil
)

const (
	userContextKey = "userID" // Key to store userID in Gin context
)

// AuthMiddleware creates a gin.HandlerFunc for JWT authentication and authorization.
// It checks for a valid JWT in the Authorization header, validates it,
// and ensures it's not blacklisted in Redis (logged out).
func AuthMiddleware(appCache cache.Cache, jwtUtil *utils.JWTUtil) gin.HandlerFunc { // Changed cache to appCache for clarity
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}
		tokenString := parts[1]

		// Validate the token signature and standard claims (like expiry)
		claims, err := jwtUtil.ValidateToken(tokenString) // Use jwtUtil
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		// Check if the token JTI is in the Redis blacklist
		// The key used in Redis is the JTI itself.
		if claims.ID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: JTI missing"})
			return
		}

		// Use a timeout for Redis operations
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		_, err = cache.Get(ctx, claims.ID)
		if err == nil {
			// Key found in Redis, meaning token is blacklisted (logged out)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked (logged out)"})
			return
		}
		if err != redis.Nil {
			// Some other Redis error occurred
			// log.Printf("Redis error checking blacklist: %v", err) // Consider logging
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not verify token status"})
			return
		}
		// If err is redis.Nil, the token JTI is not in the blacklist, so it's valid to proceed.

		// Set user information in context for downstream handlers
		c.Set(userContextKey, claims.UserID.String()) // Store as string or uuid.UUID as needed by handlers

		c.Next()
	}
}

// GetUserIDFromContext retrieves the userID from the Gin context.
// This is a helper function for handlers that need the userID after middleware validation.
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	val, exists := c.Get(userContextKey)
	if !exists {
		return "", false
	}
	userIDStr, ok := val.(string)
	if !ok {
		// This case should ideally not happen if middleware sets it correctly.
		// Consider logging this anomaly.
		return "", false
	}
	return userIDStr, true
}

// GetUserIDFromContextAsUUID retrieves the userID as uuid.UUID from the Gin context.
// Returns error if not found or if not a valid UUID string.
func GetUserIDFromContextAsUUID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := GetUserIDFromContext(c)
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	parsedUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID format in context: %w", err)
	}
	return parsedUUID, nil
}

// Need to import "errors" and "fmt" for GetUserIDFromContextAsUUID
import (
	"errors"
	"fmt"
	"github.com/google/uuid" // Ensure uuid is imported
	// "go-moon/internal/user" // Not directly needed here but good to keep track if context setting becomes user object
)
