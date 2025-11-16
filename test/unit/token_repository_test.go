package unit

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/repository"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTokenMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestCreateRefreshToken(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	refreshToken := &domain.RefreshToken{
		UserID:      1,
		Token:       "test-refresh-token",
		TokenFamily: "family-123",
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `refresh_tokens`")).
		WithArgs(
			refreshToken.UserID,
			refreshToken.Token,
			refreshToken.TokenFamily,
			refreshToken.ExpiresAt,
			sqlmock.AnyArg(), // IsRevoked
			sqlmock.AnyArg(), // RevokedAt
			sqlmock.AnyArg(), // ReplacedBy
			sqlmock.AnyArg(), // CreatedAt
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.CreateRefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRefreshTokenByToken(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	token := "test-refresh-token"
	expectedToken := &domain.RefreshToken{
		ID:          1,
		UserID:      1,
		Token:       token,
		TokenFamily: "family-123",
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		IsRevoked:   false,
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "token", "token_family", "expires_at", "is_revoked", "created_at"}).
		AddRow(expectedToken.ID, expectedToken.UserID, expectedToken.Token, expectedToken.TokenFamily, expectedToken.ExpiresAt, expectedToken.IsRevoked, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `refresh_tokens` WHERE token = ? ORDER BY `refresh_tokens`.`id` LIMIT ?")).
		WithArgs(token, 1).
		WillReturnRows(rows)

	result, err := repo.FindRefreshTokenByToken(token)
	require.NoError(t, err)
	assert.Equal(t, expectedToken.Token, result.Token)
	assert.Equal(t, expectedToken.UserID, result.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRefreshTokenByToken_NotFound(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	token := "non-existent-token"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `refresh_tokens` WHERE token = ? ORDER BY `refresh_tokens`.`id` LIMIT ?")).
		WithArgs(token, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.FindRefreshTokenByToken(token)
	assert.Error(t, err)
	assert.Equal(t, domain.ErrTokenNotFound, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindRefreshTokensByUserID(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	userID := uint(1)
	rows := sqlmock.NewRows([]string{"id", "user_id", "token", "token_family", "expires_at", "is_revoked", "created_at"}).
		AddRow(1, userID, "token1", "family1", time.Now().Add(7*24*time.Hour), false, time.Now()).
		AddRow(2, userID, "token2", "family1", time.Now().Add(7*24*time.Hour), false, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `refresh_tokens` WHERE user_id = ?")).
		WithArgs(userID).
		WillReturnRows(rows)

	tokens, err := repo.FindRefreshTokensByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeRefreshToken(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	token := "token-to-revoke"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `refresh_tokens` SET `is_revoked`=?,`revoked_at`=? WHERE token = ?")).
		WithArgs(true, sqlmock.AnyArg(), token).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.RevokeRefreshToken(token)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeAllUserRefreshTokens(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	userID := uint(1)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `refresh_tokens` SET `is_revoked`=?,`revoked_at`=? WHERE user_id = ? AND is_revoked = ?")).
		WithArgs(true, sqlmock.AnyArg(), userID, false).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	err := repo.RevokeAllUserRefreshTokens(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeTokenFamily(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	tokenFamily := "family-123"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `refresh_tokens` SET `is_revoked`=?,`revoked_at`=? WHERE token_family = ? AND is_revoked = ?")).
		WithArgs(true, sqlmock.AnyArg(), tokenFamily, false).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err := repo.RevokeTokenFamily(tokenFamily)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteExpiredRefreshTokens(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `refresh_tokens` WHERE expires_at < ?")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectCommit()

	err := repo.DeleteExpiredRefreshTokens()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddToBlacklist(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	blacklistToken := &domain.TokenBlacklist{
		Token:     "blacklisted-token",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `token_blacklist`")).
		WithArgs(
			blacklistToken.Token,
			blacklistToken.ExpiresAt,
			sqlmock.AnyArg(), // CreatedAt
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.AddToBlacklist(blacklistToken)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsTokenBlacklisted(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	t.Run("Token is blacklisted", func(t *testing.T) {
		token := "blacklisted-token"

		rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `token_blacklist` WHERE token = ? AND expires_at > ?")).
			WithArgs(token, sqlmock.AnyArg()).
			WillReturnRows(rows)

		isBlacklisted, err := repo.IsTokenBlacklisted(token)
		require.NoError(t, err)
		assert.True(t, isBlacklisted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Token is not blacklisted", func(t *testing.T) {
		token := "valid-token"

		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `token_blacklist` WHERE token = ? AND expires_at > ?")).
			WithArgs(token, sqlmock.AnyArg()).
			WillReturnRows(rows)

		isBlacklisted, err := repo.IsTokenBlacklisted(token)
		require.NoError(t, err)
		assert.False(t, isBlacklisted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteExpiredBlacklistTokens(t *testing.T) {
	db, mock := setupTokenMockDB(t)
	repo := repository.NewTokenRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `token_blacklist` WHERE expires_at < ?")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	err := repo.DeleteExpiredBlacklistTokens()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenIsValid(t *testing.T) {
	t.Run("Valid token", func(t *testing.T) {
		token := &domain.RefreshToken{
			IsRevoked: false,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.True(t, token.IsValid())
	})

	t.Run("Revoked token is invalid", func(t *testing.T) {
		token := &domain.RefreshToken{
			IsRevoked: true,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.False(t, token.IsValid())
	})

	t.Run("Expired token is invalid", func(t *testing.T) {
		token := &domain.RefreshToken{
			IsRevoked: false,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		assert.False(t, token.IsValid())
	})

	t.Run("Revoked and expired token is invalid", func(t *testing.T) {
		token := &domain.RefreshToken{
			IsRevoked: true,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		assert.False(t, token.IsValid())
	})
}
