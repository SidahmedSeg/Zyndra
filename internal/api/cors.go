package api

import (
	"net/http"
	"strings"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin || allowedOrigin == "*" {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddlewareFromEnv creates CORS middleware from environment variable
// Expects CORS_ORIGINS as comma-separated list of allowed origins
func CORSMiddlewareFromEnv(originsEnv string) func(http.Handler) http.Handler {
	origins := []string{"*"} // Default: allow all (for development)
	
	if originsEnv != "" {
		origins = strings.Split(originsEnv, ",")
		// Trim whitespace
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
	}

	return CORSMiddleware(origins)
}

