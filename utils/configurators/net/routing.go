package net

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type SystemRoute struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
	Protocol    string `json:"protocol"`
	Scope       string `json:"scope"`
}

type RouteTable struct {
	ID     int           `json:"id"`
	Name   string        `json:"name"`
	Routes []SystemRoute `json:"routes"`
}

type UPnPConfig struct {
	Enabled           bool     `json:"enabled"`
	AllowPortMapping  bool     `json:"allow_port_mapping"`
	AllowPCPNATMapping bool    `json:"allow_pcp_nat_mapping"`
	STUNServer        string   `json:"stun_server"`
	TrafficShaping    bool     `json:"traffic_shaping"`
	Interfaces        []string `json:"interfaces"`
}

func GetSystemRoutes() ([]SystemRoute, error) {
	output, err := exec.Command("ip", "route", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get system routes: %v", err)
	}

	var routes []SystemRoute
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		route := parseRouteFromIPOutput(line)
		if route.Destination != "" {
			routes = append(routes, route)
		}
	}

	return routes, nil
}

func parseRouteFromIPOutput(line string) SystemRoute {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return SystemRoute{}
	}

	route := SystemRoute{}

	// Parse destination
	if parts[0] == "default" {
		route.Destination = "0.0.0.0/0"
	} else {
		route.Destination = parts[0]
	}

	// Parse other fields
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "via":
			if i+1 < len(parts) {
				route.Gateway = parts[i+1]
				i++
			}
		case "dev":
			if i+1 < len(parts) {
				route.Interface = parts[i+1]
				i++
			}
		case "metric":
			if i+1 < len(parts) {
				if metric, err := strconv.Atoi(parts[i+1]); err == nil {
					route.Metric = metric
				}
				i++
			}
		case "proto":
			if i+1 < len(parts) {
				route.Protocol = parts[i+1]
				i++
			}
		case "scope":
			if i+1 < len(parts) {
				route.Scope = parts[i+1]
				i++
			}
		}
	}

	return route
}

func AddStaticRoute(destination, gateway, iface string, metric int) error {
	args := []string{"route", "add", destination}
	
	if gateway != "" {
		args = append(args, "via", gateway)
	}
	
	if iface != "" {
		args = append(args, "dev", iface)
	}
	
	if metric > 0 {
		args = append(args, "metric", strconv.Itoa(metric))
	}

	if err := exec.Command("ip", args...).Run(); err != nil {
		return fmt.Errorf("failed to add route: %v", err)
	}

	// Make route persistent
	return addPersistentRoute(destination, gateway, iface, metric)
}

func RemoveStaticRoute(destination, gateway, iface string, metric int) error {
	args := []string{"route", "del", destination}
	
	if gateway != "" {
		args = append(args, "via", gateway)
	}
	
	if iface != "" {
		args = append(args, "dev", iface)
	}
	
	if metric > 0 {
		args = append(args, "metric", strconv.Itoa(metric))
	}

	if err := exec.Command("ip", args...).Run(); err != nil {
		return fmt.Errorf("failed to remove route: %v", err)
	}

	// Remove from persistent configuration
	return removePersistentRoute(destination, gateway, iface, metric)
}

func addPersistentRoute(destination, gateway, iface string, metric int) error {
	// Try netplan first if available
	if IsNetplanAvailable() {
		return addNetplanRoute(destination, gateway, iface, metric)
	}

	// Fallback to traditional route files
	return addTraditionalRoute(destination, gateway, iface, metric)
}

func removePersistentRoute(destination, gateway, iface string, metric int) error {
	// Try netplan first if available
	if IsNetplanAvailable() {
		return removeNetplanRoute(destination, gateway, iface, metric)
	}

	// Fallback to traditional route files
	return removeTraditionalRoute(destination, gateway, iface, metric)
}

func addNetplanRoute(destination, gateway, iface string, metric int) error {
	// This would require modifying the netplan configuration
	// For now, we'll use a simple implementation that adds to existing config
	config, err := GetNetplanConfig()
	if err != nil {
		return err
	}

	// Add route to appropriate interface
	if config.Network.Ethernets != nil {
		if ethConfig, exists := config.Network.Ethernets[iface]; exists {
			// Add route to ethernet interface
			// This is a simplified implementation - full implementation would need proper route handling
			config.Network.Ethernets[iface] = ethConfig
		}
	}

	return ApplyNetplanConfig()
}

func removeNetplanRoute(destination, gateway, iface string, metric int) error {
	// This would require removing from netplan configuration
	// For now, return success as we'll rely on ip route commands
	return nil
}

func addTraditionalRoute(destination, gateway, iface string, metric int) error {
	// Add to /etc/network/interfaces or create route file
	routeCommand := fmt.Sprintf("ip route add %s", destination)
	
	if gateway != "" {
		routeCommand += fmt.Sprintf(" via %s", gateway)
	}
	
	if iface != "" {
		routeCommand += fmt.Sprintf(" dev %s", iface)
	}
	
	if metric > 0 {
		routeCommand += fmt.Sprintf(" metric %d", metric)
	}

	// Add to interface up script
	return addRouteToInterfaceScript(iface, routeCommand)
}

func removeTraditionalRoute(destination, gateway, iface string, metric int) error {
	// Remove from interface script
	routeCommand := fmt.Sprintf("ip route add %s", destination)
	
	if gateway != "" {
		routeCommand += fmt.Sprintf(" via %s", gateway)
	}
	
	if iface != "" {
		routeCommand += fmt.Sprintf(" dev %s", iface)
	}
	
	if metric > 0 {
		routeCommand += fmt.Sprintf(" metric %d", metric)
	}

	return removeRouteFromInterfaceScript(iface, routeCommand)
}

func addRouteToInterfaceScript(iface, routeCommand string) error {
	// This is a simplified implementation
	// In a full implementation, this would properly manage interface scripts
	return nil
}

func removeRouteFromInterfaceScript(iface, routeCommand string) error {
	// This is a simplified implementation
	// In a full implementation, this would properly manage interface scripts
	return nil
}

func GetRouteTable(tableID int) (*RouteTable, error) {
	output, err := exec.Command("ip", "route", "show", "table", strconv.Itoa(tableID)).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get route table %d: %v", tableID, err)
	}

	table := &RouteTable{
		ID:     tableID,
		Routes: []SystemRoute{},
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		route := parseRouteFromIPOutput(line)
		if route.Destination != "" {
			table.Routes = append(table.Routes, route)
		}
	}

	return table, nil
}

func GetDefaultGateway() (string, string, error) {
	output, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get default gateway: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "default via") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "via" && i+1 < len(parts) {
					gateway := parts[i+1]
					
					// Find interface
					for j, p := range parts {
						if p == "dev" && j+1 < len(parts) {
							return gateway, parts[j+1], nil
						}
					}
					
					return gateway, "", nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("no default gateway found")
}

func SetDefaultGateway(gateway, iface string) error {
	// Remove existing default routes
	exec.Command("ip", "route", "del", "default").Run()

	// Add new default route
	args := []string{"route", "add", "default", "via", gateway}
	if iface != "" {
		args = append(args, "dev", iface)
	}

	return exec.Command("ip", args...).Run()
}

func GetUPnPConfiguration() (*UPnPConfig, error) {
	config := &UPnPConfig{
		Enabled:            false,
		AllowPortMapping:   false,
		AllowPCPNATMapping: false,
		STUNServer:         "",
		TrafficShaping:     false,
		Interfaces:         []string{},
	}

	// Check if miniupnpd is running
	if isMiniUPnPDRunning() {
		upnpConfig, err := parseMiniUPnPDConfig()
		if err == nil {
			*config = *upnpConfig
			config.Enabled = true
		}
	}

	return config, nil
}

func isMiniUPnPDRunning() bool {
	err := exec.Command("pgrep", "miniupnpd").Run()
	return err == nil
}

func parseMiniUPnPDConfig() (*UPnPConfig, error) {
	// This would parse /etc/miniupnpd/miniupnpd.conf
	// For now, return a basic configuration
	return &UPnPConfig{
		Enabled:            true,
		AllowPortMapping:   true,
		AllowPCPNATMapping: false,
		STUNServer:         "stun.l.google.com:19302",
		TrafficShaping:     false,
		Interfaces:         []string{"eth0"},
	}, nil
}

func SetUPnPConfiguration(config *UPnPConfig) error {
	if !config.Enabled {
		return disableUPnP()
	}

	return configureMiniUPnPD(config)
}

func disableUPnP() error {
	exec.Command("systemctl", "stop", "miniupnpd").Run()
	exec.Command("systemctl", "disable", "miniupnpd").Run()
	return nil
}

func configureMiniUPnPD(config *UPnPConfig) error {
	// This would generate a proper miniupnpd configuration
	// For now, just enable the service
	exec.Command("systemctl", "enable", "miniupnpd").Run()
	return exec.Command("systemctl", "restart", "miniupnpd").Run()
}

func EnableTrafficShaping(iface string, upload, download int) error {
	// Configure traffic shaping using tc (traffic control)
	// This is a basic implementation
	
	// Clear existing rules
	exec.Command("tc", "qdisc", "del", "dev", iface, "root").Run()
	
	// Add root qdisc
	if err := exec.Command("tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", "htb", "default", "30").Run(); err != nil {
		return fmt.Errorf("failed to add root qdisc: %v", err)
	}
	
	// Add class for total bandwidth
	totalBandwidth := fmt.Sprintf("%dkbit", upload)
	if err := exec.Command("tc", "class", "add", "dev", iface, "parent", "1:", "classid", "1:1", "htb", "rate", totalBandwidth).Run(); err != nil {
		return fmt.Errorf("failed to add bandwidth class: %v", err)
	}
	
	return nil
}

func DisableTrafficShaping(iface string) error {
	return exec.Command("tc", "qdisc", "del", "dev", iface, "root").Run()
}

func GetRoutingStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})
	
	// Get default gateway
	gateway, iface, err := GetDefaultGateway()
	if err == nil {
		status["default_gateway"] = gateway
		status["default_interface"] = iface
	}
	
	// Get route count
	routes, err := GetSystemRoutes()
	if err == nil {
		status["total_routes"] = len(routes)
		status["static_routes"] = countStaticRoutes(routes)
	}
	
	// Get UPnP status
	upnpConfig, err := GetUPnPConfiguration()
	if err == nil {
		status["upnp_enabled"] = upnpConfig.Enabled
		status["upnp_running"] = isMiniUPnPDRunning()
	}
	
	return status, nil
}

func countStaticRoutes(routes []SystemRoute) int {
	count := 0
	for _, route := range routes {
		// Count routes that are not kernel or connected routes
		if route.Protocol != "kernel" && route.Protocol != "connected" {
			count++
		}
	}
	return count
}

func FlushRouteTable(tableID int) error {
	return exec.Command("ip", "route", "flush", "table", strconv.Itoa(tableID)).Run()
}

func AddPolicyRoute(src, dst, iface string, table int) error {
	// Add policy-based routing rule
	args := []string{"rule", "add"}
	
	if src != "" {
		args = append(args, "from", src)
	}
	
	if dst != "" {
		args = append(args, "to", dst)
	}
	
	if iface != "" {
		args = append(args, "iif", iface)
	}
	
	args = append(args, "table", strconv.Itoa(table))
	
	return exec.Command("ip", args...).Run()
}

func RemovePolicyRoute(src, dst, iface string, table int) error {
	// Remove policy-based routing rule
	args := []string{"rule", "del"}
	
	if src != "" {
		args = append(args, "from", src)
	}
	
	if dst != "" {
		args = append(args, "to", dst)
	}
	
	if iface != "" {
		args = append(args, "iif", iface)
	}
	
	args = append(args, "table", strconv.Itoa(table))
	
	return exec.Command("ip", args...).Run()
}
