package core

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Her0x27/x-routersbc/utils"
	"github.com/labstack/echo/v4"
)

// User represents a system user
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	FirstLogin   bool      `json:"first_login"`
	CreatedAt    time.Time `json:"created_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthService handles authentication operations
type AuthService struct{}

// NewAuthService creates a new auth service instance
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Login authenticates a user and creates a session
func (a *AuthService) Login(username, password string) (*Session, error) {
	// Get user from database
	user, err := a.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if !a.VerifyPassword(password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create session
	session := &Session{
		ID:        utils.GenerateSessionID(),
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Save session to database
	if err := a.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return session, nil
}

// GetUserByUsername retrieves a user by username
func (a *AuthService) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := "SELECT id, username, password_hash, first_login, created_at FROM users WHERE username = ?"
	
	row := db.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.FirstLogin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// VerifyPassword checks if the provided password matches the hash
func (a *AuthService) VerifyPassword(password, hash string) bool {
	return utils.HashPassword(password) == hash
}

// SaveSession saves a session to the database
func (a *AuthService) SaveSession(session *Session) error {
	query := "INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, session.ID, session.UserID, session.CreatedAt, session.ExpiresAt)
	return err
}

// GetSession retrieves a session by ID
func (a *AuthService) GetSession(sessionID string) (*Session, error) {
	session := &Session{}
	query := "SELECT id, user_id, created_at, expires_at FROM sessions WHERE id = ? AND expires_at > ?"
	
	row := db.QueryRow(query, sessionID, time.Now())
	err := row.Scan(&session.ID, &session.UserID, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// DeleteSession removes a session from the database
func (a *AuthService) DeleteSession(sessionID string) error {
	query := "DELETE FROM sessions WHERE id = ?"
	_, err := db.Exec(query, sessionID)
	return err
}

// ChangePassword updates user password
func (a *AuthService) ChangePassword(userID int, newPassword string) error {
	hashedPassword := utils.HashPassword(newPassword)
	query := "UPDATE users SET password_hash = ?, first_login = 0 WHERE id = ?"
	_, err := db.Exec(query, hashedPassword, userID)
	return err
}

// CreateDefaultUser creates the default admin user if it doesn't exist
func (a *AuthService) CreateDefaultUser() error {
	// Check if default user exists
	_, err := a.GetUserByUsername("sbc")
	if err == nil {
		return nil // User already exists
	}

	// Create default user with hashed password "sbc"
	hashedPassword := utils.HashPassword("sbc")
	query := "INSERT INTO users (username, password_hash, first_login, created_at) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(query, "sbc", hashedPassword, true, time.Now())
	return err
}

// AuthMiddleware checks for valid session
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip auth for login page and static files
		path := c.Request().URL.Path
		if path == "/login" || path == "/static" {
			return next(c)
		}

		// Get session cookie
		cookie, err := c.Cookie("session_id")
		if err != nil {
			return c.Redirect(302, "/login")
		}

		// Validate session
		authService := NewAuthService()
		session, err := authService.GetSession(cookie.Value)
		if err != nil {
			return c.Redirect(302, "/login")
		}

		// Store session in context
		c.Set("session", session)
		return next(c)
	}
}
