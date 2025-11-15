package migrations

import (
	"gojwt-rest-api/internal/domain"

	"gorm.io/gorm"
)

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
	)
}
