package handlers

import (
	"net/http"

	"github.com/Her0x27/x-routersbc/core"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authManager *core.AuthManager
	database    *core.Database
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authManager *core.AuthManager, database *core.Database) *AuthHandler {
	return &AuthHandler{
		authManager: authManager,
		database:    database,
	}
}

// ShowLogin shows the login page
func (h *AuthHandler) ShowLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"title": "Login - RouterSBC",
	})
}

// Login handles login requests
func (h *AuthHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	
	if username == "" || password == "" {
		return c.Render(http.StatusBadRequest, "login.html", map[string]interface{}{
			"title": "Login - RouterSBC",
			"error": "Username and password are required",
		})
	}
	
	session, user, err := h.authManager.Login(username, password)
	if err != nil {
		return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{
			"title": "Login - RouterSBC",
			"error": "Invalid credentials",
		})
	}
	
	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true if using HTTPS
		MaxAge:   86400, // 24 hours
	}
	c.SetCookie(cookie)
	
	// Check if first login to show password change alert
	if user.FirstLogin {
		return c.Redirect(http.StatusFound, "/system?first_login=true")
	}
	
	return c.Redirect(http.StatusFound, "/")
}

// Logout handles logout requests
func (h *AuthHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err == nil {
		h.authManager.Logout(cookie.Value)
	}
	
	// Clear session cookie
	cookie = &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	c.SetCookie(cookie)
	
	return c.Redirect(http.StatusFound, "/login")
}

// ChangePassword handles password change requests
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	user := c.Get("user").(*core.User)
	
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
	if !core.VerifyPassword(currentPassword, user.PasswordHash) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Current password is incorrect",
		})
	}
	
	// Change password
	if err := h.authManager.ChangePassword(user.ID, newPassword); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to change password",
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// RequireAuth middleware to require authentication
func (h *AuthHandler) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}
		
		user, err := h.authManager.ValidateSession(cookie.Value)
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}
		
		c.Set("user", user)
		return next(c)
	}
}
