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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	welcomeMessage   = "Welcome to Go JWT REST API"
	apiVersion       = "1.0.0"
	serverStatus     = "running"
	healthEndpoint   = "/health"
	registerEndpoint = "/api/v1/auth/register"
	loginEndpoint    = "/api/v1/auth/login"
	usersEndpoint    = "/api/v1/users (requires auth)"
	documentationURL = "https://github.com/prassaaa/gojwt-rest-api"
)

func main() {
	// Initialize logger
	appLogger := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatal("Failed to load configuration:", err)
	}

	// Set Gin mode
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := config.NewDatabase(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := migrations.Migrate(db); err != nil {
		appLogger.Fatal("Failed to run migrations:", err)
	}
	appLogger.Info("Database migrations completed successfully")

	// Initialize dependencies
		validator, err := validator.New()
	if err != nil {
		appLogger.Fatal("Failed to create validator:", err)
	}
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiration)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, validator)
	userHandler := handler.NewUserHandler(userService, validator)
	profileHandler := handler.NewProfileHandler(userService, validator)

	// Initialize Gin router
	router := gin.Default()

	// Apply global middlewares
	router.Use(middleware.CORSMiddleware(cfg.CORS))
	router.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Welcome endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": welcomeMessage,
			"version": apiVersion,
			"status":  serverStatus,
			"endpoints": gin.H{
				"health":   healthEndpoint,
				"register": registerEndpoint,
				"login":    loginEndpoint,
				"users":    usersEndpoint,
			},
			"documentation": documentationURL,
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

		// Profile routes (protected - user self-service)
		profile := v1.Group("/profile")
		profile.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			profile.GET("", profileHandler.GetOwnProfile)
			profile.PUT("", profileHandler.UpdateOwnProfile)
			profile.PUT("/password", profileHandler.ChangePassword)
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			users.GET("/profile", userHandler.GetProfile)
			// Admin-only routes
			admin := users.Group("")
			admin.Use(middleware.AdminMiddleware(userService))
			{
				admin.GET("", userHandler.GetAllUsers)
				admin.GET("/:id", userHandler.GetUserByID)
				admin.PUT("/:id", userHandler.UpdateUser)
				admin.DELETE("/:id", userHandler.DeleteUser)
			}
		}
	}

	// Create server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Infof("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("Failed to start server: %v", err)
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
		appLogger.Fatal("Server forced to shutdown:", err)
	}

	// Close database connection
	if err := config.CloseDatabase(db); err != nil {
		appLogger.Error("Error closing database:", err)
	}

	appLogger.Info("Server stopped gracefully")
}
