# Security Hardening Implementation

## Overview

This document describes the security hardening measures implemented in Phase 7 of the Click-to-Deploy project.

## Implemented Features

### 1. Rate Limiting

**Location:** `internal/api/ratelimit.go`

**Features:**
- Token bucket rate limiting algorithm
- Per-user rate limiting (for authenticated requests)
- Per-IP rate limiting (for unauthenticated requests)
- Configurable rate limits via environment variables
- Automatic cleanup of old visitor entries

**Configuration:**
```bash
RATE_LIMIT_REQUESTS=100  # Number of requests allowed per window
RATE_LIMIT_WINDOW=60     # Time window in seconds
```

**Usage:**
- API routes: 100 requests per minute per authenticated user
- Health check: 10 requests per minute per IP
- Metrics endpoint: 60 requests per minute per IP

**Implementation Details:**
- Uses in-memory storage (can be upgraded to Redis for distributed systems)
- Thread-safe with mutex locks
- Automatic token refill based on time window
- Returns `429 Too Many Requests` with `Retry-After` header when limit exceeded

### 2. Security Headers

**Location:** `internal/api/security_headers.go`

**Headers Added:**
- `X-Frame-Options: DENY` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME type sniffing
- `X-XSS-Protection: 1; mode=block` - Enables XSS protection
- `Content-Security-Policy` - Restricts resource loading
- `Referrer-Policy: strict-origin-when-cross-origin` - Controls referrer information
- `Permissions-Policy` - Restricts browser features
- `Strict-Transport-Security` - Enforces HTTPS (when TLS is enabled)

**CSP Configuration:**
- Allows same-origin resources
- Allows WebSocket connections (for Centrifugo)
- Allows HTTPS connections
- Blocks frame embedding
- Allows inline styles/scripts (required for Next.js)

### 3. Input Sanitization

**Location:** `internal/api/sanitize.go`

**Sanitization Functions:**
- `SanitizeString()` - Removes null bytes, trims whitespace, unescapes HTML
- `SanitizeURL()` - Validates and sanitizes URLs (only http/https)
- `SanitizeHostname()` - Removes protocol, path, and port from hostnames
- `SanitizeDomain()` - Sanitizes domain names
- `SanitizeFilename()` - Prevents path traversal attacks
- `SanitizeSQLIdentifier()` - Sanitizes SQL identifiers (extra safety)
- `SanitizeEnvironmentVariableKey()` - Ensures valid env var key format
- `SanitizeGitBranch()` - Sanitizes git branch names
- `SanitizeCommitSHA()` - Validates commit SHA format

**Integration:**
- Integrated into `ValidateString()` function
- Applied to all user inputs in API handlers:
  - Project names
  - Service names
  - Custom domains
  - Environment variable keys

### 4. Validation Integration

**Location:** `internal/api/validation.go`

**Enhancements:**
- `ValidateString()` now automatically sanitizes input before validation
- All string validations include sanitization step

## Middleware Stack

The middleware is applied in the following order:

1. **Logger** - Request logging
2. **Recoverer** - Panic recovery
3. **RequestID** - Request ID generation
4. **SecurityHeaders** - Security headers (all routes)
5. **Compression** - Response compression
6. **Authentication** - JWT validation (API routes)
7. **Rate Limiting** - Per-user rate limiting (API routes)

## Configuration

### Environment Variables

```bash
# Rate Limiting
RATE_LIMIT_REQUESTS=100  # Default: 100 requests
RATE_LIMIT_WINDOW=60     # Default: 60 seconds
```

### Default Limits

- **API Routes (Authenticated):** 100 requests per minute per user
- **Health Check:** 10 requests per minute per IP
- **Metrics Endpoint:** 60 requests per minute per IP

## Security Best Practices

### 1. Defense in Depth
- Multiple layers of security (rate limiting, input sanitization, validation)
- Security headers provide additional browser-level protection

### 2. Input Validation
- All user inputs are sanitized before processing
- Validation occurs after sanitization
- SQL injection protection via prepared statements (existing)

### 3. Rate Limiting
- Prevents brute force attacks
- Protects against DDoS
- Configurable per endpoint type

### 4. Security Headers
- Prevents common web vulnerabilities
- Enforces secure communication
- Restricts resource loading

## Future Enhancements

### 1. Distributed Rate Limiting
- Migrate to Redis-based rate limiting for multi-instance deployments
- Shared rate limit state across instances

### 2. Advanced Input Validation
- Add regex-based validation for specific fields
- Implement custom validators for domain-specific rules

### 3. Token Encryption
- Encrypt OAuth tokens at rest (currently stored in plaintext)
- Use AES-256-GCM encryption

### 4. Password Encryption
- Encrypt database passwords before storage
- Use secure key management

### 5. CSRF Protection
- Add CSRF tokens for state-changing operations
- Validate CSRF tokens in OAuth callbacks

### 6. Request Size Limits
- Add maximum request body size limits
- Prevent large payload attacks

## Testing

### Rate Limiting Test
```bash
# Test rate limiting (should fail after 100 requests)
for i in {1..110}; do
  curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/v1/click-deploy/projects
done
```

### Security Headers Test
```bash
# Check security headers
curl -I http://localhost:8080/health
```

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
- [Rate Limiting Best Practices](https://cloud.google.com/architecture/rate-limiting-strategies-techniques)

