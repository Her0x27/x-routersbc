package handlers

import (
	"net/http"

	"github.com/Her0x27/x-routersbc/core"
	"github.com/Her0x27/x-routersbc/services"
	"github.com/labstack/echo/v4"
)

// SystemHandler handles system-related requests
type SystemHandler struct {
	systemService *services.SystemService
}

// NewSystemHandler creates a new system handler
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		systemService: services.NewSystemService(),
	}
}

// ShowSystemIndex shows the system overview page
func (h *SystemHandler) ShowSystemIndex(c echo.Context) error {
	user := c.Get("user").(*core.User)
	showFirstLoginAlert := c.QueryParam("first_login") == "true"
	
	return c.Render(http.StatusOK, "system/index.html", map[string]interface{}{
		"title":               "System - RouterSBC",
		"user":                user,
		"showFirstLoginAlert": showFirstLoginAlert,
	})
}

// ShowAboutDevices shows the device information page
func (h *SystemHandler) ShowAboutDevices(c echo.Context) error {
	deviceInfo, err := h.systemService.GetDeviceInformation()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "system/about-devices.html", map[string]interface{}{
			"title": "About Device - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "system/about-devices.html", map[string]interface{}{
		"title":      "About Device - RouterSBC",
		"deviceInfo": deviceInfo,
	})
}

// ShowPortableDevices shows the portable devices page
func (h *SystemHandler) ShowPortableDevices(c echo.Context) error {
	usbDevices, err := h.systemService.GetUSBDevices()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "system/portable-devices.html", map[string]interface{}{
			"title": "Portable Devices - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "system/portable-devices.html", map[string]interface{}{
		"title":      "Portable Devices - RouterSBC",
		"usbDevices": usbDevices,
	})
}

// GetSystemStatus returns system status as JSON
func (h *SystemHandler) GetSystemStatus(c echo.Context) error {
	status, err := h.systemService.GetSystemStatus()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, status)
}

// CreateBackup creates a system backup
func (h *SystemHandler) CreateBackup(c echo.Context) error {
	backupPath, err := h.systemService.CreateBackup()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message":    "Backup created successfully",
		"backupPath": backupPath,
	})
}

// RestoreBackup restores from a backup
func (h *SystemHandler) RestoreBackup(c echo.Context) error {
	// Handle file upload
	file, err := c.FormFile("backup")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No backup file provided",
		})
	}
	
	if err := h.systemService.RestoreBackup(file); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup restored successfully",
	})
}
