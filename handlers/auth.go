package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Her0x27/x-routersbc/core"
	"github.com/Her0x27/x-routersbc/utils"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) ShowLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Title": "Login - Router SBC",
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	
	if username == "" || password == "" {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"Title": "Login - Router SBC",
			"Error": "Username and password are required",
		})
	}
	
	// Get user from database
	var userID int
	var hashedPassword string
	var firstLogin bool
	
	err := h.db.QueryRow(`
		SELECT id, password_hash, first_login 
		FROM users 
		WHERE username = ?
	`, username).Scan(&userID, &hashedPassword, &firstLogin)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Render(http.StatusOK, "login.html", map[string]interface{}{
				"Title": "Login - Router SBC",
				"Error": "Invalid username or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database error",
		})
	}
	
	// Verify password
	if !utils.CheckPasswordHash(password, hashedPassword) {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"Title": "Login - Router SBC",
			"Error": "Invalid username or password",
		})
	}
	
	// Create session
	session, err := core.CreateSession(h.db, userID, username, firstLogin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create session",
		})
	}
	
	// Set cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   24 * 60 * 60, // 24 hours
	}
	c.SetCookie(cookie)
	
	return c.Redirect(http.StatusFound, "/network")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("session_id")
	if err == nil {
		core.DeleteSession(h.db, cookie.Value)
	}
	
	// Clear cookie
	cookie = &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	c.SetCookie(cookie)
	
	return c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	if c.Request().Method == "GET" {
		return c.Render(http.StatusOK, "system/index.html", map[string]interface{}{
			"Title":       "System Settings - Router SBC",
			"Session":     session,
			"ShowPasswordAlert": session.FirstLogin,
		})
	}
	
	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")
	
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "All fields are required",
		})
	}
	
	if newPassword != confirmPassword {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "New passwords do not match",
		})
	}
	
	// Verify current password
	var hashedPassword string
	err := h.db.QueryRow(`
		SELECT password_hash FROM users WHERE id = ?
	`, session.UserID).Scan(&hashedPassword)
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database error",
		})
	}
	
	if !utils.CheckPasswordHash(currentPassword, hashedPassword) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Current password is incorrect",
		})
	}
	
	// Hash new password
	newHashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to hash password",
		})
	}
	
	// Update password and mark first login as complete
	_, err = h.db.Exec(`
		UPDATE users 
		SET password_hash = ?, first_login = FALSE 
		WHERE id = ?
	`, newHashedPassword, session.UserID)
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update password",
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}
