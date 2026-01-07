package store

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID                uuid.UUID
	CasdoorOrgID      string
	Name              string
	Slug              string
	Description       sql.NullString
	OpenStackTenantID string
	OpenStackNetworkID sql.NullString
	DefaultRegion     sql.NullString
	AutoDeploy        bool
	CreatedBy         sql.NullString
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (db *DB) CreateProject(ctx context.Context, p *Project) error {
	// Generate UUID if not set (for SQLite compatibility)
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO projects (
				id, casdoor_org_id, name, slug, description,
				openstack_tenant_id, openstack_network_id,
				default_region, auto_deploy, created_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = db.ExecContext(ctx, query,
			p.ID.String(), p.CasdoorOrgID, p.Name, p.Slug, p.Description,
			p.OpenStackTenantID, p.OpenStackNetworkID,
			p.DefaultRegion, p.AutoDeploy, p.CreatedBy,
		)
		if err != nil {
			return err
		}
		// Get timestamps
		err = db.QueryRowContext(ctx, "SELECT created_at, updated_at FROM projects WHERE id = $1", p.ID.String()).
			Scan(&p.CreatedAt, &p.UpdatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO projects (
			casdoor_org_id, name, slug, description,
			openstack_tenant_id, openstack_network_id,
			default_region, auto_deploy, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err = db.QueryRowContext(ctx, query,
		p.CasdoorOrgID, p.Name, p.Slug, p.Description,
		p.OpenStackTenantID, p.OpenStackNetworkID,
		p.DefaultRegion, p.AutoDeploy, p.CreatedBy,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	return err
}

func (db *DB) GetProject(ctx context.Context, id uuid.UUID) (*Project, error) {
	var p Project
	query := `SELECT * FROM projects WHERE id = $1`

	err := db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.CasdoorOrgID, &p.Name, &p.Slug, &p.Description,
		&p.OpenStackTenantID, &p.OpenStackNetworkID,
		&p.DefaultRegion, &p.AutoDeploy, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &p, err
}

func (db *DB) ListProjectsByOrg(ctx context.Context, orgID string) ([]*Project, error) {
	query := `SELECT * FROM projects WHERE casdoor_org_id = $1 ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		err := rows.Scan(
			&p.ID, &p.CasdoorOrgID, &p.Name, &p.Slug, &p.Description,
			&p.OpenStackTenantID, &p.OpenStackNetworkID,
			&p.DefaultRegion, &p.AutoDeploy, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}

	return projects, rows.Err()
}

// UpdateProject updates an existing project
func (db *DB) UpdateProject(ctx context.Context, id uuid.UUID, updates *Project) error {
	query := `
		UPDATE projects 
		SET name = $1,
		    slug = $2,
		    description = $3,
		    default_region = $4,
		    auto_deploy = $5,
		    updated_at = now()
		WHERE id = $6 AND casdoor_org_id = $7
		RETURNING updated_at
	`

	err := db.QueryRowContext(ctx, query,
		updates.Name,
		updates.Slug,
		updates.Description,
		updates.DefaultRegion,
		updates.AutoDeploy,
		id,
		updates.CasdoorOrgID,
	).Scan(&updates.UpdatedAt)

	return err
}

// DeleteProject deletes a project and all its resources (cascade)
func (db *DB) DeleteProject(ctx context.Context, id uuid.UUID, orgID string) error {
	query := `DELETE FROM projects WHERE id = $1 AND casdoor_org_id = $2`
	
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

// ProjectExists checks if a project exists and belongs to the organization
func (db *DB) ProjectExists(ctx context.Context, id uuid.UUID, orgID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND casdoor_org_id = $2)`
	
	err := db.QueryRowContext(ctx, query, id, orgID).Scan(&exists)
	return exists, err
}

// GenerateSlug generates a URL-friendly slug from a project name
func GenerateSlug(name string) string {
	// Simple slug generation - convert to lowercase and replace spaces with hyphens
	// In production, use a proper slug library
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove special characters (keep only alphanumeric and hyphens)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

