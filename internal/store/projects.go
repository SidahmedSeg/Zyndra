package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID                 uuid.UUID
	CasdoorOrgID       string
	Name               string
	Slug               string
	Description        sql.NullString
	OpenStackTenantID  string
	OpenStackNetworkID sql.NullString
	DefaultRegion      sql.NullString
	AutoDeploy         bool
	CreatedBy          sql.NullString
	CreatedAt          time.Time
	UpdatedAt          time.Time
	OrgID              uuid.NullUUID // Custom auth organization ID
	UserID             uuid.NullUUID // Custom auth user ID
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
				default_region, auto_deploy, created_by, org_id, user_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err = db.ExecContext(ctx, query,
			p.ID.String(), p.CasdoorOrgID, p.Name, p.Slug, p.Description,
			p.OpenStackTenantID, p.OpenStackNetworkID,
			p.DefaultRegion, p.AutoDeploy, p.CreatedBy, p.OrgID, p.UserID,
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
			default_region, auto_deploy, created_by, org_id, user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err = db.QueryRowContext(ctx, query,
		p.CasdoorOrgID, p.Name, p.Slug, p.Description,
		p.OpenStackTenantID, p.OpenStackNetworkID,
		p.DefaultRegion, p.AutoDeploy, p.CreatedBy, p.OrgID, p.UserID,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	return err
}

func (db *DB) GetProject(ctx context.Context, id uuid.UUID) (*Project, error) {
	var p Project
	query := `SELECT id, casdoor_org_id, name, slug, description, openstack_tenant_id, openstack_network_id, default_region, auto_deploy, created_by, created_at, updated_at, org_id, user_id FROM projects WHERE id = $1`

	err := db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.CasdoorOrgID, &p.Name, &p.Slug, &p.Description,
		&p.OpenStackTenantID, &p.OpenStackNetworkID,
		&p.DefaultRegion, &p.AutoDeploy, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt, &p.OrgID, &p.UserID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &p, err
}

func (db *DB) ListProjectsByOrg(ctx context.Context, orgID string) ([]*Project, error) {
	query := `SELECT id, casdoor_org_id, name, slug, description, openstack_tenant_id, openstack_network_id, default_region, auto_deploy, created_by, created_at, updated_at, org_id, user_id FROM projects WHERE casdoor_org_id = $1 ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query, orgID)
	if err != nil {
		// Check if it's a "table does not exist" error
		errStr := err.Error()
		if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "relation") && strings.Contains(errStr, "not found") {
			return nil, fmt.Errorf("database table 'projects' does not exist. Please run migrations first: %w", err)
		}
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		err := rows.Scan(
			&p.ID, &p.CasdoorOrgID, &p.Name, &p.Slug, &p.Description,
			&p.OpenStackTenantID, &p.OpenStackNetworkID,
			&p.DefaultRegion, &p.AutoDeploy, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.OrgID, &p.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}
		projects = append(projects, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project rows: %w", err)
	}

	return projects, nil
}

// ListProjectsByOrgID lists projects by the new org_id column (for custom auth)
func (db *DB) ListProjectsByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Project, error) {
	query := `SELECT id, casdoor_org_id, name, slug, description, openstack_tenant_id, openstack_network_id, default_region, auto_deploy, created_by, created_at, updated_at, org_id, user_id FROM projects WHERE org_id = $1 ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		err := rows.Scan(
			&p.ID, &p.CasdoorOrgID, &p.Name, &p.Slug, &p.Description,
			&p.OpenStackTenantID, &p.OpenStackNetworkID,
			&p.DefaultRegion, &p.AutoDeploy, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.OrgID, &p.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}
		projects = append(projects, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project rows: %w", err)
	}

	return projects, nil
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

