package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-super-secret-key-for-jwt-!@#$%^&*()_+"
const testInvalidSecret = "invalid-test-secret-key"

func TestNewJWTUtil(t *testing.T) {
	jwtUtil := NewJWTUtil(testSecret)
	require.NotNil(t, jwtUtil)
	assert.Equal(t, []byte(testSecret), jwtUtil.secretKey)
}

func TestJWTUtil_GenerateAndValidateAccessToken(t *testing.T) {
	jwtUtil := NewJWTUtil(testSecret)
	userID := uuid.New()

	tokenString, err := jwtUtil.GenerateAccessToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	// Validate the token
	claims, err := jwtUtil.ValidateToken(tokenString)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, userID, claims.UserID)
	assert.NotEmpty(t, claims.ID) // JTI
	assert.Equal(t, "go-moon-app", claims.Issuer)
	assert.WithinDuration(t, time.Now().Add(60*time.Minute), claims.ExpiresAt.Time, 10*time.Second) // Allow 10s skew
	assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Second)))
	assert.True(t, claims.NotBefore.Time.Before(time.Now().Add(time.Second)))
}

func TestJWTUtil_GenerateAndValidateRefreshToken(t *testing.T) {
	jwtUtil := NewJWTUtil(testSecret)
	userID := uuid.New()

	tokenString, err := jwtUtil.GenerateRefreshToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	// Validate the token
	claims, err := jwtUtil.ValidateToken(tokenString)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, userID, claims.UserID)
	assert.NotEmpty(t, claims.ID) // JTI
	assert.Equal(t, "go-moon-app", claims.Issuer)
	assert.WithinDuration(t, time.Now().Add(24*7*time.Hour), claims.ExpiresAt.Time, 10*time.Second) // Allow 10s skew
}

func TestJWTUtil_ValidateToken_Expired(t *testing.T) {
	jwtUtil := NewJWTUtil(testSecret)
	userID := uuid.New()

	// Generate a token that expired 1 hour ago
	expiredToken, err := jwtUtil.generateToken(userID, uuid.NewString(), time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	claims, err := jwtUtil.ValidateToken(expiredToken)
	require.Error(t, err) // Expect an error due to expiry
	assert.Nil(t, claims)
	// The error from `jwt.ParseWithClaims` for an expired token usually contains "token is expired"
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWTUtil_ValidateToken_InvalidSignature(t *testing.T) {
	jwtUtil := NewJWTUtil(testSecret)
	jwtUtilInvalidSig := NewJWTUtil(testInvalidSecret)
	userID := uuid.New()

	// Generate token with one utility
	tokenString, err := jwtUtil.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Try to validate with another utility (different secret)
	claims, err := jwtUtilInvalidSig.ValidateToken(tokenString)
	require.Error(t, err)
	assert.Nil(t, claims)
	// Error should be about signature validity
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestJWTUtil_ValidateToken_Malformed(t *testing.T) {
	jwtUtil := NewJWTUtil(testSecret)
	malformedToken := "this.is.not.a.jwt"

	claims, err := jwtUtil.ValidateToken(malformedToken)
	require.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is malformed")
}

func TestJWTUtil_ValidateToken_FutureNotBefore(t *testing.T) {
	jwtUtil := NewJWTUtil(testSecret)
	userID := uuid.New()

	// Generate a token that is not yet valid
	claimsInput := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)), // Issued 1 min ago
			NotBefore: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // Not valid for 30 mins
			ID:        uuid.NewString(),
			Issuer:    "go-moon-app",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsInput)
	tokenString, errToken := token.SignedString(jwtUtil.secretKey)
	require.NoError(t, errToken)

	claims, err := jwtUtil.ValidateToken(tokenString)
	require.Error(t, err) // Expect an error due to NBF
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is not valid yet")
}
