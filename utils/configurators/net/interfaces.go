package net

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// InterfacesConfigurator handles /etc/network/interfaces configuration
type InterfacesConfigurator struct {
	configPath string
}

// NewInterfacesConfigurator creates a new interfaces configurator
func NewInterfacesConfigurator() *InterfacesConfigurator {
	return &InterfacesConfigurator{
		configPath: "/etc/network/interfaces",
	}
}

// IsAvailable checks if /etc/network/interfaces is available
func (i *InterfacesConfigurator) IsAvailable() bool {
	_, err := os.Stat(i.configPath)
	return !os.IsNotExist(err)
}

// Interface represents a network interface configuration
type Interface struct {
	Name      string
	Family    string // inet, inet6
	Method    string // dhcp, static, manual
	Address   string
	Netmask   string
	Gateway   string
	DNS       []string
	PreUp     []string
	PostUp    []string
	PreDown   []string
	PostDown  []string
	Options   map[string]string
}

// GetInterfaces reads all interfaces from /etc/network/interfaces
func (i *InterfacesConfigurator) GetInterfaces() ([]Interface, error) {
	file, err := os.Open(i.configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var interfaces []Interface
	var currentInterface *Interface
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "auto":
			// Auto interface declaration
			continue
		case "allow-hotplug":
			// Hotplug interface declaration
			continue
		case "iface":
			// Save previous interface
			if currentInterface != nil {
				interfaces = append(interfaces, *currentInterface)
			}
			
			// New interface definition
			if len(parts) >= 4 {
				currentInterface = &Interface{
					Name:    parts[1],
					Family:  parts[2],
					Method:  parts[3],
					Options: make(map[string]string),
				}
			}
		default:
			if currentInterface != nil {
				// Interface options
				if len(parts) >= 2 {
					key := parts[0]
					value := strings.Join(parts[1:], " ")
					
					switch key {
					case "address":
						currentInterface.Address = value
					case "netmask":
						currentInterface.Netmask = value
					case "gateway":
						currentInterface.Gateway = value
					case "dns-nameservers":
						currentInterface.DNS = parts[1:]
					case "pre-up":
						currentInterface.PreUp = append(currentInterface.PreUp, value)
					case "post-up", "up":
						currentInterface.PostUp = append(currentInterface.PostUp, value)
					case "pre-down":
						currentInterface.PreDown = append(currentInterface.PreDown, value)
					case "post-down", "down":
						currentInterface.PostDown = append(currentInterface.PostDown, value)
					default:
						currentInterface.Options[key] = value
					}
				}
			}
		}
	}

	// Add last interface
	if currentInterface != nil {
		interfaces = append(interfaces, *currentInterface)
	}

	return interfaces, scanner.Err()
}

// WriteInterfaces writes interfaces configuration to file
func (i *InterfacesConfigurator) WriteInterfaces(interfaces []Interface) error {
	var content strings.Builder
	
	content.WriteString("# This file describes the network interfaces available on your system\n")
	content.WriteString("# and how to activate them. For more information, see interfaces(5).\n\n")
	content.WriteString("source /etc/network/interfaces.d/*\n\n")
	content.WriteString("# The loopback network interface\n")
	content.WriteString("auto lo\n")
	content.WriteString("iface lo inet loopback\n\n")

	for _, iface := range interfaces {
		if iface.Name == "lo" {
			continue // Skip loopback, already handled
		}

		content.WriteString(fmt.Sprintf("# Interface %s\n", iface.Name))
		content.WriteString(fmt.Sprintf("auto %s\n", iface.Name))
		content.WriteString(fmt.Sprintf("iface %s %s %s\n", iface.Name, iface.Family, iface.Method))

		if iface.Address != "" {
			content.WriteString(fmt.Sprintf("    address %s\n", iface.Address))
		}
		if iface.Netmask != "" {
			content.WriteString(fmt.Sprintf("    netmask %s\n", iface.Netmask))
		}
		if iface.Gateway != "" {
			content.WriteString(fmt.Sprintf("    gateway %s\n", iface.Gateway))
		}
		if len(iface.DNS) > 0 {
			content.WriteString(fmt.Sprintf("    dns-nameservers %s\n", strings.Join(iface.DNS, " ")))
		}

		for _, cmd := range iface.PreUp {
			content.WriteString(fmt.Sprintf("    pre-up %s\n", cmd))
		}
		for _, cmd := range iface.PostUp {
			content.WriteString(fmt.Sprintf("    post-up %s\n", cmd))
		}
		for _, cmd := range iface.PreDown {
			content.WriteString(fmt.Sprintf("    pre-down %s\n", cmd))
		}
		for _, cmd := range iface.PostDown {
			content.WriteString(fmt.Sprintf("    post-down %s\n", cmd))
		}

		for key, value := range iface.Options {
			content.WriteString(fmt.Sprintf("    %s %s\n", key, value))
		}

		content.WriteString("\n")
	}

	return ioutil.WriteFile(i.configPath, []byte(content.String()), 0644)
}

// ConfigureInterface configures a specific interface
func (i *InterfacesConfigurator) ConfigureInterface(name, interfaceType, configJSON string) error {
	interfaces, err := i.GetInterfaces()
	if err != nil {
		return err
	}

	// Find or create interface
	var targetInterface *Interface
	for idx := range interfaces {
		if interfaces[idx].Name == name {
			targetInterface = &interfaces[idx]
			break
		}
	}

	if targetInterface == nil {
		// Create new interface
		newInterface := Interface{
			Name:    name,
			Family:  "inet",
			Method:  "dhcp",
			Options: make(map[string]string),
		}
		interfaces = append(interfaces, newInterface)
		targetInterface = &interfaces[len(interfaces)-1]
	}

	// Configure based on type
	switch interfaceType {
	case "static":
		targetInterface.Method = "static"
		// TODO: Parse configJSON to set IP, netmask, gateway
	case "dhcp":
		targetInterface.Method = "dhcp"
	case "bridge":
		targetInterface.Method = "static"
		targetInterface.Options["bridge_ports"] = "none"
		targetInterface.Options["bridge_stp"] = "off"
		targetInterface.Options["bridge_fd"] = "0"
	}

	return i.WriteInterfaces(interfaces)
}

// RemoveInterface removes an interface configuration
func (i *InterfacesConfigurator) RemoveInterface(name string) error {
	interfaces, err := i.GetInterfaces()
	if err != nil {
		return err
	}

	// Filter out the interface
	var filteredInterfaces []Interface
	for _, iface := range interfaces {
		if iface.Name != name {
			filteredInterfaces = append(filteredInterfaces, iface)
		}
	}

	return i.WriteInterfaces(filteredInterfaces)
}

// SetStaticIP sets static IP configuration for an interface
func (i *InterfacesConfigurator) SetStaticIP(interfaceName, ip, netmask, gateway string, dnsServers []string) error {
	interfaces, err := i.GetInterfaces()
	if err != nil {
		return err
	}

	// Find or create interface
	var targetInterface *Interface
	for idx := range interfaces {
		if interfaces[idx].Name == interfaceName {
			targetInterface = &interfaces[idx]
			break
		}
	}

	if targetInterface == nil {
		// Create new interface
		newInterface := Interface{
			Name:    interfaceName,
			Family:  "inet",
			Options: make(map[string]string),
		}
		interfaces = append(interfaces, newInterface)
		targetInterface = &interfaces[len(interfaces)-1]
	}

	targetInterface.Method = "static"
	targetInterface.Address = ip
	targetInterface.Netmask = netmask
	targetInterface.Gateway = gateway
	targetInterface.DNS = dnsServers

	return i.WriteInterfaces(interfaces)
}

// EnableDHCP enables DHCP for an interface
func (i *InterfacesConfigurator) EnableDHCP(interfaceName string) error {
	interfaces, err := i.GetInterfaces()
	if err != nil {
		return err
	}

	// Find or create interface
	var targetInterface *Interface
	for idx := range interfaces {
		if interfaces[idx].Name == interfaceName {
			targetInterface = &interfaces[idx]
			break
		}
	}

	if targetInterface == nil {
		// Create new interface
		newInterface := Interface{
			Name:    interfaceName,
			Family:  "inet",
			Options: make(map[string]string),
		}
		interfaces = append(interfaces, newInterface)
		targetInterface = &interfaces[len(interfaces)-1]
	}

	targetInterface.Method = "dhcp"
	targetInterface.Address = ""
	targetInterface.Netmask = ""
	targetInterface.Gateway = ""
	targetInterface.DNS = nil

	return i.WriteInterfaces(interfaces)
}

// RestartNetworking restarts network service
func (i *InterfacesConfigurator) RestartNetworking() error {
	// Try systemctl first
	if err := exec.Command("systemctl", "restart", "networking").Run(); err != nil {
		// Fallback to service command
		return exec.Command("service", "networking", "restart").Run()
	}
	return nil
}

// BringInterfaceUp brings an interface up
func (i *InterfacesConfigurator) BringInterfaceUp(name string) error {
	return exec.Command("ifup", name).Run()
}

// BringInterfaceDown brings an interface down
func (i *InterfacesConfigurator) BringInterfaceDown(name string) error {
	return exec.Command("ifdown", name).Run()
}
