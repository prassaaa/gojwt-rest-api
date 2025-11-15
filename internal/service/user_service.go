package service

import (
	"errors"
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/repository"
	"gojwt-rest-api/internal/utils"
	"time"
)

// UserService defines the interface for user business logic
type UserService interface {
	Register(req *domain.RegisterRequest) (*domain.User, error)
	Login(req *domain.LoginRequest) (*domain.LoginResponse, string, error)
	GetUserByID(id uint) (*domain.User, error)
	GetAllUsers(pagination *domain.PaginationQuery) ([]*domain.User, int64, error)
	UpdateUser(id uint, req *domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(id uint) error
}

// userServiceImpl is the implementation of UserService
type userServiceImpl struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) UserService {
	return &userServiceImpl{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Register registers a new user
func (s *userServiceImpl) Register(req *domain.RegisterRequest) (*domain.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *userServiceImpl) Login(req *domain.LoginRequest) (*domain.LoginResponse, string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	// Check password
	if err := utils.CheckPassword(user.Password, req.Password); err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	response := &domain.LoginResponse{
		User:  user.ToResponse(),
		Token: token,
	}

	return response, token, nil
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
		existingUser, _ := s.userRepo.FindByEmail(req.Email)
		if existingUser != nil {
			return nil, errors.New("email already in use")
		}
		user.Email = req.Email
	}

	// Update name if provided
	if req.Name != "" {
		user.Name = req.Name
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, errors.New("failed to update user")
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *userServiceImpl) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}
