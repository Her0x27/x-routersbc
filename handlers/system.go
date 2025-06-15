package handlers

import (
	"net/http"

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

// ShowSystemIndex displays the main system page
func (h *SystemHandler) ShowSystemIndex(c echo.Context) error {
	systemInfo, err := h.systemService.GetSystemInfo()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get system information",
		})
	}

	return c.Render(http.StatusOK, "system/index.html", map[string]interface{}{
		"Title":      "System - RouterSBC",
		"SystemInfo": systemInfo,
		"FirstLogin": c.QueryParam("first_login") == "true",
	})
}

// ShowAboutDevices displays device information
func (h *SystemHandler) ShowAboutDevices(c echo.Context) error {
	deviceInfo, err := h.systemService.GetDeviceInfo()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get device information",
		})
	}

	return c.Render(http.StatusOK, "system/about-devices.html", map[string]interface{}{
		"Title":      "About Devices - RouterSBC",
		"DeviceInfo": deviceInfo,
	})
}

// ShowPortableDevices displays connected USB devices
func (h *SystemHandler) ShowPortableDevices(c echo.Context) error {
	portableDevices, err := h.systemService.GetPortableDevices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get portable devices information",
		})
	}

	return c.Render(http.StatusOK, "system/portable-devices.html", map[string]interface{}{
		"Title":           "Portable Devices - RouterSBC",
		"PortableDevices": portableDevices,
	})
}

// UpdateTimeZone updates the system timezone
func (h *SystemHandler) UpdateTimeZone(c echo.Context) error {
	timezone := c.FormValue("timezone")
	if timezone == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Timezone is required",
		})
	}

	if err := h.systemService.SetTimeZone(timezone); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update timezone: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Timezone updated successfully",
	})
}

// CreateBackup creates a system backup
func (h *SystemHandler) CreateBackup(c echo.Context) error {
	backupPath, err := h.systemService.CreateBackup()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create backup: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Backup created successfully",
		"backup_path": backupPath,
	})
}

// RestoreBackup restores from a backup file
func (h *SystemHandler) RestoreBackup(c echo.Context) error {
	// Handle file upload
	file, err := c.FormFile("backup_file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No backup file provided",
		})
	}

	if err := h.systemService.RestoreBackup(file); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to restore backup: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup restored successfully",
	})
}

// GetSystemInfoAPI returns system information as JSON
func (h *SystemHandler) GetSystemInfoAPI(c echo.Context) error {
	systemInfo, err := h.systemService.GetSystemInfo()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get system information",
		})
	}
	return c.JSON(http.StatusOK, systemInfo)
}

// RefreshPortableDevices rescans for portable devices
func (h *SystemHandler) RefreshPortableDevices(c echo.Context) error {
	portableDevices, err := h.systemService.RefreshPortableDevices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to refresh portable devices: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, portableDevices)
}
