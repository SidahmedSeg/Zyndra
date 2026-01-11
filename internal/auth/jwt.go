package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string
	AccessExpiry    time.Duration
	RefreshExpiry   time.Duration
	Issuer          string
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig(secret string) JWTConfig {
	return JWTConfig{
		Secret:        secret,
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour, // 7 days
		Issuer:        "zyndra",
	}
}

// ZyndraClaims represents the JWT claims for Zyndra's own auth
type ZyndraClaims struct {
	UserID  string   `json:"user_id"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	OrgID   string   `json:"org_id"`
	OrgSlug string   `json:"org_slug"`
	Role    string   `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	config JWTConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{config: config}
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// GenerateTokenPair generates access and refresh tokens for a user
func (s *JWTService) GenerateTokenPair(userID, email, name, orgID, orgSlug, role string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.config.AccessExpiry)
	
	// Access token claims
	accessClaims := &ZyndraClaims{
		UserID:  userID,
		Email:   email,
		Name:    name,
		OrgID:   orgID,
		OrgSlug: orgSlug,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   userID,
		},
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token (random string)
	refreshTokenString, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*ZyndraClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ZyndraClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*ZyndraClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshExpiry returns the refresh token expiry duration
func (s *JWTService) RefreshExpiry() time.Duration {
	return s.config.RefreshExpiry
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// JWTValidator implements ValidatorInterface for custom JWT auth
type JWTValidator struct {
	jwtService *JWTService
}

// NewJWTValidator creates a new JWT validator that implements ValidatorInterface
func NewJWTValidator(jwtService *JWTService) *JWTValidator {
	return &JWTValidator{jwtService: jwtService}
}

// ValidateToken validates a token and returns CasdoorClaims for backwards compatibility
func (v *JWTValidator) ValidateToken(tokenString string) (*CasdoorClaims, error) {
	claims, err := v.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Convert ZyndraClaims to CasdoorClaims for backwards compatibility
	return &CasdoorClaims{
		Sub:   claims.UserID,
		Name:  claims.Name,
		Owner: claims.OrgID, // OrgID maps to Owner for backwards compatibility
		Roles: []string{claims.Role},
		RegisteredClaims: claims.RegisteredClaims,
	}, nil
}

