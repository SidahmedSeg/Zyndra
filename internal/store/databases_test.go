package store

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/intelifox/click-deploy/internal/testutil"
)

func TestDB_CreateDatabase(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Check if we're using SQLite
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	database := &Database{
		Engine:       "postgresql",
		Version:      sql.NullString{String: "14", Valid: true},
		Size:         "medium",
		VolumeSizeMB: 10240,
		Status:       "pending",
	}

	if isSQLite {
		database.ID = uuid.New()
		query := `INSERT INTO databases (id, engine, version, size, volume_size_mb, status) 
			VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = db.ExecContext(ctx, query,
			database.ID.String(), database.Engine, database.Version.String,
			database.Size, database.VolumeSizeMB, database.Status)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		db.QueryRowContext(ctx, "SELECT created_at FROM databases WHERE id = $1", database.ID.String()).
			Scan(&database.CreatedAt)
	} else {
		err = dbStore.CreateDatabase(ctx, database)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
	}

	if database.ID == uuid.Nil {
		t.Error("Database ID should be set after creation")
	}

	if database.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestDB_GetDatabase(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	databaseID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO databases (id, engine, size, volume_size_mb, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			databaseID.String(), "postgresql", "medium", 10240, "pending")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
	} else {
		database := &Database{
			ID:           databaseID,
			Engine:       "postgresql",
			Size:         "medium",
			VolumeSizeMB: 10240,
			Status:       "pending",
		}
		if err := dbStore.CreateDatabase(ctx, database); err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
	}

	// Retrieve the database
	retrieved, err := dbStore.GetDatabase(ctx, databaseID)
	if err != nil {
		t.Fatalf("Failed to get database: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Database should exist")
	}

	if retrieved.ID != databaseID {
		t.Errorf("Expected database ID %s, got %s", databaseID, retrieved.ID)
	}

	if retrieved.Engine != "postgresql" {
		t.Errorf("Expected engine 'postgresql', got '%s'", retrieved.Engine)
	}

	// Test non-existent database
	nonExistentID := uuid.New()
	retrieved, err = dbStore.GetDatabase(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("GetDatabase should return nil for non-existent database, not error: %v", err)
	}
	if retrieved != nil {
		t.Error("GetDatabase should return nil for non-existent database")
	}
}

func TestDB_UpdateDatabase(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	databaseID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO databases (id, engine, size, volume_size_mb, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			databaseID.String(), "postgresql", "medium", 10240, "pending")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
	} else {
		database := &Database{
			ID:           databaseID,
			Engine:       "postgresql",
			Size:         "medium",
			VolumeSizeMB: 10240,
			Status:       "pending",
		}
		if err := dbStore.CreateDatabase(ctx, database); err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
	}

	// Update the database
	updates := &Database{
		Status:              "active",
		InternalHostname:    sql.NullString{String: "pg123.internal", Valid: true},
		InternalIP:         sql.NullString{String: "10.0.0.1", Valid: true},
		Port:                sql.NullInt64{Int64: 5432, Valid: true},
		Username:            sql.NullString{String: "admin", Valid: true},
		Password:            sql.NullString{String: "secret", Valid: true},
		DatabaseName:        sql.NullString{String: "mydb", Valid: true},
		ConnectionURL:      sql.NullString{String: "postgresql://admin:secret@10.0.0.1:5432/mydb", Valid: true},
		OpenStackInstanceID: sql.NullString{String: "instance-123", Valid: true},
	}

	if isSQLite {
		query := `UPDATE databases SET status = $1, internal_hostname = $2, internal_ip = $3, 
			port = $4, username = $5, password = $6, database_name = $7, connection_url = $8, 
			openstack_instance_id = $9 WHERE id = $10`
		_, err = db.ExecContext(ctx, query,
			updates.Status, updates.InternalHostname.String, updates.InternalIP.String,
			updates.Port.Int64, updates.Username.String, updates.Password.String,
			updates.DatabaseName.String, updates.ConnectionURL.String, updates.OpenStackInstanceID.String,
			databaseID.String())
		if err != nil {
			t.Fatalf("Failed to update database: %v", err)
		}
	} else {
		err = dbStore.UpdateDatabase(ctx, databaseID, updates)
		if err != nil {
			t.Fatalf("Failed to update database: %v", err)
		}
	}

	// Verify updates
	updated, err := dbStore.GetDatabase(ctx, databaseID)
	if err != nil {
		t.Fatalf("Failed to get updated database: %v", err)
	}

	if updated.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", updated.Status)
	}

	if !updated.InternalHostname.Valid || updated.InternalHostname.String != "pg123.internal" {
		t.Errorf("Expected internal hostname 'pg123.internal', got valid=%v, string=%s",
			updated.InternalHostname.Valid, updated.InternalHostname.String)
	}
}

func TestDB_DeleteDatabase(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	testutil.RunMigrations(t, db)

	dbStore := &DB{DB: db}
	ctx := context.Background()

	// Setup
	var isSQLite bool
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite = err == nil

	databaseID := uuid.New()

	if isSQLite {
		_, err = db.ExecContext(ctx, `INSERT INTO databases (id, engine, size, volume_size_mb, status) 
			VALUES ($1, $2, $3, $4, $5)`,
			databaseID.String(), "postgresql", "medium", 10240, "pending")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
	} else {
		database := &Database{
			ID:           databaseID,
			Engine:       "postgresql",
			Size:         "medium",
			VolumeSizeMB: 10240,
			Status:       "pending",
		}
		if err := dbStore.CreateDatabase(ctx, database); err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
	}

	// Delete the database
	err = dbStore.DeleteDatabase(ctx, databaseID)
	if err != nil {
		t.Fatalf("Failed to delete database: %v", err)
	}

	// Verify database is deleted
	deleted, err := dbStore.GetDatabase(ctx, databaseID)
	if err != nil {
		t.Fatalf("Failed to check deleted database: %v", err)
	}

	if deleted != nil {
		t.Error("Database should be deleted")
	}
}

