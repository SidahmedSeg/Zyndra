package auth

// ValidatorInterface defines the interface for token validators
type ValidatorInterface interface {
	ValidateToken(tokenString string) (*CasdoorClaims, error)
}

