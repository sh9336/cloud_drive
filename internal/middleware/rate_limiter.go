// internal/middleware/rate_limiter.go
package middleware

import (
	"net/http"
	"sync"

	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// RateLimit middleware by IP address
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := utils.GetClientIP(c)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			utils.ErrorResponse(c, http.StatusTooManyRequests, utils.ErrBadRequest)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser middleware by user ID
func (rl *RateLimiter) RateLimitByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetUserID(c)
		if err != nil {
			c.Next()
			return
		}

		limiter := rl.getLimiter(userID.String())
		if !limiter.Allow() {
			utils.ErrorResponse(c, http.StatusTooManyRequests, utils.ErrBadRequest)
			c.Abort()
			return
		}

		c.Next()
	}
}
