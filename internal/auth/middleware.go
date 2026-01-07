package auth

import (
	"context"
	"net/http"
	"strings"
)

// Middleware creates an authentication middleware
func Middleware(validator *Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			// For development, use simple validation
			// In production, use ValidateToken with proper JWKS
			claims, err := validator.ValidateTokenSimple(tokenString)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Add user context to request
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.Sub)
			ctx = context.WithValue(ctx, OrgIDKey, claims.Owner)
			ctx = context.WithValue(ctx, RolesKey, claims.Roles)
			ctx = context.WithValue(ctx, NameKey, claims.Name)

			// Continue with authenticated request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalMiddleware creates a middleware that allows optional authentication
// If token is present, it validates and adds context
// If token is missing, request continues without user context
func OptionalMiddleware(validator *Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No auth header, continue without user context
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Invalid format, continue without user context
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]
			claims, err := validator.ValidateTokenSimple(tokenString)
			if err != nil {
				// Invalid token, continue without user context
				next.ServeHTTP(w, r)
				return
			}

			// Add user context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.Sub)
			ctx = context.WithValue(ctx, OrgIDKey, claims.Owner)
			ctx = context.WithValue(ctx, RolesKey, claims.Roles)
			ctx = context.WithValue(ctx, NameKey, claims.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetOrgID extracts organization ID from context
func GetOrgID(ctx context.Context) string {
	if orgID, ok := ctx.Value(OrgIDKey).(string); ok {
		return orgID
	}
	return ""
}

// GetRoles extracts user roles from context
func GetRoles(ctx context.Context) []string {
	if roles, ok := ctx.Value(RolesKey).([]string); ok {
		return roles
	}
	return []string{}
}

// GetUserName extracts username from context
func GetUserName(ctx context.Context) string {
	if name, ok := ctx.Value(NameKey).(string); ok {
		return name
	}
	return ""
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := GetRoles(r.Context())
			
			hasRole := false
			for _, r := range roles {
				if r == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

