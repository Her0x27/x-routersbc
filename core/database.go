package core

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection
type Database struct {
	conn *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase() (*Database, error) {
	conn, err := sql.Open("sqlite3", "routersbc.sqlitedb")
	if err != nil {
		return nil, err
	}
	
	db := &Database{conn: conn}
	
	// Initialize tables
	if err := db.initTables(); err != nil {
		return nil, err
	}
	
	return db, nil
}

// initTables creates the necessary tables
func (db *Database) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME,
			first_login BOOLEAN DEFAULT TRUE
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS network_interfaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			ip_address TEXT,
			netmask TEXT,
			gateway TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS firewall_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			chain TEXT NOT NULL,
			action TEXT NOT NULL,
			protocol TEXT,
			source TEXT,
			destination TEXT,
			port TEXT,
			enabled BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	
	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}
	
	// Create default admin user if not exists
	return db.createDefaultAdmin()
}

// createDefaultAdmin creates the default admin user
func (db *Database) createDefaultAdmin() error {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "sbc").Scan(&count)
	if err != nil {
		return err
	}
	
	if count == 0 {
		// Hash the default password "sbc"
		hashedPassword := HashPassword("sbc")
		_, err = db.conn.Exec(
			"INSERT INTO users (username, password_hash) VALUES (?, ?)",
			"sbc", hashedPassword,
		)
		return err
	}
	
	return nil
}

// GetConnection returns the database connection
func (db *Database) GetConnection() *sql.DB {
	return db.conn
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}

// Session represents a user session
type Session struct {
	ID        string
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
}

// User represents a user account
type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	LastLogin    *time.Time
	FirstLogin   bool
}
