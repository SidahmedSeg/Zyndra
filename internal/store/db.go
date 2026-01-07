package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

// PoolConfig holds database connection pool configuration
type PoolConfig struct {
	MaxOpenConns    int // Maximum number of open connections
	MaxIdleConns    int // Maximum number of idle connections
	ConnMaxLifetime int // Connection max lifetime in seconds
	ConnMaxIdleTime int // Connection max idle time in seconds
}

// DefaultPoolConfig returns sensible defaults for connection pooling
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300, // 5 minutes
		ConnMaxIdleTime: 600, // 10 minutes
	}
}

// New creates a new database connection with optimized connection pool settings
func New(databaseURL string) (*DB, error) {
	return NewWithConfig(databaseURL, DefaultPoolConfig())
}

// NewWithConfig creates a new database connection with custom pool configuration
func NewWithConfig(databaseURL string, config PoolConfig) (*DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for optimal performance
	// MaxOpenConns: Maximum number of open connections to the database
	// Prevents connection exhaustion and leaves room for other connections
	db.SetMaxOpenConns(config.MaxOpenConns)

	// MaxIdleConns: Maximum number of idle connections in the pool
	// Maintains a pool of ready connections for faster response times
	db.SetMaxIdleConns(config.MaxIdleConns)

	// ConnMaxLifetime: Maximum amount of time a connection may be reused
	// Prevents stale connections after network issues or database restarts
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	// ConnMaxIdleTime: Maximum amount of time a connection may be idle
	// Closes idle connections to free resources
	db.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

