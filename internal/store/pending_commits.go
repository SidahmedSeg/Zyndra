package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// PendingCommit represents a commit waiting to be deployed
type PendingCommit struct {
	ID            uuid.UUID `json:"id"`
	ServiceID     uuid.UUID `json:"service_id"`
	CommitSHA     string    `json:"commit_sha"`
	CommitMessage string    `json:"commit_message,omitempty"`
	CommitAuthor  string    `json:"commit_author,omitempty"`
	CommitURL     string    `json:"commit_url,omitempty"`
	Branch        string    `json:"branch,omitempty"`
	PushedAt      time.Time `json:"pushed_at"`
	Acknowledged  bool      `json:"acknowledged"`
	Deployed      bool      `json:"deployed"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreatePendingCommit creates a new pending commit
func (db *DB) CreatePendingCommit(ctx context.Context, pc *PendingCommit) error {
	query := `
		INSERT INTO pending_commits (service_id, commit_sha, commit_message, commit_author, commit_url, branch, pushed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := db.QueryRowContext(ctx, query,
		pc.ServiceID,
		pc.CommitSHA,
		pc.CommitMessage,
		pc.CommitAuthor,
		pc.CommitURL,
		pc.Branch,
		pc.PushedAt,
	).Scan(&pc.ID, &pc.CreatedAt)

	if err != nil {
		return err
	}

	// Update pending count on service
	_, err = db.ExecContext(ctx, `
		UPDATE services 
		SET pending_changes_count = (
			SELECT COUNT(*) FROM pending_commits 
			WHERE service_id = $1 AND acknowledged = FALSE
		)
		WHERE id = $1
	`, pc.ServiceID)

	return err
}

// GetPendingCommit retrieves a pending commit by ID
func (db *DB) GetPendingCommit(ctx context.Context, id uuid.UUID) (*PendingCommit, error) {
	query := `
		SELECT id, service_id, commit_sha, commit_message, commit_author, commit_url, 
		       branch, pushed_at, acknowledged, deployed, created_at
		FROM pending_commits
		WHERE id = $1
	`

	var pc PendingCommit
	var commitMessage, commitAuthor, commitURL, branch sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&pc.ID,
		&pc.ServiceID,
		&pc.CommitSHA,
		&commitMessage,
		&commitAuthor,
		&commitURL,
		&branch,
		&pc.PushedAt,
		&pc.Acknowledged,
		&pc.Deployed,
		&pc.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	pc.CommitMessage = commitMessage.String
	pc.CommitAuthor = commitAuthor.String
	pc.CommitURL = commitURL.String
	pc.Branch = branch.String

	return &pc, nil
}

// ListPendingCommitsByService lists all pending commits for a service
func (db *DB) ListPendingCommitsByService(ctx context.Context, serviceID uuid.UUID) ([]*PendingCommit, error) {
	query := `
		SELECT id, service_id, commit_sha, commit_message, commit_author, commit_url,
		       branch, pushed_at, acknowledged, deployed, created_at
		FROM pending_commits
		WHERE service_id = $1
		ORDER BY pushed_at DESC
	`

	rows, err := db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []*PendingCommit
	for rows.Next() {
		var pc PendingCommit
		var commitMessage, commitAuthor, commitURL, branch sql.NullString

		err := rows.Scan(
			&pc.ID,
			&pc.ServiceID,
			&pc.CommitSHA,
			&commitMessage,
			&commitAuthor,
			&commitURL,
			&branch,
			&pc.PushedAt,
			&pc.Acknowledged,
			&pc.Deployed,
			&pc.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		pc.CommitMessage = commitMessage.String
		pc.CommitAuthor = commitAuthor.String
		pc.CommitURL = commitURL.String
		pc.Branch = branch.String

		commits = append(commits, &pc)
	}

	return commits, nil
}

// ListUnacknowledgedCommits lists commits that haven't been acknowledged yet
func (db *DB) ListUnacknowledgedCommits(ctx context.Context, serviceID uuid.UUID) ([]*PendingCommit, error) {
	query := `
		SELECT id, service_id, commit_sha, commit_message, commit_author, commit_url,
		       branch, pushed_at, acknowledged, deployed, created_at
		FROM pending_commits
		WHERE service_id = $1 AND acknowledged = FALSE
		ORDER BY pushed_at ASC
	`

	rows, err := db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []*PendingCommit
	for rows.Next() {
		var pc PendingCommit
		var commitMessage, commitAuthor, commitURL, branch sql.NullString

		err := rows.Scan(
			&pc.ID,
			&pc.ServiceID,
			&pc.CommitSHA,
			&commitMessage,
			&commitAuthor,
			&commitURL,
			&branch,
			&pc.PushedAt,
			&pc.Acknowledged,
			&pc.Deployed,
			&pc.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		pc.CommitMessage = commitMessage.String
		pc.CommitAuthor = commitAuthor.String
		pc.CommitURL = commitURL.String
		pc.Branch = branch.String

		commits = append(commits, &pc)
	}

	return commits, nil
}

// AcknowledgePendingCommits marks commits as acknowledged
func (db *DB) AcknowledgePendingCommits(ctx context.Context, serviceID uuid.UUID, commitIDs []uuid.UUID) error {
	if len(commitIDs) == 0 {
		return nil
	}

	// Build query with IN clause
	query := `UPDATE pending_commits SET acknowledged = TRUE WHERE service_id = $1 AND id = ANY($2)`

	_, err := db.ExecContext(ctx, query, serviceID, commitIDs)
	if err != nil {
		return err
	}

	// Update pending count
	_, err = db.ExecContext(ctx, `
		UPDATE services 
		SET pending_changes_count = (
			SELECT COUNT(*) FROM pending_commits 
			WHERE service_id = $1 AND acknowledged = FALSE
		)
		WHERE id = $1
	`, serviceID)

	return err
}

// AcknowledgeAllPendingCommits marks all commits for a service as acknowledged
func (db *DB) AcknowledgeAllPendingCommits(ctx context.Context, serviceID uuid.UUID) error {
	query := `UPDATE pending_commits SET acknowledged = TRUE WHERE service_id = $1 AND acknowledged = FALSE`
	_, err := db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return err
	}

	// Reset pending count
	_, err = db.ExecContext(ctx, `UPDATE services SET pending_changes_count = 0 WHERE id = $1`, serviceID)
	return err
}

// MarkCommitsAsDeployed marks specific commits as deployed
func (db *DB) MarkCommitsAsDeployed(ctx context.Context, serviceID uuid.UUID, upToCommitSHA string) error {
	// Mark all commits up to and including this SHA as deployed
	query := `
		UPDATE pending_commits 
		SET deployed = TRUE, acknowledged = TRUE
		WHERE service_id = $1 AND pushed_at <= (
			SELECT pushed_at FROM pending_commits WHERE service_id = $1 AND commit_sha = $2
		)
	`
	_, err := db.ExecContext(ctx, query, serviceID, upToCommitSHA)
	if err != nil {
		return err
	}

	// Update pending count
	_, err = db.ExecContext(ctx, `
		UPDATE services 
		SET pending_changes_count = (
			SELECT COUNT(*) FROM pending_commits 
			WHERE service_id = $1 AND acknowledged = FALSE
		)
		WHERE id = $1
	`, serviceID)

	return err
}

// GetPendingChangesCount returns the number of pending changes for a service
func (db *DB) GetPendingChangesCount(ctx context.Context, serviceID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM pending_commits WHERE service_id = $1 AND acknowledged = FALSE`
	var count int
	err := db.QueryRowContext(ctx, query, serviceID).Scan(&count)
	return count, err
}

// DeletePendingCommit deletes a pending commit
func (db *DB) DeletePendingCommit(ctx context.Context, id uuid.UUID) error {
	// Get service ID first for count update
	var serviceID uuid.UUID
	err := db.QueryRowContext(ctx, `SELECT service_id FROM pending_commits WHERE id = $1`, id).Scan(&serviceID)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `DELETE FROM pending_commits WHERE id = $1`, id)
	if err != nil {
		return err
	}

	// Update count
	_, err = db.ExecContext(ctx, `
		UPDATE services 
		SET pending_changes_count = (
			SELECT COUNT(*) FROM pending_commits 
			WHERE service_id = $1 AND acknowledged = FALSE
		)
		WHERE id = $1
	`, serviceID)

	return err
}

// CleanupOldPendingCommits removes old deployed commits (keep last 50)
func (db *DB) CleanupOldPendingCommits(ctx context.Context, serviceID uuid.UUID) error {
	query := `
		DELETE FROM pending_commits
		WHERE service_id = $1 AND deployed = TRUE
		AND id NOT IN (
			SELECT id FROM pending_commits 
			WHERE service_id = $1 
			ORDER BY pushed_at DESC 
			LIMIT 50
		)
	`
	_, err := db.ExecContext(ctx, query, serviceID)
	return err
}

