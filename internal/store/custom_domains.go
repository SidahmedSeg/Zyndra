package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CustomDomain struct {
	ID              uuid.UUID
	ServiceID       uuid.UUID
	Domain          string
	Status          string // pending, verified, active, error
	CNAME           sql.NullString
	CNAMETarget     sql.NullString
	SSLEnabled      bool
	SSLCertStatus   sql.NullString
	SSLCertExpiry   sql.NullTime
	ValidationToken sql.NullString
	CreatedAt       time.Time
	UpdatedAt       time.Time
	VerifiedAt      sql.NullTime
}

// CreateCustomDomain creates a new custom domain record
func (db *DB) CreateCustomDomain(ctx context.Context, d *CustomDomain) error {
	// Generate UUID if not set (for SQLite compatibility)
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	var cnameTarget sql.NullString
	if d.CNAMETarget.Valid {
		cnameTarget = d.CNAMETarget
	}

	var validationToken sql.NullString
	if d.ValidationToken.Valid {
		validationToken = d.ValidationToken
	}

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO custom_domains (
				id, service_id, domain, status, cname_target,
				ssl_enabled, validation_token
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		sslEnabled := 0
		if d.SSLEnabled {
			sslEnabled = 1
		}
		_, err = db.ExecContext(ctx, query,
			d.ID.String(), d.ServiceID.String(), d.Domain, d.Status,
			cnameTarget, sslEnabled, validationToken,
		)
		if err != nil {
			return err
		}
		// Get timestamps
		err = db.QueryRowContext(ctx, "SELECT created_at, updated_at FROM custom_domains WHERE id = $1", d.ID.String()).
			Scan(&d.CreatedAt, &d.UpdatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO custom_domains (
			service_id, domain, status, cname_target,
			ssl_enabled, validation_token
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err = db.QueryRowContext(ctx, query,
		d.ServiceID,
		d.Domain,
		d.Status,
		cnameTarget,
		d.SSLEnabled,
		validationToken,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)

	return err
}

// GetCustomDomain retrieves a custom domain by ID
func (db *DB) GetCustomDomain(ctx context.Context, id uuid.UUID) (*CustomDomain, error) {
	var d CustomDomain
	query := `
		SELECT id, service_id, domain, status, cname, cname_target,
		       ssl_enabled, ssl_cert_status, ssl_cert_expiry,
		       validation_token, created_at, updated_at, verified_at
		FROM custom_domains
		WHERE id = $1
	`

	var cname sql.NullString
	var cnameTarget sql.NullString
	var sslCertStatus sql.NullString
	var sslCertExpiry sql.NullTime
	var validationToken sql.NullString
	var verifiedAt sql.NullTime

	err := db.QueryRowContext(ctx, query, id).Scan(
		&d.ID,
		&d.ServiceID,
		&d.Domain,
		&d.Status,
		&cname,
		&cnameTarget,
		&d.SSLEnabled,
		&sslCertStatus,
		&sslCertExpiry,
		&validationToken,
		&d.CreatedAt,
		&d.UpdatedAt,
		&verifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	d.CNAME = cname
	d.CNAMETarget = cnameTarget
	d.SSLCertStatus = sslCertStatus
	d.SSLCertExpiry = sslCertExpiry
	d.ValidationToken = validationToken
	d.VerifiedAt = verifiedAt

	return &d, nil
}

// ListCustomDomainsByService lists custom domains for a service
func (db *DB) ListCustomDomainsByService(ctx context.Context, serviceID uuid.UUID) ([]*CustomDomain, error) {
	query := `
		SELECT id, service_id, domain, status, cname, cname_target,
		       ssl_enabled, ssl_cert_status, ssl_cert_expiry,
		       validation_token, created_at, updated_at, verified_at
		FROM custom_domains
		WHERE service_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*CustomDomain
	for rows.Next() {
		var d CustomDomain
		var cname sql.NullString
		var cnameTarget sql.NullString
		var sslCertStatus sql.NullString
		var sslCertExpiry sql.NullTime
		var validationToken sql.NullString
		var verifiedAt sql.NullTime

		err := rows.Scan(
			&d.ID,
			&d.ServiceID,
			&d.Domain,
			&d.Status,
			&cname,
			&cnameTarget,
			&d.SSLEnabled,
			&sslCertStatus,
			&sslCertExpiry,
			&validationToken,
			&d.CreatedAt,
			&d.UpdatedAt,
			&verifiedAt,
		)
		if err != nil {
			return nil, err
		}

		d.CNAME = cname
		d.CNAMETarget = cnameTarget
		d.SSLCertStatus = sslCertStatus
		d.SSLCertExpiry = sslCertExpiry
		d.ValidationToken = validationToken
		d.VerifiedAt = verifiedAt

		domains = append(domains, &d)
	}

	return domains, rows.Err()
}

// UpdateCustomDomain updates a custom domain
func (db *DB) UpdateCustomDomain(ctx context.Context, id uuid.UUID, updates *CustomDomain) error {
	query := `
		UPDATE custom_domains
		SET status = $1,
		    cname = $2,
		    cname_target = $3,
		    ssl_cert_status = $4,
		    ssl_cert_expiry = $5,
		    verified_at = $6,
		    updated_at = now()
		WHERE id = $7
		RETURNING updated_at
	`

	var cname sql.NullString
	if updates.CNAME.Valid {
		cname = updates.CNAME
	}

	var cnameTarget sql.NullString
	if updates.CNAMETarget.Valid {
		cnameTarget = updates.CNAMETarget
	}

	var sslCertStatus sql.NullString
	if updates.SSLCertStatus.Valid {
		sslCertStatus = updates.SSLCertStatus
	}

	var sslCertExpiry sql.NullTime
	if updates.SSLCertExpiry.Valid {
		sslCertExpiry = updates.SSLCertExpiry
	}

	var verifiedAt sql.NullTime
	if updates.VerifiedAt.Valid {
		verifiedAt = updates.VerifiedAt
	}

	err := db.QueryRowContext(ctx, query,
		updates.Status,
		cname,
		cnameTarget,
		sslCertStatus,
		sslCertExpiry,
		verifiedAt,
		id,
	).Scan(&updates.UpdatedAt)

	return err
}

// DeleteCustomDomain deletes a custom domain
func (db *DB) DeleteCustomDomain(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM custom_domains WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

