package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/Her0x27/x-routersbc/handlers"
)

func SetupRoutes(e *echo.Echo, authHandler *handlers.AuthHandler, networkHandler *handlers.NetworkHandler, systemHandler *handlers.SystemHandler) {
	// Auth routes
	e.GET("/", authHandler.ShowLogin)
	e.GET("/login", authHandler.ShowLogin)
	e.POST("/login", authHandler.Login)
	e.POST("/logout", authHandler.Logout)
	e.GET("/change-password", authHandler.ChangePassword)
	e.POST("/change-password", authHandler.ChangePassword)
	
	// Network routes
	network := e.Group("/network")
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
		api := network.Group("/api")
		{
			// Interface management
			api.POST("/interfaces", networkHandler.CreateInterface)
			api.PUT("/interfaces/:id", networkHandler.UpdateInterface)
			api.DELETE("/interfaces/:id", networkHandler.DeleteInterface)
			
			// Static routes
			api.POST("/routes", networkHandler.CreateStaticRoute)
			api.PUT("/routes/:id", networkHandler.UpdateStaticRoute)
			api.DELETE("/routes/:id", networkHandler.DeleteStaticRoute)
			
			// Firewall rules
			api.POST("/firewall/rules", networkHandler.CreateFirewallRule)
		}
	}
	
	// System routes
	system := e.Group("/system")
	{
		system.GET("", systemHandler.ShowSystemIndex)
		system.GET("/", systemHandler.ShowSystemIndex)
		system.GET("/about-devices", systemHandler.ShowAboutDevices)
		system.GET("/portable-devices", systemHandler.ShowPortableDevices)
		
		// API endpoints for system management
		api := system.Group("/api")
		{
			api.GET("/time", systemHandler.GetSystemTime)
			api.POST("/time", systemHandler.SetSystemTime)
			api.POST("/backup", systemHandler.CreateBackup)
			api.POST("/restore", systemHandler.RestoreBackup)
			api.GET("/devices/:type/status", systemHandler.GetPortableDeviceStatus)
			api.POST("/devices/:type/drivers", systemHandler.InstallDrivers)
		}
	}
}
