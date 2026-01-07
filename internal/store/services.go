package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID                  uuid.UUID
	ProjectID           uuid.UUID
	GitSourceID         sql.NullString
	Name                string
	Type                string // app, database, volume
	Status              string
	InstanceSize        string
	Port                int
	OpenStackInstanceID sql.NullString
	OpenStackFIPID      sql.NullString
	OpenStackFIPAddress sql.NullString
	SecurityGroupID     sql.NullString
	Subdomain           sql.NullString
	GeneratedURL        sql.NullString
	CurrentImageTag     sql.NullString
	CanvasX             int
	CanvasY             int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// CreateService creates a new service
func (db *DB) CreateService(ctx context.Context, s *Service) error {
	// Generate UUID if not set (for SQLite compatibility)
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	var gitSourceID interface{}
	if s.GitSourceID.Valid {
		gitSourceID = s.GitSourceID.String
	} else {
		gitSourceID = nil
	}

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO services (
				id, project_id, git_source_id, name, type, status,
				instance_size, port, canvas_x, canvas_y
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = db.ExecContext(ctx, query,
			s.ID.String(), s.ProjectID.String(), gitSourceID, s.Name, s.Type, s.Status,
			s.InstanceSize, s.Port, s.CanvasX, s.CanvasY,
		)
		if err != nil {
			return err
		}
		// Get timestamps
		err = db.QueryRowContext(ctx, "SELECT created_at, updated_at FROM services WHERE id = $1", s.ID.String()).
			Scan(&s.CreatedAt, &s.UpdatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO services (
			project_id, git_source_id, name, type, status,
			instance_size, port, canvas_x, canvas_y
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err = db.QueryRowContext(ctx, query,
		s.ProjectID,
		gitSourceID,
		s.Name,
		s.Type,
		s.Status,
		s.InstanceSize,
		s.Port,
		s.CanvasX,
		s.CanvasY,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	return err
}

// GetService retrieves a service by ID
func (db *DB) GetService(ctx context.Context, id uuid.UUID) (*Service, error) {
	var s Service
	query := `
		SELECT id, project_id, git_source_id, name, type, status,
		       instance_size, port, openstack_instance_id, openstack_fip_id,
		       openstack_fip_address, security_group_id, subdomain,
		       generated_url, current_image_tag, canvas_x, canvas_y,
		       created_at, updated_at
		FROM services
		WHERE id = $1
	`

	var gitSourceID sql.NullString
	var openstackInstanceID sql.NullString
	var openstackFIPID sql.NullString
	var openstackFIPAddress sql.NullString
	var securityGroupID sql.NullString
	var subdomain sql.NullString
	var generatedURL sql.NullString
	var currentImageTag sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.ProjectID,
		&gitSourceID,
		&s.Name,
		&s.Type,
		&s.Status,
		&s.InstanceSize,
		&s.Port,
		&openstackInstanceID,
		&openstackFIPID,
		&openstackFIPAddress,
		&securityGroupID,
		&subdomain,
		&generatedURL,
		&currentImageTag,
		&s.CanvasX,
		&s.CanvasY,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	s.GitSourceID = gitSourceID
	s.OpenStackInstanceID = openstackInstanceID
	s.OpenStackFIPID = openstackFIPID
	s.OpenStackFIPAddress = openstackFIPAddress
	s.SecurityGroupID = securityGroupID
	s.Subdomain = subdomain
	s.GeneratedURL = generatedURL
	s.CurrentImageTag = currentImageTag

	return &s, nil
}

// ListServicesByProject lists all services in a project
func (db *DB) ListServicesByProject(ctx context.Context, projectID uuid.UUID) ([]*Service, error) {
	query := `
		SELECT id, project_id, git_source_id, name, type, status,
		       instance_size, port, openstack_instance_id, openstack_fip_id,
		       openstack_fip_address, security_group_id, subdomain,
		       generated_url, current_image_tag, canvas_x, canvas_y,
		       created_at, updated_at
		FROM services
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		var s Service
		var gitSourceID sql.NullString
		var openstackInstanceID sql.NullString
		var openstackFIPID sql.NullString
		var openstackFIPAddress sql.NullString
		var securityGroupID sql.NullString
		var subdomain sql.NullString
		var generatedURL sql.NullString
		var currentImageTag sql.NullString

		err := rows.Scan(
			&s.ID,
			&s.ProjectID,
			&gitSourceID,
			&s.Name,
			&s.Type,
			&s.Status,
			&s.InstanceSize,
			&s.Port,
			&openstackInstanceID,
			&openstackFIPID,
			&openstackFIPAddress,
			&securityGroupID,
			&subdomain,
			&generatedURL,
			&currentImageTag,
			&s.CanvasX,
			&s.CanvasY,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		s.GitSourceID = gitSourceID
		s.OpenStackInstanceID = openstackInstanceID
		s.OpenStackFIPID = openstackFIPID
		s.OpenStackFIPAddress = openstackFIPAddress
		s.SecurityGroupID = securityGroupID
		s.Subdomain = subdomain
		s.GeneratedURL = generatedURL
		s.CurrentImageTag = currentImageTag

		services = append(services, &s)
	}

	return services, rows.Err()
}

// UpdateService updates a service
func (db *DB) UpdateService(ctx context.Context, id uuid.UUID, updates *Service) error {
	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	var query string
	if isSQLite {
		var fipAddress interface{}
		if updates.OpenStackFIPAddress.Valid {
			fipAddress = updates.OpenStackFIPAddress.String
		}
		query = `
			UPDATE services
			SET name = $1,
			    type = $2,
			    instance_size = $3,
			    port = $4,
			    status = $5,
			    canvas_x = $6,
			    canvas_y = $7,
			    openstack_fip_address = $8,
			    updated_at = datetime('now')
			WHERE id = $9
		`
		_, err = db.ExecContext(ctx, query,
			updates.Name,
			updates.Type,
			updates.InstanceSize,
			updates.Port,
			updates.Status,
			updates.CanvasX,
			updates.CanvasY,
			fipAddress,
			id.String(),
		)
		if err != nil {
			return err
		}
		// Get updated timestamp
		err = db.QueryRowContext(ctx, "SELECT updated_at FROM services WHERE id = $1", id.String()).
			Scan(&updates.UpdatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	var fipAddress interface{}
	if updates.OpenStackFIPAddress.Valid {
		fipAddress = updates.OpenStackFIPAddress.String
	}
	query = `
		UPDATE services
		SET name = $1,
		    type = $2,
		    instance_size = $3,
		    port = $4,
		    status = $5,
		    canvas_x = $6,
		    canvas_y = $7,
		    openstack_fip_address = $8,
		    updated_at = now()
		WHERE id = $9
		RETURNING updated_at
	`

	err = db.QueryRowContext(ctx, query,
		updates.Name,
		updates.Type,
		updates.InstanceSize,
		updates.Port,
		updates.Status,
		updates.CanvasX,
		updates.CanvasY,
		fipAddress,
		id,
	).Scan(&updates.UpdatedAt)

	return err
}

// UpdateServicePosition updates the canvas position of a service
func (db *DB) UpdateServicePosition(ctx context.Context, id uuid.UUID, x, y int) error {
	query := `
		UPDATE services
		SET canvas_x = $1,
		    canvas_y = $2,
		    updated_at = now()
		WHERE id = $3
	`

	_, err := db.ExecContext(ctx, query, x, y, id)
	return err
}

// DeleteService deletes a service
func (db *DB) DeleteService(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM services WHERE id = $1`

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

// ServiceExists checks if a service exists and belongs to a project
func (db *DB) ServiceExists(ctx context.Context, id uuid.UUID, projectID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM services WHERE id = $1 AND project_id = $2)`

	err := db.QueryRowContext(ctx, query, id, projectID).Scan(&exists)
	return exists, err
}

// GetProjectIDForService gets the project ID for a service
func (db *DB) GetProjectIDForService(ctx context.Context, serviceID uuid.UUID) (uuid.UUID, error) {
	var projectID uuid.UUID
	query := `SELECT project_id FROM services WHERE id = $1`

	err := db.QueryRowContext(ctx, query, serviceID).Scan(&projectID)
	return projectID, err
}

