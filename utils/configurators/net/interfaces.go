package net

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InterfaceConfigurator handles network interface configuration
type InterfaceConfigurator struct {
	configPath string
}

// NewInterfaceConfigurator creates a new interface configurator
func NewInterfaceConfigurator() *InterfaceConfigurator {
	return &InterfaceConfigurator{
		configPath: "/etc/network/interfaces",
	}
}

// SaveInterface saves or updates a network interface configuration
func (ic *InterfaceConfigurator) SaveInterface(name, ifaceType, ipAddress, netmask, gateway string, enabled bool) error {
	// Determine configuration method based on system
	if ic.hasNetplan() {
		return ic.saveNetplanInterface(name, ifaceType, ipAddress, netmask, gateway, enabled)
	} else {
		return ic.saveInterfacesInterface(name, ifaceType, ipAddress, netmask, gateway, enabled)
	}
}

// DeleteInterface removes a network interface configuration
func (ic *InterfaceConfigurator) DeleteInterface(name string) error {
	if ic.hasNetplan() {
		return ic.deleteNetplanInterface(name)
	} else {
		return ic.deleteInterfacesInterface(name)
	}
}

// hasNetplan checks if the system uses netplan
func (ic *InterfaceConfigurator) hasNetplan() bool {
	_, err := os.Stat("/etc/netplan")
	return err == nil
}

// saveNetplanInterface saves interface configuration using netplan
func (ic *InterfaceConfigurator) saveNetplanInterface(name, ifaceType, ipAddress, netmask, gateway string, enabled bool) error {
	configFile := "/etc/netplan/01-routersbc.yaml"
	
	// Generate netplan configuration
	config := fmt.Sprintf(`network:
  version: 2
  renderer: networkd
  ethernets:
    %s:
      dhcp4: %t
`, name, ipAddress == "")
	
	if ipAddress != "" && netmask != "" {
		config += fmt.Sprintf(`      addresses:
        - %s/%s
`, ipAddress, ic.netmaskToCIDR(netmask))
		
		if gateway != "" {
			config += fmt.Sprintf(`      gateway4: %s
`, gateway)
		}
	}
	
	// Write configuration to file
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write netplan config: %v", err)
	}
	
	// Apply configuration
	cmd := exec.Command("netplan", "apply")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply netplan config: %v", err)
	}
	
	return nil
}

// saveInterfacesInterface saves interface configuration using /etc/network/interfaces
func (ic *InterfaceConfigurator) saveInterfacesInterface(name, ifaceType, ipAddress, netmask, gateway string, enabled bool) error {
	// Read current configuration
	content := ""
	if data, err := os.ReadFile(ic.configPath); err == nil {
		content = string(data)
	}
	
	// Remove existing configuration for this interface
	lines := strings.Split(content, "\n")
	var newLines []string
	skipInterface := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check if this is the start of our interface configuration
		if strings.HasPrefix(trimmed, "auto "+name) || strings.HasPrefix(trimmed, "iface "+name) {
			skipInterface = true
			continue
		}
		
		// Check if this is the start of another interface
		if strings.HasPrefix(trimmed, "auto ") || strings.HasPrefix(trimmed, "iface ") {
			skipInterface = false
		}
		
		// Skip lines that are part of our interface configuration
		if skipInterface && (strings.HasPrefix(trimmed, "address") ||
			strings.HasPrefix(trimmed, "netmask") ||
			strings.HasPrefix(trimmed, "gateway") ||
			strings.HasPrefix(trimmed, "broadcast") ||
			trimmed == "") {
			continue
		}
		
		if !skipInterface {
			newLines = append(newLines, line)
		}
	}
	
	// Add new interface configuration
	if enabled {
		newLines = append(newLines, "")
		newLines = append(newLines, "auto "+name)
		
		if ipAddress != "" && netmask != "" {
			newLines = append(newLines, "iface "+name+" inet static")
			newLines = append(newLines, "    address "+ipAddress)
			newLines = append(newLines, "    netmask "+netmask)
			if gateway != "" {
				newLines = append(newLines, "    gateway "+gateway)
			}
		} else {
			newLines = append(newLines, "iface "+name+" inet dhcp")
		}
	}
	
	// Write updated configuration
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(ic.configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write interfaces config: %v", err)
	}
	
	// Restart networking
	cmd := exec.Command("systemctl", "restart", "networking")
	if err := cmd.Run(); err != nil {
		// Try alternative restart method
		cmd = exec.Command("service", "networking", "restart")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart networking: %v", err)
		}
	}
	
	return nil
}

// deleteNetplanInterface removes interface configuration from netplan
func (ic *InterfaceConfigurator) deleteNetplanInterface(name string) error {
	configFile := "/etc/netplan/01-routersbc.yaml"
	
	// For simplicity, regenerate the config without the interface
	// In a real implementation, you'd parse and modify the YAML
	config := `network:
  version: 2
  renderer: networkd
`
	
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write netplan config: %v", err)
	}
	
	// Apply configuration
	cmd := exec.Command("netplan", "apply")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply netplan config: %v", err)
	}
	
	return nil
}

// deleteInterfacesInterface removes interface configuration from /etc/network/interfaces
func (ic *InterfaceConfigurator) deleteInterfacesInterface(name string) error {
	// Read current configuration
	content := ""
	if data, err := os.ReadFile(ic.configPath); err == nil {
		content = string(data)
	}
	
	// Remove configuration for this interface
	lines := strings.Split(content, "\n")
	var newLines []string
	skipInterface := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check if this is the start of our interface configuration
		if strings.HasPrefix(trimmed, "auto "+name) || strings.HasPrefix(trimmed, "iface "+name) {
			skipInterface = true
			continue
		}
		
		// Check if this is the start of another interface
		if strings.HasPrefix(trimmed, "auto ") || strings.HasPrefix(trimmed, "iface ") {
			skipInterface = false
		}
		
		if !skipInterface {
			newLines = append(newLines, line)
		}
	}
	
	// Write updated configuration
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(ic.configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write interfaces config: %v", err)
	}
	
	// Restart networking
	cmd := exec.Command("systemctl", "restart", "networking")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("service", "networking", "restart")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart networking: %v", err)
		}
	}
	
	return nil
}

// netmaskToCIDR converts netmask to CIDR notation
func (ic *InterfaceConfigurator) netmaskToCIDR(netmask string) string {
	// Simple conversion for common netmasks
	switch netmask {
	case "255.255.255.0":
		return "24"
	case "255.255.0.0":
		return "16"
	case "255.0.0.0":
		return "8"
	case "255.255.255.128":
		return "25"
	case "255.255.255.192":
		return "26"
	case "255.255.255.224":
		return "27"
	case "255.255.255.240":
		return "28"
	case "255.255.255.248":
		return "29"
	case "255.255.255.252":
		return "30"
	default:
		return "24" // Default to /24
	}
}
