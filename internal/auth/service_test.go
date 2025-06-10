package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/go-moon/internal/cache"
	cacheMocks "github.com/your-username/go-moon/internal/cache/mocks" // Mock for Cache
	"github.com/your-username/go-moon/pkg/utils"
	utilMocks "github.com/your-username/go-moon/pkg/utils/mocks" // Mock for JWTUtil
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthService_Logout(t *testing.T) {
	mockCache := new(cacheMocks.MockCache)
	mockJWTUtil := new(utilMocks.MockJWTUtil)
	authService := NewAuthService(mockCache, mockJWTUtil)

	ctx := context.Background()
	tokenString := "valid.token.string"
	userID := uuid.New()
	jti := uuid.New().String()
	expiry := time.Now().Add(1 * time.Hour)

	claims := &utils.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", tokenString).Return(claims, nil).Once()
		mockCache.On("Set", ctx, jti, "revoked", mock.AnythingOfType("time.Duration")).Return(nil).Once()

		err := authService.Logout(ctx, tokenString)
		assert.NoError(t, err)

		mockJWTUtil.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("ValidateToken_Error", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", tokenString).Return(nil, errors.New("validation failed")).Once()

		err := authService.Logout(ctx, tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")

		mockJWTUtil.AssertExpectations(t)
		// No cache call expected
		mockCache.AssertNotCalled(t, "Set", ctx, jti, "revoked", mock.AnythingOfType("time.Duration"))
	})

	t.Run("Missing_JTI", func(t *testing.T) {
		claimsWithoutJTI := &utils.Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expiry)}, // JTI is empty
		}
		mockJWTUtil.On("ValidateToken", tokenString).Return(claimsWithoutJTI, nil).Once()

		err := authService.Logout(ctx, tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token JTI (ID) is missing")
		mockJWTUtil.AssertExpectations(t)
	})

	t.Run("Token_Already_Expired_On_Logout", func(t *testing.T) {
		expiredClaims := &utils.Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        jti,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)), // Already expired
			},
		}
		mockJWTUtil.On("ValidateToken", tokenString).Return(expiredClaims, nil).Once()
		// No Set should be called on cache if token is already expired by the time we check remaining validity

		err := authService.Logout(ctx, tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is already expired")
		mockJWTUtil.AssertExpectations(t)
		mockCache.AssertNotCalled(t, "Set")
	})


	t.Run("Cache_Set_Error", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", tokenString).Return(claims, nil).Once()
		mockCache.On("Set", ctx, jti, "revoked", mock.AnythingOfType("time.Duration")).Return(errors.New("cache set failed")).Once()

		err := authService.Logout(ctx, tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store token JTI in cache")
		mockJWTUtil.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	mockCache := new(cacheMocks.MockCache)
	mockJWTUtil := new(utilMocks.MockJWTUtil)
	authService := NewAuthService(mockCache, mockJWTUtil)

	ctx := context.Background()
	oldRefreshTokenString := "old.refresh.token"
	userID := uuid.New()
	oldJti := uuid.New().String()
	oldExpiry := time.Now().Add(7 * 24 * time.Hour)

	oldClaims := &utils.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        oldJti,
			ExpiresAt: jwt.NewNumericDate(oldExpiry),
		},
	}
	newAccessToken := "new.access.token"
	newRefreshToken := "new.refresh.token"

	t.Run("Success", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", oldRefreshTokenString).Return(oldClaims, nil).Once()
		mockCache.On("Get", ctx, "blacklist_jti:"+oldJti).Return("", cache.ErrNotFound).Once() // Not blacklisted
		mockCache.On("Set", ctx, "blacklist_jti:"+oldJti, "revoked_refresh", mock.AnythingOfType("time.Duration")).Return(nil).Once()
		mockJWTUtil.On("GenerateAccessToken", userID).Return(newAccessToken, nil).Once()
		mockJWTUtil.On("GenerateRefreshToken", userID).Return(newRefreshToken, nil).Once()

		accessToken, refreshToken, err := authService.RefreshToken(ctx, oldRefreshTokenString)
		assert.NoError(t, err)
		assert.Equal(t, newAccessToken, accessToken)
		assert.Equal(t, newRefreshToken, refreshToken)

		mockJWTUtil.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("Invalid_RefreshToken", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", oldRefreshTokenString).Return(nil, errors.New("validation failed")).Once()

		_, _, err := authService.RefreshToken(ctx, oldRefreshTokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token validation failed")
		mockJWTUtil.AssertExpectations(t)
	})

	t.Run("RefreshToken_JTI_Blacklisted", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", oldRefreshTokenString).Return(oldClaims, nil).Once()
		mockCache.On("Get", ctx, "blacklist_jti:"+oldJti).Return("revoked", nil).Once() // Found in cache (blacklisted)

		_, _, err := authService.RefreshToken(ctx, oldRefreshTokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token has been used or revoked")
		mockJWTUtil.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("RefreshToken_Cache_Get_Error", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", oldRefreshTokenString).Return(oldClaims, nil).Once()
		mockCache.On("Get", ctx, "blacklist_jti:"+oldJti).Return("", errors.New("some redis error")).Once()

		_, _, err := authService.RefreshToken(ctx, oldRefreshTokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check refresh token blacklist")
		mockJWTUtil.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})


	t.Run("GenerateAccessToken_Error", func(t *testing.T) {
		mockJWTUtil.On("ValidateToken", oldRefreshTokenString).Return(oldClaims, nil).Once()
		mockCache.On("Get", ctx, "blacklist_jti:"+oldJti).Return("", cache.ErrNotFound).Once()
		mockCache.On("Set", ctx, "blacklist_jti:"+oldJti, "revoked_refresh", mock.AnythingOfType("time.Duration")).Return(nil).Once()
		mockJWTUtil.On("GenerateAccessToken", userID).Return("", errors.New("access token gen failed")).Once()

		_, _, err := authService.RefreshToken(ctx, oldRefreshTokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate new access token")
		mockJWTUtil.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}
