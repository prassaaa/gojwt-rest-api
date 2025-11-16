package middleware

import (
	"gojwt-rest-api/internal/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowCredentials = "Access-Control-Allow-Credentials"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	allowMethods           = "POST, OPTIONS, GET, PUT, DELETE, PATCH"
	allowHeaders           = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
)

// CORSMiddleware handles CORS
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set(headerAllowOrigin, cfg.AllowedOrigins)
		c.Writer.Header().Set(headerAllowCredentials, "true")
		c.Writer.Header().Set(headerAllowHeaders, allowHeaders)
		c.Writer.Header().Set(headerAllowMethods, allowMethods)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
