package unit

import (
	"errors"
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/internal/utils"
	"gojwt-rest-api/test/helpers"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_Register(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successful registration", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := helpers.CreateRegisterRequest("John Doe", "john@example.com", "password123")

		// Mock: email doesn't exist
		mockRepo.On("FindByEmail", req.Email).Return(nil, domain.ErrUserNotFound)
		// Mock: user creation succeeds
		mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := userService.Register(req)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Email, user.Email)
		assert.NotEqual(t, req.Password, user.Password, "Password should be hashed")
		assert.False(t, user.IsAdmin, "New users should not be admin by default")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Registration with existing email", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := helpers.CreateRegisterRequest("John Doe", "existing@example.com", "password123")
		existingUser := helpers.CreateTestUser(1, req.Email)

		// Mock: email already exists
		mockRepo.On("FindByEmail", req.Email).Return(existingUser, nil)

		user, err := userService.Register(req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Registration with database error on email check", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := helpers.CreateRegisterRequest("John Doe", "john@example.com", "password123")
		dbError := errors.New("database connection error")

		// Mock: database error
		mockRepo.On("FindByEmail", req.Email).Return(nil, dbError)

		user, err := userService.Register(req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, dbError, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Registration with create error", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := helpers.CreateRegisterRequest("John Doe", "john@example.com", "password123")

		// Mock: email doesn't exist
		mockRepo.On("FindByEmail", req.Email).Return(nil, domain.ErrUserNotFound)
		// Mock: user creation fails
		mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

		user, err := userService.Register(req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrFailedToCreateUser)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Login(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successful login", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		password := "password123"
		hashedPassword, _ := utils.HashPassword(password)
		user := &domain.User{
			ID:       1,
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: hashedPassword,
			IsAdmin:  false,
		}

		req := helpers.CreateLoginRequest(user.Email, password)

		// Mock: user found
		mockRepo.On("FindByEmail", req.Email).Return(user, nil)
		// Mock: token creation
		mockTokenRepo.On("CreateRefreshToken", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

		response, err := userService.Login(req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, user.ID, response.User.ID)
		assert.Equal(t, user.Email, response.User.Email)

		// Verify token is valid
		claims, err := utils.ValidateToken(response.AccessToken, jwtSecret)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)

		mockRepo.AssertExpectations(t)
		mockTokenRepo.AssertExpectations(t)
	})

	t.Run("Login with non-existent email", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := helpers.CreateLoginRequest("nonexistent@example.com", "password123")

		// Mock: user not found
		mockRepo.On("FindByEmail", req.Email).Return(nil, domain.ErrUserNotFound)

		response, err := userService.Login(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		correctPassword := "password123"
		wrongPassword := "wrongpassword"
		hashedPassword, _ := utils.HashPassword(correctPassword)
		user := &domain.User{
			ID:       1,
			Email:    "john@example.com",
			Password: hashedPassword,
		}

		req := helpers.CreateLoginRequest(user.Email, wrongPassword)

		// Mock: user found
		mockRepo.On("FindByEmail", req.Email).Return(user, nil)

		response, err := userService.Login(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successfully get user by ID", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		expectedUser := helpers.CreateTestUser(1, "john@example.com")

		// Mock: user found
		mockRepo.On("FindByID", uint(1)).Return(expectedUser, nil)

		user, err := userService.GetUserByID(1)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)

		mockRepo.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		// Mock: user not found
		mockRepo.On("FindByID", uint(999)).Return(nil, domain.ErrUserNotFound)

		user, err := userService.GetUserByID(999)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetAllUsers(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successfully get all users with valid pagination", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		expectedUsers := []*domain.User{
			helpers.CreateTestUser(1, "user1@example.com"),
			helpers.CreateTestUser(2, "user2@example.com"),
		}
		pagination := helpers.CreatePaginationQuery(1, 10, "")

		// Mock: return users
		mockRepo.On("FindAll", pagination).Return(expectedUsers, int64(2), nil)

		users, total, err := userService.GetAllUsers(pagination)

		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, expectedUsers, users)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Set default pagination values", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		pagination := helpers.CreatePaginationQuery(0, 0, "")

		// Mock: should be called with default values
		mockRepo.On("FindAll", mock.MatchedBy(func(p *domain.PaginationQuery) bool {
			return p.Page == 1 && p.PageSize == 10
		})).Return([]*domain.User{}, int64(0), nil)

		users, total, err := userService.GetAllUsers(pagination)

		require.NoError(t, err)
		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 10, pagination.PageSize)
		assert.Len(t, users, 0)
		assert.Equal(t, int64(0), total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Limit max page size to 100", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		pagination := helpers.CreatePaginationQuery(1, 200, "")

		// Mock: should be called with limited page size
		mockRepo.On("FindAll", mock.MatchedBy(func(p *domain.PaginationQuery) bool {
			return p.PageSize == 100
		})).Return([]*domain.User{}, int64(0), nil)

		_, _, err := userService.GetAllUsers(pagination)

		require.NoError(t, err)
		assert.Equal(t, 100, pagination.PageSize)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successfully update user name", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		existingUser := helpers.CreateTestUser(1, "john@example.com")
		req := helpers.CreateUpdateUserRequest("John Updated", "")

		// Mock: user found
		mockRepo.On("FindByID", uint(1)).Return(existingUser, nil)
		// Mock: update succeeds
		mockRepo.On("Update", mock.MatchedBy(func(u *domain.User) bool {
			return u.Name == "John Updated"
		})).Return(nil)

		user, err := userService.UpdateUser(1, req)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "John Updated", user.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Successfully update user email", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		existingUser := helpers.CreateTestUser(1, "old@example.com")
		newEmail := "new@example.com"
		req := helpers.CreateUpdateUserRequest("", newEmail)

		// Mock: user found
		mockRepo.On("FindByID", uint(1)).Return(existingUser, nil)
		// Mock: new email doesn't exist
		mockRepo.On("FindByEmail", newEmail).Return(nil, domain.ErrUserNotFound)
		// Mock: update succeeds
		mockRepo.On("Update", mock.MatchedBy(func(u *domain.User) bool {
			return u.Email == newEmail
		})).Return(nil)

		user, err := userService.UpdateUser(1, req)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, newEmail, user.Email)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Update fails when user not found", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := helpers.CreateUpdateUserRequest("John Updated", "")

		// Mock: user not found
		mockRepo.On("FindByID", uint(999)).Return(nil, domain.ErrUserNotFound)

		user, err := userService.UpdateUser(999, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Update fails when email already in use", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		existingUser := helpers.CreateTestUser(1, "john@example.com")
		anotherUser := helpers.CreateTestUser(2, "taken@example.com")
		req := helpers.CreateUpdateUserRequest("", "taken@example.com")

		// Mock: user found
		mockRepo.On("FindByID", uint(1)).Return(existingUser, nil)
		// Mock: email already taken
		mockRepo.On("FindByEmail", "taken@example.com").Return(anotherUser, nil)

		user, err := userService.UpdateUser(1, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyInUse)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Update with database error", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		existingUser := helpers.CreateTestUser(1, "john@example.com")
		req := helpers.CreateUpdateUserRequest("John Updated", "")

		// Mock: user found
		mockRepo.On("FindByID", uint(1)).Return(existingUser, nil)
		// Mock: update fails
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

		user, err := userService.UpdateUser(1, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrFailedToUpdateUser)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successfully delete user", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		// Mock: delete succeeds
		mockRepo.On("Delete", uint(1)).Return(nil)

		err := userService.DeleteUser(1)

		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete with database error", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		dbError := errors.New("database error")

		// Mock: delete fails
		mockRepo.On("Delete", uint(1)).Return(dbError)

		err := userService.DeleteUser(1)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successfully change password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		user := helpers.CreateTestUser(1, "john@example.com")
		req := &domain.ChangePasswordRequest{
			OldPassword: "password123",
			NewPassword: "newpassword123",
		}

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: update user
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

		err := userService.ChangePassword(1, req)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Change password with wrong old password", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		user := helpers.CreateTestUser(1, "john@example.com")
		req := &domain.ChangePasswordRequest{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword123",
		}

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)

		err := userService.ChangePassword(1, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Change password for non-existent user", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := &domain.ChangePasswordRequest{
			OldPassword: "password123",
			NewPassword: "newpassword123",
		}

		// Mock: user not found
		mockRepo.On("FindByID", uint(999)).Return(nil, domain.ErrUserNotFound)

		err := userService.ChangePassword(999, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Change password with database error on update", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		user := helpers.CreateTestUser(1, "john@example.com")
		req := &domain.ChangePasswordRequest{
			OldPassword: "password123",
			NewPassword: "newpassword123",
		}

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: update fails
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

		err := userService.ChangePassword(1, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFailedToUpdateUser)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateOwnProfile(t *testing.T) {
	jwtSecret := "test-secret"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	t.Run("Successfully update own profile", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		user := helpers.CreateTestUser(1, "john@example.com")
		req := &domain.UpdateProfileRequest{
			Name:  "John Updated",
			Email: "johnupdated@example.com",
		}

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: email doesn't exist (checking for duplicate)
		mockRepo.On("FindByEmail", req.Email).Return(nil, domain.ErrUserNotFound)
		// Mock: update succeeds
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

		updatedUser, err := userService.UpdateOwnProfile(1, req)

		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, req.Name, updatedUser.Name)
		assert.Equal(t, req.Email, updatedUser.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update own profile - name only", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		user := helpers.CreateTestUser(1, "john@example.com")
		req := &domain.UpdateProfileRequest{
			Name: "John Updated",
		}

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: update succeeds
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(nil)

		updatedUser, err := userService.UpdateOwnProfile(1, req)

		require.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, req.Name, updatedUser.Name)
		assert.Equal(t, user.Email, updatedUser.Email) // Email unchanged
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update own profile with duplicate email", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		user := helpers.CreateTestUser(1, "john@example.com")
		existingUser := helpers.CreateTestUser(2, "existing@example.com")
		req := &domain.UpdateProfileRequest{
			Name:  "John Updated",
			Email: "existing@example.com",
		}

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: email already exists
		mockRepo.On("FindByEmail", req.Email).Return(existingUser, nil)

		updatedUser, err := userService.UpdateOwnProfile(1, req)

		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyInUse)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update own profile for non-existent user", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		req := &domain.UpdateProfileRequest{
			Name: "John Updated",
		}

		// Mock: user not found
		mockRepo.On("FindByID", uint(999)).Return(nil, domain.ErrUserNotFound)

		updatedUser, err := userService.UpdateOwnProfile(999, req)

		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update own profile with database error", func(t *testing.T) {
		mockRepo := new(helpers.MockUserRepository)
		mockTokenRepo := new(helpers.MockTokenRepository)
		userService := service.NewUserService(mockRepo, mockTokenRepo, jwtSecret, accessExpiry, refreshExpiry)

		user := helpers.CreateTestUser(1, "john@example.com")
		req := &domain.UpdateProfileRequest{
			Name: "John Updated",
		}

		// Mock: find user
		mockRepo.On("FindByID", uint(1)).Return(user, nil)
		// Mock: update fails
		mockRepo.On("Update", mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

		updatedUser, err := userService.UpdateOwnProfile(1, req)

		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.ErrorIs(t, err, domain.ErrFailedToUpdateUser)
		mockRepo.AssertExpectations(t)
	})
}
