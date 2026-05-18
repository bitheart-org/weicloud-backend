package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws: wss:; img-src 'self' data:; style-src 'self' 'unsafe-inline';")
		c.Next()
	}
}

type rateWindow struct {
	Count      int
	WindowEnds time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateWindow
	limit   int
	window  time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		entries: make(map[string]rateWindow),
		limit:   limit,
		window:  window,
	}
}

func (r *rateLimiter) allow(key string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.entries[key]
	if !ok || now.After(entry.WindowEnds) {
		r.entries[key] = rateWindow{
			Count:      1,
			WindowEnds: now.Add(r.window),
		}
		return true
	}

	if entry.Count >= r.limit {
		return false
	}

	entry.Count++
	r.entries[key] = entry
	return true
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.allow(key, time.Now()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    42901,
				"message": "rate limit exceeded",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
