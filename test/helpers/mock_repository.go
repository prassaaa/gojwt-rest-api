package helpers

import (
	"gojwt-rest-api/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

// MockTokenRepository is a mock implementation of repository.TokenRepository
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uint) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindAll(pagination *domain.PaginationQuery) ([]*domain.User, int64, error) {
	args := m.Called(pagination)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockTokenRepository methods
func (m *MockTokenRepository) CreateRefreshToken(token *domain.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenRepository) FindRefreshTokenByToken(token string) (*domain.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockTokenRepository) FindRefreshTokensByUserID(userID uint) ([]*domain.RefreshToken, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.RefreshToken), args.Error(1)
}

func (m *MockTokenRepository) UpdateRefreshToken(token *domain.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenRepository) RevokeRefreshToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenRepository) RevokeAllUserRefreshTokens(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockTokenRepository) RevokeTokenFamily(tokenFamily string) error {
	args := m.Called(tokenFamily)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteExpiredRefreshTokens() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTokenRepository) AddToBlacklist(token *domain.TokenBlacklist) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenRepository) IsTokenBlacklisted(token string) (bool, error) {
	args := m.Called(token)
	return args.Bool(0), args.Error(1)
}

func (m *MockTokenRepository) DeleteExpiredBlacklistTokens() error {
	args := m.Called()
	return args.Error(0)
}
