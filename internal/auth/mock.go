package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MockValidator is a simple validator for development/testing
type MockValidator struct{}

// NewMockValidator creates a new mock validator
func NewMockValidator() *MockValidator {
	return &MockValidator{}
}

// ValidateToken validates a mock token (just checks if it's a valid JWT structure)
func (v *MockValidator) ValidateToken(tokenString string) (*CasdoorClaims, error) {
	// Parse token without verification (for mock)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CasdoorClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CasdoorClaims); ok {
		// Ensure required fields are present
		if claims.Owner == "" {
			claims.Owner = "mock-org-1"
		}
		if claims.Sub == "" {
			claims.Sub = "mock-user-1"
		}
		if claims.Name == "" {
			claims.Name = "Mock User"
		}
		if claims.Roles == nil {
			claims.Roles = []string{"admin", "user"}
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// GenerateMockToken generates a mock JWT token for testing
func GenerateMockToken(userID, orgID, name string, roles []string) (string, error) {
	if userID == "" {
		userID = "mock-user-1"
	}
	if orgID == "" {
		orgID = "mock-org-1"
	}
	if name == "" {
		name = "Mock User"
	}
	if roles == nil || len(roles) == 0 {
		roles = []string{"admin", "user"}
	}

	claims := &CasdoorClaims{
		Sub:   userID,
		Name:  name,
		Owner: orgID,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mock-auth",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Use a simple secret for mock tokens
	secret := []byte("mock-secret-key-for-development-only")
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateMockTokenJSON generates a mock token and returns it as JSON
func GenerateMockTokenJSON(userID, orgID, name string, roles []string) ([]byte, error) {
	token, err := GenerateMockToken(userID, orgID, name, roles)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400, // 24 hours
	}

	return json.Marshal(response)
}

