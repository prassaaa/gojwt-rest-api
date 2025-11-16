package e2e

import (
	"fmt"
	"gojwt-rest-api/internal/handler"
	"gojwt-rest-api/internal/middleware"
	"gojwt-rest-api/internal/repository"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/migrations"
	"gojwt-rest-api/pkg/validator"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestServer(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"root",
		"",
		"localhost",
		"3306",
		"gojwt_db_test",
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping e2e test: database connection failed: %v", err)
		return nil, nil
	}

	// Run migrations
	migrations.Migrate(db)

	// Clean up tables before each test
	db.Exec("DELETE FROM token_blacklist")
	db.Exec("DELETE FROM refresh_tokens")
	db.Exec("DELETE FROM users")

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Setup services
	jwtSecret := "test-jwt-secret-key-for-testing"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour
	userService := service.NewUserService(userRepo, tokenRepo, jwtSecret, accessExpiry, refreshExpiry)

	// Setup validator
	v, _ := validator.New()

	// Setup handlers
	authHandler := handler.NewAuthHandler(userService, v)
	profileHandler := handler.NewProfileHandler(userService, v)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Auth routes
	router.POST("/api/v1/auth/register", authHandler.Register)
	router.POST("/api/v1/auth/login", authHandler.Login)
	router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		protected.POST("/auth/logout", authHandler.Logout)
		protected.GET("/profile", profileHandler.GetOwnProfile)
		protected.PUT("/profile", profileHandler.UpdateOwnProfile)
		protected.PUT("/profile/password", profileHandler.ChangePassword)
	}

	return router, db
}
