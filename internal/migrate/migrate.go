package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed all:../../migrations
var migrationsFS embed.FS

// RunMigrations runs all migration files in the specified directory
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Detect database type from connection string or by testing
	dbType := detectDatabaseType(db)
	
	log.Printf("Detected database type: %s", dbType)
	
	// Get migration files from embedded filesystem
	migrationPath := fmt.Sprintf("migrations/%s", dbType)
	
	// Read all migration files
	files, err := migrationsFS.ReadDir(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}
	
	// Filter and sort migration files
	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	
	sort.Strings(migrationFiles)
	
	log.Printf("Found %d migration files", len(migrationFiles))
	
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// Run each migration
	for _, filename := range migrationFiles {
		migrationName := strings.TrimSuffix(filename, ".up.sql")
		
		// Check if migration already ran
		hasRun, err := hasMigrationRun(db, migrationName)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}
		
		if hasRun {
			log.Printf("Migration %s already applied, skipping", migrationName)
			continue
		}
		
		// Read migration file from embedded filesystem
		migrationFilePath := filepath.Join(migrationPath, filename)
		sql, err := migrationsFS.ReadFile(migrationFilePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}
		
		// Execute migration
		log.Printf("Running migration: %s", migrationName)
		if err := runMigration(db, string(sql)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migrationName, err)
		}
		
		// Record migration
		if err := recordMigration(db, migrationName); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}
		
		log.Printf("Migration %s completed successfully", migrationName)
	}
	
	log.Println("All migrations completed successfully")
	return nil
}

func detectDatabaseType(db *sql.DB) string {
	// Try PostgreSQL
	var version string
	err := db.QueryRow("SELECT version()").Scan(&version)
	if err == nil && strings.Contains(strings.ToLower(version), "postgresql") {
		return "postgres"
	}
	
	// Try SQLite
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master").Scan(&count)
	if err == nil {
		return "sqlite"
	}
	
	// Try MySQL/MariaDB
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err == nil {
		return "mysql"
	}
	
	// Default to postgres
	return "postgres"
}

func createMigrationsTable(db *sql.DB) error {
	// Try PostgreSQL syntax first
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT now()
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		// Try SQLite syntax
		query = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version TEXT PRIMARY KEY,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`
		_, err = db.Exec(query)
		if err != nil {
			// Try MySQL syntax
			query = `
				CREATE TABLE IF NOT EXISTS schema_migrations (
					version VARCHAR(255) PRIMARY KEY,
					applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			_, err = db.Exec(query)
		}
	}
	return err
}

func hasMigrationRun(db *sql.DB, version string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
	if err != nil {
		// Try with ? placeholder for SQLite/MySQL
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	}
	return count > 0, err
}

func recordMigration(db *sql.DB, version string) error {
	_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		// Try with ? placeholder
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
	}
	return err
}

func runMigration(db *sql.DB, sql string) error {
	// Split by semicolon and execute each statement
	statements := strings.Split(sql, ";")
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nStatement: %s", err, stmt)
		}
	}
	
	return nil
}

