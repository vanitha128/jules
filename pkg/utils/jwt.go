package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTUtil is a helper struct for JWT operations.
type JWTUtil struct {
	secretKey []byte
}

// NewJWTUtil creates a new JWTUtil with the given secret.
func NewJWTUtil(secret string) *JWTUtil {
	return &JWTUtil{secretKey: []byte(secret)}
}

// Claims defines the JWT claims structure, embedding standard claims and adding UserID.
// The `ID` field from `jwt.RegisteredClaims` will be used as the JTI.
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a new JWT access token for a user.
func (j *JWTUtil) GenerateAccessToken(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute) // Token expires in 60 minutes
	jti := uuid.NewString()
	return j.generateToken(userID, jti, expirationTime)
}

// GenerateRefreshToken creates a new JWT refresh token for a user.
func (j *JWTUtil) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(24 * 7 * time.Hour) // Token expires in 7 days
	jti := uuid.NewString()
	return j.generateToken(userID, jti, expirationTime)
}

func (j *JWTUtil) generateToken(userID uuid.UUID, jti string, expirationTime time.Time) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-moon-app",    // Optional: an identifier for the issuer
			ID:        jti,              // Set the JWT ID (JTI)
			Subject:   userID.String(), // Optional: can use UserID as subject
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken parses and validates a JWT token string.
// It returns the custom claims if the token is valid.
func (j *JWTUtil) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}
