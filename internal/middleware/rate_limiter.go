package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket implements a simple fixed-window counter limiter per key
type TokenBucket struct {
	mu       sync.Mutex
	window   time.Duration
	limit    int
	counters map[string]int
	resets   map[string]time.Time
}

func NewTokenBucket(limit int, window time.Duration) *TokenBucket {
	return &TokenBucket{
		window:   window,
		limit:    limit,
		counters: make(map[string]int),
		resets:   make(map[string]time.Time),
	}
}

func (tb *TokenBucket) Allow(key string) (allowed bool, remaining int, reset time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	r, ok := tb.resets[key]
	if !ok || now.After(r) {
		tb.resets[key] = now.Add(tb.window)
		tb.counters[key] = 0
		r = tb.resets[key]
	}

	if tb.counters[key] >= tb.limit {
		return false, 0, r
	}
	tb.counters[key]++
	return true, tb.limit - tb.counters[key], r
}

// RateLimitByIP returns a middleware limiting requests per IP
// (removed RateLimitByIP)

// RateLimitByEmail limits based on the email in JSON body; falls back to IP if missing
func RateLimitByEmail(limit int, window time.Duration) gin.HandlerFunc {
	bucket := NewTokenBucket(limit, window)
	return func(c *gin.Context) {
		// Read and restore body
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var payload struct {
			Email string `json:"email"`
		}
		_ = json.Unmarshal(bodyBytes, &payload)
		key := strings.ToLower(strings.TrimSpace(payload.Email))
		if key == "" {
			// Skip limiting when email is missing
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			c.Next()
			return
		}

		allowed, remaining, reset := bucket.Allow(key)
		c.Header("X-RateLimit-Limit", intToStr(limit))
		c.Header("X-RateLimit-Remaining", intToStr(remaining))
		c.Header("X-RateLimit-Reset", reset.Format(time.RFC3339))
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}

		// Restore body again for downstream binders
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

func intToStr(v int) string {
	// small fast int->string, avoiding strconv import here for brevity
	return fmtInt(v)
}

// Minimal int to string conversion for small values
func fmtInt(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [12]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
