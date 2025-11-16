package e2e

import (
	"bytes"
	"encoding/json"
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/handler"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/pkg/validator"
	"gojwt-rest-api/test/helpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("Successfully register new user", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/register", authHandler.Register)

		reqBody := map[string]string{
			"name":     "John Doe",
			"email":    "john@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: email doesn't exist
		mockRepo.On("FindByEmail", "john@example.com").Return(nil, domain.ErrUserNotFound)
		// Mock: user creation succeeds
		mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "registered")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Register with invalid email format", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/register", authHandler.Register)

		reqBody := map[string]string{
			"name":     "John Doe",
			"email":    "invalid-email",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})

	t.Run("Register with short password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/register", authHandler.Register)

		reqBody := map[string]string{
			"name":     "John Doe",
			"email":    "john@example.com",
			"password": "12345", // Too short (< 6 chars)
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})

	t.Run("Register with existing email", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/register", authHandler.Register)

		existingUser := helpers.CreateTestUser(1, "existing@example.com")
		reqBody := map[string]string{
			"name":     "John Doe",
			"email":    "existing@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: email already exists
		mockRepo.On("FindByEmail", "existing@example.com").Return(existingUser, nil)

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "already exists")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Register with missing fields", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/register", authHandler.Register)

		reqBody := map[string]string{
			"email": "john@example.com",
			// Missing name and password
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("Successfully login with valid credentials", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		jwtSecret := "test-secret"
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/login", authHandler.Login)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"email":    "john@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: user found
		mockRepo.On("FindByEmail", "john@example.com").Return(user, nil)
		// Mock: token creation
		mockTokenRepo.On("CreateRefreshToken", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["data"])

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
		assert.NotNil(t, data["user"])

		mockRepo.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("Login with non-existent email", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/login", authHandler.Login)

		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: user not found
		mockRepo.On("FindByEmail", "nonexistent@example.com").Return(nil, domain.ErrUserNotFound)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Message, "invalid")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/login", authHandler.Login)

		user := helpers.CreateTestUser(1, "john@example.com")
		reqBody := map[string]string{
			"email":    "john@example.com",
			"password": "wrongpassword",
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock: user found
		mockRepo.On("FindByEmail", "john@example.com").Return(user, nil)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Login with invalid email format", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/login", authHandler.Login)

		reqBody := map[string]string{
			"email":    "invalid-email",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})

	t.Run("Login with empty body", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)
		v, _ := validator.New()
		authHandler := handler.NewAuthHandler(userService, v)

		router := setupRouter()
		router.POST("/login", authHandler.Login)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response domain.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
	})
}
