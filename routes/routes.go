package routes

import (
	"github.com/Her0x27/x-routersbc/core"
	"github.com/Her0x27/x-routersbc/handlers"
	"github.com/labstack/echo/v4"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	e *echo.Echo,
	authHandler *handlers.AuthHandler,
	networkHandler *handlers.NetworkHandler,
	systemHandler *handlers.SystemHandler,
	wsManager *core.WebSocketManager,
) {
	// Public routes (no authentication required)
	e.GET("/login", authHandler.ShowLogin)
	e.POST("/login", authHandler.Login)
	e.GET("/logout", authHandler.Logout)
	
	// WebSocket endpoint
	e.GET("/ws", wsManager.HandleWebSocket)
	
	// Protected routes (authentication required)
	protected := e.Group("", authHandler.RequireAuth)
	
	// Dashboard/Home
	protected.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/network")
	})
	
	// Network routes
	network := protected.Group("/network")
	{
		network.GET("", networkHandler.ShowNetworkIndex)
		network.GET("/", networkHandler.ShowNetworkIndex)
		network.GET("/interfaces", networkHandler.ShowInterfaces)
		network.GET("/wan", networkHandler.ShowWAN)
		network.GET("/lan", networkHandler.ShowLAN)
		network.GET("/wireless", networkHandler.ShowWireless)
		network.GET("/routing", networkHandler.ShowRouting)
		network.GET("/firewall", networkHandler.ShowFirewall)
		
		// API endpoints for network configuration
		network.POST("/interfaces", networkHandler.SaveInterface)
		network.DELETE("/interfaces/:name", networkHandler.DeleteInterface)
	}
	
	// System routes
	system := protected.Group("/system")
	{
		system.GET("", systemHandler.ShowSystemIndex)
		system.GET("/", systemHandler.ShowSystemIndex)
		system.GET("/about-devices", systemHandler.ShowAboutDevices)
		system.GET("/portable-devices", systemHandler.ShowPortableDevices)
		
		// API endpoints for system management
		system.GET("/status", systemHandler.GetSystemStatus)
		system.POST("/change-password", authHandler.ChangePassword)
		system.POST("/backup", systemHandler.CreateBackup)
		system.POST("/restore", systemHandler.RestoreBackup)
	}
}
