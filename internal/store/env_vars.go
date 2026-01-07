package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EnvVar struct {
	ID              uuid.UUID
	ServiceID       uuid.UUID
	Key             string
	Value           sql.NullString // NULL if linked to database
	IsSecret        bool
	LinkedDatabaseID sql.NullString
	LinkType        sql.NullString // connection_url, host, port, username, password, database
	CreatedAt       time.Time
}

// CreateEnvVar creates a new environment variable
func (db *DB) CreateEnvVar(ctx context.Context, ev *EnvVar) error {
	// Generate UUID if not set (for SQLite compatibility)
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	var value interface{}
	if ev.Value.Valid {
		value = ev.Value.String
	}

	var linkedDatabaseID interface{}
	if ev.LinkedDatabaseID.Valid {
		linkedDatabaseID = ev.LinkedDatabaseID.String
	}

	var linkType interface{}
	if ev.LinkType.Valid {
		linkType = ev.LinkType.String
	}

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		isSecret := 0
		if ev.IsSecret {
			isSecret = 1
		}
		query := `
			INSERT INTO env_vars (id, service_id, key, value, is_secret, linked_database_id, link_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = db.ExecContext(ctx, query,
			ev.ID.String(), ev.ServiceID.String(), ev.Key, value, isSecret, linkedDatabaseID, linkType,
		)
		if err != nil {
			return err
		}
		// Get timestamp
		err = db.QueryRowContext(ctx, "SELECT created_at FROM env_vars WHERE id = $1", ev.ID.String()).
			Scan(&ev.CreatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO env_vars (service_id, key, value, is_secret, linked_database_id, link_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err = db.QueryRowContext(ctx, query,
		ev.ServiceID,
		ev.Key,
		value,
		ev.IsSecret,
		linkedDatabaseID,
		linkType,
	).Scan(&ev.ID, &ev.CreatedAt)

	return err
}

// GetEnvVar retrieves an environment variable by ID
func (db *DB) GetEnvVar(ctx context.Context, id uuid.UUID) (*EnvVar, error) {
	query := `
		SELECT id, service_id, key, value, is_secret,
		       linked_database_id, link_type, created_at
		FROM env_vars
		WHERE id = $1
	`

	var ev EnvVar
	var value sql.NullString
	var linkedDatabaseID sql.NullString
	var linkType sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&ev.ID,
		&ev.ServiceID,
		&ev.Key,
		&value,
		&ev.IsSecret,
		&linkedDatabaseID,
		&linkType,
		&ev.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	ev.Value = value
	ev.LinkedDatabaseID = linkedDatabaseID
	ev.LinkType = linkType

	return &ev, nil
}

// ListEnvVarsByService lists environment variables for a service
func (db *DB) ListEnvVarsByService(ctx context.Context, serviceID uuid.UUID) ([]*EnvVar, error) {
	query := `
		SELECT id, service_id, key, value, is_secret,
		       linked_database_id, link_type, created_at
		FROM env_vars
		WHERE service_id = $1
		ORDER BY key ASC
	`

	rows, err := db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []*EnvVar
	for rows.Next() {
		var ev EnvVar
		var value sql.NullString
		var linkedDatabaseID sql.NullString
		var linkType sql.NullString

		err := rows.Scan(
			&ev.ID,
			&ev.ServiceID,
			&ev.Key,
			&value,
			&ev.IsSecret,
			&linkedDatabaseID,
			&linkType,
			&ev.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		ev.Value = value
		ev.LinkedDatabaseID = linkedDatabaseID
		ev.LinkType = linkType

		envVars = append(envVars, &ev)
	}

	return envVars, rows.Err()
}

// UpdateEnvVar updates an environment variable
func (db *DB) UpdateEnvVar(ctx context.Context, id uuid.UUID, ev *EnvVar) error {
	query := `
		UPDATE env_vars
		SET value = $1, is_secret = $2, linked_database_id = $3, link_type = $4
		WHERE id = $5
	`

	var value interface{}
	if ev.Value.Valid {
		value = ev.Value.String
	}

	var linkedDatabaseID interface{}
	if ev.LinkedDatabaseID.Valid {
		linkedDatabaseID = ev.LinkedDatabaseID.String
	}

	var linkType interface{}
	if ev.LinkType.Valid {
		linkType = ev.LinkType.String
	}

	_, err := db.ExecContext(ctx, query,
		value,
		ev.IsSecret,
		linkedDatabaseID,
		linkType,
		id,
	)

	return err
}

// DeleteEnvVar deletes an environment variable
func (db *DB) DeleteEnvVar(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM env_vars WHERE id = $1`

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

// ResolveEnvVars resolves environment variables for a service
// This includes resolving linked database values
func (db *DB) ResolveEnvVars(ctx context.Context, serviceID uuid.UUID) (map[string]string, error) {
	envVars, err := db.ListEnvVarsByService(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	resolved := make(map[string]string)
	for _, ev := range envVars {
		if ev.LinkedDatabaseID.Valid {
			// Resolve from linked database
			databaseID, err := uuid.Parse(ev.LinkedDatabaseID.String)
			if err != nil {
				continue // Skip invalid database ID
			}

			database, err := db.GetDatabase(ctx, databaseID)
			if err != nil || database == nil {
				continue // Skip if database not found
			}

			// Resolve based on link type
			switch ev.LinkType.String {
			case "connection_url":
				if database.ConnectionURL.Valid {
					resolved[ev.Key] = database.ConnectionURL.String
				}
			case "host":
				if database.InternalHostname.Valid {
					resolved[ev.Key] = database.InternalHostname.String
				}
			case "port":
				if database.Port.Valid {
					resolved[ev.Key] = fmt.Sprintf("%d", database.Port.Int64)
				}
			case "username":
				if database.Username.Valid {
					resolved[ev.Key] = database.Username.String
				}
			case "password":
				if database.Password.Valid {
					resolved[ev.Key] = database.Password.String // TODO: Decrypt
				}
			case "database":
				if database.DatabaseName.Valid {
					resolved[ev.Key] = database.DatabaseName.String
				}
			}
		} else if ev.Value.Valid {
			// Direct value
			resolved[ev.Key] = ev.Value.String
		}
	}

	return resolved, nil
}

