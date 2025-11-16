package repository

import (
	"gojwt-rest-api/internal/domain"
)

// TokenRepository defines the interface for token operations
type TokenRepository interface {
	// Refresh Token operations
	CreateRefreshToken(token *domain.RefreshToken) error
	FindRefreshTokenByToken(token string) (*domain.RefreshToken, error)
	FindRefreshTokensByUserID(userID uint) ([]*domain.RefreshToken, error)
	UpdateRefreshToken(token *domain.RefreshToken) error
	RevokeRefreshToken(token string) error
	RevokeAllUserRefreshTokens(userID uint) error
	RevokeTokenFamily(tokenFamily string) error
	DeleteExpiredRefreshTokens() error

	// Token Blacklist operations
	AddToBlacklist(token *domain.TokenBlacklist) error
	IsTokenBlacklisted(token string) (bool, error)
	DeleteExpiredBlacklistTokens() error
}
