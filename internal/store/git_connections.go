package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type GitConnection struct {
	ID            uuid.UUID
	CasdoorOrgID  string
	Provider      string // github, gitlab
	AccessToken   string // encrypted
	RefreshToken  sql.NullString
	TokenExpiresAt sql.NullTime
	AccountName   sql.NullString
	AccountID     sql.NullString
	ConnectedBy   sql.NullString
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreateGitConnection creates a new git connection
func (db *DB) CreateGitConnection(ctx context.Context, gc *GitConnection) error {
	// Generate UUID if not set (for SQLite compatibility)
	if gc.ID == uuid.Nil {
		gc.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO git_connections (
				id, casdoor_org_id, provider, access_token, refresh_token,
				token_expires_at, account_name, account_id, connected_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = db.ExecContext(ctx, query,
			gc.ID.String(), gc.CasdoorOrgID, gc.Provider, gc.AccessToken,
			gc.RefreshToken, gc.TokenExpiresAt, gc.AccountName, gc.AccountID, gc.ConnectedBy,
		)
		if err != nil {
			return err
		}
		// Get timestamps
		err = db.QueryRowContext(ctx, "SELECT created_at, updated_at FROM git_connections WHERE id = $1", gc.ID.String()).
			Scan(&gc.CreatedAt, &gc.UpdatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO git_connections (
			casdoor_org_id, provider, access_token, refresh_token,
			token_expires_at, account_name, account_id, connected_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err = db.QueryRowContext(ctx, query,
		gc.CasdoorOrgID,
		gc.Provider,
		gc.AccessToken,
		gc.RefreshToken,
		gc.TokenExpiresAt,
		gc.AccountName,
		gc.AccountID,
		gc.ConnectedBy,
	).Scan(&gc.ID, &gc.CreatedAt, &gc.UpdatedAt)

	return err
}

// GetGitConnection retrieves a git connection by ID
func (db *DB) GetGitConnection(ctx context.Context, id uuid.UUID) (*GitConnection, error) {
	var gc GitConnection
	query := `
		SELECT id, casdoor_org_id, provider, access_token, refresh_token,
		       token_expires_at, account_name, account_id, connected_by,
		       created_at, updated_at
		FROM git_connections
		WHERE id = $1
	`

	var refreshToken sql.NullString
	var tokenExpiresAt sql.NullTime
	var accountName sql.NullString
	var accountID sql.NullString
	var connectedBy sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&gc.ID,
		&gc.CasdoorOrgID,
		&gc.Provider,
		&gc.AccessToken,
		&refreshToken,
		&tokenExpiresAt,
		&accountName,
		&accountID,
		&connectedBy,
		&gc.CreatedAt,
		&gc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	gc.RefreshToken = refreshToken
	gc.TokenExpiresAt = tokenExpiresAt
	gc.AccountName = accountName
	gc.AccountID = accountID
	gc.ConnectedBy = connectedBy

	return &gc, nil
}

// ListGitConnectionsByOrg lists all git connections for an organization
func (db *DB) ListGitConnectionsByOrg(ctx context.Context, orgID string) ([]*GitConnection, error) {
	query := `
		SELECT id, casdoor_org_id, provider, access_token, refresh_token,
		       token_expires_at, account_name, account_id, connected_by,
		       created_at, updated_at
		FROM git_connections
		WHERE casdoor_org_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*GitConnection
	for rows.Next() {
		var gc GitConnection
		var refreshToken sql.NullString
		var tokenExpiresAt sql.NullTime
		var accountName sql.NullString
		var accountID sql.NullString
		var connectedBy sql.NullString

		err := rows.Scan(
			&gc.ID,
			&gc.CasdoorOrgID,
			&gc.Provider,
			&gc.AccessToken,
			&refreshToken,
			&tokenExpiresAt,
			&accountName,
			&accountID,
			&connectedBy,
			&gc.CreatedAt,
			&gc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		gc.RefreshToken = refreshToken
		gc.TokenExpiresAt = tokenExpiresAt
		gc.AccountName = accountName
		gc.AccountID = accountID
		gc.ConnectedBy = connectedBy

		connections = append(connections, &gc)
	}

	return connections, rows.Err()
}

// GetGitConnectionByOrgAndProvider gets a git connection by org and provider
func (db *DB) GetGitConnectionByOrgAndProvider(ctx context.Context, orgID, provider string) (*GitConnection, error) {
	var gc GitConnection
	query := `
		SELECT id, casdoor_org_id, provider, access_token, refresh_token,
		       token_expires_at, account_name, account_id, connected_by,
		       created_at, updated_at
		FROM git_connections
		WHERE casdoor_org_id = $1 AND provider = $2
		LIMIT 1
	`

	var refreshToken sql.NullString
	var tokenExpiresAt sql.NullTime
	var accountName sql.NullString
	var accountID sql.NullString
	var connectedBy sql.NullString

	err := db.QueryRowContext(ctx, query, orgID, provider).Scan(
		&gc.ID,
		&gc.CasdoorOrgID,
		&gc.Provider,
		&gc.AccessToken,
		&refreshToken,
		&tokenExpiresAt,
		&accountName,
		&accountID,
		&connectedBy,
		&gc.CreatedAt,
		&gc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	gc.RefreshToken = refreshToken
	gc.TokenExpiresAt = tokenExpiresAt
	gc.AccountName = accountName
	gc.AccountID = accountID
	gc.ConnectedBy = connectedBy

	return &gc, nil
}

// UpdateGitConnection updates a git connection
func (db *DB) UpdateGitConnection(ctx context.Context, id uuid.UUID, gc *GitConnection) error {
	query := `
		UPDATE git_connections
		SET access_token = $1,
		    refresh_token = $2,
		    token_expires_at = $3,
		    account_name = $4,
		    account_id = $5,
		    updated_at = now()
		WHERE id = $6
		RETURNING updated_at
	`

	err := db.QueryRowContext(ctx, query,
		gc.AccessToken,
		gc.RefreshToken,
		gc.TokenExpiresAt,
		gc.AccountName,
		gc.AccountID,
		id,
	).Scan(&gc.UpdatedAt)

	return err
}

// DeleteGitConnection deletes a git connection
func (db *DB) DeleteGitConnection(ctx context.Context, id uuid.UUID, orgID string) error {
	query := `DELETE FROM git_connections WHERE id = $1 AND casdoor_org_id = $2`

	result, err := db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

