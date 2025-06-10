package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/your-username/go-moon/internal/cache"
	"github.com/your-username/go-moon/pkg/utils"
)

// AuthService defines the interface for authentication-related business logic.
// Note: This is duplicated from handler.go for now. Consider a shared location or one importing the other.
// For this exercise, we'll keep it, but in a larger project, this would be refined.
type AuthService interface {
	Logout(ctx context.Context, tokenString string) error
	RefreshToken(ctx context.Context, refreshTokenString string) (newAccessToken string, newRefreshToken string, err error)
}

type authService struct {
	jwtUtil  *utils.JWTUtil
	cache    cache.Cache
	userRepo user.UserRepository // Added UserRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(cache cache.Cache, jwtUtil *utils.JWTUtil, userRepo user.UserRepository) AuthService { // Added userRepo
	return &authService{cache: cache, jwtUtil: jwtUtil, userRepo: userRepo} // Store userRepo
}

// Logout validates a token and stores its JTI in Redis to invalidate it.
func (s *authService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.jwtUtil.ValidateToken(tokenString) // Use s.jwtUtil
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if claims.ID == "" {
		return fmt.Errorf("token JTI (ID) is missing")
	}

	// Calculate remaining time until token expiry.
	// The token is stored in cache until it naturally expires.
	// This prevents the cache from growing indefinitely with old JTIs.
	var expiresAtTime time.Time
	if claims.ExpiresAt != nil {
		expiresAtTime = claims.ExpiresAt.Time
	} else {
		// Should not happen for tokens we generate, as they always have expiry.
		return fmt.Errorf("token expiry is missing")
	}

	remainingValidity := time.Until(expiresAtTime)
	if remainingValidity <= 0 {
		// Token is already expired, no need to store it.
		return fmt.Errorf("token is already expired")
	}

	// Store the JTI in Redis with an expiry equal to the token's remaining validity.
	// The key could be prefixed, e.g., "blacklist_jti:<jti_value>"
	err = s.cache.Set(ctx, claims.ID, "revoked", remainingValidity)
	if err != nil {
		return fmt.Errorf("failed to store token JTI in cache: %w", err)
	}

	return nil
}

// RefreshToken validates a refresh token, generates new tokens, and blacklists the used refresh token.
func (s *authService) RefreshToken(ctx context.Context, refreshTokenString string) (string, string, error) {
	// 1. Validate the refresh token
	claims, err := s.jwtUtil.ValidateToken(refreshTokenString) // Use s.jwtUtil
	if err != nil {
		return "", "", fmt.Errorf("refresh token validation failed: %w", err)
	}

	if claims.ID == "" { // JTI check
		return "", "", fmt.Errorf("refresh token JTI (ID) is missing")
	}

	// 2. Check if the refresh token's JTI is blacklisted in Redis
	// Use a prefixed key for refresh token blacklist to distinguish from access token blacklist if necessary.
	// For simplicity here, we use the JTI directly, assuming JTIs are unique across token types or
	// that the logout mechanism for access tokens doesn't conflict.
	redisKey := "blacklist_jti:" + claims.ID
	_, err = s.cache.Get(ctx, redisKey)
	if err == nil {
		// Token JTI found in Redis, meaning it's blacklisted
		return "", "", fmt.Errorf("refresh token has been used or revoked")
	}
	if err != cache.ErrNotFound { // Check against the standardized cache.ErrNotFound
		// Some other Redis error occurred (that isn't ErrNotFound)
		return "", "", fmt.Errorf("failed to check refresh token blacklist: %w", err)
	}
	// If err is cache.ErrNotFound, the token JTI is not in the blacklist.

	// (New Step) 2.5. Verify user still exists (optional, but good practice for refresh)
	// This step is illustrative; actual error handling might differ.
	// _, errUser := s.userRepo.GetUserByID(ctx, claims.UserID)
	// if errUser != nil {
	// 	// If user not found, invalidate refresh token (or handle as per security policy)
	// 	// This prevents generating tokens for a deleted user.
	// 	// log.Printf("User %s not found during refresh token validation", claims.UserID)
	// 	return "", "", fmt.Errorf("user associated with refresh token not found")
	// }


	// 3. (Optional but recommended) Blacklist the used refresh token to prevent its reuse.
	// The expiry for the blacklisted refresh token should be its original expiry.
	var expiresAtTime time.Time
	if claims.ExpiresAt != nil {
		expiresAtTime = claims.ExpiresAt.Time
	} else {
		return "", "", fmt.Errorf("refresh token expiry is missing")
	}
	remainingValidity := time.Until(expiresAtTime)
	if remainingValidity <= 0 {
		return "", "", fmt.Errorf("refresh token is already expired (should have been caught by ValidateToken)")
	}

	err = s.cache.Set(ctx, redisKey, "revoked_refresh", remainingValidity)
	if err != nil {
		// Log this error but proceed to issue new tokens as the core validation passed.
		// Depending on policy, you might choose to fail here.
		// log.Printf("Warning: failed to blacklist used refresh token %s: %v", claims.ID, err)
	}

	// 4. Generate new access and refresh tokens
	newAccessToken, err := s.jwtUtil.GenerateAccessToken(claims.UserID) // Use s.jwtUtil
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := s.jwtUtil.GenerateRefreshToken(claims.UserID) // Use s.jwtUtil
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// A specific error for cache not found, if not provided by the cache package.
// For example, if your cache.Get returns a specific error type:
// var ErrNotFound = errors.New("key not found")
// You would then check against this in Get and RefreshToken.
// The go-redis client returns redis.Nil for Get when key is not found.
// We should ensure cache.Get normalizes this or the service layer checks for redis.Nil.
// For this implementation, I'll assume cache.Get returns an error that can be checked,
// or I'll modify the cache interface/impl later if needed.
// For now, checking err.Error() == "redis: nil" is a common way if redis.Nil is not exported by cache.
// Let's assume our cache.Get might return a generic cache.ErrNotFound or redis.Nil itself.
// The current cache.Get in `internal/cache/redis.go` returns `redis.Nil` directly.
// So, the check `err.Error() != "redis: nil"` is okay, or `err != redis.Nil` if redis is imported here.
// For cleaner code, it's better if cache.Cache defines its own ErrNotFound.
// Let's add that to internal/cache/redis.go
