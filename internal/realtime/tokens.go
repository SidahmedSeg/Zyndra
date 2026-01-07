package realtime

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateConnectionToken(hmacSecret string, userID string, ttl time.Duration) (string, error) {
	if hmacSecret == "" {
		return "", fmt.Errorf("missing centrifugo token hmac secret")
	}
	if userID == "" {
		return "", fmt.Errorf("missing userID")
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(hmacSecret))
}

func GenerateSubscriptionToken(hmacSecret string, userID string, channel string, ttl time.Duration) (string, error) {
	if hmacSecret == "" {
		return "", fmt.Errorf("missing centrifugo token hmac secret")
	}
	if userID == "" {
		return "", fmt.Errorf("missing userID")
	}
	if channel == "" {
		return "", fmt.Errorf("missing channel")
	}

	claims := jwt.MapClaims{
		"sub":     userID,
		"channel": channel,
		"exp":     time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(hmacSecret))
}


