package e2e

import (
	"bytes"
	"encoding/json"
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/handler"
	"gojwt-rest-api/internal/middleware"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/internal/utils"
	"gojwt-rest-api/pkg/validator"
	"gojwt-rest-api/test/helpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProfileHandler_GetOwnProfile(t *testing.T) {
	t.Run("Successfully get own profile", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.GET("/profile", profileHandler.GetOwnProfile)

		user := helpers.CreateTestUser(1, "john@example.com")

		// Mock: user found
		mockRepo.On("FindByID", uint(1)).Return(user, nil)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]interface{})
		assert.Equal(t, user.Email, data["email"])
		assert.Equal(t, user.Name, data["name"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("Get profile without authentication", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.GET("/profile", profileHandler.GetOwnProfile)

		req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestProfileHandler_UpdateOwnProfile(t *testing.T) {
	t.Run("Successfully update own profile", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile", profileHandler.UpdateOwnProfile)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"name":  "John Updated",
			"email": "johnupdated@example.com",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: email doesn't exist (checking for duplicate)
		mockRepo.On("FindByEmail", "johnupdated@example.com").Return(nil, domain.ErrUserNotFound)
		// Mock: update succeeds
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response["message"], "updated")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Update profile with invalid email format", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile", profileHandler.UpdateOwnProfile)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"name":  "John Updated",
			"email": "invalid-email",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})

	t.Run("Update profile with duplicate email", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile", profileHandler.UpdateOwnProfile)

		user := helpers.CreateTestUser(1, "john@example.com")
		existingUser := helpers.CreateTestUser(2, "existing@example.com")
		reqBody := map[string]string{
			"name":  "John Updated",
			"email": "existing@example.com",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: email already exists
		mockRepo.On("FindByEmail", "existing@example.com").Return(existingUser, nil)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "already in use")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Update profile without authentication", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile", profileHandler.UpdateOwnProfile)

		reqBody := map[string]string{
			"name": "John Updated",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestProfileHandler_ChangePassword(t *testing.T) {
	t.Run("Successfully change password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile/password", profileHandler.ChangePassword)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"old_password": "password123",
			"new_password": "newpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: update succeeds
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodPut, "/profile/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Password changed")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Change password with wrong old password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile/password", profileHandler.ChangePassword)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"old_password": "wrongpassword",
			"new_password": "newpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodPut, "/profile/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "incorrect")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Change password with short new password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile/password", profileHandler.ChangePassword)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"old_password": "password123",
			"new_password": "12345", // Too short
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodPut, "/profile/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})

	t.Run("Change password without authentication", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile/password", profileHandler.ChangePassword)

		reqBody := map[string]string{
			"old_password": "password123",
			"new_password": "newpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPut, "/profile/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Change password with missing fields", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, jwtSecret, 24*time.Hour)
		v, _ := validator.New()
		profileHandler := handler.NewProfileHandler(userService, v)

		router := setupRouter()
		router.Use(middleware.AuthMiddleware(jwtSecret))
		router.PUT("/profile/password", profileHandler.ChangePassword)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"old_password": "password123",
			// Missing new_password
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Generate valid token
		token, _ := utils.GenerateToken(user.ID, user.Email, jwtSecret, 24*time.Hour)

		req, _ := http.NewRequest(http.MethodPut, "/profile/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})
}
