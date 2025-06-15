package core

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDatabase initializes the SQLite database
func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "routersbc.sqlitedb")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	// Create default admin user
	authService := NewAuthService()
	if err = authService.CreateDefaultUser(); err != nil {
		return fmt.Errorf("failed to create default user: %v", err)
	}

	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			first_login BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
		`CREATE TABLE IF NOT EXISTS network_interfaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL,
			enabled BOOLEAN DEFAULT 1,
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS network_routes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			interface TEXT NOT NULL,
			destination TEXT NOT NULL,
			gateway TEXT NOT NULL,
			metric INTEGER DEFAULT 0,
			enabled BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS firewall_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chain TEXT NOT NULL,
			rule_text TEXT NOT NULL,
			position INTEGER DEFAULT 0,
			enabled BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS system_config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}
