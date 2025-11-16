package middleware

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware checks if the user is an admin
func AdminMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse("user not authenticated", nil))
			return
		}

		user, err := userService.GetUserByID(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, domain.ErrorResponse("user not found", nil))
			return
		}

		if !user.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, domain.ErrorResponse("admin access required", nil))
			return
		}

		c.Next()
	}
}
