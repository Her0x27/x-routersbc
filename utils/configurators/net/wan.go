package net

import (
	"fmt"
	"os/exec"
	"strings"
)

type WANConfig struct {
	Interface string                 `json:"interface"`
	Type      string                 `json:"type"`
	Settings  map[string]interface{} `json:"settings"`
}

func GetWANConfiguration() (*WANConfig, error) {
	// Get default route to determine WAN interface
	output, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return &WANConfig{
			Interface: "",
			Type:      "none",
			Settings:  make(map[string]interface{}),
		}, nil
	}
	
	// Parse default route output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "default via") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "dev" && i+1 < len(parts) {
					interface_name := parts[i+1]
					
					// Determine connection type
					connType := "wired"
					if strings.HasPrefix(interface_name, "wl") {
						connType = "wireless"
					}
					
					return &WANConfig{
						Interface: interface_name,
						Type:      connType,
						Settings: map[string]interface{}{
							"gateway": getGatewayFromRoute(parts),
							"method":  "dhcp", // Default assumption
						},
					}, nil
				}
			}
		}
	}
	
	return &WANConfig{
		Interface: "",
		Type:      "none",
		Settings:  make(map[string]interface{}),
	}, nil
}

func getGatewayFromRoute(routeParts []string) string {
	for i, part := range routeParts {
		if part == "via" && i+1 < len(routeParts) {
			return routeParts[i+1]
		}
	}
	return ""
}

func SetWANInterface(interfaceName, connectionType string) error {
	// Validate interface exists
	interfaces, err := GetSystemInterfaces()
	if err != nil {
		return fmt.Errorf("failed to get system interfaces: %v", err)
	}
	
	found := false
	for _, iface := range interfaces {
		if iface.Name == interfaceName {
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("interface %s not found", interfaceName)
	}
	
	// Configure interface for WAN use
	switch connectionType {
	case "dhcp":
		return configureWANDHCP(interfaceName)
	case "static":
		return fmt.Errorf("static WAN configuration not implemented yet")
	case "pppoe":
		return fmt.Errorf("PPPoE WAN configuration not implemented yet")
	default:
		return fmt.Errorf("unsupported WAN connection type: %s", connectionType)
	}
}

func configureWANDHCP(interfaceName string) error {
	// Configure interface for DHCP
	if IsNetplanAvailable() {
		return AddEthernetInterface(interfaceName, EthernetConfig{
			DHCP4: true,
			DHCP6: false,
		})
	}
	
	// Fallback to traditional configuration
	return applyTraditionalInterfaceConfig(interfaceName, "ethernet", `{"method": "dhcp"}`)
}

func GetAvailableWANInterfaces() ([]SystemInterface, error) {
	interfaces, err := GetSystemInterfaces()
	if err != nil {
		return nil, err
	}
	
	var wanInterfaces []SystemInterface
	for _, iface := range interfaces {
		// Filter interfaces that can be used as WAN
		if iface.Type == "ethernet" || iface.Type == "wifi" {
			wanInterfaces = append(wanInterfaces, iface)
		}
	}
	
	return wanInterfaces, nil
}

func EnableMultiWAN(interfaces []string, method string) error {
	switch method {
	case "load_balance":
		return configureLoadBalancing(interfaces)
	case "failover":
		return configureFailover(interfaces)
	default:
		return fmt.Errorf("unsupported multi-WAN method: %s", method)
	}
}

func configureLoadBalancing(interfaces []string) error {
	// This would require advanced routing configuration
	// For now, return not implemented
	return fmt.Errorf("load balancing not implemented yet")
}

func configureFailover(interfaces []string) error {
	// This would require failover routing configuration
	// For now, return not implemented
	return fmt.Errorf("failover not implemented yet")
}

func GetWANStatus() (map[string]interface{}, error) {
	config, err := GetWANConfiguration()
	if err != nil {
		return nil, err
	}
	
	status := map[string]interface{}{
		"interface": config.Interface,
		"type":      config.Type,
		"connected": false,
		"ip":        "",
		"gateway":   "",
		"dns":       []string{},
	}
	
	if config.Interface != "" {
		// Check if interface is up and has IP
		interfaces, err := GetSystemInterfaces()
		if err == nil {
			for _, iface := range interfaces {
				if iface.Name == config.Interface {
					status["connected"] = iface.IsUp && iface.IP != ""
					status["ip"] = iface.IP
					break
				}
			}
		}
		
		// Get gateway from config
		if gateway, ok := config.Settings["gateway"].(string); ok {
			status["gateway"] = gateway
		}
		
		// Get DNS servers
		if dns := getCurrentDNSServers(); len(dns) > 0 {
			status["dns"] = dns
		}
	}
	
	return status, nil
}

func getCurrentDNSServers() []string {
	// Read DNS servers from resolv.conf
	output, err := exec.Command("cat", "/etc/resolv.conf").Output()
	if err != nil {
		return []string{}
	}
	
	var servers []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				servers = append(servers, parts[1])
			}
		}
	}
	
	return servers
}
