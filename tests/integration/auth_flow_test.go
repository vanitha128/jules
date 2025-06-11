package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-username/go-moon/internal/user" // For request/response structs
	// "github.com/your-username/go-moon/pkg/utils" // For utils.Claims if direct token inspection is needed
)

// Helper to make requests and return recorder and parsed body
func performRequest(router *gin.Engine, method, path string, body interface{}) (*httptest.ResponseRecorder, map[string]interface{}) {
	var reqBodyBytes []byte
	if body != nil {
		reqBodyBytes, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(reqBodyBytes))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var responseBody map[string]interface{}
	if rr.Body.Len() > 0 {
		_ = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	}
	return rr, responseBody
}


func TestAuthFlow_UserRegistrationAndLogin(t *testing.T) {
	router := baseRouter // Use the router from TestMain

	t.Run("Register_New_User_Success", func(t *testing.T) {
		clearDatabaseTables() // Clean slate
		// clearRedisCache() // If Redis is used for registration flow (e.g. rate limiting)

		dob, _ := time.Parse("2006-01-02", "1990-01-01")
		registerReq := user.RegisterRequest{
			Email:     "testuser@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			DOB:       dob,
		}

		rr, body := performRequest(router, http.MethodPost, "/users/register", registerReq)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "User registered successfully", body["message"])
		assert.NotEmpty(t, body["userID"])
	})

	t.Run("Register_Existing_User_Conflict", func(t *testing.T) {
		clearDatabaseTables()
		dob, _ := time.Parse("2006-01-02", "1990-01-01")
		registerReq := user.RegisterRequest{
			Email:     "existinguser@example.com",
			Password:  "password123",
			FirstName: "Existing",
			LastName:  "User",
			DOB:       dob,
		}
		// First registration
		performRequest(router, http.MethodPost, "/users/register", registerReq)

		// Attempt to register same email again
		rr, body := performRequest(router, http.MethodPost, "/users/register", registerReq)

		// Expecting a conflict or bad request due to unique email constraint
		// GORM returns a generic error which our service maps to "failed to register user" and HTTP 500
		// A more specific error handling (e.g. checking for unique constraint violation) would return 409 or 400.
		// Current User service's Register method does not check for email existence before CreateUser.
		// The database unique constraint on email will cause CreateUser in repo to fail.
		// The service now returns ErrEmailAlreadyExists, and the handler maps this to HTTP 409.
		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Equal(t, user.ErrEmailAlreadyExists.Error(), body["error"])
	})

	t.Run("Login_Success", func(t *testing.T) {
		clearDatabaseTables()
		email := "loginuser@example.com"
		password := "password123"
		dob, _ := time.Parse("2006-01-02", "1990-01-01")
		registerReq := user.RegisterRequest{
			Email: email, Password: password, FirstName: "Login", LastName: "User", DOB: dob,
		}
		performRequest(router, http.MethodPost, "/users/register", registerReq) // Register first

		loginReq := user.LoginRequest{Email: email, Password: password}
		rr, body := performRequest(router, http.MethodPost, "/users/login", loginReq)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, body["access_token"])
		assert.NotEmpty(t, body["refresh_token"])
	})

	t.Run("Login_Incorrect_Credentials_Wrong_Password", func(t *testing.T) {
		clearDatabaseTables()
		email := "wrongpass@example.com"
		password := "password123"
		dob, _ := time.Parse("2006-01-02", "1990-01-01")
		registerReq := user.RegisterRequest{
			Email: email, Password: password, FirstName: "Wrong", LastName: "Pass", DOB: dob,
		}
		performRequest(router, http.MethodPost, "/users/register", registerReq)

		loginReq := user.LoginRequest{Email: email, Password: "wrongpassword"}
		rr, body := performRequest(router, http.MethodPost, "/users/login", loginReq)

		// userService.Login returns ErrInvalidCredentials, which handler maps to StatusUnauthorized
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, body["error"], "Invalid credentials")
	})

	t.Run("Login_Incorrect_Credentials_User_Not_Found", func(t *testing.T) {
		clearDatabaseTables() // Ensure no users from other tests
		loginReq := user.LoginRequest{Email: "nonexistent@example.com", Password: "password123"}
		rr, body := performRequest(router, http.MethodPost, "/users/login", loginReq)

		assert.Equal(t, http.StatusUnauthorized, rr.Code) // userService.Login returns ErrInvalidCredentials
		assert.Contains(t, body["error"], "Invalid credentials")
	})
}

// Helper function to register and login a user for auth-related tests, returns tokens and user ID
func registerAndLoginForAuthTest(t *testing.T, router *gin.Engine, email, password string) (accessToken, refreshToken, userID string) {
	clearDatabaseTables() // Ensure clean state before this combined operation
	// clearRedisCache() // If Redis state needs clearing for this specific helper

	dob, _ := time.Parse("2006-01-02", "1990-01-01")
	registerReq := user.RegisterRequest{
		Email: email, Password: password, FirstName: "AuthFlow", LastName: "User", DOB: dob,
	}
	rrReg, bodyReg := performRequest(router, http.MethodPost, "/users/register", registerReq)
	require.Equal(t, http.StatusCreated, rrReg.Code, "Registration failed in helper")
	userID = bodyReg["userID"].(string)
	require.NotEmpty(t, userID, "UserID not found in registration response in helper")

	loginReq := user.LoginRequest{Email: email, Password: password}
	rrLogin, bodyLogin := performRequest(router, http.MethodPost, "/users/login", loginReq)
	require.Equal(t, http.StatusOK, rrLogin.Code, "Login failed in helper")

	accessTokenVal, okAccess := bodyLogin["access_token"].(string)
	require.True(t, okAccess, "Access token not found or not a string in login response")
	refreshTokenVal, okRefresh := bodyLogin["refresh_token"].(string)
	require.True(t, okRefresh, "Refresh token not found or not a string in login response")

	require.NotEmpty(t, accessTokenVal, "Access token is empty in helper")
	require.NotEmpty(t, refreshTokenVal, "Refresh token is empty in helper")

	return accessTokenVal, refreshTokenVal, userID
}

func TestAuthFlow_Logout(t *testing.T) {
	router := baseRouter // Use the router from TestMain
	email := "logout@example.com"
	password := "password123"

	accessToken, _, _ := registerAndLoginForAuthTest(t, router, email, password)

	// 1. Access a protected route with the access token
	// The performRequest helper sends JSON, for GET with only headers, make request directly.
	reqProfile, _ := http.NewRequest(http.MethodGet, "/users/me/profile", nil)
	reqProfile.Header.Set("Authorization", "Bearer "+accessToken)

	rrProfile := httptest.NewRecorder()
	router.ServeHTTP(rrProfile, reqProfile)
	assert.Equal(t, http.StatusOK, rrProfile.Code, "Accessing protected route should succeed before logout")

	// 2. Call /auth/logout with the access token
	reqLogout, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	reqLogout.Header.Set("Authorization", "Bearer "+accessToken)
	rrLogout := httptest.NewRecorder()
	router.ServeHTTP(rrLogout, reqLogout)
	assert.Equal(t, http.StatusOK, rrLogout.Code, "Logout request should succeed")
	// Service returns success message, so 200. If it were 204, no body parsing.

	// 3. Try to access the protected route again with the same access token
	reqProfileAgain, _ := http.NewRequest(http.MethodGet, "/users/me/profile", nil)
	reqProfileAgain.Header.Set("Authorization", "Bearer "+accessToken)
	rrProfileAgain := httptest.NewRecorder()
	router.ServeHTTP(rrProfileAgain, reqProfileAgain)

	assert.Equal(t, http.StatusUnauthorized, rrProfileAgain.Code, "Accessing protected route should fail after logout")
	var bodyLogoutAttempt map[string]interface{}
	err := json.Unmarshal(rrProfileAgain.Body.Bytes(), &bodyLogoutAttempt)
	require.NoError(t, err, "Failed to parse response body after logout attempt")
	assert.Contains(t, bodyLogoutAttempt["error"], "Token has been revoked", "Error message should indicate token revocation")
}

func TestAuthFlow_RefreshToken(t *testing.T) {
	router := baseRouter // Use the router from TestMain
	email := "refresh@example.com"
	password := "password123"

	oldAccessToken, oldRefreshToken, _ := registerAndLoginForAuthTest(t, router, email, password)

	// Use the refresh token to get a new set of tokens
	refreshReqPayload := gin.H{"refresh_token": oldRefreshToken} // Use gin.H for simple JSON bodies
	rrRefresh, bodyRefresh := performRequest(router, http.MethodPost, "/auth/refresh", refreshReqPayload)
	require.Equal(t, http.StatusOK, rrRefresh.Code, "Refresh token request failed")

	newAccessToken, okAccess := bodyRefresh["access_token"].(string)
	require.True(t, okAccess, "New access token not found or not a string")
	newRefreshToken, okRefresh := bodyRefresh["refresh_token"].(string)
	require.True(t, okRefresh, "New refresh token not found or not a string")

	require.NotEmpty(t, newAccessToken, "New access token is empty")
	require.NotEmpty(t, newRefreshToken, "New refresh token is empty")
	assert.NotEqual(t, oldAccessToken, newAccessToken, "New access token should be different from old")
	assert.NotEqual(t, oldRefreshToken, newRefreshToken, "New refresh token should be different from old")

	// Use the *old* refresh token again, should fail as it's blacklisted/invalidated
	rrOldRefresh, bodyOldRefresh := performRequest(router, http.MethodPost, "/auth/refresh", gin.H{"refresh_token": oldRefreshToken})
	assert.Equal(t, http.StatusUnauthorized, rrOldRefresh.Code, "Using old refresh token should fail")
	assert.Contains(t, bodyOldRefresh["error"], "refresh token has been used or revoked", "Error message for used refresh token mismatch")

	// Use the new access token to access a protected route
	reqProfile, _ := http.NewRequest(http.MethodGet, "/users/me/profile", nil)
	reqProfile.Header.Set("Authorization", "Bearer "+newAccessToken)
	rrProfile := httptest.NewRecorder()
	router.ServeHTTP(rrProfile, reqProfile)
	assert.Equal(t, http.StatusOK, rrProfile.Code, "Accessing protected route with new access token should succeed")

	// Optional: Try to use an access token as a refresh token
	rrAccessAsRefresh, bodyAccessAsRefresh := performRequest(router, http.MethodPost, "/auth/refresh", gin.H{"refresh_token": newAccessToken})
	assert.Equal(t, http.StatusUnauthorized, rrAccessAsRefresh.Code)
	assert.Contains(t, bodyAccessAsRefresh["error"], "refresh token validation failed")

	// Optional: Try to use a malformed refresh token
	rrMalformed, bodyMalformed := performRequest(router, http.MethodPost, "/auth/refresh", gin.H{"refresh_token": "this.is.malformed"})
	assert.Equal(t, http.StatusUnauthorized, rrMalformed.Code)
	assert.Contains(t, bodyMalformed["error"], "refresh token validation failed")
}
