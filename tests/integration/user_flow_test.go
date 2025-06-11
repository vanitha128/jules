package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-moon/internal/user"
)

// Helper to register and login a unique user for user flow tests
func registerAndLoginUser(t *testing.T, router *gin.Engine, uniqueEmail, password string) (userID, accessToken, refreshToken string) {
	clearDatabaseTables() // Clear DB for each user registration in this helper to ensure isolation

	dob, _ := time.Parse("2006-01-02", "1990-01-01")
	registerReq := user.RegisterRequest{
		Email:     uniqueEmail,
		Password:  password,
		FirstName: "UserFlow",
		LastName:  "Test",
		DOB:       dob,
	}
	rrReg, bodyReg := performRequest(router, http.MethodPost, "/users/register", registerReq)
	require.Equal(t, http.StatusCreated, rrReg.Code, "Helper: Registration failed")

	userIDRaw, ok := bodyReg["userID"]
	require.True(t, ok, "Helper: UserID not in registration response")
	userID = userIDRaw.(string)
	require.NotEmpty(t, userID, "Helper: UserID is empty")

	loginReq := user.LoginRequest{Email: uniqueEmail, Password: password}
	rrLogin, bodyLogin := performRequest(router, http.MethodPost, "/users/login", loginReq)
	require.Equal(t, http.StatusOK, rrLogin.Code, "Helper: Login failed")

	accessToken, okAccess := bodyLogin["access_token"].(string)
	require.True(t, okAccess, "Helper: Access token not found or not string")
	refreshToken, okRefresh := bodyLogin["refresh_token"].(string)
	require.True(t, okRefresh, "Helper: Refresh token not found or not string")

	require.NotEmpty(t, accessToken, "Helper: Access token is empty")
	require.NotEmpty(t, refreshToken, "Helper: Refresh token is empty")

	return userID, accessToken, refreshToken
}

func TestUserFlow_ChangePassword(t *testing.T) {
	router := baseRouter
	uniqueEmail := fmt.Sprintf("changepass_%s@example.com", uuid.NewString()[:8])
	oldPassword := "oldPassword123"
	newPassword := "newPassword456"

	_, accessToken, _ := registerAndLoginUser(t, router, uniqueEmail, oldPassword)

	t.Run("Success", func(t *testing.T) {
		changePassReq := user.ChangePasswordRequest{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}
		jsonBody, _ := json.Marshal(changePassReq)
		req, _ := http.NewRequest(http.MethodPost, "/users/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var respBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, "Password changed successfully (placeholder)", respBody["message"]) // Current placeholder

		// Try to log in with the old password - should fail
		loginWithOldPassReq := user.LoginRequest{Email: uniqueEmail, Password: oldPassword}
		rrOldLogin, _ := performRequest(router, http.MethodPost, "/users/login", loginWithOldPassReq)
		assert.Equal(t, http.StatusUnauthorized, rrOldLogin.Code)

		// Log in with the new password - should succeed
		loginWithNewPassReq := user.LoginRequest{Email: uniqueEmail, Password: newPassword}
		rrNewLogin, bodyNewLogin := performRequest(router, http.MethodPost, "/users/login", loginWithNewPassReq)
		assert.Equal(t, http.StatusOK, rrNewLogin.Code)
		assert.NotEmpty(t, bodyNewLogin["access_token"]) // Store new tokens if needed for further actions
	})

	t.Run("Incorrect_Old_Password", func(t *testing.T) {
		// Note: registerAndLoginUser clears tables, so we need a fresh user or re-login the existing one if state is preserved.
		// For simplicity, let's assume the accessToken from the parent test is still valid for this sub-test
		// if the DB state for this user hasn't been wiped by another registerAndLoginUser call.
		// Better: Get a fresh token if unsure, or ensure subtests don't interfere with auth state.
		// The current registerAndLoginUser clears DB, so previous accessToken is for a now-deleted user.
		// We need to re-register/login for this sub-test or ensure ChangePassword doesn't depend on prior subtest login.
		// Let's re-register for true isolation.
		freshEmail := fmt.Sprintf("changepass_fail_%s@example.com", uuid.NewString()[:8])
		_, freshAccessToken, _ := registerAndLoginUser(t, router, freshEmail, oldPassword)


		changePassReq := user.ChangePasswordRequest{
			OldPassword: "completelyWrongOldPassword",
			NewPassword: newPassword,
		}
		jsonBody, _ := json.Marshal(changePassReq)
		req, _ := http.NewRequest(http.MethodPost, "/users/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+freshAccessToken)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		// User service returns ErrInvalidCredentials, handler maps to 401
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var respBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Contains(t, respBody["error"], "Invalid credentials")
	})
}


func TestUserFlow_UpdateProfile(t *testing.T) {
	router := baseRouter
	uniqueEmail := fmt.Sprintf("updateprofile_%s@example.com", uuid.NewString()[:8])
	password := "password123"

	userID, accessToken, _ := registerAndLoginUser(t, router, uniqueEmail, password)

	t.Run("Success", func(t *testing.T) {
		updateReq := user.UpdateUserRequest{
			FirstName: "UpdatedFirst",
			LastName:  "UpdatedLast",
		}
		jsonBody, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		// Current placeholder response:
		assert.Equal(t, "Profile updated successfully (placeholder)", respBody["message"])
		// Ideally, it would return the updated user data or the fields updated.
		// For now, checking placeholder.

		// Verify by fetching profile
		reqProfile, _ := http.NewRequest(http.MethodGet, "/users/me/profile", nil)
		reqProfile.Header.Set("Authorization", "Bearer "+accessToken)
		rrProfile := httptest.NewRecorder()
		router.ServeHTTP(rrProfile, reqProfile)

		assert.Equal(t, http.StatusOK, rrProfile.Code)
		var profileBody map[string]interface{}
		json.Unmarshal(rrProfile.Body.Bytes(), &profileBody)

		assert.Equal(t, userID, profileBody["id"])
		assert.Equal(t, updateReq.FirstName, profileBody["firstName"])
		assert.Equal(t, updateReq.LastName, profileBody["lastName"])
		assert.Equal(t, uniqueEmail, profileBody["email"]) // Email should not change
	})

	t.Run("Update_Only_FirstName", func(t *testing.T) {
		// Re-register and login to ensure a known starting state for FirstName/LastName
		freshEmail := fmt.Sprintf("updatepartial_%s@example.com", uuid.NewString()[:8])
		freshUserID, freshAccessToken, _ := registerAndLoginUser(t, router, freshEmail, password)

		updateReq := user.UpdateUserRequest{
			FirstName: "PartialUpdateFirst",
		}
		jsonBody, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+freshAccessToken)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify by fetching profile
		reqProfile, _ := http.NewRequest(http.MethodGet, "/users/me/profile", nil)
		reqProfile.Header.Set("Authorization", "Bearer "+freshAccessToken)
		rrProfile := httptest.NewRecorder()
		router.ServeHTTP(rrProfile, reqProfile)

		assert.Equal(t, http.StatusOK, rrProfile.Code)
		var profileBody map[string]interface{}
		json.Unmarshal(rrProfile.Body.Bytes(), &profileBody)

		assert.Equal(t, freshUserID, profileBody["id"])
		assert.Equal(t, updateReq.FirstName, profileBody["firstName"])
		assert.Equal(t, "Test", profileBody["lastName"]) // Original LastName from registerAndLoginUser
	})

	t.Run("Update_With_Empty_Payload_No_Change", func(t *testing.T) {
		// Re-register to ensure known state.
		currentEmail := fmt.Sprintf("emptyupdate_%s@example.com", uuid.NewString()[:8])
		currentUserID, currentAccessToken, _ := registerAndLoginUser(t, router, currentEmail, password)

		// Fetch initial profile to get original names
		reqOrigProfile, _ := http.NewRequest(http.MethodGet, "/users/me/profile", nil)
		reqOrigProfile.Header.Set("Authorization", "Bearer "+currentAccessToken)
		rrOrigProfile := httptest.NewRecorder()
		router.ServeHTTP(rrOrigProfile, reqOrigProfile)
		require.Equal(t, http.StatusOK, rrOrigProfile.Code)
		var origProfileBody map[string]interface{}
		json.Unmarshal(rrOrigProfile.Body.Bytes(), &origProfileBody)
		originalFirstName := origProfileBody["firstName"].(string)
		originalLastName := origProfileBody["lastName"].(string)


		updateReq := user.UpdateUserRequest{} // Empty payload
		jsonBody, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+currentAccessToken)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		// The handler has validation: `if req.FirstName == "" && req.LastName == "" { c.JSON(http.StatusBadRequest...)}`
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var respBody map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Contains(t, respBody["error"], "No fields provided for update")


		// Verify profile unchanged by fetching again
		reqProfile, _ := http.NewRequest(http.MethodGet, "/users/me/profile", nil)
		reqProfile.Header.Set("Authorization", "Bearer "+currentAccessToken)
		rrProfile := httptest.NewRecorder()
		router.ServeHTTP(rrProfile, reqProfile)
		assert.Equal(t, http.StatusOK, rrProfile.Code)
		var profileBody map[string]interface{}
		json.Unmarshal(rrProfile.Body.Bytes(), &profileBody)

		assert.Equal(t, currentUserID, profileBody["id"])
		assert.Equal(t, originalFirstName, profileBody["firstName"]) // Should be original
		assert.Equal(t, originalLastName, profileBody["lastName"])   // Should be original
	})
}
