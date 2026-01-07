package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/intelifox/click-deploy/internal/auth"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per duration
	duration time.Duration // time window
	cleanup  *time.Ticker
}

type visitor struct {
	lastSeen time.Time
	tokens   int
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed
// duration: time window (e.g., time.Minute for per-minute rate)
func NewRateLimiter(rate int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		duration: duration,
		cleanup:  time.NewTicker(5 * time.Minute), // Clean up old visitors every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors removes old visitor entries
func (rl *RateLimiter) cleanupVisitors() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if now.Sub(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:   rl.rate,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}
	rl.mu.Unlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	// Refill tokens if enough time has passed
	if now.Sub(v.lastSeen) >= rl.duration {
		v.tokens = rl.rate
		v.lastSeen = now
	}

	if v.tokens > 0 {
		v.tokens--
		v.lastSeen = now
		return true
	}

	return false
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rate int, duration time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(rate, duration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", duration.String())
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	// Check X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	return ip
}

// PerUserRateLimitMiddleware creates a rate limiter based on user ID (from JWT)
// This is more accurate than IP-based limiting for authenticated users
func PerUserRateLimitMiddleware(rate int, duration time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(rate, duration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get user ID from context (set by auth middleware)
			userID := getUserIDFromContext(r)
			identifier := userID

			// Fall back to IP if no user ID
			if identifier == "" {
				identifier = getClientIP(r)
			}

			if !limiter.Allow(identifier) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", duration.String())
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(r *http.Request) string {
	// Use the auth package helper function
	return auth.GetUserID(r.Context())
}

