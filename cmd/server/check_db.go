package main

import (
	"database/sql"
	"log"
)

// checkDatabaseTables verifies that required tables exist
func checkDatabaseTables(db *sql.DB) error {
	// Check if projects table exists
	var tableName string
	err := db.QueryRow(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'projects'
	`).Scan(&tableName)
	
	if err == sql.ErrNoRows {
		log.Println("WARNING: Database tables not found. Please run migrations:")
		log.Println("  Using golang-migrate:")
		log.Println("    migrate -path migrations/postgres -database $DATABASE_URL up")
		log.Println("  Or manually run the SQL files in migrations/postgres/")
		return err
	}
	
	if err != nil {
		// Try SQLite check
		var count int
		err2 := db.QueryRow(`
			SELECT COUNT(*) 
			FROM sqlite_master 
			WHERE type='table' 
			AND name='projects'
		`).Scan(&count)
		
		if err2 != nil || count == 0 {
			log.Println("WARNING: Database tables not found. Please run migrations:")
			log.Println("  Using golang-migrate:")
			log.Println("    migrate -path migrations/sqlite -database $DATABASE_URL up")
			return err
		}
	}
	
	log.Println("Database tables verified")
	return nil
}

