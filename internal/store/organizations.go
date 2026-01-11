package store

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Organization represents an organization in the system
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrgMember represents a member of an organization
type OrgMember struct {
	OrgID    string    `json:"org_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// OrgMemberWithUser includes user details
type OrgMemberWithUser struct {
	OrgMember
	Email string `json:"email"`
	Name  string `json:"name"`
}

// CreateOrganization creates a new organization
func (db *DB) CreateOrganization(ctx context.Context, name, ownerID string) (*Organization, error) {
	slug := generateSlug(name)
	
	query := `
		INSERT INTO organizations (name, slug, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, owner_id, created_at, updated_at
	`

	var org Organization
	err := db.QueryRowContext(ctx, query, name, slug, ownerID).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add owner as a member with 'owner' role
	_, err = db.AddOrgMember(ctx, org.ID, ownerID, "owner")
	if err != nil {
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	return &org, nil
}

// GetOrganizationByID retrieves an organization by ID
func (db *DB) GetOrganizationByID(ctx context.Context, id string) (*Organization, error) {
	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	var org Organization
	err := db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// GetOrganizationBySlug retrieves an organization by slug
func (db *DB) GetOrganizationBySlug(ctx context.Context, slug string) (*Organization, error) {
	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`

	var org Organization
	err := db.QueryRowContext(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// ListUserOrganizations lists all organizations a user belongs to
func (db *DB) ListUserOrganizations(ctx context.Context, userID string) ([]*Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.owner_id, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN org_members om ON o.id = om.org_id
		WHERE om.user_id = $1
		ORDER BY o.created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.OwnerID,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, &org)
	}

	return orgs, nil
}

// UpdateOrganization updates an organization
func (db *DB) UpdateOrganization(ctx context.Context, id, name string) (*Organization, error) {
	query := `
		UPDATE organizations
		SET name = COALESCE(NULLIF($2, ''), name),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, name, slug, owner_id, created_at, updated_at
	`

	var org Organization
	err := db.QueryRowContext(ctx, query, id, name).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &org, nil
}

// DeleteOrganization deletes an organization
func (db *DB) DeleteOrganization(ctx context.Context, id string) error {
	query := `DELETE FROM organizations WHERE id = $1`
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

// AddOrgMember adds a user to an organization
func (db *DB) AddOrgMember(ctx context.Context, orgID, userID, role string) (*OrgMember, error) {
	query := `
		INSERT INTO org_members (org_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (org_id, user_id) DO UPDATE SET role = $3
		RETURNING org_id, user_id, role, joined_at
	`

	var member OrgMember
	err := db.QueryRowContext(ctx, query, orgID, userID, role).Scan(
		&member.OrgID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add org member: %w", err)
	}

	return &member, nil
}

// GetOrgMember retrieves a member's role in an organization
func (db *DB) GetOrgMember(ctx context.Context, orgID, userID string) (*OrgMember, error) {
	query := `
		SELECT org_id, user_id, role, joined_at
		FROM org_members
		WHERE org_id = $1 AND user_id = $2
	`

	var member OrgMember
	err := db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&member.OrgID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("member not found")
		}
		return nil, fmt.Errorf("failed to get org member: %w", err)
	}

	return &member, nil
}

// ListOrgMembers lists all members of an organization
func (db *DB) ListOrgMembers(ctx context.Context, orgID string) ([]*OrgMemberWithUser, error) {
	query := `
		SELECT om.org_id, om.user_id, om.role, om.joined_at, u.email, u.name
		FROM org_members om
		INNER JOIN users u ON om.user_id = u.id
		WHERE om.org_id = $1
		ORDER BY om.joined_at ASC
	`

	rows, err := db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list org members: %w", err)
	}
	defer rows.Close()

	var members []*OrgMemberWithUser
	for rows.Next() {
		var member OrgMemberWithUser
		err := rows.Scan(
			&member.OrgID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.Email,
			&member.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan org member: %w", err)
		}
		members = append(members, &member)
	}

	return members, nil
}

// RemoveOrgMember removes a user from an organization
func (db *DB) RemoveOrgMember(ctx context.Context, orgID, userID string) error {
	query := `DELETE FROM org_members WHERE org_id = $1 AND user_id = $2`
	result, err := db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove org member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// IsUserInOrg checks if a user is a member of an organization
func (db *DB) IsUserInOrg(ctx context.Context, orgID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM org_members WHERE org_id = $1 AND user_id = $2)`
	var exists bool
	err := db.QueryRowContext(ctx, query, orgID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check org membership: %w", err)
	}
	return exists, nil
}

// GetUserRoleInOrg gets a user's role in an organization
func (db *DB) GetUserRoleInOrg(ctx context.Context, orgID, userID string) (string, error) {
	query := `SELECT role FROM org_members WHERE org_id = $1 AND user_id = $2`
	var role string
	err := db.QueryRowContext(ctx, query, orgID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("member not found")
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	return role, nil
}

// generateSlug creates a URL-safe slug from a name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
	}
	return slug
}

