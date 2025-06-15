package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/Her0x27/x-routersbc/core"
	"github.com/Her0x27/x-routersbc/services"
)

type NetworkHandler struct {
	db             *sql.DB
	networkService *services.NetworkService
}

func NewNetworkHandler(db *sql.DB) *NetworkHandler {
	return &NetworkHandler{
		db:             db,
		networkService: services.NewNetworkService(db),
	}
}

func (h *NetworkHandler) ShowNetworkIndex(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	return c.Render(http.StatusOK, "network/index.html", map[string]interface{}{
		"Title":   "Network Configuration - Router SBC",
		"Session": session,
	})
}

func (h *NetworkHandler) ShowInterfaces(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	interfaces, err := h.networkService.GetNetworkInterfaces()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get network interfaces",
		})
	}
	
	return c.Render(http.StatusOK, "network/interfaces.html", map[string]interface{}{
		"Title":      "Network Interfaces - Router SBC",
		"Session":    session,
		"Interfaces": interfaces,
	})
}

func (h *NetworkHandler) ShowWAN(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	wanConfig, err := h.networkService.GetWANConfiguration()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get WAN configuration",
		})
	}
	
	return c.Render(http.StatusOK, "network/wan.html", map[string]interface{}{
		"Title":     "WAN Configuration - Router SBC",
		"Session":   session,
		"WANConfig": wanConfig,
	})
}

func (h *NetworkHandler) ShowLAN(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	lanConfig, err := h.networkService.GetLANConfiguration()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get LAN configuration",
		})
	}
	
	return c.Render(http.StatusOK, "network/lan.html", map[string]interface{}{
		"Title":     "LAN Configuration - Router SBC",
		"Session":   session,
		"LANConfig": lanConfig,
	})
}

func (h *NetworkHandler) ShowWireless(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	wirelessConfig, err := h.networkService.GetWirelessConfiguration()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get wireless configuration",
		})
	}
	
	return c.Render(http.StatusOK, "network/wireless.html", map[string]interface{}{
		"Title":          "Wireless Configuration - Router SBC",
		"Session":        session,
		"WirelessConfig": wirelessConfig,
	})
}

func (h *NetworkHandler) ShowRouting(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	routes, err := h.networkService.GetStaticRoutes()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get routing configuration",
		})
	}
	
	return c.Render(http.StatusOK, "network/routing.html", map[string]interface{}{
		"Title":   "Routing Configuration - Router SBC",
		"Session": session,
		"Routes":  routes,
	})
}

func (h *NetworkHandler) ShowFirewall(c echo.Context) error {
	session := c.Get("session").(*core.Session)
	
	firewallConfig, err := h.networkService.GetFirewallConfiguration()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get firewall configuration",
		})
	}
	
	// Determine which firewall template to use
	template := "network/firewall_new.html"
	if firewallConfig.Backend == "iptables" {
		template = "network/firewall_classic.html"
	}
	
	return c.Render(http.StatusOK, template, map[string]interface{}{
		"Title":          "Firewall Configuration - Router SBC",
		"Session":        session,
		"FirewallConfig": firewallConfig,
	})
}

// API endpoints for AJAX requests

func (h *NetworkHandler) CreateInterface(c echo.Context) error {
	var req services.NetworkInterface
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}
	
	if err := h.networkService.CreateNetworkInterface(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create network interface: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("interface_created", req)
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Network interface created successfully",
	})
}

func (h *NetworkHandler) UpdateInterface(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid interface ID",
		})
	}
	
	var req services.NetworkInterface
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}
	
	req.ID = id
	if err := h.networkService.UpdateNetworkInterface(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update network interface: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("interface_updated", req)
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Network interface updated successfully",
	})
}

func (h *NetworkHandler) DeleteInterface(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid interface ID",
		})
	}
	
	if err := h.networkService.DeleteNetworkInterface(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete network interface: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("interface_deleted", map[string]int{"id": id})
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Network interface deleted successfully",
	})
}

func (h *NetworkHandler) CreateStaticRoute(c echo.Context) error {
	var req services.StaticRoute
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}
	
	if err := h.networkService.CreateStaticRoute(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create static route: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("route_created", req)
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Static route created successfully",
	})
}

func (h *NetworkHandler) UpdateStaticRoute(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid route ID",
		})
	}
	
	var req services.StaticRoute
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}
	
	req.ID = id
	if err := h.networkService.UpdateStaticRoute(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update static route: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("route_updated", req)
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Static route updated successfully",
	})
}

func (h *NetworkHandler) DeleteStaticRoute(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid route ID",
		})
	}
	
	if err := h.networkService.DeleteStaticRoute(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete static route: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("route_deleted", map[string]int{"id": id})
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Static route deleted successfully",
	})
}

func (h *NetworkHandler) CreateFirewallRule(c echo.Context) error {
	var req services.FirewallRule
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}
	
	if err := h.networkService.CreateFirewallRule(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create firewall rule: " + err.Error(),
		})
	}
	
	// Broadcast update via WebSocket
	core.BroadcastMessage("firewall_rule_created", req)
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Firewall rule created successfully",
	})
}
