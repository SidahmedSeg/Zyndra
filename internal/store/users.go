package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"` // Never expose password hash
	Name          string    `json:"name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateUser creates a new user
func (db *DB) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, avatar_url, email_verified, created_at, updated_at
	`

	var user User
	var avatarURL sql.NullString
	err := db.QueryRowContext(ctx, query, email, passwordHash, name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatarURL,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.PasswordHash = passwordHash // Keep for internal use
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, email_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	var avatarURL sql.NullString
	err := db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&avatarURL,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	var avatarURL sql.NullString
	err := db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&avatarURL,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

// UpdateUser updates a user's information
func (db *DB) UpdateUser(ctx context.Context, id, name, avatarURL string) (*User, error) {
	query := `
		UPDATE users
		SET name = COALESCE(NULLIF($2, ''), name),
		    avatar_url = COALESCE(NULLIF($3, ''), avatar_url),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, email, name, avatar_url, email_verified, created_at, updated_at
	`

	var user User
	var avatar sql.NullString
	err := db.QueryRowContext(ctx, query, id, name, avatarURL).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatar,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if avatar.Valid {
		user.AvatarURL = avatar.String
	}

	return &user, nil
}

// DeleteUser deletes a user
func (db *DB) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UserExistsByEmail checks if a user exists by email
func (db *DB) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// CreateUserWithVerifiedEmail creates a new user with email already verified
func (db *DB) CreateUserWithVerifiedEmail(ctx context.Context, email, password, name string) (*User, error) {
	// Hash password
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (email, password_hash, name, email_verified)
		VALUES ($1, $2, $3, true)
		RETURNING id, email, name, avatar_url, email_verified, created_at, updated_at
	`

	var user User
	var avatarURL sql.NullString
	err = db.QueryRowContext(ctx, query, email, passwordHash, name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&avatarURL,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.PasswordHash = passwordHash
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return &user, nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

