package core

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func InitDatabase() (*sql.DB, error) {
	dbPath := "routersbc.sqlitedb"
	
	// Check if database exists
	_, err := os.Stat(dbPath)
	isNewDB := os.IsNotExist(err)
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	if isNewDB {
		if err := createTables(db); err != nil {
			return nil, err
		}
		if err := createDefaultUser(db); err != nil {
			return nil, err
		}
	}
	
	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			first_login BOOLEAN DEFAULT TRUE,
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
			enabled BOOLEAN DEFAULT TRUE,
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS firewall_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			chain TEXT NOT NULL,
			rule_text TEXT NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			position INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS static_routes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			interface TEXT NOT NULL,
			destination TEXT NOT NULL,
			gateway TEXT NOT NULL,
			metric INTEGER DEFAULT 0,
			enabled BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS dns_zones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			zone TEXT NOT NULL,
			record_type TEXT NOT NULL,
			value TEXT NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	
	return nil
}

func createDefaultUser(db *sql.DB) error {
	hashedPassword, err := HashPassword("sbc")
	if err != nil {
		return err
	}
	
	_, err = db.Exec(`
		INSERT INTO users (username, password_hash, first_login) 
		VALUES (?, ?, ?)
	`, "sbc", hashedPassword, true)
	
	return err
}

func CleanExpiredSessions(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}
