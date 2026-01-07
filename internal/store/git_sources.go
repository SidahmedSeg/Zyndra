package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type GitSource struct {
	ID              uuid.UUID
	ServiceID       uuid.UUID
	GitConnectionID uuid.UUID
	Provider        string // github, gitlab
	RepoOwner       string
	RepoName        string
	Branch          string
	RootDir         sql.NullString
	WebhookID       sql.NullString
	WebhookSecret   sql.NullString
	CreatedAt       time.Time
}

// CreateGitSource creates a new git source
func (db *DB) CreateGitSource(ctx context.Context, gs *GitSource) error {
	// Generate UUID if not set (for SQLite compatibility)
	if gs.ID == uuid.Nil {
		gs.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO git_sources (
				id, service_id, git_connection_id, provider, repo_owner,
				repo_name, branch, root_dir, webhook_id, webhook_secret
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = db.ExecContext(ctx, query,
			gs.ID.String(), gs.ServiceID.String(), gs.GitConnectionID.String(), gs.Provider,
			gs.RepoOwner, gs.RepoName, gs.Branch, gs.RootDir, gs.WebhookID, gs.WebhookSecret,
		)
		if err != nil {
			return err
		}
		// Get timestamp
		err = db.QueryRowContext(ctx, "SELECT created_at FROM git_sources WHERE id = $1", gs.ID.String()).
			Scan(&gs.CreatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO git_sources (
			service_id, git_connection_id, provider, repo_owner,
			repo_name, branch, root_dir, webhook_id, webhook_secret
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`

	err = db.QueryRowContext(ctx, query,
		gs.ServiceID,
		gs.GitConnectionID,
		gs.Provider,
		gs.RepoOwner,
		gs.RepoName,
		gs.Branch,
		gs.RootDir,
		gs.WebhookID,
		gs.WebhookSecret,
	).Scan(&gs.ID, &gs.CreatedAt)

	return err
}

// GetGitSource retrieves a git source by ID
func (db *DB) GetGitSource(ctx context.Context, id uuid.UUID) (*GitSource, error) {
	var gs GitSource
	query := `
		SELECT id, service_id, git_connection_id, provider, repo_owner,
		       repo_name, branch, root_dir, webhook_id, webhook_secret,
		       created_at
		FROM git_sources
		WHERE id = $1
	`

	var rootDir sql.NullString
	var webhookID sql.NullString
	var webhookSecret sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&gs.ID,
		&gs.ServiceID,
		&gs.GitConnectionID,
		&gs.Provider,
		&gs.RepoOwner,
		&gs.RepoName,
		&gs.Branch,
		&rootDir,
		&webhookID,
		&webhookSecret,
		&gs.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	gs.RootDir = rootDir
	gs.WebhookID = webhookID
	gs.WebhookSecret = webhookSecret

	return &gs, nil
}

// GetGitSourceByService retrieves git sources for a service
func (db *DB) GetGitSourceByService(ctx context.Context, serviceID uuid.UUID) (*GitSource, error) {
	var gs GitSource
	query := `
		SELECT id, service_id, git_connection_id, provider, repo_owner,
		       repo_name, branch, root_dir, webhook_id, webhook_secret,
		       created_at
		FROM git_sources
		WHERE service_id = $1
		LIMIT 1
	`

	var rootDir sql.NullString
	var webhookID sql.NullString
	var webhookSecret sql.NullString

	err := db.QueryRowContext(ctx, query, serviceID).Scan(
		&gs.ID,
		&gs.ServiceID,
		&gs.GitConnectionID,
		&gs.Provider,
		&gs.RepoOwner,
		&gs.RepoName,
		&gs.Branch,
		&rootDir,
		&webhookID,
		&webhookSecret,
		&gs.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	gs.RootDir = rootDir
	gs.WebhookID = webhookID
	gs.WebhookSecret = webhookSecret

	return &gs, nil
}

// UpdateGitSource updates a git source
func (db *DB) UpdateGitSource(ctx context.Context, id uuid.UUID, gs *GitSource) error {
	query := `
		UPDATE git_sources
		SET branch = $1, root_dir = $2, webhook_id = $3, webhook_secret = $4
		WHERE id = $5
	`

	_, err := db.ExecContext(ctx, query,
		gs.Branch,
		gs.RootDir,
		gs.WebhookID,
		gs.WebhookSecret,
		id,
	)

	return err
}

// DeleteGitSource deletes a git source
func (db *DB) DeleteGitSource(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM git_sources WHERE id = $1`

	result, err := db.ExecContext(ctx, query, id)
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

