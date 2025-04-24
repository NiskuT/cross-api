package repository

import (
	"database/sql"
	"fmt"

	"github.com/NiskuT/cross-api/internal/config"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// NewDatabaseConnection creates a new database connection
func NewDatabaseConnection(cfg *config.Config) (*sql.DB, error) {
	// Connect to the database using the configuration
	db, err := sql.Open("mysql", cfg.Database.Uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// InitializeDatabase sets up the database schema
func InitializeDatabase(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(CreateUsersTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create competitions table
	_, err = db.Exec(CreateCompetitionsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create competitions table: %w", err)
	}

	// Create participants table
	_, err = db.Exec(CreateParticipantsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create participants table: %w", err)
	}

	// Create scales table
	_, err = db.Exec(CreateScalesTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create scales table: %w", err)
	}

	// Create runs table
	_, err = db.Exec(CreateRunsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create runs table: %w", err)
	}

	// Create liverankings table
	_, err = db.Exec(CreateLiverankingsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create liverankings table: %w", err)
	}

	return nil
}
