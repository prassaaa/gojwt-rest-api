package repository

import (
	"gojwt-rest-api/internal/domain"
	"time"

	"gorm.io/gorm"
)

// tokenRepositoryImpl is the implementation of TokenRepository
type tokenRepositoryImpl struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepositoryImpl{db: db}
}

// CreateRefreshToken creates a new refresh token
func (r *tokenRepositoryImpl) CreateRefreshToken(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

// FindRefreshTokenByToken finds a refresh token by token string
func (r *tokenRepositoryImpl) FindRefreshTokenByToken(token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	err := r.db.Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}
	return &refreshToken, nil
}

// FindRefreshTokensByUserID finds all refresh tokens for a user
func (r *tokenRepositoryImpl) FindRefreshTokensByUserID(userID uint) ([]*domain.RefreshToken, error) {
	var tokens []*domain.RefreshToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// UpdateRefreshToken updates a refresh token
func (r *tokenRepositoryImpl) UpdateRefreshToken(token *domain.RefreshToken) error {
	return r.db.Save(token).Error
}

// RevokeRefreshToken revokes a specific refresh token
func (r *tokenRepositoryImpl) RevokeRefreshToken(token string) error {
	now := time.Now()
	return r.db.Model(&domain.RefreshToken{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		}).Error
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func (r *tokenRepositoryImpl) RevokeAllUserRefreshTokens(userID uint) error {
	now := time.Now()
	return r.db.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		}).Error
}

// RevokeTokenFamily revokes all tokens in a token family (for security breach detection)
func (r *tokenRepositoryImpl) RevokeTokenFamily(tokenFamily string) error {
	now := time.Now()
	return r.db.Model(&domain.RefreshToken{}).
		Where("token_family = ? AND is_revoked = ?", tokenFamily, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		}).Error
}

// DeleteExpiredRefreshTokens deletes expired refresh tokens (cleanup)
func (r *tokenRepositoryImpl) DeleteExpiredRefreshTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&domain.RefreshToken{}).Error
}

// AddToBlacklist adds a token to the blacklist
func (r *tokenRepositoryImpl) AddToBlacklist(token *domain.TokenBlacklist) error {
	return r.db.Create(token).Error
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *tokenRepositoryImpl) IsTokenBlacklisted(token string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.TokenBlacklist{}).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		Count(&count).Error
	return count > 0, err
}

// DeleteExpiredBlacklistTokens deletes expired blacklisted tokens (cleanup)
func (r *tokenRepositoryImpl) DeleteExpiredBlacklistTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&domain.TokenBlacklist{}).Error
}
