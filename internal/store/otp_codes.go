package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OTPCode represents an OTP code for email verification
type OTPCode struct {
	ID          uuid.UUID
	Email       string
	Code        string
	Purpose     string // "registration", "password_reset", "login"
	Attempts    int
	MaxAttempts int
	ExpiresAt   time.Time
	VerifiedAt  sql.NullTime
	CreatedAt   time.Time
}

// OTPPurpose defines the purpose of an OTP
type OTPPurpose string

const (
	OTPPurposeRegistration  OTPPurpose = "registration"
	OTPPurposePasswordReset OTPPurpose = "password_reset"
	OTPPurposeLogin         OTPPurpose = "login"
)

// GenerateOTPCode generates a random 6-digit OTP code
func GenerateOTPCode() (string, error) {
	// Generate 6 random digits
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Convert to 6 digits (000000-999999)
	code := fmt.Sprintf("%06d", (int(b[0])<<16|int(b[1])<<8|int(b[2]))%1000000)
	return code, nil
}

// CreateOTPCode creates a new OTP code for an email
func (db *DB) CreateOTPCode(ctx context.Context, email string, purpose OTPPurpose, expiresIn time.Duration) (*OTPCode, error) {
	// Delete any existing OTPs for this email/purpose first
	_, err := db.ExecContext(ctx, 
		"DELETE FROM otp_codes WHERE email = $1 AND purpose = $2",
		email, string(purpose))
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup old OTPs: %w", err)
	}

	code, err := GenerateOTPCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	otp := &OTPCode{
		ID:          uuid.New(),
		Email:       email,
		Code:        code,
		Purpose:     string(purpose),
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(expiresIn),
		CreatedAt:   time.Now(),
	}

	query := `
		INSERT INTO otp_codes (id, email, code, purpose, attempts, max_attempts, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = db.ExecContext(ctx, query,
		otp.ID, otp.Email, otp.Code, otp.Purpose, otp.Attempts, otp.MaxAttempts, otp.ExpiresAt, otp.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	return otp, nil
}

// VerifyOTPCode verifies an OTP code and marks it as used
func (db *DB) VerifyOTPCode(ctx context.Context, email, code string, purpose OTPPurpose) (bool, error) {
	var otp OTPCode
	query := `
		SELECT id, email, code, purpose, attempts, max_attempts, expires_at, verified_at, created_at
		FROM otp_codes
		WHERE email = $1 AND purpose = $2 AND verified_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := db.QueryRowContext(ctx, query, email, string(purpose)).Scan(
		&otp.ID, &otp.Email, &otp.Code, &otp.Purpose,
		&otp.Attempts, &otp.MaxAttempts, &otp.ExpiresAt,
		&otp.VerifiedAt, &otp.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("no OTP found for this email")
	}
	if err != nil {
		return false, fmt.Errorf("failed to get OTP: %w", err)
	}

	// Check if expired
	if time.Now().After(otp.ExpiresAt) {
		return false, fmt.Errorf("OTP has expired")
	}

	// Check max attempts
	if otp.Attempts >= otp.MaxAttempts {
		return false, fmt.Errorf("maximum attempts exceeded")
	}

	// Increment attempts
	_, err = db.ExecContext(ctx, 
		"UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1",
		otp.ID)
	if err != nil {
		return false, fmt.Errorf("failed to update attempts: %w", err)
	}

	// Verify code
	if otp.Code != code {
		return false, fmt.Errorf("invalid OTP code")
	}

	// Mark as verified
	_, err = db.ExecContext(ctx,
		"UPDATE otp_codes SET verified_at = now() WHERE id = $1",
		otp.ID)
	if err != nil {
		return false, fmt.Errorf("failed to mark OTP as verified: %w", err)
	}

	return true, nil
}

// IsOTPVerified checks if there's a verified OTP for an email/purpose
func (db *DB) IsOTPVerified(ctx context.Context, email string, purpose OTPPurpose) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM otp_codes 
			WHERE email = $1 AND purpose = $2 AND verified_at IS NOT NULL
			AND verified_at > now() - interval '30 minutes'
		)
	`
	err := db.QueryRowContext(ctx, query, email, string(purpose)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// DeleteOTPCodes deletes all OTP codes for an email/purpose
func (db *DB) DeleteOTPCodes(ctx context.Context, email string, purpose OTPPurpose) error {
	_, err := db.ExecContext(ctx,
		"DELETE FROM otp_codes WHERE email = $1 AND purpose = $2",
		email, string(purpose))
	return err
}

// CleanupExpiredOTPs deletes all expired OTP codes
func (db *DB) CleanupExpiredOTPs(ctx context.Context) (int64, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM otp_codes WHERE expires_at < now()")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

