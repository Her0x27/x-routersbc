package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Her0x27/x-routersbc/core"
	"github.com/Her0x27/x-routersbc/services"
)

type SystemHandler struct {
	db            *sql.DB
	systemService *services.SystemService
}

func NewSystemHandler(db *sql.DB) *SystemHandler {
	return &SystemHandler{
		db:            db,
		systemService: services.NewSystemService(db),
	}
}

func (h *SystemHandler) ShowSystemIndex(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	return c.Render(http.StatusOK, "system/index.html", map[string]interface{}{
		"Title":             "System Settings - Router SBC",
		"Session":           session,
		"ShowPasswordAlert": session.FirstLogin,
	})
}

func (h *SystemHandler) ShowAboutDevices(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	deviceInfo, err := h.systemService.GetDeviceInformation()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get device information",
		})
	}
	
	return c.Render(http.StatusOK, "system/about-devices.html", map[string]interface{}{
		"Title":      "About Device - Router SBC",
		"Session":    session,
		"DeviceInfo": deviceInfo,
	})
}

func (h *SystemHandler) ShowPortableDevices(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	portableDevices, err := h.systemService.GetPortableDevices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get portable devices information",
		})
	}
	
	return c.Render(http.StatusOK, "system/portable-devices.html", map[string]interface{}{
		"Title":           "Portable Devices - Router SBC",
		"Session":         session,
		"PortableDevices": portableDevices,
	})
}

func (h *SystemHandler) GetSystemTime(c echo.Context) error {
	timeInfo, err := h.systemService.GetSystemTime()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get system time",
		})
	}
	
	return c.JSON(http.StatusOK, timeInfo)
}

func (h *SystemHandler) SetSystemTime(c echo.Context) error {
	var req struct {
		Timezone string `json:"timezone"`
		NTPEnabled bool `json:"ntp_enabled"`
		NTPServers []string `json:"ntp_servers"`
	}
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}
	
	if err := h.systemService.SetSystemTime(req.Timezone, req.NTPEnabled, req.NTPServers); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to set system time: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("system_time_updated", req)
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "System time updated successfully",
	})
}

func (h *SystemHandler) CreateBackup(c echo.Context) error {
	backupPath, err := h.systemService.CreateBackup()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create backup: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup created successfully",
		"path":    backupPath,
	})
}

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
			"error": "Failed to restore backup: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Backup restored successfully. System will restart.",
	})
}

func (h *SystemHandler) GetPortableDeviceStatus(c echo.Context) error {
	deviceType := c.Param("type")
	
	status, err := h.systemService.GetPortableDeviceStatus(deviceType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get device status: " + err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, status)
}

func (h *SystemHandler) InstallDrivers(c echo.Context) error {
	deviceType := c.Param("type")
	
	if err := h.systemService.InstallDrivers(deviceType); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to install drivers: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("drivers_installed", map[string]string{"device_type": deviceType})
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Driver installation started. Check status for updates.",
	})
}
