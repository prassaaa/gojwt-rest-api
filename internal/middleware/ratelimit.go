package middleware

import (
	"gojwt-rest-api/internal/config"
	"gojwt-rest-api/internal/domain"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a simple in-memory rate limiter.
// NOTE: This implementation is not suitable for a distributed environment
// with multiple server instances. For production, consider using a
// distributed rate limiter with a shared data store like Redis.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	duration time.Duration
}

// visitor represents a client visitor
type visitor struct {
	count      int
	lastAccess time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg config.RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     cfg.RequestsPerDuration,
		duration: cfg.Duration,
	}

	// Start cleanup goroutine
	go rl.cleanup(cfg.CleanupInterval)

	return rl
}

// cleanup removes old visitors periodically
func (rl *RateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastAccess) > rl.duration {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// allow checks if the request is allowed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			count:      1,
			lastAccess: now,
		}
		return true
	}

	// Reset count if duration has passed
	if now.Sub(v.lastAccess) > rl.duration {
		v.count = 1
		v.lastAccess = now
		return true
	}

	// Check if rate limit exceeded
	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastAccess = now
	return true
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, domain.ErrorResponse(domain.ErrRateLimitExceeded.Error(), nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
