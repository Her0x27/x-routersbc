package net

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// WANConfigurator handles WAN configuration
type WANConfigurator struct {
	netplanConfig     *NetplanConfigurator
	interfacesConfig  *InterfacesConfigurator
}

// NewWANConfigurator creates a new WAN configurator
func NewWANConfigurator() *WANConfigurator {
	return &WANConfigurator{
		netplanConfig:    NewNetplanConfigurator(),
		interfacesConfig: NewInterfacesConfigurator(),
	}
}

// WANConfiguration represents WAN settings
type WANConfiguration struct {
	Interface      string   `json:"interface"`
	ConnectionType string   `json:"connection_type"` // dhcp, static, pppoe
	IP             string   `json:"ip,omitempty"`
	Netmask        string   `json:"netmask,omitempty"`
	Gateway        string   `json:"gateway,omitempty"`
	DNS1           string   `json:"dns1,omitempty"`
	DNS2           string   `json:"dns2,omitempty"`
	Username       string   `json:"username,omitempty"` // For PPPoE
	Password       string   `json:"password,omitempty"` // For PPPoE
	MTU            int      `json:"mtu,omitempty"`
	Enabled        bool     `json:"enabled"`
}

// MultiWANConfiguration represents multi-WAN load balancing
type MultiWANConfiguration struct {
	Enabled     bool                  `json:"enabled"`
	Interfaces  []WANInterface        `json:"interfaces"`
	LoadBalance LoadBalanceConfig     `json:"load_balance"`
	Failover    FailoverConfig        `json:"failover"`
}

// WANInterface represents a WAN interface for multi-WAN
type WANInterface struct {
	Name     string `json:"name"`
	Weight   int    `json:"weight"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

// LoadBalanceConfig represents load balancing configuration
type LoadBalanceConfig struct {
	Method    string `json:"method"` // round-robin, weighted, least-conn
	Sticky    bool   `json:"sticky"`
	Threshold int    `json:"threshold"`
}

// FailoverConfig represents failover configuration
type FailoverConfig struct {
	Enabled       bool   `json:"enabled"`
	PingTarget    string `json:"ping_target"`
	PingInterval  int    `json:"ping_interval"`
	PingTimeout   int    `json:"ping_timeout"`
	PingFailures  int    `json:"ping_failures"`
	RecoveryDelay int    `json:"recovery_delay"`
}

// GetWANConfiguration gets current WAN configuration
func (w *WANConfigurator) GetWANConfiguration() (*WANConfiguration, error) {
	// Try to read from system configuration
	config := &WANConfiguration{
		Interface:      "eth0",
		ConnectionType: "dhcp",
		Enabled:        true,
	}

	// Check if we have a saved configuration
	if data, err := os.ReadFile("/etc/routersbc/wan.json"); err == nil {
		if err := json.Unmarshal(data, config); err == nil {
			return config, nil
		}
	}

	// Try to detect current configuration from system
	return w.detectCurrentWANConfig()
}

// detectCurrentWANConfig detects current WAN configuration from system
func (w *WANConfigurator) detectCurrentWANConfig() (*WANConfiguration, error) {
	config := &WANConfiguration{
		Interface:      "eth0",
		ConnectionType: "dhcp",
		Enabled:        true,
	}

	// Get default route to determine WAN interface
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err == nil {
		// Parse output to get interface
		// Format: default via 192.168.1.1 dev eth0 proto dhcp src 192.168.1.100 metric 100
		parts := parseRouteOutput(string(output))
		if len(parts) > 0 {
			config.Interface = parts["dev"]
			config.Gateway = parts["via"]
		}
	}

	// Get interface IP configuration
	cmd = exec.Command("ip", "addr", "show", config.Interface)
	output, err = cmd.Output()
	if err == nil {
		ipInfo := parseIPOutput(string(output))
		config.IP = ipInfo["ip"]
		config.Netmask = ipInfo["netmask"]
	}

	// Check if interface uses DHCP
	if w.interfaceUsesDHCP(config.Interface) {
		config.ConnectionType = "dhcp"
	} else {
		config.ConnectionType = "static"
	}

	return config, nil
}

// SetWANConfiguration applies WAN configuration
func (w *WANConfigurator) SetWANConfiguration(config *WANConfiguration) error {
	// Save configuration
	if err := w.saveWANConfig(config); err != nil {
		return fmt.Errorf("failed to save WAN config: %v", err)
	}

	// Apply configuration based on available system
	if w.netplanConfig.IsAvailable() {
		return w.configureWANNetplan(config)
	} else if w.interfacesConfig.IsAvailable() {
		return w.configureWANInterfaces(config)
	}

	return fmt.Errorf("no supported network configuration system found")
}

// configureWANNetplan configures WAN using netplan
func (w *WANConfigurator) configureWANNetplan(config *WANConfiguration) error {
	netplanConfig, err := w.netplanConfig.GetConfiguration()
	if err != nil {
		return err
	}

	if netplanConfig.Network.Ethernets == nil {
		netplanConfig.Network.Ethernets = make(map[string]NetplanEthernet)
	}

	var ethernet NetplanEthernet

	switch config.ConnectionType {
	case "dhcp":
		ethernet = NetplanEthernet{
			DHCP4: true,
		}
	case "static":
		cidr := fmt.Sprintf("%s/%s", config.IP, config.Netmask)
		dnsServers := []string{}
		if config.DNS1 != "" {
			dnsServers = append(dnsServers, config.DNS1)
		}
		if config.DNS2 != "" {
			dnsServers = append(dnsServers, config.DNS2)
		}

		ethernet = NetplanEthernet{
			DHCP4:     false,
			Addresses: []string{cidr},
			Gateway4:  config.Gateway,
			Nameservers: NetplanNameservers{
				Addresses: dnsServers,
			},
		}
	case "pppoe":
		// PPPoE configuration would require additional setup
		return fmt.Errorf("PPPoE configuration not yet implemented")
	}

	netplanConfig.Network.Ethernets[config.Interface] = ethernet
	return w.netplanConfig.WriteConfiguration(netplanConfig)
}

// configureWANInterfaces configures WAN using /etc/network/interfaces
func (w *WANConfigurator) configureWANInterfaces(config *WANConfiguration) error {
	switch config.ConnectionType {
	case "dhcp":
		return w.interfacesConfig.EnableDHCP(config.Interface)
	case "static":
		dnsServers := []string{}
		if config.DNS1 != "" {
			dnsServers = append(dnsServers, config.DNS1)
		}
		if config.DNS2 != "" {
			dnsServers = append(dnsServers, config.DNS2)
		}
		return w.interfacesConfig.SetStaticIP(config.Interface, config.IP, config.Netmask, config.Gateway, dnsServers)
	case "pppoe":
		return fmt.Errorf("PPPoE configuration not yet implemented")
	}

	return fmt.Errorf("unsupported connection type: %s", config.ConnectionType)
}

// saveWANConfig saves WAN configuration to file
func (w *WANConfigurator) saveWANConfig(config *WANConfiguration) error {
	// Ensure config directory exists
	if err := os.MkdirAll("/etc/routersbc", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("/etc/routersbc/wan.json", data, 0644)
}

// interfaceUsesDHCP checks if an interface is configured for DHCP
func (w *WANConfigurator) interfaceUsesDHCP(interfaceName string) bool {
	// Check if DHCP client is running for this interface
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("dhclient.*%s", interfaceName))
	return cmd.Run() == nil
}

// GetMultiWANConfiguration gets multi-WAN configuration
func (w *WANConfigurator) GetMultiWANConfiguration() (*MultiWANConfiguration, error) {
	config := &MultiWANConfiguration{
		Enabled: false,
		LoadBalance: LoadBalanceConfig{
			Method:    "round-robin",
			Sticky:    false,
			Threshold: 80,
		},
		Failover: FailoverConfig{
			Enabled:       true,
			PingTarget:    "8.8.8.8",
			PingInterval:  30,
			PingTimeout:   5,
			PingFailures:  3,
			RecoveryDelay: 60,
		},
	}

	// Try to load from configuration file
	if data, err := os.ReadFile("/etc/routersbc/multiwan.json"); err == nil {
		json.Unmarshal(data, config)
	}

	return config, nil
}

// SetMultiWANConfiguration sets multi-WAN configuration
func (w *WANConfigurator) SetMultiWANConfiguration(config *MultiWANConfiguration) error {
	// Save configuration
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll("/etc/routersbc", 0755); err != nil {
		return err
	}

	if err := os.WriteFile("/etc/routersbc/multiwan.json", data, 0644); err != nil {
		return err
	}

	// Apply configuration
	if config.Enabled {
		return w.setupMultiWAN(config)
	} else {
		return w.disableMultiWAN()
	}
}

// setupMultiWAN sets up multi-WAN routing
func (w *WANConfigurator) setupMultiWAN(config *MultiWANConfiguration) error {
	// This would involve complex routing table manipulation
	// Implementation would depend on the specific multi-WAN setup
	
	// Create routing tables for each WAN interface
	for i, iface := range config.Interfaces {
		if !iface.Enabled {
			continue
		}

		tableID := 100 + i
		tableName := fmt.Sprintf("wan%d", i+1)

		// Add routing table to /etc/iproute2/rt_tables if not exists
		if err := w.addRoutingTable(tableID, tableName); err != nil {
			return err
		}

		// Configure routing rules
		if err := w.configureWANRoutingTable(iface.Name, tableID); err != nil {
			return err
		}
	}

	// Set up load balancing
	return w.configureLoadBalancing(config)
}

// disableMultiWAN disables multi-WAN configuration
func (w *WANConfigurator) disableMultiWAN() error {
	// Remove custom routing tables and rules
	// Restore single WAN configuration
	return nil
}

// addRoutingTable adds a custom routing table
func (w *WANConfigurator) addRoutingTable(tableID int, tableName string) error {
	// Check if table already exists
	cmd := exec.Command("grep", "-q", fmt.Sprintf("%d.*%s", tableID, tableName), "/etc/iproute2/rt_tables")
	if cmd.Run() == nil {
		return nil // Table already exists
	}

	// Add table
	tableEntry := fmt.Sprintf("%d\t%s\n", tableID, tableName)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("echo '%s' >> /etc/iproute2/rt_tables", tableEntry))
	return cmd.Run()
}

// configureWANRoutingTable configures routing table for a WAN interface
func (w *WANConfigurator) configureWANRoutingTable(interfaceName string, tableID int) error {
	// Get interface gateway
	gateway, err := w.getInterfaceGateway(interfaceName)
	if err != nil {
		return err
	}

	// Add default route to custom table
	cmd := exec.Command("ip", "route", "add", "default", "via", gateway, "dev", interfaceName, "table", fmt.Sprintf("%d", tableID))
	return cmd.Run()
}

// configureLoadBalancing sets up load balancing between WAN interfaces
func (w *WANConfigurator) configureLoadBalancing(config *MultiWANConfiguration) error {
	// Remove existing multipath route
	exec.Command("ip", "route", "del", "default").Run()

	// Build multipath route command
	var nexthops []string
	for _, iface := range config.Interfaces {
		if !iface.Enabled {
			continue
		}

		gateway, err := w.getInterfaceGateway(iface.Name)
		if err != nil {
			continue
		}

		nexthop := fmt.Sprintf("nexthop via %s dev %s weight %d", gateway, iface.Name, iface.Weight)
		nexthops = append(nexthops, nexthop)
	}

	if len(nexthops) == 0 {
		return fmt.Errorf("no enabled WAN interfaces found")
	}

	// Create multipath route
	args := append([]string{"route", "add", "default"}, nexthops...)
	cmd := exec.Command("ip", args...)
	return cmd.Run()
}

// getInterfaceGateway gets the gateway for an interface
func (w *WANConfigurator) getInterfaceGateway(interfaceName string) (string, error) {
	cmd := exec.Command("ip", "route", "show", "dev", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	routes := parseRouteOutput(string(output))
	if gateway, ok := routes["via"]; ok {
		return gateway, nil
	}

	return "", fmt.Errorf("no gateway found for interface %s", interfaceName)
}

// Helper functions for parsing command output

func parseRouteOutput(output string) map[string]string {
	result := make(map[string]string)
	parts := splitFields(output)
	
	for i, part := range parts {
		switch part {
		case "via":
			if i+1 < len(parts) {
				result["via"] = parts[i+1]
			}
		case "dev":
			if i+1 < len(parts) {
				result["dev"] = parts[i+1]
			}
		case "src":
			if i+1 < len(parts) {
				result["src"] = parts[i+1]
			}
		}
	}
	
	return result
}

func parseIPOutput(output string) map[string]string {
	result := make(map[string]string)
	lines := splitLines(output)
	
	for _, line := range lines {
		if containsString(line, "inet ") && !containsString(line, "inet6") {
			parts := splitFields(trimSpace(line))
			if len(parts) >= 2 {
				ipCidr := parts[1]
				ipParts := splitString(ipCidr, "/")
				if len(ipParts) >= 2 {
					result["ip"] = ipParts[0]
					result["netmask"] = ipParts[1]
				}
			}
		}
	}
	
	return result
}

// Helper functions (simplified implementations)
func splitFields(s string) []string {
	return strings.Fields(s)
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func splitString(s, sep string) []string {
	return strings.Split(s, sep)
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
