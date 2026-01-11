package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Volume struct {
	ID                  uuid.UUID
	ProjectID           uuid.UUID
	Name                string
	SizeMB              int
	MountPath           sql.NullString
	AttachedToServiceID sql.NullString
	AttachedToDatabaseID sql.NullString
	OpenStackVolumeID   sql.NullString
	OpenStackAttachmentID sql.NullString
	Status              string // pending, available, attached, error
	VolumeType          string // user, database_auto
	CreatedAt           time.Time
}

// CreateVolume creates a new volume
func (db *DB) CreateVolume(ctx context.Context, v *Volume) error {
	// Generate UUID if not set (for SQLite compatibility)
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	var mountPath interface{}
	if v.MountPath.Valid {
		mountPath = v.MountPath.String
	}

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO volumes (
				id, project_id, name, size_mb, mount_path, volume_type, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = db.ExecContext(ctx, query,
			v.ID.String(), v.ProjectID.String(), v.Name, v.SizeMB,
			mountPath, v.VolumeType, v.Status,
		)
		if err != nil {
			return err
		}
		// Get timestamp
		err = db.QueryRowContext(ctx, "SELECT created_at FROM volumes WHERE id = $1", v.ID.String()).
			Scan(&v.CreatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO volumes (
			project_id, name, size_mb, mount_path, volume_type, status
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err = db.QueryRowContext(ctx, query,
		v.ProjectID,
		v.Name,
		v.SizeMB,
		mountPath,
		v.VolumeType,
		v.Status,
	).Scan(&v.ID, &v.CreatedAt)

	return err
}

// GetVolume retrieves a volume by ID
func (db *DB) GetVolume(ctx context.Context, id uuid.UUID) (*Volume, error) {
	query := `
		SELECT id, project_id, name, size_mb, mount_path,
		       attached_to_service_id, attached_to_database_id,
		       openstack_volume_id, openstack_attachment_id,
		       status, volume_type, created_at
		FROM volumes
		WHERE id = $1
	`

	var v Volume
	var mountPath sql.NullString
	var attachedToServiceID sql.NullString
	var attachedToDatabaseID sql.NullString
	var openstackVolumeID sql.NullString
	var openstackAttachmentID sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&v.ID,
		&v.ProjectID,
		&v.Name,
		&v.SizeMB,
		&mountPath,
		&attachedToServiceID,
		&attachedToDatabaseID,
		&openstackVolumeID,
		&openstackAttachmentID,
		&v.Status,
		&v.VolumeType,
		&v.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	v.MountPath = mountPath
	v.AttachedToServiceID = attachedToServiceID
	v.AttachedToDatabaseID = attachedToDatabaseID
	v.OpenStackVolumeID = openstackVolumeID
	v.OpenStackAttachmentID = openstackAttachmentID

	return &v, nil
}

// ListVolumesByProject lists volumes for a project
func (db *DB) ListVolumesByProject(ctx context.Context, projectID uuid.UUID) ([]*Volume, error) {
	query := `
		SELECT id, project_id, name, size_mb, mount_path,
		       attached_to_service_id, attached_to_database_id,
		       openstack_volume_id, openstack_attachment_id,
		       status, volume_type, created_at
		FROM volumes
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []*Volume
	for rows.Next() {
		var v Volume
		var mountPath sql.NullString
		var attachedToServiceID sql.NullString
		var attachedToDatabaseID sql.NullString
		var openstackVolumeID sql.NullString
		var openstackAttachmentID sql.NullString

		err := rows.Scan(
			&v.ID,
			&v.ProjectID,
			&v.Name,
			&v.SizeMB,
			&mountPath,
			&attachedToServiceID,
			&attachedToDatabaseID,
			&openstackVolumeID,
			&openstackAttachmentID,
			&v.Status,
			&v.VolumeType,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		v.MountPath = mountPath
		v.AttachedToServiceID = attachedToServiceID
		v.AttachedToDatabaseID = attachedToDatabaseID
		v.OpenStackVolumeID = openstackVolumeID
		v.OpenStackAttachmentID = openstackAttachmentID

		volumes = append(volumes, &v)
	}

	return volumes, rows.Err()
}

// UpdateVolume updates a volume
func (db *DB) UpdateVolume(ctx context.Context, id uuid.UUID, v *Volume) error {
	query := `
		UPDATE volumes
		SET mount_path = $1, attached_to_service_id = $2, attached_to_database_id = $3,
		    openstack_volume_id = $4, openstack_attachment_id = $5, status = $6
		WHERE id = $7
	`

	var mountPath interface{}
	if v.MountPath.Valid {
		mountPath = v.MountPath.String
	}

	var attachedToServiceID interface{}
	if v.AttachedToServiceID.Valid {
		attachedToServiceID = v.AttachedToServiceID.String
	}

	var attachedToDatabaseID interface{}
	if v.AttachedToDatabaseID.Valid {
		attachedToDatabaseID = v.AttachedToDatabaseID.String
	}

	var openstackVolumeID interface{}
	if v.OpenStackVolumeID.Valid {
		openstackVolumeID = v.OpenStackVolumeID.String
	}

	var openstackAttachmentID interface{}
	if v.OpenStackAttachmentID.Valid {
		openstackAttachmentID = v.OpenStackAttachmentID.String
	}

	_, err := db.ExecContext(ctx, query,
		mountPath,
		attachedToServiceID,
		attachedToDatabaseID,
		openstackVolumeID,
		openstackAttachmentID,
		v.Status,
		id,
	)

	return err
}

// AttachVolumeToService attaches a volume to a service
func (db *DB) AttachVolumeToService(ctx context.Context, volumeID uuid.UUID, serviceID uuid.UUID, mountPath string) error {
	query := `
		UPDATE volumes
		SET attached_to_service_id = $1, mount_path = $2, status = 'attached'
		WHERE id = $3
	`

	_, err := db.ExecContext(ctx, query, serviceID.String(), mountPath, volumeID)
	return err
}

// DetachVolumeFromService detaches a volume from a service
func (db *DB) DetachVolumeFromService(ctx context.Context, volumeID uuid.UUID) error {
	query := `
		UPDATE volumes
		SET attached_to_service_id = NULL, mount_path = NULL, status = 'available'
		WHERE id = $1
	`

	_, err := db.ExecContext(ctx, query, volumeID)
	return err
}

// DeleteVolume deletes a volume
func (db *DB) DeleteVolume(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM volumes WHERE id = $1`

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

// UpdateVolumeStatus updates just the status field of a volume
func (db *DB) UpdateVolumeStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE volumes SET status = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, status, id)
	return err
}

// UpdateVolumeSize updates the size of a volume
func (db *DB) UpdateVolumeSize(ctx context.Context, id uuid.UUID, sizeMB int) error {
	query := `UPDATE volumes SET size_mb = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, sizeMB, id)
	return err
}

// AttachVolume attaches a volume to a service or database
func (db *DB) AttachVolume(ctx context.Context, volumeID uuid.UUID, serviceID *uuid.UUID, databaseID *uuid.UUID) error {
	query := `
		UPDATE volumes
		SET attached_to_service_id = $1, attached_to_database_id = $2, status = 'attached'
		WHERE id = $3
	`

	var svcID, dbID interface{}
	if serviceID != nil {
		svcID = serviceID.String()
	}
	if databaseID != nil {
		dbID = databaseID.String()
	}

	_, err := db.ExecContext(ctx, query, svcID, dbID, volumeID)
	return err
}

// DetachVolume detaches a volume from any service or database
func (db *DB) DetachVolume(ctx context.Context, volumeID uuid.UUID) error {
	query := `
		UPDATE volumes
		SET attached_to_service_id = NULL, attached_to_database_id = NULL, status = 'available'
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, volumeID)
	return err
}

