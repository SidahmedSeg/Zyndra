package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed ../../migrations
var migrationsFS embed.FS

// RunMigrations runs all migration files in the specified directory
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Detect database type from connection string or by testing
	dbType := detectDatabaseType(db)
	
	log.Printf("=== Starting migrations ===")
	log.Printf("Detected database type: %s", dbType)
	
	// Try to get migration files - first from embed, then from filesystem
	var files []fs.DirEntry
	var migrationPath string
	var useEmbed bool
	var err error
	
	// Try embedded filesystem first
	embedPath := fmt.Sprintf("migrations/%s", dbType)
	log.Printf("Trying embedded filesystem path: %s", embedPath)
	files, err = migrationsFS.ReadDir(embedPath)
	if err != nil {
		log.Printf("Embedded filesystem failed: %v", err)
		log.Printf("Falling back to filesystem...")
		
		// Fallback to filesystem (for Docker where migrations are copied)
		fsPath := filepath.Join("/app", "migrations", dbType)
		log.Printf("Trying filesystem path: %s", fsPath)
		files, err = os.ReadDir(fsPath)
		if err != nil {
			// Try relative path
			fsPath = filepath.Join("migrations", dbType)
			log.Printf("Trying relative filesystem path: %s", fsPath)
			files, err = os.ReadDir(fsPath)
		}
		if err != nil {
			log.Printf("❌ Both embedded and filesystem failed")
			log.Printf("Embed error: %v", err)
			return fmt.Errorf("failed to read migrations from both embed and filesystem: %w", err)
		}
		migrationPath = fsPath
		useEmbed = false
		log.Printf("✓ Using filesystem migrations from: %s", migrationPath)
	} else {
		migrationPath = embedPath
		useEmbed = true
		log.Printf("✓ Using embedded migrations from: %s", migrationPath)
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
	if len(migrationFiles) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationPath)
	}
	
	// Create migrations table if it doesn't exist
	log.Printf("Creating schema_migrations table if it doesn't exist...")
	if err := createMigrationsTable(db); err != nil {
		log.Printf("ERROR: Failed to create migrations table: %v", err)
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	log.Printf("Schema_migrations table ready")
	
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
		
		// Read migration file
		migrationFilePath := filepath.Join(migrationPath, filename)
		var sqlBytes []byte
		
		if useEmbed {
			// Read from embedded filesystem
			sqlBytes, err = migrationsFS.ReadFile(migrationFilePath)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s from embed: %w", filename, err)
			}
		} else {
			// Read from filesystem
			sqlBytes, err = os.ReadFile(migrationFilePath)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s from filesystem: %w", filename, err)
			}
		}
		
		// Execute migration
		log.Printf("Running migration: %s (file: %s)", migrationName, filename)
		log.Printf("Migration SQL size: %d bytes", len(sqlBytes))
		if err := runMigration(db, string(sqlBytes)); err != nil {
			log.Printf("ERROR: Failed to run migration %s: %v", migrationName, err)
			return fmt.Errorf("failed to run migration %s: %w", migrationName, err)
		}
		
		// Record migration
		log.Printf("Recording migration %s in schema_migrations...", migrationName)
		if err := recordMigration(db, migrationName); err != nil {
			log.Printf("ERROR: Failed to record migration %s: %v", migrationName, err)
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}
		
		log.Printf("✓ Migration %s completed successfully", migrationName)
	}
	
	log.Println("All migrations completed successfully")
	return nil
}

func detectDatabaseType(db *sql.DB) string {
	log.Printf("Detecting database type...")
	
	// Try PostgreSQL
	var version string
	err := db.QueryRow("SELECT version()").Scan(&version)
	if err == nil {
		log.Printf("PostgreSQL version query succeeded: %s", version[:min(50, len(version))])
		if strings.Contains(strings.ToLower(version), "postgresql") {
			log.Printf("Detected: PostgreSQL")
			return "postgres"
		}
	} else {
		log.Printf("PostgreSQL detection failed: %v", err)
	}
	
	// Try SQLite
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master").Scan(&count)
	if err == nil {
		log.Printf("Detected: SQLite")
		return "sqlite"
	} else {
		log.Printf("SQLite detection failed: %v", err)
	}
	
	// Try MySQL/MariaDB
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err == nil {
		log.Printf("Detected: MySQL/MariaDB")
		return "mysql"
	} else {
		log.Printf("MySQL detection failed: %v", err)
	}
	
	// Default to postgres
	log.Printf("Defaulting to: PostgreSQL")
	return "postgres"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	// For PostgreSQL, we can execute the entire SQL as one transaction
	// This is better than splitting by semicolon which can break on functions/procedures
	
	// Remove comments and empty lines for cleaner execution
	lines := strings.Split(sql, "\n")
	var cleanSQL strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and single-line comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		cleanSQL.WriteString(line)
		cleanSQL.WriteString("\n")
	}
	
	finalSQL := cleanSQL.String()
	if strings.TrimSpace(finalSQL) == "" {
		log.Printf("Warning: Migration SQL is empty after cleaning")
		return nil
	}
	
	// Execute the migration
	log.Printf("Executing migration SQL (%d characters)...", len(finalSQL))
	if _, err := db.Exec(finalSQL); err != nil {
		log.Printf("ERROR: SQL execution failed: %v", err)
		log.Printf("First 500 chars of SQL: %s", finalSQL[:min(500, len(finalSQL))])
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	
	log.Printf("Migration SQL executed successfully")
	return nil
}
