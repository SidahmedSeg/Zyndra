package git

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ValidateGitHubWebhookSignature validates a GitHub webhook signature
func ValidateGitHubWebhookSignature(secret string, payload []byte, signature string) bool {
	// GitHub sends signature as: sha256=<hash>
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedHash := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	actualHash := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedHash), []byte(actualHash))
}

// ValidateGitLabWebhookSignature validates a GitLab webhook signature
// GitLab uses X-Gitlab-Token header for simple token validation
func ValidateGitLabWebhookSignature(secret, token string) bool {
	return secret == token
}

// ExtractGitHubSignature extracts the signature from the X-Hub-Signature-256 header
func ExtractGitHubSignature(header string) string {
	return header
}

// ExtractGitLabToken extracts the token from the X-Gitlab-Token header
func ExtractGitLabToken(header string) string {
	return header
}

// ParseGitHubEvent parses GitHub webhook event type
func ParseGitHubEvent(eventType string) (string, error) {
	switch eventType {
	case "push", "ping", "pull_request":
		return eventType, nil
	default:
		return "", fmt.Errorf("unsupported event type: %s", eventType)
	}
}

// ParseGitLabEvent parses GitLab webhook event type
func ParseGitLabEvent(eventType string) (string, error) {
	switch eventType {
	case "Push Hook", "Tag Push Hook", "Merge Request Hook":
		return eventType, nil
	default:
		return "", fmt.Errorf("unsupported event type: %s", eventType)
	}
}

