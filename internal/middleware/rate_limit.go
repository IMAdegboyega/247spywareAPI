package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	tokens    float64
	lastSeen  time.Time
	maxTokens float64
	rate      float64 // tokens per second
}

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     float64 // requests per second
	burst    int     // max burst
}

func NewRateLimiter(requestsPerMinute int, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    burst,
	}

	// Clean up stale entries every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-10 * time.Minute)
	for ip, v := range rl.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists {
		rl.visitors[key] = &visitor{
			tokens:    float64(rl.burst) - 1,
			lastSeen:  now,
			maxTokens: float64(rl.burst),
			rate:      rl.rate,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens += elapsed * v.rate
	if v.tokens > v.maxTokens {
		v.tokens = v.maxTokens
	}
	v.lastSeen = now

	if v.tokens >= 1 {
		v.tokens--
		return true
	}

	return false
}

// RateLimit creates a Gin middleware that limits requests per IP.
// requestsPerMinute: sustained rate, burst: max burst size.
func RateLimit(requestsPerMinute int, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerMinute, burst)

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimit is a tighter limiter for sensitive endpoints (login, subscribe).
func StrictRateLimit(requestsPerMinute int, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerMinute, burst)

	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + c.FullPath()

		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many attempts, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
