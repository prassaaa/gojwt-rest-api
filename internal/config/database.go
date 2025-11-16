package config

import (
	"fmt"
	"time"

	"gojwt-rest-api/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection
func NewDatabase(cfg *Config, appLogger *logger.Logger) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	var gormLogger gormlogger.Interface
	if cfg.AppEnv == "production" {
		gormLogger = gormlogger.Default.LogMode(gormlogger.Silent)
	} else {
		gormLogger = gormlogger.Default.LogMode(gormlogger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	appLogger.Info("Database connection established successfully")

	return db, nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
