package e2e

import (
	"bytes"
	"encoding/json"
	"gojwt-rest-api/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenEndpoint(t *testing.T) {
	router, _ := setupTestServer(t)

	// Register and login first
	registerReq := domain.RegisterRequest{
		Name:     "Refresh Test User",
		Email:    "refresh@test.com",
		Password: "password123",
	}

	registerBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Login to get tokens
	loginReq := domain.LoginRequest{
		Email:    "refresh@test.com",
		Password: "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp struct {
		Success bool                  `json:"success"`
		Message string                `json:"message"`
		Data    domain.LoginResponse  `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err)
	require.True(t, loginResp.Success)
	require.NotEmpty(t, loginResp.Data.AccessToken)
	require.NotEmpty(t, loginResp.Data.RefreshToken)

	originalAccessToken := loginResp.Data.AccessToken
	originalRefreshToken := loginResp.Data.RefreshToken

	t.Run("Successfully refresh token", func(t *testing.T) {
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: originalRefreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var refreshResp struct {
			Success bool                          `json:"success"`
			Message string                        `json:"message"`
			Data    domain.RefreshTokenResponse   `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &refreshResp)
		require.NoError(t, err)
		assert.True(t, refreshResp.Success)
		assert.NotEmpty(t, refreshResp.Data.AccessToken)
		assert.NotEmpty(t, refreshResp.Data.RefreshToken)
		assert.Equal(t, "Bearer", refreshResp.Data.TokenType)
		assert.Greater(t, refreshResp.Data.ExpiresIn, int64(0))

		// New tokens should be different from original
		assert.NotEqual(t, originalAccessToken, refreshResp.Data.AccessToken)
		assert.NotEqual(t, originalRefreshToken, refreshResp.Data.RefreshToken)
	})

	t.Run("Token rotation - old refresh token becomes invalid", func(t *testing.T) {
		// Try to use the original refresh token again (should fail)
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: originalRefreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.False(t, errorResp.Success)
	})

	t.Run("Fail with invalid refresh token", func(t *testing.T) {
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: "invalid-token",
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Fail with empty refresh token", func(t *testing.T) {
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: "",
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogoutEndpoint(t *testing.T) {
	router, _ := setupTestServer(t)

	// Register and login
	registerReq := domain.RegisterRequest{
		Name:     "Logout Test User",
		Email:    "logout@test.com",
		Password: "password123",
	}

	registerBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Login
	loginReq := domain.LoginRequest{
		Email:    "logout@test.com",
		Password: "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var loginResp struct {
		Success bool                  `json:"success"`
		Data    domain.LoginResponse  `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	accessToken := loginResp.Data.AccessToken
	refreshToken := loginResp.Data.RefreshToken

	t.Run("Successfully logout", func(t *testing.T) {
		logoutReq := domain.LogoutRequest{
			RefreshToken: refreshToken,
		}

		logoutBody, _ := json.Marshal(logoutReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(logoutBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var logoutResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &logoutResp)
		require.NoError(t, err)
		assert.True(t, logoutResp.Success)
	})

	t.Run("Cannot refresh token after logout", func(t *testing.T) {
		// Try to use refresh token after logout
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Fail logout without auth token", func(t *testing.T) {
		logoutReq := domain.LogoutRequest{
			RefreshToken: refreshToken,
		}

		logoutBody, _ := json.Marshal(logoutReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(logoutBody))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestTokenReuseDetection(t *testing.T) {
	router, _ := setupTestServer(t)

	// Register and login
	registerReq := domain.RegisterRequest{
		Name:     "Reuse Test User",
		Email:    "reuse@test.com",
		Password: "password123",
	}

	registerBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Login
	loginReq := domain.LoginRequest{
		Email:    "reuse@test.com",
		Password: "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var loginResp struct {
		Data domain.LoginResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	refreshToken1 := loginResp.Data.RefreshToken

	t.Run("Token reuse detection scenario", func(t *testing.T) {
		// First refresh - should work
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: refreshToken1,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var refreshResp struct {
			Data domain.RefreshTokenResponse `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &refreshResp)
		refreshToken2 := refreshResp.Data.RefreshToken

		// Second refresh with SAME token (refreshToken1) - should fail (token reuse)
		refreshReq = domain.RefreshTokenRequest{
			RefreshToken: refreshToken1,
		}

		refreshBody, _ = json.Marshal(refreshReq)
		w = httptest.NewRecorder()
		req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should fail with 401 - token already used
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.False(t, errorResp.Success)

		// Even the new token (refreshToken2) should be invalid now due to family revocation
		// This protects against token theft scenarios
		refreshReq = domain.RefreshTokenRequest{
			RefreshToken: refreshToken2,
		}

		refreshBody, _ = json.Marshal(refreshReq)
		w = httptest.NewRecorder()
		req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestTokenExpiration(t *testing.T) {
	// This test would require setting very short token expiration
	// and waiting for it to expire. Skip in regular test runs.
	t.Skip("Token expiration test requires waiting for actual expiration")

	// Would need to configure server with very short token expiration (e.g., 1 second)
	// Then test that token becomes invalid after expiration
}

func TestMultipleRefreshChain(t *testing.T) {
	router, _ := setupTestServer(t)

	// Register and login
	registerReq := domain.RegisterRequest{
		Name:     "Chain Test User",
		Email:    "chain@test.com",
		Password: "password123",
	}

	registerBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Login
	loginReq := domain.LoginRequest{
		Email:    "chain@test.com",
		Password: "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var loginResp struct {
		Data domain.LoginResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	t.Run("Chain multiple refresh operations", func(t *testing.T) {
		currentRefreshToken := loginResp.Data.RefreshToken
		var tokens []string

		// Perform 5 refresh operations in a chain
		for i := 0; i < 5; i++ {
			refreshReq := domain.RefreshTokenRequest{
				RefreshToken: currentRefreshToken,
			}

			refreshBody, _ := json.Marshal(refreshReq)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Refresh %d should succeed", i+1)

			var refreshResp struct {
				Data domain.RefreshTokenResponse `json:"data"`
			}
			json.Unmarshal(w.Body.Bytes(), &refreshResp)

			tokens = append(tokens, refreshResp.Data.RefreshToken)
			currentRefreshToken = refreshResp.Data.RefreshToken

			// Small delay to ensure different timestamps
			time.Sleep(10 * time.Millisecond)
		}

		// Verify all tokens are unique
		tokenSet := make(map[string]bool)
		for _, token := range tokens {
			assert.False(t, tokenSet[token], "Found duplicate token in chain")
			tokenSet[token] = true
		}

		// Old tokens should not work
		firstRefreshToken := tokens[0]
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: firstRefreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Old token should not work")
	})
}

func TestAccessTokenUsageWithRefresh(t *testing.T) {
	router, _ := setupTestServer(t)

	// Register and login
	registerReq := domain.RegisterRequest{
		Name:     "Access Token Test",
		Email:    "accesstoken@test.com",
		Password: "password123",
	}

	registerBody, _ := json.Marshal(registerReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Login
	loginReq := domain.LoginRequest{
		Email:    "accesstoken@test.com",
		Password: "password123",
	}

	loginBody, _ := json.Marshal(loginReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var loginResp struct {
		Data domain.LoginResponse `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	t.Run("Use access token for protected endpoint", func(t *testing.T) {
		// Try to access protected endpoint with access token
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Access token still valid after refresh", func(t *testing.T) {
		// Refresh to get new tokens
		refreshReq := domain.RefreshTokenRequest{
			RefreshToken: loginResp.Data.RefreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		var refreshResp struct {
			Data domain.RefreshTokenResponse `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &refreshResp)

		// Old access token should still work (until it expires)
		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+loginResp.Data.AccessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// New access token should also work
		w = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+refreshResp.Data.AccessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
