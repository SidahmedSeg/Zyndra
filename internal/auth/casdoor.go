package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// CasdoorClaims represents the JWT claims from Casdoor
type CasdoorClaims struct {
	Sub   string   `json:"sub"`   // User ID
	Name  string   `json:"name"`   // Username
	Owner string   `json:"owner"`  // Organization ID
	Roles []string `json:"roles"`  // User roles in org
	jwt.RegisteredClaims
}

// ContextKey is a type for context keys
type ContextKey string

const (
	UserIDKey ContextKey = "user_id"
	OrgIDKey  ContextKey = "org_id"
	RolesKey  ContextKey = "roles"
	NameKey   ContextKey = "name"
)

// Validator validates JWT tokens from Casdoor
type Validator struct {
	casdoorEndpoint string
	clientID        string
	jwksURL         string
}

// NewValidator creates a new JWT validator
func NewValidator(casdoorEndpoint, clientID string) *Validator {
	// Construct JWKS URL from Casdoor endpoint
	jwksURL := fmt.Sprintf("%s/.well-known/jwks", casdoorEndpoint)
	
	return &Validator{
		casdoorEndpoint: casdoorEndpoint,
		clientID:        clientID,
		jwksURL:         jwksURL,
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *Validator) ValidateToken(tokenString string) (*CasdoorClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &CasdoorClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		// For now, we'll use a simple approach
		// In production, fetch public key from JWKS endpoint
		// For development, we can skip verification or use a test key
		return []byte(""), nil // Placeholder - will be replaced with actual key fetching
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	// Extract claims
	if claims, ok := token.Claims.(*CasdoorClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token claims")
}

// ValidateTokenSimple is a simplified version for development
// It validates the token structure without full cryptographic verification
func (v *Validator) ValidateTokenSimple(tokenString string) (*CasdoorClaims, error) {
	// For development, we'll parse without verification
	// In production, use ValidateToken with proper JWKS
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CasdoorClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	if claims, ok := token.Claims.(*CasdoorClaims); ok {
		// Basic validation
		if claims.Owner == "" {
			return nil, fmt.Errorf("missing organization ID in token")
		}
		if claims.Sub == "" {
			return nil, fmt.Errorf("missing user ID in token")
		}
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token claims")
}

// FetchPublicKey fetches the public key from Casdoor JWKS endpoint
func (v *Validator) FetchPublicKey(kid string) (interface{}, error) {
	// TODO: Implement JWKS fetching
	// This will fetch the public key from Casdoor's JWKS endpoint
	// For now, return error to indicate it needs implementation
	return nil, fmt.Errorf("JWKS fetching not yet implemented")
}

