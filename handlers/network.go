package handlers

import (
	"net/http"
	"strconv"

	"github.com/Her0x27/x-routersbc/services"
	"github.com/labstack/echo/v4"
)

// NetworkHandler handles network-related requests
type NetworkHandler struct {
	networkService *services.NetworkService
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler() *NetworkHandler {
	return &NetworkHandler{
		networkService: services.NewNetworkService(),
	}
}

// ShowNetworkIndex displays the main network page
func (h *NetworkHandler) ShowNetworkIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "network/index.html", map[string]interface{}{
		"Title": "Network - RouterSBC",
	})
}

// ShowInterfaces displays the network interfaces page
func (h *NetworkHandler) ShowInterfaces(c echo.Context) error {
	interfaces, err := h.networkService.GetInterfaces()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get network interfaces",
		})
	}

	return c.Render(http.StatusOK, "network/interfaces.html", map[string]interface{}{
		"Title":      "Network Interfaces - RouterSBC",
		"Interfaces": interfaces,
	})
}

// ShowWAN displays the WAN configuration page
func (h *NetworkHandler) ShowWAN(c echo.Context) error {
	wanConfig, err := h.networkService.GetWANConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get WAN configuration",
		})
	}

	return c.Render(http.StatusOK, "network/wan.html", map[string]interface{}{
		"Title":     "WAN Configuration - RouterSBC",
		"WANConfig": wanConfig,
	})
}

// ShowLAN displays the LAN configuration page
func (h *NetworkHandler) ShowLAN(c echo.Context) error {
	lanConfig, err := h.networkService.GetLANConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get LAN configuration",
		})
	}

	return c.Render(http.StatusOK, "network/lan.html", map[string]interface{}{
		"Title":     "LAN Configuration - RouterSBC",
		"LANConfig": lanConfig,
	})
}

// ShowWireless displays the wireless configuration page
func (h *NetworkHandler) ShowWireless(c echo.Context) error {
	wirelessConfig, err := h.networkService.GetWirelessConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get wireless configuration",
		})
	}

	return c.Render(http.StatusOK, "network/wireless.html", map[string]interface{}{
		"Title":          "Wireless Configuration - RouterSBC",
		"WirelessConfig": wirelessConfig,
	})
}

// ShowRouting displays the routing configuration page
func (h *NetworkHandler) ShowRouting(c echo.Context) error {
	routes, err := h.networkService.GetRoutes()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get routing table",
		})
	}

	return c.Render(http.StatusOK, "network/routing.html", map[string]interface{}{
		"Title":  "Routing Configuration - RouterSBC",
		"Routes": routes,
	})
}

// ShowFirewall displays the firewall configuration page
func (h *NetworkHandler) ShowFirewall(c echo.Context) error {
	firewallConfig, err := h.networkService.GetFirewallConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get firewall configuration",
		})
	}

	// Determine which template to use based on firewall backend
	template := "network/firewall_new.html" // NFTables by default
	if firewallConfig.Backend == "iptables" {
		template = "network/firewall_classic.html"
	}

	return c.Render(http.StatusOK, template, map[string]interface{}{
		"Title":          "Firewall Configuration - RouterSBC",
		"FirewallConfig": firewallConfig,
	})
}

// API Endpoints for AJAX requests

// GetInterfacesAPI returns interfaces as JSON
func (h *NetworkHandler) GetInterfacesAPI(c echo.Context) error {
	interfaces, err := h.networkService.GetInterfaces()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get network interfaces",
		})
	}
	return c.JSON(http.StatusOK, interfaces)
}

// CreateInterface creates a new network interface
func (h *NetworkHandler) CreateInterface(c echo.Context) error {
	var req services.InterfaceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if err := h.networkService.CreateInterface(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create interface: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Interface created successfully",
	})
}

// UpdateInterface updates an existing network interface
func (h *NetworkHandler) UpdateInterface(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid interface ID",
		})
	}

	var req services.InterfaceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if err := h.networkService.UpdateInterface(id, req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update interface: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Interface updated successfully",
	})
}

// DeleteInterface removes a network interface
func (h *NetworkHandler) DeleteInterface(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid interface ID",
		})
	}

	if err := h.networkService.DeleteInterface(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete interface: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Interface deleted successfully",
	})
}

// CreateRoute creates a new static route
func (h *NetworkHandler) CreateRoute(c echo.Context) error {
	var req services.RouteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if err := h.networkService.CreateRoute(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create route: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Route created successfully",
	})
}

// UpdateWANConfig updates WAN configuration
func (h *NetworkHandler) UpdateWANConfig(c echo.Context) error {
	var req services.WANConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if err := h.networkService.UpdateWANConfig(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update WAN configuration: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "WAN configuration updated successfully",
	})
}
