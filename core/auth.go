package core

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

// AuthManager handles authentication and session management
type AuthManager struct {
	db *Database
}

// NewAuthManager creates a new auth manager
func NewAuthManager(db *Database) *AuthManager {
	return &AuthManager{db: db}
}

// Login authenticates a user and creates a session
func (am *AuthManager) Login(username, password string) (*Session, *User, error) {
	// Get user from database
	user, err := am.getUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("invalid credentials")
		}
		return nil, nil, err
	}
	
	// Verify password
	if !VerifyPassword(password, user.PasswordHash) {
		return nil, nil, errors.New("invalid credentials")
	}
	
	// Create session
	session, err := am.createSession(user.ID)
	if err != nil {
		return nil, nil, err
	}
	
	// Update last login
	_, err = am.db.conn.Exec(
		"UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?",
		user.ID,
	)
	if err != nil {
		return nil, nil, err
	}
	
	return session, user, nil
}

// ValidateSession validates a session token
func (am *AuthManager) ValidateSession(sessionID string) (*User, error) {
	var userID int
	var expiresAt time.Time
	
	err := am.db.conn.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&userID, &expiresAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid session")
		}
		return nil, err
	}
	
	// Check if session is expired
	if time.Now().After(expiresAt) {
		// Delete expired session
		am.db.conn.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		return nil, errors.New("session expired")
	}
	
	// Get user details
	user, err := am.getUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// Logout removes a session
func (am *AuthManager) Logout(sessionID string) error {
	_, err := am.db.conn.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// ChangePassword changes a user's password
func (am *AuthManager) ChangePassword(userID int, newPassword string) error {
	hashedPassword := HashPassword(newPassword)
	_, err := am.db.conn.Exec(
		"UPDATE users SET password_hash = ?, first_login = FALSE WHERE id = ?",
		hashedPassword, userID,
	)
	return err
}

// getUserByUsername gets a user by username
func (am *AuthManager) getUserByUsername(username string) (*User, error) {
	user := &User{}
	var lastLogin sql.NullTime
	
	err := am.db.conn.QueryRow(
		"SELECT id, username, password_hash, created_at, last_login, first_login FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &lastLogin, &user.FirstLogin)
	
	if err != nil {
		return nil, err
	}
	
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	
	return user, nil
}

// getUserByID gets a user by ID
func (am *AuthManager) getUserByID(id int) (*User, error) {
	user := &User{}
	var lastLogin sql.NullTime
	
	err := am.db.conn.QueryRow(
		"SELECT id, username, password_hash, created_at, last_login, first_login FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &lastLogin, &user.FirstLogin)
	
	if err != nil {
		return nil, err
	}
	
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	
	return user, nil
}

// createSession creates a new session
func (am *AuthManager) createSession(userID int) (*Session, error) {
	// Generate session ID
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	sessionID := hex.EncodeToString(bytes)
	
	// Session expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)
	
	// Insert session
	_, err := am.db.conn.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		return nil, err
	}
	
	return &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}, nil
}
