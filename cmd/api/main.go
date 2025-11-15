package main

import (
	"context"
	"fmt"
	"gojwt-rest-api/internal/config"
	"gojwt-rest-api/internal/handler"
	"gojwt-rest-api/internal/middleware"
	"gojwt-rest-api/internal/repository"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/migrations"
	"gojwt-rest-api/pkg/logger"
	"gojwt-rest-api/pkg/validator"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	appLogger := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Set Gin mode
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := migrations.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	appLogger.Info("Database migrations completed successfully")

	// Initialize dependencies
	validator := validator.New()
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerDuration, cfg.RateLimit.Duration)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiration)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, validator)
	userHandler := handler.NewUserHandler(userService, validator)

	// Initialize Gin router
	router := gin.Default()

	// Apply global middlewares
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Welcome endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Go JWT REST API",
			"version": "1.0.0",
			"status":  "running",
			"endpoints": gin.H{
				"health":   "/health",
				"register": "/api/v1/auth/register",
				"login":    "/api/v1/auth/login",
				"users":    "/api/v1/users (requires auth)",
			},
			"documentation": "https://github.com/prassaaa/gojwt-rest-api",
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// Create server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Infof("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close database connection
	if err := config.CloseDatabase(db); err != nil {
		appLogger.Error("Error closing database:", err)
	}

	appLogger.Info("Server stopped gracefully")
}
