package routes

import (
	"github.com/Her0x27/x-routersbc/core"
	"github.com/Her0x27/x-routersbc/handlers"
	"github.com/labstack/echo/v4"
)

// SetupRoutes configures all application routes
func SetupRoutes(e *echo.Echo) {
	// Initialize WebSocket hub
	core.InitWebSocket()

	// Create handlers
	authHandler := handlers.NewAuthHandler()
	networkHandler := handlers.NewNetworkHandler()
	systemHandler := handlers.NewSystemHandler()

	// Public routes (no authentication required)
	e.GET("/login", authHandler.ShowLogin)
	e.POST("/login", authHandler.Login)

	// Protected routes (authentication required)
	protected := e.Group("", core.AuthMiddleware)

	// Root redirect to system page
	protected.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/system")
	})

	// Authentication routes
	protected.POST("/logout", authHandler.Logout)
	protected.POST("/change-password", authHandler.ChangePassword)

	// System routes
	systemRoutes := protected.Group("/system")
	systemRoutes.GET("", systemHandler.ShowSystemIndex)
	systemRoutes.GET("/", systemHandler.ShowSystemIndex)
	systemRoutes.GET("/about-devices", systemHandler.ShowAboutDevices)
	systemRoutes.GET("/portable-devices", systemHandler.ShowPortableDevices)
	
	// System API routes
	systemAPI := systemRoutes.Group("/api")
	systemAPI.GET("/info", systemHandler.GetSystemInfoAPI)
	systemAPI.POST("/timezone", systemHandler.UpdateTimeZone)
	systemAPI.POST("/backup", systemHandler.CreateBackup)
	systemAPI.POST("/restore", systemHandler.RestoreBackup)
	systemAPI.POST("/portable-devices/refresh", systemHandler.RefreshPortableDevices)

	// Network routes
	networkRoutes := protected.Group("/network")
	networkRoutes.GET("", networkHandler.ShowNetworkIndex)
	networkRoutes.GET("/", networkHandler.ShowNetworkIndex)
	networkRoutes.GET("/interfaces", networkHandler.ShowInterfaces)
	networkRoutes.GET("/wan", networkHandler.ShowWAN)
	networkRoutes.GET("/lan", networkHandler.ShowLAN)
	networkRoutes.GET("/wireless", networkHandler.ShowWireless)
	networkRoutes.GET("/routing", networkHandler.ShowRouting)
	networkRoutes.GET("/firewall", networkHandler.ShowFirewall)

	// Network API routes
	networkAPI := networkRoutes.Group("/api")
	networkAPI.GET("/interfaces", networkHandler.GetInterfacesAPI)
	networkAPI.POST("/interfaces", networkHandler.CreateInterface)
	networkAPI.PUT("/interfaces/:id", networkHandler.UpdateInterface)
	networkAPI.DELETE("/interfaces/:id", networkHandler.DeleteInterface)
	networkAPI.POST("/routes", networkHandler.CreateRoute)
	networkAPI.PUT("/wan", networkHandler.UpdateWANConfig)

	// WebSocket route
	protected.GET("/ws", core.HandleWebSocket(core.GlobalWSHub))
}
