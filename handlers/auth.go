package handlers

import (
	"net/http"
	"time"

	"github.com/Her0x27/x-routersbc/core"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService *core.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: core.NewAuthService(),
	}
}

// ShowLogin displays the login page
func (h *AuthHandler) ShowLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Title": "Login - RouterSBC",
	})
}

// Login handles login form submission
func (h *AuthHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.Render(http.StatusBadRequest, "login.html", map[string]interface{}{
			"Title": "Login - RouterSBC",
			"Error": "Username and password are required",
		})
	}

	session, err := h.authService.Login(username, password)
	if err != nil {
		return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{
			"Title": "Login - RouterSBC",
			"Error": "Invalid credentials",
		})
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true for HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}
	c.SetCookie(cookie)

	// Check if first login
	user, err := h.authService.GetUserByUsername(username)
	if err == nil && user.FirstLogin {
		// Redirect to password change with alert
		return c.Redirect(http.StatusFound, "/?first_login=true")
	}

	return c.Redirect(http.StatusFound, "/")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get session cookie
	cookie, err := c.Cookie("session_id")
	if err == nil {
		// Delete session from database
		h.authService.DeleteSession(cookie.Value)
	}

	// Clear session cookie
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

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	session := c.Get("session").(*core.Session)
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
	user, err := h.authService.GetUserByUsername("sbc") // For now, assuming single user
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user information",
		})
	}

	if !h.authService.VerifyPassword(currentPassword, user.PasswordHash) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Current password is incorrect",
		})
	}

	// Change password
	if err := h.authService.ChangePassword(session.UserID, newPassword); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to change password",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}
