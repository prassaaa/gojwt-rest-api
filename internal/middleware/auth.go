package middleware

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contextUserIDKey   = "user_id"
	contextUserEmailKey = "user_email"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrAuthHeaderRequired.Error(), nil))
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrInvalidAuthHeaderFormat.Error(), nil))
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrInvalidOrExpiredToken.Error(), err))
			c.Abort()
			return
		}

		// Set user information in context
		c.Set(contextUserIDKey, claims.UserID)
		c.Set(contextUserEmailKey, claims.Email)

		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(contextUserIDKey)
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// GetUserEmail retrieves user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(contextUserEmailKey)
	if !exists {
		return "", false
	}
	return email.(string), true
}
