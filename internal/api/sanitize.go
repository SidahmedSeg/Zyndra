package api

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
)

// SanitizeString sanitizes a string input
func SanitizeString(s string) string {
	// Trim whitespace
	s = strings.TrimSpace(s)

	// Unescape HTML entities (in case of double encoding)
	s = html.UnescapeString(s)

	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	return s
}

// SanitizeURL sanitizes and validates a URL
func SanitizeURL(u string) (string, error) {
	u = strings.TrimSpace(u)

	// Parse URL to validate
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("only http and https schemes are allowed")
	}

	return parsed.String(), nil
}

// SanitizeHostname sanitizes a hostname
func SanitizeHostname(hostname string) string {
	hostname = strings.TrimSpace(hostname)
	hostname = strings.ToLower(hostname)

	// Remove any protocol prefix
	hostname = strings.TrimPrefix(hostname, "http://")
	hostname = strings.TrimPrefix(hostname, "https://")

	// Remove trailing slash
	hostname = strings.TrimSuffix(hostname, "/")

	// Remove path if present
	if idx := strings.Index(hostname, "/"); idx != -1 {
		hostname = hostname[:idx]
	}

	return hostname
}

// SanitizeDomain sanitizes a domain name
func SanitizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.ToLower(domain)

	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Remove trailing slash
	domain = strings.TrimSuffix(domain, "/")

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// SanitizeFilename sanitizes a filename to prevent path traversal
func SanitizeFilename(filename string) string {
	filename = strings.TrimSpace(filename)

	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	return filename
}

// SanitizeSQLIdentifier sanitizes a SQL identifier (table/column name)
// This is for extra safety, but prepared statements should already prevent SQL injection
func SanitizeSQLIdentifier(identifier string) string {
	// Only allow alphanumeric, underscore, and hyphen
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return re.ReplaceAllString(identifier, "")
}

// SanitizeEnvironmentVariableKey sanitizes an environment variable key
func SanitizeEnvironmentVariableKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.ToUpper(key)

	// Only allow alphanumeric and underscore
	re := regexp.MustCompile(`[^A-Z0-9_]`)
	key = re.ReplaceAllString(key, "")

	return key
}

// SanitizeGitBranch sanitizes a git branch name
func SanitizeGitBranch(branch string) string {
	branch = strings.TrimSpace(branch)

	// Remove dangerous characters
	re := regexp.MustCompile(`[^a-zA-Z0-9/._-]`)
	branch = re.ReplaceAllString(branch, "")

	// Remove leading/trailing dots and slashes
	branch = strings.Trim(branch, "./")

	return branch
}

// SanitizeCommitSHA sanitizes a git commit SHA
func SanitizeCommitSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	sha = strings.ToLower(sha)

	// Only allow hexadecimal characters
	re := regexp.MustCompile(`[^a-f0-9]`)
	sha = re.ReplaceAllString(sha, "")

	// Limit length (SHA-1 is 40 chars, SHA-256 is 64 chars)
	if len(sha) > 64 {
		sha = sha[:64]
	}

	return sha
}

// Note: ValidationError is defined in validation.go

