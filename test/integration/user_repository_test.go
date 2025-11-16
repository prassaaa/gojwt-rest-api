package integration

import (
	"database/sql"
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

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("Successfully create user", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		user := &domain.User{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "hashedpassword",
			IsAdmin:  false,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
			WithArgs(
				sqlmock.AnyArg(), // name
				sqlmock.AnyArg(), // email
				sqlmock.AnyArg(), // password
				sqlmock.AnyArg(), // is_admin
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(user)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Create with database error", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		user := &domain.User{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "hashedpassword",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Create(user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	t.Run("Successfully find user by ID", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "is_admin", "created_at", "updated_at"}).
			AddRow(1, "John Doe", "john@example.com", "hashedpassword", false, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? ORDER BY `users`.`id` LIMIT ?")).
			WithArgs(1, 1).
			WillReturnRows(rows)

		user, err := repo.FindByID(1)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User not found", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? ORDER BY `users`.`id` LIMIT ?")).
			WithArgs(999, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByID(999)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	t.Run("Successfully find user by email", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "is_admin", "created_at", "updated_at"}).
			AddRow(1, "John Doe", "john@example.com", "hashedpassword", false, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
			WithArgs("john@example.com", 1).
			WillReturnRows(rows)

		user, err := repo.FindByEmail("john@example.com")

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "john@example.com", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User not found by email", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
			WithArgs("nonexistent@example.com", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByEmail("nonexistent@example.com")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_FindAll(t *testing.T) {
	t.Run("Successfully find all users with pagination", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		pagination := &domain.PaginationQuery{
			Page:     1,
			PageSize: 10,
			Search:   "",
		}

		now := time.Now()

		// Mock count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users`")).
			WillReturnRows(countRows)

		// Mock find query
		userRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "is_admin", "created_at", "updated_at"}).
			AddRow(1, "User 1", "user1@example.com", "hash1", false, now, now).
			AddRow(2, "User 2", "user2@example.com", "hash2", false, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` LIMIT ?")).
			WithArgs(10).
			WillReturnRows(userRows)

		users, total, err := repo.FindAll(pagination)

		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, int64(2), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Find all with search filter", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		pagination := &domain.PaginationQuery{
			Page:     1,
			PageSize: 10,
			Search:   "john",
		}

		now := time.Now()

		// Mock count query with search
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE name LIKE ? OR email LIKE ?")).
			WithArgs("%john%", "%john%").
			WillReturnRows(countRows)

		// Mock find query with search
		userRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "is_admin", "created_at", "updated_at"}).
			AddRow(1, "John Doe", "john@example.com", "hash1", false, now, now)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE name LIKE ? OR email LIKE ? LIMIT ?")).
			WithArgs("%john%", "%john%", 10).
			WillReturnRows(userRows)

		users, total, err := repo.FindAll(pagination)

		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "John Doe", users[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Find all with offset pagination", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		pagination := &domain.PaginationQuery{
			Page:     2,
			PageSize: 5,
			Search:   "",
		}

		// Mock count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users`")).
			WillReturnRows(countRows)

		// Mock find query with offset
		userRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "is_admin", "created_at", "updated_at"})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` LIMIT ? OFFSET ?")).
			WithArgs(5, 5).
			WillReturnRows(userRows)

		_, total, err := repo.FindAll(pagination)

		require.NoError(t, err)
		assert.Equal(t, int64(10), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("Successfully update user", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		user := &domain.User{
			ID:       1,
			Name:     "John Updated",
			Email:    "johnupdated@example.com",
			Password: "hashedpassword",
			IsAdmin:  false,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `users`")).
			WithArgs(
				sqlmock.AnyArg(), // name
				sqlmock.AnyArg(), // email
				sqlmock.AnyArg(), // password
				sqlmock.AnyArg(), // is_admin
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // id
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(user)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Update with database error", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		user := &domain.User{
			ID:    1,
			Name:  "John Updated",
			Email: "johnupdated@example.com",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `users`")).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Update(user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Delete(t *testing.T) {
	t.Run("Successfully delete user", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `users` WHERE `users`.`id` = ?")).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Delete(1)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Delete non-existent user", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `users` WHERE `users`.`id` = ?")).
			WithArgs(999).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.Delete(999)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Delete with database error", func(t *testing.T) {
		db, mock, cleanup := setupMockDB(t)
		defer cleanup()

		repo := repository.NewUserRepository(db)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `users` WHERE `users`.`id` = ?")).
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Delete(1)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
