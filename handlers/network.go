package handlers

import (
	"net/http"

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

// ShowNetworkIndex shows the network overview page
func (h *NetworkHandler) ShowNetworkIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "network/index.html", map[string]interface{}{
		"title": "Network - RouterSBC",
	})
}

// ShowInterfaces shows the network interfaces page
func (h *NetworkHandler) ShowInterfaces(c echo.Context) error {
	interfaces, err := h.networkService.GetNetworkInterfaces()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "network/interfaces.html", map[string]interface{}{
			"title": "Network Interfaces - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "network/interfaces.html", map[string]interface{}{
		"title":      "Network Interfaces - RouterSBC",
		"interfaces": interfaces,
	})
}

// ShowWAN shows the WAN configuration page
func (h *NetworkHandler) ShowWAN(c echo.Context) error {
	wanConfig, err := h.networkService.GetWANConfiguration()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "network/wan.html", map[string]interface{}{
			"title": "WAN Configuration - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "network/wan.html", map[string]interface{}{
		"title":  "WAN Configuration - RouterSBC",
		"config": wanConfig,
	})
}

// ShowLAN shows the LAN configuration page
func (h *NetworkHandler) ShowLAN(c echo.Context) error {
	lanConfig, err := h.networkService.GetLANConfiguration()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "network/lan.html", map[string]interface{}{
			"title": "LAN Configuration - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "network/lan.html", map[string]interface{}{
		"title":  "LAN Configuration - RouterSBC",
		"config": lanConfig,
	})
}

// ShowWireless shows the wireless configuration page
func (h *NetworkHandler) ShowWireless(c echo.Context) error {
	wirelessConfig, err := h.networkService.GetWirelessConfiguration()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "network/wireless.html", map[string]interface{}{
			"title": "Wireless Configuration - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "network/wireless.html", map[string]interface{}{
		"title":  "Wireless Configuration - RouterSBC",
		"config": wirelessConfig,
	})
}

// ShowRouting shows the routing configuration page
func (h *NetworkHandler) ShowRouting(c echo.Context) error {
	routes, err := h.networkService.GetStaticRoutes()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "network/routing.html", map[string]interface{}{
			"title": "Routing Configuration - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "network/routing.html", map[string]interface{}{
		"title":  "Routing Configuration - RouterSBC",
		"routes": routes,
	})
}

// ShowFirewall shows the firewall configuration page
func (h *NetworkHandler) ShowFirewall(c echo.Context) error {
	firewallConfig, err := h.networkService.GetFirewallConfiguration()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "network/firewall.html", map[string]interface{}{
			"title": "Firewall Configuration - RouterSBC",
			"error": err.Error(),
		})
	}
	
	return c.Render(http.StatusOK, "network/firewall.html", map[string]interface{}{
		"title":  "Firewall Configuration - RouterSBC",
		"config": firewallConfig,
	})
}

// SaveInterface saves or updates a network interface
func (h *NetworkHandler) SaveInterface(c echo.Context) error {
	var req services.InterfaceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data",
		})
	}
	
	if err := h.networkService.SaveInterface(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Interface saved successfully",
	})
}

// DeleteInterface deletes a network interface
func (h *NetworkHandler) DeleteInterface(c echo.Context) error {
	interfaceName := c.Param("name")
	
	if err := h.networkService.DeleteInterface(interfaceName); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Interface deleted successfully",
	})
}
