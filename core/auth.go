package core

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Session struct {
	ID        string
	UserID    int
	Username  string
	FirstLogin bool
	CreatedAt time.Time
	ExpiresAt time.Time
}

func SessionMiddleware(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip auth for login page and static files
			path := c.Request().URL.Path
			if path == "/login" || path == "/" || 
			   len(path) >= 7 && path[:7] == "/static" {
				return next(c)
			}
			
			cookie, err := c.Cookie("session_id")
			if err != nil {
				return c.Redirect(http.StatusFound, "/login")
			}
			
			session, err := GetSession(db, cookie.Value)
			if err != nil || session == nil {
				return c.Redirect(http.StatusFound, "/login")
			}
			
			if session.ExpiresAt.Before(time.Now()) {
				DeleteSession(db, cookie.Value)
				return c.Redirect(http.StatusFound, "/login")
			}
			
			c.Set("session", session)
			return next(c)
		}
	}
}

func CreateSession(db *sql.DB, userID int, username string, firstLogin bool) (*Session, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	
	_, err := db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at) 
		VALUES (?, ?, ?)
	`, sessionID, userID, expiresAt)
	
	if err != nil {
		return nil, err
	}
	
	return &Session{
		ID:        sessionID,
		UserID:    userID,
		Username:  username,
		FirstLogin: firstLogin,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}, nil
}

func GetSession(db *sql.DB, sessionID string) (*Session, error) {
	var session Session
	var userID int
	var username string
	var firstLogin bool
	
	err := db.QueryRow(`
		SELECT s.id, s.user_id, s.created_at, s.expires_at, u.username, u.first_login
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = ?
	`, sessionID).Scan(&session.ID, &userID, &session.CreatedAt, &session.ExpiresAt, &username, &firstLogin)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	session.UserID = userID
	session.Username = username
	session.FirstLogin = firstLogin
	
	return &session, nil
}

func DeleteSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

func MarkFirstLoginComplete(db *sql.DB, userID int) error {
	_, err := db.Exec("UPDATE users SET first_login = FALSE WHERE id = ?", userID)
	return err
}
