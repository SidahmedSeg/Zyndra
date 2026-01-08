package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations runs all migration files in the specified directory
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Detect database type from connection string or by testing
	dbType := detectDatabaseType(db)
	
	log.Printf("=== Starting migrations ===")
	log.Printf("Detected database type: %s", dbType)
	
	// Read migration files from filesystem
	var files []os.DirEntry
	var migrationPath string
	var err error
	
	// Try filesystem paths (Docker uses /app/migrations, local dev uses ./migrations)
	fsPaths := []string{
		filepath.Join("/app", "migrations", dbType), // Docker path
		filepath.Join("migrations", dbType),        // Local dev path
		filepath.Join(migrationsDir, dbType),       // Explicit migrationsDir parameter
	}
	
	for _, fsPath := range fsPaths {
		log.Printf("Trying filesystem path: %s", fsPath)
		files, err = os.ReadDir(fsPath)
		if err == nil {
			migrationPath = fsPath
			log.Printf("✓ Using filesystem migrations from: %s", migrationPath)
			break
		}
		log.Printf("  Failed: %v", err)
	}
	
	if err != nil {
		log.Printf("❌ Failed to find migrations in any location")
		return fmt.Errorf("failed to read migrations from filesystem: %w", err)
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
		
		// Read migration file from filesystem
		migrationFilePath := filepath.Join(migrationPath, filename)
		sqlBytes, err := os.ReadFile(migrationFilePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
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

func runMigration(db *sql.DB, sqlContent string) error {
	// For PostgreSQL, we need to handle statements carefully
	// Some statements like CREATE EXTENSION can't be in a transaction
	// We'll execute statements one by one, but try to be smart about it
	
	log.Printf("Preparing to execute migration SQL (%d characters)...", len(sqlContent))
	
	// Split by semicolon, but be careful not to split inside strings or functions
	statements := splitSQLStatements(sqlContent)
	
	log.Printf("Split into %d statements", len(statements))
	
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		
		// Skip comments
		if strings.HasPrefix(stmt, "--") {
			continue
		}
		
		log.Printf("Executing statement %d/%d (%d chars): %s...", 
			i+1, len(statements), len(stmt), 
			strings.ReplaceAll(stmt[:min(100, len(stmt))], "\n", " "))
		
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("ERROR: SQL execution failed on statement %d: %v", i+1, err)
			log.Printf("Failed statement: %s", stmt[:min(500, len(stmt))])
			return fmt.Errorf("failed to execute statement %d: %w\nStatement: %s", i+1, err, stmt[:min(200, len(stmt))])
		}
		
		log.Printf("✓ Statement %d executed successfully", i+1)
	}
	
	log.Printf("All statements executed successfully")
	return nil
}

func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := byte(0)
	
	lines := strings.Split(sql, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		
		// Process line character by character
		for i := 0; i < len(line); i++ {
			char := line[i]
			
			// Track string literals
			if (char == '\'' || char == '"') && (i == 0 || line[i-1] != '\\') {
				if !inString {
					inString = true
					stringChar = char
				} else if char == stringChar {
					inString = false
					stringChar = 0
				}
				current.WriteByte(char)
				continue
			}
			
			// If we find a semicolon outside a string, it's a statement separator
			if char == ';' && !inString {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
				continue
			}
			
			current.WriteByte(char)
		}
		
		// Add newline for proper formatting
		current.WriteByte('\n')
	}
	
	// Add final statement if any
	final := strings.TrimSpace(current.String())
	if final != "" {
		statements = append(statements, final)
	}
	
	return statements
}
