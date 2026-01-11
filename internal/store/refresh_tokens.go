package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"` // Never expose
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

// hashToken creates a SHA-256 hash of a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateRefreshToken creates a new refresh token
func (db *DB) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (*RefreshToken, error) {
	tokenHash := hashToken(token)
	
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, revoked, created_at
	`

	var rt RefreshToken
	err := db.QueryRowContext(ctx, query, userID, tokenHash, expiresAt).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.ExpiresAt,
		&rt.Revoked,
		&rt.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &rt, nil
}

// GetRefreshToken retrieves a refresh token by its value (hashes and looks up)
func (db *DB) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	tokenHash := hashToken(token)
	
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var rt RefreshToken
	err := db.QueryRowContext(ctx, query, tokenHash).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.ExpiresAt,
		&rt.Revoked,
		&rt.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &rt, nil
}

// ValidateRefreshToken checks if a refresh token is valid (exists, not expired, not revoked)
func (db *DB) ValidateRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	rt, err := db.GetRefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if rt.Revoked {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	return rt, nil
}

// RevokeRefreshToken revokes a refresh token
func (db *DB) RevokeRefreshToken(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	
	query := `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`
	result, err := db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func (db *DB) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user refresh tokens: %w", err)
	}
	return nil
}

// DeleteExpiredRefreshTokens deletes expired refresh tokens (for cleanup)
func (db *DB) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < now()`
	result, err := db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// RotateRefreshToken creates a new refresh token and revokes the old one
func (db *DB) RotateRefreshToken(ctx context.Context, oldToken, newToken string, userID string, expiresAt time.Time) (*RefreshToken, error) {
	// Revoke old token
	err := db.RevokeRefreshToken(ctx, oldToken)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Create new token
	return db.CreateRefreshToken(ctx, userID, newToken, expiresAt)
}

