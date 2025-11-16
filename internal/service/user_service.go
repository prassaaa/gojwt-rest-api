package service

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/repository"
	"gojwt-rest-api/internal/utils"
	"time"
)

// UserService defines the interface for user business logic
type UserService interface {
	Register(req *domain.RegisterRequest) (*domain.User, error)
	Login(req *domain.LoginRequest) (*domain.LoginResponse, error)
	RefreshToken(req *domain.RefreshTokenRequest) (*domain.RefreshTokenResponse, error)
	Logout(userID uint, req *domain.LogoutRequest) error
	GetUserByID(id uint) (*domain.User, error)
	GetAllUsers(pagination *domain.PaginationQuery) ([]*domain.User, int64, error)
	UpdateUser(id uint, req *domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(id uint) error
	// Self-service methods
	ChangePassword(userID uint, req *domain.ChangePasswordRequest) error
	UpdateOwnProfile(userID uint, req *domain.UpdateProfileRequest) (*domain.User, error)
}

// userServiceImpl is the implementation of UserService
type userServiceImpl struct {
	userRepo          repository.UserRepository
	tokenRepo         repository.TokenRepository
	jwtSecret         string
	accessTokenExpiry time.Duration
	refreshTokenExpiry time.Duration
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtSecret string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) UserService {
	return &userServiceImpl{
		userRepo:           userRepo,
		tokenRepo:          tokenRepo,
		jwtSecret:          jwtSecret,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Register registers a new user
func (s *userServiceImpl) Register(req *domain.RegisterRequest) (*domain.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		// Handle potential database errors
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, domain.ErrFailedToHashPassword
	}

	// Create user
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, domain.ErrFailedToCreateUser
	}

	return user, nil
}

// Login authenticates a user and returns JWT tokens
func (s *userServiceImpl) Login(req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Check password
	if err := utils.CheckPassword(user.Password, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate JWT token pair
	tokenPair, tokenFamily, err := utils.GenerateTokenPair(
		user.ID,
		user.Email,
		s.jwtSecret,
		s.accessTokenExpiry,
		s.refreshTokenExpiry,
	)
	if err != nil {
		return nil, domain.ErrFailedToGenerateToken
	}

	// Store refresh token in database
	refreshToken := &domain.RefreshToken{
		UserID:      user.ID,
		Token:       tokenPair.RefreshToken,
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(s.refreshTokenExpiry),
	}

	if err := s.tokenRepo.CreateRefreshToken(refreshToken); err != nil {
		return nil, domain.ErrFailedToCreateRefreshToken
	}

	response := &domain.LoginResponse{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}

	return response, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *userServiceImpl) RefreshToken(req *domain.RefreshTokenRequest) (*domain.RefreshTokenResponse, error) {
	// Find refresh token in database
	storedToken, err := s.tokenRepo.FindRefreshTokenByToken(req.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	// Check if token is valid
	if !storedToken.IsValid() {
		if storedToken.IsRevoked {
			// Token reuse detected - revoke entire token family
			_ = s.tokenRepo.RevokeTokenFamily(storedToken.TokenFamily)
			return nil, domain.ErrTokenReused
		}
		return nil, domain.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.FindByID(storedToken.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Generate new token pair (token rotation)
	newTokenPair, _, err := utils.GenerateTokenPair(
		user.ID,
		user.Email,
		s.jwtSecret,
		s.accessTokenExpiry,
		s.refreshTokenExpiry,
	)
	if err != nil {
		return nil, domain.ErrFailedToGenerateToken
	}

	// Revoke old refresh token
	now := time.Now()
	storedToken.IsRevoked = true
	storedToken.RevokedAt = &now
	replacedBy := newTokenPair.RefreshToken
	storedToken.ReplacedBy = &replacedBy

	if err := s.tokenRepo.UpdateRefreshToken(storedToken); err != nil {
		return nil, err
	}

	// Store new refresh token with same family (for rotation tracking)
	newRefreshToken := &domain.RefreshToken{
		UserID:      user.ID,
		Token:       newTokenPair.RefreshToken,
		TokenFamily: storedToken.TokenFamily, // Same family for rotation tracking
		ExpiresAt:   time.Now().Add(s.refreshTokenExpiry),
	}

	if err := s.tokenRepo.CreateRefreshToken(newRefreshToken); err != nil {
		return nil, domain.ErrFailedToCreateRefreshToken
	}

	response := &domain.RefreshTokenResponse{
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		ExpiresIn:    newTokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}

	return response, nil
}

// Logout revokes refresh token and blacklists access token
func (s *userServiceImpl) Logout(userID uint, req *domain.LogoutRequest) error {
	// Revoke refresh token if provided
	if req.RefreshToken != "" {
		if err := s.tokenRepo.RevokeRefreshToken(req.RefreshToken); err != nil {
			// Don't fail logout if refresh token is already revoked or not found
			// Just log and continue
		}
	}

	// Optionally: revoke all user's refresh tokens for "logout from all devices"
	// Uncomment below to enable:
	// if err := s.tokenRepo.RevokeAllUserRefreshTokens(userID); err != nil {
	// 	return err
	// }

	return nil
}

// GetUserByID retrieves a user by ID
func (s *userServiceImpl) GetUserByID(id uint) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetAllUsers retrieves all users with pagination
func (s *userServiceImpl) GetAllUsers(pagination *domain.PaginationQuery) ([]*domain.User, int64, error) {
	// Set default pagination values
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}
	if pagination.PageSize > 100 {
		pagination.PageSize = 100 // Max page size
	}

	return s.userRepo.FindAll(pagination)
}

// UpdateUser updates a user
func (s *userServiceImpl) UpdateUser(id uint, req *domain.UpdateUserRequest) (*domain.User, error) {
	// Find existing user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and if it's already taken
	if req.Email != "" && req.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(req.Email)
		if err != nil && err != domain.ErrUserNotFound {
			// Handle potential database errors
			return nil, err
		}
		if existingUser != nil {
			return nil, domain.ErrEmailAlreadyInUse
		}
		user.Email = req.Email
	}

	// Update name if provided
	if req.Name != "" {
		user.Name = req.Name
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, domain.ErrFailedToUpdateUser
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *userServiceImpl) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}

// ChangePassword allows a user to change their own password
func (s *userServiceImpl) ChangePassword(userID uint, req *domain.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := utils.CheckPassword(user.Password, req.OldPassword); err != nil {
		return domain.ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return domain.ErrFailedToHashPassword
	}

	// Update password
	user.Password = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return domain.ErrFailedToUpdateUser
	}

	return nil
}

// UpdateOwnProfile allows a user to update their own profile
func (s *userServiceImpl) UpdateOwnProfile(userID uint, req *domain.UpdateProfileRequest) (*domain.User, error) {
	// Find existing user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and if it's already taken
	if req.Email != "" && req.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(req.Email)
		if err != nil && err != domain.ErrUserNotFound {
			// Handle potential database errors
			return nil, err
		}
		if existingUser != nil {
			return nil, domain.ErrEmailAlreadyInUse
		}
		user.Email = req.Email
	}

	// Update name if provided
	if req.Name != "" {
		user.Name = req.Name
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, domain.ErrFailedToUpdateUser
	}

	return user, nil
}
