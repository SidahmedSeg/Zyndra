package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor
	DefaultCost = 12
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return "", fmt.Errorf("password must be at most 72 characters")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if a password meets the requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 72 {
		return fmt.Errorf("password must be at most 72 characters")
	}
	
	// Could add more requirements here:
	// - At least one uppercase letter
	// - At least one lowercase letter
	// - At least one number
	// - At least one special character
	
	return nil
}

