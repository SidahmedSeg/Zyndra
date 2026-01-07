package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Deployment struct {
	ID            uuid.UUID
	ServiceID     uuid.UUID
	CommitSHA     sql.NullString
	CommitMessage sql.NullString
	CommitAuthor  sql.NullString
	Status        string // queued, building, pushing, deploying, success, failed, cancelled
	ImageTag      sql.NullString
	BuildDuration sql.NullInt64 // seconds
	DeployDuration sql.NullInt64 // seconds
	ErrorMessage  sql.NullString
	TriggeredBy   string // webhook, manual, rollback
	StartedAt     sql.NullTime
	FinishedAt    sql.NullTime
	CreatedAt     time.Time
}

// CreateDeployment creates a new deployment record
func (db *DB) CreateDeployment(ctx context.Context, d *Deployment) error {
	// Generate UUID if not set (for SQLite compatibility)
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	var commitSHA sql.NullString
	if d.CommitSHA.Valid {
		commitSHA = d.CommitSHA
	}

	var commitMessage sql.NullString
	if d.CommitMessage.Valid {
		commitMessage = d.CommitMessage
	}

	var commitAuthor sql.NullString
	if d.CommitAuthor.Valid {
		commitAuthor = d.CommitAuthor
	}

	var imageTag sql.NullString
	if d.ImageTag.Valid {
		imageTag = d.ImageTag
	}

	var startedAt sql.NullTime
	if d.StartedAt.Valid {
		startedAt = d.StartedAt
	}

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO deployments (
				id, service_id, commit_sha, commit_message, commit_author,
				status, image_tag, triggered_by, started_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = db.ExecContext(ctx, query,
			d.ID.String(), d.ServiceID.String(), commitSHA, commitMessage, commitAuthor,
			d.Status, imageTag, d.TriggeredBy, startedAt,
		)
		if err != nil {
			return err
		}
		// Get timestamp
		err = db.QueryRowContext(ctx, "SELECT created_at FROM deployments WHERE id = $1", d.ID.String()).
			Scan(&d.CreatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO deployments (
			service_id, commit_sha, commit_message, commit_author,
			status, image_tag, triggered_by, started_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	err = db.QueryRowContext(ctx, query,
		d.ServiceID,
		commitSHA,
		commitMessage,
		commitAuthor,
		d.Status,
		imageTag,
		d.TriggeredBy,
		startedAt,
	).Scan(&d.ID, &d.CreatedAt)

	return err
}

// GetDeployment retrieves a deployment by ID
func (db *DB) GetDeployment(ctx context.Context, id uuid.UUID) (*Deployment, error) {
	var d Deployment
	query := `
		SELECT id, service_id, commit_sha, commit_message, commit_author,
		       status, image_tag, build_duration, deploy_duration,
		       error_message, triggered_by, started_at, finished_at, created_at
		FROM deployments
		WHERE id = $1
	`

	var commitSHA sql.NullString
	var commitMessage sql.NullString
	var commitAuthor sql.NullString
	var imageTag sql.NullString
	var buildDuration sql.NullInt64
	var deployDuration sql.NullInt64
	var errorMessage sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime

	err := db.QueryRowContext(ctx, query, id).Scan(
		&d.ID,
		&d.ServiceID,
		&commitSHA,
		&commitMessage,
		&commitAuthor,
		&d.Status,
		&imageTag,
		&buildDuration,
		&deployDuration,
		&errorMessage,
		&d.TriggeredBy,
		&startedAt,
		&finishedAt,
		&d.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	d.CommitSHA = commitSHA
	d.CommitMessage = commitMessage
	d.CommitAuthor = commitAuthor
	d.ImageTag = imageTag
	d.BuildDuration = buildDuration
	d.DeployDuration = deployDuration
	d.ErrorMessage = errorMessage
	d.StartedAt = startedAt
	d.FinishedAt = finishedAt

	return &d, nil
}

// ListDeploymentsByService lists deployments for a service, ordered by created_at DESC
func (db *DB) ListDeploymentsByService(ctx context.Context, serviceID uuid.UUID, limit, offset int) ([]*Deployment, error) {
	query := `
		SELECT id, service_id, commit_sha, commit_message, commit_author,
		       status, image_tag, build_duration, deploy_duration,
		       error_message, triggered_by, started_at, finished_at, created_at
		FROM deployments
		WHERE service_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.QueryContext(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		var d Deployment
		var commitSHA sql.NullString
		var commitMessage sql.NullString
		var commitAuthor sql.NullString
		var imageTag sql.NullString
		var buildDuration sql.NullInt64
		var deployDuration sql.NullInt64
		var errorMessage sql.NullString
		var startedAt sql.NullTime
		var finishedAt sql.NullTime

		err := rows.Scan(
			&d.ID,
			&d.ServiceID,
			&commitSHA,
			&commitMessage,
			&commitAuthor,
			&d.Status,
			&imageTag,
			&buildDuration,
			&deployDuration,
			&errorMessage,
			&d.TriggeredBy,
			&startedAt,
			&finishedAt,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		d.CommitSHA = commitSHA
		d.CommitMessage = commitMessage
		d.CommitAuthor = commitAuthor
		d.ImageTag = imageTag
		d.BuildDuration = buildDuration
		d.DeployDuration = deployDuration
		d.ErrorMessage = errorMessage
		d.StartedAt = startedAt
		d.FinishedAt = finishedAt

		deployments = append(deployments, &d)
	}

	return deployments, rows.Err()
}

// GetSuccessfulDeploymentsByService gets successful deployments for a service (for rollback)
func (db *DB) GetSuccessfulDeploymentsByService(ctx context.Context, serviceID uuid.UUID, limit int) ([]*Deployment, error) {
	query := `
		SELECT id, service_id, commit_sha, commit_message, commit_author,
		       status, image_tag, build_duration, deploy_duration,
		       error_message, triggered_by, started_at, finished_at, created_at
		FROM deployments
		WHERE service_id = $1 AND status = 'success' AND image_tag IS NOT NULL
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, serviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		var d Deployment
		var commitSHA sql.NullString
		var commitMessage sql.NullString
		var commitAuthor sql.NullString
		var imageTag sql.NullString
		var buildDuration sql.NullInt64
		var deployDuration sql.NullInt64
		var errorMessage sql.NullString
		var startedAt sql.NullTime
		var finishedAt sql.NullTime

		err := rows.Scan(
			&d.ID,
			&d.ServiceID,
			&commitSHA,
			&commitMessage,
			&commitAuthor,
			&d.Status,
			&imageTag,
			&buildDuration,
			&deployDuration,
			&errorMessage,
			&d.TriggeredBy,
			&startedAt,
			&finishedAt,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		d.CommitSHA = commitSHA
		d.CommitMessage = commitMessage
		d.CommitAuthor = commitAuthor
		d.ImageTag = imageTag
		d.BuildDuration = buildDuration
		d.DeployDuration = deployDuration
		d.ErrorMessage = errorMessage
		d.StartedAt = startedAt
		d.FinishedAt = finishedAt

		deployments = append(deployments, &d)
	}

	return deployments, rows.Err()
}

// UpdateDeploymentStatus updates the status of a deployment
func (db *DB) UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE deployments SET status = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, status, id)
	return err
}

// UpdateDeploymentProgress updates deployment progress fields
func (db *DB) UpdateDeploymentProgress(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	var setParts []string
	var args []interface{}
	argIndex := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if imageTag, ok := updates["image_tag"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("image_tag = $%d", argIndex))
		args = append(args, imageTag)
		argIndex++
	}

	if buildDuration, ok := updates["build_duration"].(int64); ok {
		setParts = append(setParts, fmt.Sprintf("build_duration = $%d", argIndex))
		args = append(args, buildDuration)
		argIndex++
	}

	if deployDuration, ok := updates["deploy_duration"].(int64); ok {
		setParts = append(setParts, fmt.Sprintf("deploy_duration = $%d", argIndex))
		args = append(args, deployDuration)
		argIndex++
	}

	if errorMessage, ok := updates["error_message"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("error_message = $%d", argIndex))
		args = append(args, errorMessage)
		argIndex++
	}

	if startedAt, ok := updates["started_at"].(time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("started_at = $%d", argIndex))
		args = append(args, startedAt)
		argIndex++
	}

	if finishedAt, ok := updates["finished_at"].(time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("finished_at = $%d", argIndex))
		args = append(args, finishedAt)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil // No updates
	}

	query := fmt.Sprintf("UPDATE deployments SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// AddDeploymentLog adds a log entry for a deployment
func (db *DB) AddDeploymentLog(ctx context.Context, deploymentID uuid.UUID, phase, level, message string, metadata map[string]interface{}) error {
	var metadataJSON sql.NullString
	if metadata != nil {
		jsonBytes, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = sql.NullString{String: string(jsonBytes), Valid: true}
	}

	query := `
		INSERT INTO deployment_logs (deployment_id, phase, level, message, metadata)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := db.ExecContext(ctx, query, deploymentID, phase, level, message, metadataJSON)
	return err
}

// GetDeploymentLogs retrieves logs for a deployment
func (db *DB) GetDeploymentLogs(ctx context.Context, deploymentID uuid.UUID, limit int) ([]*DeploymentLog, error) {
	query := `
		SELECT id, deployment_id, timestamp, phase, level, message, metadata
		FROM deployment_logs
		WHERE deployment_id = $1
		ORDER BY timestamp ASC
		LIMIT $2
	`

	rows, err := db.QueryContext(ctx, query, deploymentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*DeploymentLog
	for rows.Next() {
		var log DeploymentLog
		var metadataJSON sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.DeploymentID,
			&log.Timestamp,
			&log.Phase,
			&log.Level,
			&log.Message,
			&metadataJSON,
		)
		if err != nil {
			return nil, err
		}

		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &log.Metadata)
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

type DeploymentLog struct {
	ID           int64
	DeploymentID uuid.UUID
	Timestamp    time.Time
	Phase        string
	Level        string
	Message      string
	Metadata     map[string]interface{}
}
