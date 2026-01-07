package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Database struct {
	ID                  uuid.UUID
	ServiceID           sql.NullString // Optional: linked to a service
	Name                string
	Engine              string // postgresql, mysql, redis
	Version             sql.NullString
	Size                string // small, medium, large
	VolumeID            sql.NullString
	VolumeSizeMB        int
	InternalHostname    sql.NullString // e.g., pg7743.internal.armonika.cloud
	InternalIP          sql.NullString
	Port                sql.NullInt64
	Username            sql.NullString
	Password            sql.NullString // encrypted
	DatabaseName        sql.NullString
	ConnectionURL       sql.NullString // Generated connection URL
	OpenStackInstanceID sql.NullString
	OpenStackPortID     sql.NullString
	SecurityGroupID     sql.NullString
	Status              string // pending, provisioning, active, error
	CreatedAt           time.Time
}

// CreateDatabase creates a new database
func (db *DB) CreateDatabase(ctx context.Context, d *Database) error {
	// Generate UUID if not set (for SQLite compatibility)
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	// Check if we're using SQLite (for compatibility)
	var isSQLite bool
	var versionStr string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&versionStr)
	isSQLite = err == nil

	var serviceID interface{}
	if d.ServiceID.Valid {
		serviceID = d.ServiceID.String
	}

	var version sql.NullString
	if d.Version.Valid {
		version = d.Version
	}

	var volumeID interface{}
	if d.VolumeID.Valid {
		volumeID = d.VolumeID.String
	}

	if isSQLite {
		// SQLite: Insert with explicit UUID (no RETURNING support in older versions)
		query := `
			INSERT INTO databases (
				id, service_id, engine, version, size,
				volume_id, volume_size_mb, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = db.ExecContext(ctx, query,
			d.ID.String(), serviceID, d.Engine, version, d.Size,
			volumeID, d.VolumeSizeMB, d.Status,
		)
		if err != nil {
			return err
		}
		// Get timestamp
		err = db.QueryRowContext(ctx, "SELECT created_at FROM databases WHERE id = $1", d.ID.String()).
			Scan(&d.CreatedAt)
		return err
	}

	// PostgreSQL: Use RETURNING clause
	query := `
		INSERT INTO databases (
			service_id, engine, version, size,
			volume_id, volume_size_mb, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err = db.QueryRowContext(ctx, query,
		serviceID,
		d.Engine,
		version,
		d.Size,
		volumeID,
		d.VolumeSizeMB,
		d.Status,
	).Scan(&d.ID, &d.CreatedAt)

	return err
}

// GetDatabase retrieves a database by ID
func (db *DB) GetDatabase(ctx context.Context, id uuid.UUID) (*Database, error) {
	query := `
		SELECT id, service_id, engine, version, size,
		       volume_id, volume_size_mb, internal_hostname, internal_ip, port,
		       username, password, database_name, connection_url,
		       openstack_instance_id, openstack_port_id, security_group_id,
		       status, created_at
		FROM databases
		WHERE id = $1
	`

	var d Database
	var serviceID sql.NullString
	var version sql.NullString
	var volumeID sql.NullString
	var internalHostname sql.NullString
	var internalIP sql.NullString
	var port sql.NullInt64
	var username sql.NullString
	var password sql.NullString
	var databaseName sql.NullString
	var connectionURL sql.NullString
	var openstackInstanceID sql.NullString
	var openstackPortID sql.NullString
	var securityGroupID sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&d.ID,
		&serviceID,
		&d.Engine,
		&version,
		&d.Size,
		&volumeID,
		&d.VolumeSizeMB,
		&internalHostname,
		&internalIP,
		&port,
		&username,
		&password,
		&databaseName,
		&connectionURL,
		&openstackInstanceID,
		&openstackPortID,
		&securityGroupID,
		&d.Status,
		&d.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	d.ServiceID = serviceID
	d.Version = version
	d.VolumeID = volumeID
	d.InternalHostname = internalHostname
	d.InternalIP = internalIP
	d.Port = port
	d.Username = username
	d.Password = password
	d.DatabaseName = databaseName
	d.ConnectionURL = connectionURL
	d.OpenStackInstanceID = openstackInstanceID
	d.OpenStackPortID = openstackPortID
	d.SecurityGroupID = securityGroupID

	return &d, nil
}

// ListDatabasesByService lists databases for a service
func (db *DB) ListDatabasesByService(ctx context.Context, serviceID uuid.UUID) ([]*Database, error) {
	query := `
		SELECT id, service_id, engine, version, size,
		       volume_id, volume_size_mb, internal_hostname, internal_ip, port,
		       username, password, database_name, connection_url,
		       openstack_instance_id, openstack_port_id, security_group_id,
		       status, created_at
		FROM databases
		WHERE service_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []*Database
	for rows.Next() {
		var d Database
		var serviceID sql.NullString
		var version sql.NullString
		var volumeID sql.NullString
		var internalHostname sql.NullString
		var internalIP sql.NullString
		var port sql.NullInt64
		var username sql.NullString
		var password sql.NullString
		var databaseName sql.NullString
		var connectionURL sql.NullString
		var openstackInstanceID sql.NullString
		var openstackPortID sql.NullString
		var securityGroupID sql.NullString

		err := rows.Scan(
			&d.ID,
			&serviceID,
			&d.Engine,
			&version,
			&d.Size,
			&volumeID,
			&d.VolumeSizeMB,
			&internalHostname,
			&internalIP,
			&port,
			&username,
			&password,
			&databaseName,
			&connectionURL,
			&openstackInstanceID,
			&openstackPortID,
			&securityGroupID,
			&d.Status,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		d.ServiceID = serviceID
		d.Version = version
		d.VolumeID = volumeID
		d.InternalHostname = internalHostname
		d.InternalIP = internalIP
		d.Port = port
		d.Username = username
		d.Password = password
		d.DatabaseName = databaseName
		d.ConnectionURL = connectionURL
		d.OpenStackInstanceID = openstackInstanceID
		d.OpenStackPortID = openstackPortID
		d.SecurityGroupID = securityGroupID

		databases = append(databases, &d)
	}

	return databases, rows.Err()
}

// ListDatabasesByProject lists databases for a project (via services)
func (db *DB) ListDatabasesByProject(ctx context.Context, projectID uuid.UUID) ([]*Database, error) {
	query := `
		SELECT d.id, d.service_id, d.engine, d.version, d.size,
		       d.volume_id, d.volume_size_mb, d.internal_hostname, d.internal_ip, d.port,
		       d.username, d.password, d.database_name, d.connection_url,
		       d.openstack_instance_id, d.openstack_port_id, d.security_group_id,
		       d.status, d.created_at
		FROM databases d
		JOIN services s ON d.service_id = s.id
		WHERE s.project_id = $1
		ORDER BY d.created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []*Database
	for rows.Next() {
		var d Database
		var serviceID sql.NullString
		var version sql.NullString
		var volumeID sql.NullString
		var internalHostname sql.NullString
		var internalIP sql.NullString
		var port sql.NullInt64
		var username sql.NullString
		var password sql.NullString
		var databaseName sql.NullString
		var connectionURL sql.NullString
		var openstackInstanceID sql.NullString
		var openstackPortID sql.NullString
		var securityGroupID sql.NullString

		err := rows.Scan(
			&d.ID,
			&serviceID,
			&d.Engine,
			&version,
			&d.Size,
			&volumeID,
			&d.VolumeSizeMB,
			&internalHostname,
			&internalIP,
			&port,
			&username,
			&password,
			&databaseName,
			&connectionURL,
			&openstackInstanceID,
			&openstackPortID,
			&securityGroupID,
			&d.Status,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		d.ServiceID = serviceID
		d.Version = version
		d.VolumeID = volumeID
		d.InternalHostname = internalHostname
		d.InternalIP = internalIP
		d.Port = port
		d.Username = username
		d.Password = password
		d.DatabaseName = databaseName
		d.ConnectionURL = connectionURL
		d.OpenStackInstanceID = openstackInstanceID
		d.OpenStackPortID = openstackPortID
		d.SecurityGroupID = securityGroupID

		databases = append(databases, &d)
	}

	return databases, rows.Err()
}

// UpdateDatabase updates a database
func (db *DB) UpdateDatabase(ctx context.Context, id uuid.UUID, d *Database) error {
	query := `
		UPDATE databases
		SET internal_hostname = $1, internal_ip = $2, port = $3,
		    username = $4, password = $5, database_name = $6,
		    connection_url = $7, openstack_instance_id = $8,
		    openstack_port_id = $9, security_group_id = $10,
		    status = $11
		WHERE id = $12
	`

	var internalHostname interface{}
	if d.InternalHostname.Valid {
		internalHostname = d.InternalHostname.String
	}

	var internalIP interface{}
	if d.InternalIP.Valid {
		internalIP = d.InternalIP.String
	}

	var port interface{}
	if d.Port.Valid {
		port = d.Port.Int64
	}

	var username interface{}
	if d.Username.Valid {
		username = d.Username.String
	}

	var password interface{}
	if d.Password.Valid {
		password = d.Password.String
	}

	var databaseName interface{}
	if d.DatabaseName.Valid {
		databaseName = d.DatabaseName.String
	}

	var connectionURL interface{}
	if d.ConnectionURL.Valid {
		connectionURL = d.ConnectionURL.String
	}

	var openstackInstanceID interface{}
	if d.OpenStackInstanceID.Valid {
		openstackInstanceID = d.OpenStackInstanceID.String
	}

	var openstackPortID interface{}
	if d.OpenStackPortID.Valid {
		openstackPortID = d.OpenStackPortID.String
	}

	var securityGroupID interface{}
	if d.SecurityGroupID.Valid {
		securityGroupID = d.SecurityGroupID.String
	}

	_, err := db.ExecContext(ctx, query,
		internalHostname,
		internalIP,
		port,
		username,
		password,
		databaseName,
		connectionURL,
		openstackInstanceID,
		openstackPortID,
		securityGroupID,
		d.Status,
		id,
	)

	return err
}

// DeleteDatabase deletes a database
func (db *DB) DeleteDatabase(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM databases WHERE id = $1`

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

// GetDatabaseCredentials retrieves database credentials (for API)
func (db *DB) GetDatabaseCredentials(ctx context.Context, id uuid.UUID) (*DatabaseCredentials, error) {
	query := `
		SELECT id, engine, internal_hostname, port,
		       username, password, database_name, connection_url
		FROM databases
		WHERE id = $1
	`

	var creds DatabaseCredentials
	var internalHostname sql.NullString
	var port sql.NullInt64
	var username sql.NullString
	var password sql.NullString
	var databaseName sql.NullString
	var connectionURL sql.NullString

	err := db.QueryRowContext(ctx, query, id).Scan(
		&creds.ID,
		&creds.Engine,
		&internalHostname,
		&port,
		&username,
		&password,
		&databaseName,
		&connectionURL,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	creds.Hostname = internalHostname.String
	if port.Valid {
		creds.Port = int(port.Int64)
	}
	creds.Username = username.String
	creds.Password = password.String // TODO: Decrypt
	creds.Database = databaseName.String
	creds.ConnectionURL = connectionURL.String

	return &creds, nil
}

type DatabaseCredentials struct {
	ID            uuid.UUID
	Engine        string
	Hostname      string
	Port          int
	Username      string
	Password      string
	Database      string
	ConnectionURL string
}

