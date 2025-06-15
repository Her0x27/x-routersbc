package net

import (
	"fmt"
	"os"
	"os/exec"
)

// NetplanConfigurator handles netplan-specific network configuration
type NetplanConfigurator struct {
	configPath string
}

// NewNetplanConfigurator creates a new netplan configurator
func NewNetplanConfigurator() *NetplanConfigurator {
	return &NetplanConfigurator{
		configPath: "/etc/netplan/01-routersbc.yaml",
	}
}

// GenerateConfiguration generates a complete netplan configuration
func (nc *NetplanConfigurator) GenerateConfiguration(interfaces []NetworkInterfaceConfig) error {
	config := `network:
  version: 2
  renderer: networkd
`
	
	// Add ethernet interfaces
	hasEthernet := false
	ethernetConfig := ""
	
	for _, iface := range interfaces {
		if iface.Type == "ethernet" {
			if !hasEthernet {
				ethernetConfig += "  ethernets:\n"
				hasEthernet = true
			}
			
			ethernetConfig += fmt.Sprintf("    %s:\n", iface.Name)
			
			if iface.DHCP {
				ethernetConfig += "      dhcp4: true\n"
			} else if iface.IPAddress != "" && iface.Netmask != "" {
				ethernetConfig += "      dhcp4: false\n"
				ethernetConfig += fmt.Sprintf("      addresses:\n        - %s/%s\n", 
					iface.IPAddress, nc.netmaskToCIDR(iface.Netmask))
				
				if iface.Gateway != "" {
					ethernetConfig += fmt.Sprintf("      gateway4: %s\n", iface.Gateway)
				}
				
				if len(iface.DNSServers) > 0 {
					ethernetConfig += "      nameservers:\n        addresses:\n"
					for _, dns := range iface.DNSServers {
						ethernetConfig += fmt.Sprintf("          - %s\n", dns)
					}
				}
			}
		}
	}
	
	// Add wireless interfaces
	hasWireless := false
	wirelessConfig := ""
	
	for _, iface := range interfaces {
		if iface.Type == "wireless" {
			if !hasWireless {
				wirelessConfig += "  wifis:\n"
				hasWireless = true
			}
			
			wirelessConfig += fmt.Sprintf("    %s:\n", iface.Name)
			
			if iface.DHCP {
				wirelessConfig += "      dhcp4: true\n"
			} else if iface.IPAddress != "" && iface.Netmask != "" {
				wirelessConfig += "      dhcp4: false\n"
				wirelessConfig += fmt.Sprintf("      addresses:\n        - %s/%s\n", 
					iface.IPAddress, nc.netmaskToCIDR(iface.Netmask))
			}
			
			// Add wireless configuration
			if iface.SSID != "" {
				wirelessConfig += "      access-points:\n"
				wirelessConfig += fmt.Sprintf("        \"%s\":\n", iface.SSID)
				if iface.Password != "" {
					wirelessConfig += fmt.Sprintf("          password: \"%s\"\n", iface.Password)
				}
			}
		}
	}
	
	// Combine all configurations
	fullConfig := config + ethernetConfig + wirelessConfig
	
	// Write configuration to file
	if err := os.WriteFile(nc.configPath, []byte(fullConfig), 0644); err != nil {
		return fmt.Errorf("failed to write netplan config: %v", err)
	}
	
	// Apply configuration
	cmd := exec.Command("netplan", "apply")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply netplan config: %v", err)
	}
	
	return nil
}

// NetworkInterfaceConfig represents network interface configuration
type NetworkInterfaceConfig struct {
	Name       string
	Type       string // ethernet, wireless, bridge
	DHCP       bool
	IPAddress  string
	Netmask    string
	Gateway    string
	DNSServers []string
	SSID       string
	Password   string
	Security   string
}

// netmaskToCIDR converts netmask to CIDR notation
func (nc *NetplanConfigurator) netmaskToCIDR(netmask string) string {
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
		return "24"
	}
}
