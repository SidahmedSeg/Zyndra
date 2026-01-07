package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID        uuid.UUID
	Type      string // build, deploy, provision_infra, etc.
	Payload   map[string]interface{}
	Status    string // queued, processing, completed, failed
	Attempts  int
	MaxAttempts int
	Error     sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
	StartedAt sql.NullTime
	FinishedAt sql.NullTime
}

// CreateJob creates a new job
func (db *DB) CreateJob(ctx context.Context, job *Job) error {
	// Generate UUID if not set (for SQLite compatibility)
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO jobs (id, type, payload, status, attempts, max_attempts)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = db.ExecContext(ctx, query,
			job.ID.String(), job.Type, payloadJSON, job.Status, job.Attempts, job.MaxAttempts,
		)
		if err != nil {
			return err
		}
		// Get timestamps
		err = db.QueryRowContext(ctx, "SELECT created_at, updated_at FROM jobs WHERE id = $1", job.ID.String()).
			Scan(&job.CreatedAt, &job.UpdatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO jobs (type, payload, status, attempts, max_attempts)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err = db.QueryRowContext(ctx, query,
		job.Type,
		payloadJSON,
		job.Status,
		job.Attempts,
		job.MaxAttempts,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)

	return err
}

// GetNextJob gets the next queued job using SKIP LOCKED
func (db *DB) GetNextJob(ctx context.Context) (*Job, error) {
	query := `
		SELECT id, type, payload, status, attempts, max_attempts, error,
		       created_at, updated_at, started_at, finished_at
		FROM jobs
		WHERE status = 'queued'
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	var job Job
	var payloadJSON []byte
	var errorMsg sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime

	err := db.QueryRowContext(ctx, query).Scan(
		&job.ID,
		&job.Type,
		&payloadJSON,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&errorMsg,
		&job.CreatedAt,
		&job.UpdatedAt,
		&startedAt,
		&finishedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// Parse payload
	if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
		return nil, err
	}

	job.Error = errorMsg
	job.StartedAt = startedAt
	job.FinishedAt = finishedAt

	return &job, nil
}

// UpdateJobStatus updates job status
func (db *DB) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	query := `UPDATE jobs SET status = $1, updated_at = now() WHERE id = $2`
	_, err := db.ExecContext(ctx, query, status, jobID)
	return err
}

// StartJob marks a job as processing
func (db *DB) StartJob(ctx context.Context, jobID uuid.UUID) error {
	query := `
		UPDATE jobs 
		SET status = 'processing', started_at = now(), updated_at = now()
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, jobID)
	return err
}

// CompleteJob marks a job as completed
func (db *DB) CompleteJob(ctx context.Context, jobID uuid.UUID) error {
	query := `
		UPDATE jobs 
		SET status = 'completed', finished_at = now(), updated_at = now()
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, jobID)
	return err
}

// FailJob marks a job as failed
func (db *DB) FailJob(ctx context.Context, jobID uuid.UUID, errorMsg string) error {
	query := `
		UPDATE jobs 
		SET status = 'failed', error = $1, finished_at = now(), updated_at = now()
		WHERE id = $2
	`
	_, err := db.ExecContext(ctx, query, errorMsg, jobID)
	return err
}

// IncrementJobAttempts increments job attempts
func (db *DB) IncrementJobAttempts(ctx context.Context, jobID uuid.UUID) error {
	query := `
		UPDATE jobs 
		SET attempts = attempts + 1, updated_at = now()
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, jobID)
	return err
}

// GetJob retrieves a job by ID
func (db *DB) GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error) {
	query := `
		SELECT id, type, payload, status, attempts, max_attempts, error,
		       created_at, updated_at, started_at, finished_at
		FROM jobs
		WHERE id = $1
	`

	var job Job
	var payloadJSON []byte
	var errorMsg sql.NullString
	var startedAt sql.NullTime
	var finishedAt sql.NullTime

	err := db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID,
		&job.Type,
		&payloadJSON,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&errorMsg,
		&job.CreatedAt,
		&job.UpdatedAt,
		&startedAt,
		&finishedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// Parse payload
	if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
		return nil, err
	}

	job.Error = errorMsg
	job.StartedAt = startedAt
	job.FinishedAt = finishedAt

	return &job, nil
}

