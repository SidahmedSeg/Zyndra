package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates a test database connection
// Uses SQLite for fast tests, or PostgreSQL if TEST_DATABASE_URL is set
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	var db *sql.DB
	var err error
	var cleanup func()

	// Check if we should use PostgreSQL (for integration tests)
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL != "" {
		db, err = sql.Open("pgx", testDBURL)
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}

		// Create a unique test database name
		testDBName := fmt.Sprintf("test_%d", os.Getpid())
		
		// Create test database (requires connection to postgres database)
		postgresDB, err := sql.Open("pgx", testDBURL)
		if err != nil {
			t.Fatalf("Failed to connect to postgres: %v", err)
		}
		defer postgresDB.Close()

		_, err = postgresDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
		if err != nil {
			t.Logf("Warning: Could not drop test database: %v", err)
		}

		_, err = postgresDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		// Close initial connection and reconnect to test database
		db.Close()
		testDBURL = fmt.Sprintf("%s dbname=%s", testDBURL, testDBName)
		db, err = sql.Open("pgx", testDBURL)
		if err != nil {
			t.Fatalf("Failed to connect to test database: %v", err)
		}

		cleanup = func() {
			db.Close()
			postgresDB, _ := sql.Open("pgx", os.Getenv("TEST_DATABASE_URL"))
			if postgresDB != nil {
				postgresDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
				postgresDB.Close()
			}
		}
	} else {
		// Use in-memory SQLite for fast unit tests
		db, err = sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open SQLite test database: %v", err)
		}

		cleanup = func() {
			db.Close()
		}
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db, cleanup
}

// RunMigrations runs database migrations on the test database
// For SQLite, we use simplified schema. For PostgreSQL, use actual migrations.
func RunMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	// Check if we're using SQLite (for fast unit tests)
	var version string
	err := db.QueryRow("SELECT sqlite_version()").Scan(&version)
	isSQLite := err == nil

	if isSQLite {
		// SQLite-compatible schema (simplified)
		migrations := []string{
			// Projects table
			`CREATE TABLE IF NOT EXISTS projects (
				id TEXT PRIMARY KEY,
				casdoor_org_id TEXT NOT NULL,
				name TEXT NOT NULL,
				slug TEXT NOT NULL,
				description TEXT,
				openstack_tenant_id TEXT NOT NULL,
				openstack_network_id TEXT,
				default_region TEXT,
				auto_deploy INTEGER DEFAULT 1,
				created_by TEXT,
				created_at DATETIME DEFAULT (datetime('now')),
				updated_at DATETIME DEFAULT (datetime('now')),
				UNIQUE(casdoor_org_id, slug)
			)`,
			// Services table
			`CREATE TABLE IF NOT EXISTS services (
				id TEXT PRIMARY KEY,
				project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
				git_source_id TEXT,
				name TEXT NOT NULL,
				type TEXT NOT NULL DEFAULT 'app',
				status TEXT DEFAULT 'pending',
				instance_size TEXT DEFAULT 'medium',
				port INTEGER DEFAULT 8080,
				openstack_instance_id TEXT,
				openstack_fip_id TEXT,
				openstack_fip_address TEXT,
				security_group_id TEXT,
				subdomain TEXT UNIQUE,
				generated_url TEXT,
				current_image_tag TEXT,
				canvas_x INTEGER DEFAULT 0,
				canvas_y INTEGER DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			// Jobs table
			`CREATE TABLE IF NOT EXISTS jobs (
				id TEXT PRIMARY KEY,
				type TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'pending',
				payload TEXT,
				locked_by TEXT,
				locked_until DATETIME,
				started_at DATETIME,
				completed_at DATETIME,
				attempts INTEGER DEFAULT 0,
				max_attempts INTEGER DEFAULT 3,
				error_message TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			// Databases table
			`CREATE TABLE IF NOT EXISTS databases (
				id TEXT PRIMARY KEY,
				service_id TEXT,
				name TEXT,
				engine TEXT NOT NULL,
				version TEXT,
				size TEXT NOT NULL,
				volume_id TEXT,
				volume_size_mb INTEGER DEFAULT 0,
				internal_hostname TEXT,
				internal_ip TEXT,
				port INTEGER,
				username TEXT,
				password TEXT,
				database_name TEXT,
				connection_url TEXT,
				openstack_instance_id TEXT,
				openstack_port_id TEXT,
				security_group_id TEXT,
				status TEXT DEFAULT 'pending',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			// Volumes table
			`CREATE TABLE IF NOT EXISTS volumes (
				id TEXT PRIMARY KEY,
				project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
				name TEXT NOT NULL,
				size_mb INTEGER NOT NULL,
				mount_path TEXT,
				attached_to_service_id TEXT,
				attached_to_database_id TEXT,
				openstack_volume_id TEXT,
				openstack_attachment_id TEXT,
				status TEXT DEFAULT 'pending',
				volume_type TEXT DEFAULT 'user',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			// Git connections table
			`CREATE TABLE IF NOT EXISTS git_connections (
				id TEXT PRIMARY KEY,
				casdoor_org_id TEXT NOT NULL,
				provider TEXT NOT NULL,
				access_token TEXT NOT NULL,
				refresh_token TEXT,
				token_expires_at DATETIME,
				account_name TEXT,
				account_id TEXT,
				connected_by TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			// Git sources table
			`CREATE TABLE IF NOT EXISTS git_sources (
				id TEXT PRIMARY KEY,
				service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
				git_connection_id TEXT NOT NULL REFERENCES git_connections(id) ON DELETE CASCADE,
				provider TEXT NOT NULL,
				repo_owner TEXT NOT NULL,
				repo_name TEXT NOT NULL,
				branch TEXT NOT NULL,
				root_dir TEXT,
				webhook_id TEXT,
				webhook_secret TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			// Deployments table
			`CREATE TABLE IF NOT EXISTS deployments (
				id TEXT PRIMARY KEY,
				service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
				commit_sha TEXT,
				commit_message TEXT,
				commit_author TEXT,
				status TEXT NOT NULL DEFAULT 'queued',
				image_tag TEXT,
				build_duration INTEGER,
				deploy_duration INTEGER,
				error_message TEXT,
				triggered_by TEXT NOT NULL DEFAULT 'manual',
				started_at DATETIME,
				finished_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			// Custom domains table
			`CREATE TABLE IF NOT EXISTS custom_domains (
				id TEXT PRIMARY KEY,
				service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
				domain TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'pending',
				cname TEXT,
				cname_target TEXT,
				ssl_enabled INTEGER DEFAULT 0,
				ssl_cert_status TEXT,
				ssl_cert_expiry DATETIME,
				validation_token TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				verified_at DATETIME
			)`,
			// Environment variables table
			`CREATE TABLE IF NOT EXISTS env_vars (
				id TEXT PRIMARY KEY,
				service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
				key TEXT NOT NULL,
				value TEXT,
				is_secret INTEGER DEFAULT 0,
				linked_database_id TEXT,
				link_type TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(service_id, key)
			)`,
		}

		for _, migration := range migrations {
			_, err := db.Exec(migration)
			if err != nil {
				t.Fatalf("Failed to run migration: %v\nSQL: %s", err, migration)
			}
		}
	} else {
		// PostgreSQL - use actual migration file
		// For now, create minimal schema. In production, use golang-migrate
		migrations := []string{
			`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
			// Projects table
			`CREATE TABLE IF NOT EXISTS projects (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				casdoor_org_id VARCHAR(255) NOT NULL,
				name VARCHAR(255) NOT NULL,
				slug VARCHAR(100) NOT NULL,
				description TEXT,
				openstack_tenant_id VARCHAR(255) NOT NULL,
				openstack_network_id VARCHAR(255),
				default_region VARCHAR(100),
				auto_deploy BOOLEAN DEFAULT true,
				created_by VARCHAR(255),
				created_at TIMESTAMPTZ DEFAULT now(),
				updated_at TIMESTAMPTZ DEFAULT now(),
				UNIQUE(casdoor_org_id, slug)
			)`,
			// Services table
			`CREATE TABLE IF NOT EXISTS services (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
				git_source_id UUID,
				name VARCHAR(255) NOT NULL,
				type VARCHAR(50) NOT NULL DEFAULT 'app',
				status VARCHAR(50) DEFAULT 'pending',
				instance_size VARCHAR(50) DEFAULT 'medium',
				port INT DEFAULT 8080,
				openstack_instance_id VARCHAR(255),
				openstack_fip_id VARCHAR(255),
				openstack_fip_address INET,
				security_group_id VARCHAR(255),
				subdomain VARCHAR(100) UNIQUE,
				generated_url TEXT,
				current_image_tag VARCHAR(255),
				canvas_x INT DEFAULT 0,
				canvas_y INT DEFAULT 0,
				created_at TIMESTAMPTZ DEFAULT now(),
				updated_at TIMESTAMPTZ DEFAULT now()
			)`,
			// Jobs table
			`CREATE TABLE IF NOT EXISTS jobs (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				type VARCHAR(50) NOT NULL,
				status VARCHAR(50) NOT NULL DEFAULT 'pending',
				payload JSONB,
				locked_by VARCHAR(255),
				locked_until TIMESTAMPTZ,
				started_at TIMESTAMPTZ,
				completed_at TIMESTAMPTZ,
				attempts INTEGER DEFAULT 0,
				max_attempts INTEGER DEFAULT 3,
				error_message TEXT,
				created_at TIMESTAMPTZ DEFAULT now(),
				updated_at TIMESTAMPTZ DEFAULT now()
			)`,
		}

		for _, migration := range migrations {
			_, err := db.Exec(migration)
			if err != nil {
				t.Fatalf("Failed to run migration: %v\nSQL: %s", err, migration)
			}
		}
	}
}

// TruncateTables truncates all tables in the test database
func TruncateTables(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{"volumes", "databases", "services", "projects", "jobs"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Could not truncate table %s: %v", table, err)
		}
	}
}
