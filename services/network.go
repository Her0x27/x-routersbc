package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Her0x27/x-routersbc/core"
	"github.com/Her0x27/x-routersbc/utils/configurators/net"
)

// NetworkInterface represents a network interface
type NetworkInterface struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	IP      string `json:"ip"`
	MAC     string `json:"mac"`
	Config  string `json:"config"`
}

// InterfaceRequest represents a request to create/update an interface
type InterfaceRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Config  string `json:"config"`
}

// WANConfig represents WAN configuration
type WANConfig struct {
	Interface    string `json:"interface"`
	ConnectionType string `json:"connection_type"`
	IP           string `json:"ip"`
	Netmask      string `json:"netmask"`
	Gateway      string `json:"gateway"`
	DNS1         string `json:"dns1"`
	DNS2         string `json:"dns2"`
	Enabled      bool   `json:"enabled"`
}

// WANConfigRequest represents a WAN configuration request
type WANConfigRequest struct {
	Interface    string `json:"interface"`
	ConnectionType string `json:"connection_type"`
	IP           string `json:"ip"`
	Netmask      string `json:"netmask"`
	Gateway      string `json:"gateway"`
	DNS1         string `json:"dns1"`
	DNS2         string `json:"dns2"`
	Enabled      bool   `json:"enabled"`
}

// LANConfig represents LAN configuration
type LANConfig struct {
	Interface   string `json:"interface"`
	IP          string `json:"ip"`
	Netmask     string `json:"netmask"`
	DHCPEnabled bool   `json:"dhcp_enabled"`
	DHCPStart   string `json:"dhcp_start"`
	DHCPEnd     string `json:"dhcp_end"`
	DNSProxy    bool   `json:"dns_proxy"`
	DNSServers  []string `json:"dns_servers"`
}

// WirelessConfig represents wireless configuration
type WirelessConfig struct {
	Interfaces []WirelessInterface `json:"interfaces"`
}

// WirelessInterface represents a wireless interface
type WirelessInterface struct {
	Name     string            `json:"name"`
	Mode     string            `json:"mode"`
	SSID     string            `json:"ssid"`
	Channel  int               `json:"channel"`
	Security string            `json:"security"`
	Password string            `json:"password"`
	Enabled  bool              `json:"enabled"`
	Clients  []WirelessClient  `json:"clients"`
}

// WirelessClient represents a connected wireless client
type WirelessClient struct {
	MAC    string `json:"mac"`
	IP     string `json:"ip"`
	Signal int    `json:"signal"`
}

// Route represents a network route
type Route struct {
	ID          int    `json:"id"`
	Interface   string `json:"interface"`
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Metric      int    `json:"metric"`
	Enabled     bool   `json:"enabled"`
}

// RouteRequest represents a route creation request
type RouteRequest struct {
	Interface   string `json:"interface"`
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Metric      int    `json:"metric"`
	Enabled     bool   `json:"enabled"`
}

// FirewallConfig represents firewall configuration
type FirewallConfig struct {
	Backend string          `json:"backend"`
	Rules   []FirewallRule  `json:"rules"`
	Chains  []FirewallChain `json:"chains"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID       int    `json:"id"`
	Chain    string `json:"chain"`
	RuleText string `json:"rule_text"`
	Position int    `json:"position"`
	Enabled  bool   `json:"enabled"`
}

// FirewallChain represents a firewall chain
type FirewallChain struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
	Rules  int    `json:"rules"`
}

// NetworkService handles network operations
type NetworkService struct {
	db           *core.DatabaseService
	netConfigurator *net.NetworkConfigurator
}

// NewNetworkService creates a new network service
func NewNetworkService() *NetworkService {
	return &NetworkService{
		netConfigurator: net.NewNetworkConfigurator(),
	}
}

// GetInterfaces returns all network interfaces
func (s *NetworkService) GetInterfaces() ([]NetworkInterface, error) {
	interfaces := []NetworkInterface{}
	
	// Get system interfaces
	systemInterfaces, err := s.getSystemInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get system interfaces: %v", err)
	}

	// Get database interfaces
	dbInterfaces, err := s.getDatabaseInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get database interfaces: %v", err)
	}

	// Merge system and database interfaces
	interfaceMap := make(map[string]*NetworkInterface)
	
	// Add system interfaces
	for _, iface := range systemInterfaces {
		interfaceMap[iface.Name] = &iface
	}

	// Update with database information
	for _, dbIface := range dbInterfaces {
		if existing, ok := interfaceMap[dbIface.Name]; ok {
			existing.ID = dbIface.ID
			existing.Type = dbIface.Type
			existing.Enabled = dbIface.Enabled
			existing.Config = dbIface.Config
		} else {
			// Interface exists in database but not in system
			dbIface.Status = "down"
			interfaceMap[dbIface.Name] = &dbIface
		}
	}

	for _, iface := range interfaceMap {
		interfaces = append(interfaces, *iface)
	}

	return interfaces, nil
}

// getSystemInterfaces gets actual system network interfaces
func (s *NetworkService) getSystemInterfaces() ([]NetworkInterface, error) {
	interfaces := []NetworkInterface{}

	// Use ip command to get interfaces
	cmd := exec.Command("ip", "link", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ": ") && !strings.HasPrefix(line, " ") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				name := strings.Split(parts[1], "@")[0] // Remove @ and everything after
				
				// Get interface details
				status := "down"
				if strings.Contains(line, "UP") {
					status = "up"
				}

				mac := s.getMACAddress(name)
				ip := s.getIPAddress(name)

				interfaces = append(interfaces, NetworkInterface{
					Name:   name,
					Status: status,
					MAC:    mac,
					IP:     ip,
				})
			}
		}
	}

	return interfaces, nil
}

// getDatabaseInterfaces gets interfaces from database
func (s *NetworkService) getDatabaseInterfaces() ([]NetworkInterface, error) {
	interfaces := []NetworkInterface{}
	
	db := core.GetDB()
	rows, err := db.Query("SELECT id, name, type, enabled, config FROM network_interfaces")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var iface NetworkInterface
		err := rows.Scan(&iface.ID, &iface.Name, &iface.Type, &iface.Enabled, &iface.Config)
		if err != nil {
			continue
		}
		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// getMACAddress gets MAC address for an interface
func (s *NetworkService) getMACAddress(interfaceName string) string {
	cmd := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/address", interfaceName))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getIPAddress gets IP address for an interface
func (s *NetworkService) getIPAddress(interfaceName string) string {
	cmd := exec.Command("ip", "addr", "show", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 2 {
				return strings.Split(parts[1], "/")[0]
			}
		}
	}
	return ""
}

// CreateInterface creates a new network interface
func (s *NetworkService) CreateInterface(req InterfaceRequest) error {
	db := core.GetDB()
	
	// Insert into database
	query := "INSERT INTO network_interfaces (name, type, enabled, config) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, req.Name, req.Type, req.Enabled, req.Config)
	if err != nil {
		return err
	}

	// Apply configuration using appropriate configurator
	return s.netConfigurator.CreateInterface(req.Name, req.Type, req.Config)
}

// UpdateInterface updates an existing network interface
func (s *NetworkService) UpdateInterface(id int, req InterfaceRequest) error {
	db := core.GetDB()
	
	// Update database
	query := "UPDATE network_interfaces SET name = ?, type = ?, enabled = ?, config = ? WHERE id = ?"
	_, err := db.Exec(query, req.Name, req.Type, req.Enabled, req.Config, id)
	if err != nil {
		return err
	}

	// Apply configuration
	return s.netConfigurator.UpdateInterface(req.Name, req.Type, req.Config)
}

// DeleteInterface removes a network interface
func (s *NetworkService) DeleteInterface(id int) error {
	db := core.GetDB()
	
	// Get interface name first
	var name string
	err := db.QueryRow("SELECT name FROM network_interfaces WHERE id = ?", id).Scan(&name)
	if err != nil {
		return err
	}

	// Delete from database
	_, err = db.Exec("DELETE FROM network_interfaces WHERE id = ?", id)
	if err != nil {
		return err
	}

	// Remove from system
	return s.netConfigurator.DeleteInterface(name)
}

// GetWANConfig returns WAN configuration
func (s *NetworkService) GetWANConfig() (*WANConfig, error) {
	return s.netConfigurator.GetWANConfig()
}

// UpdateWANConfig updates WAN configuration
func (s *NetworkService) UpdateWANConfig(req WANConfigRequest) error {
	config := &WANConfig{
		Interface:      req.Interface,
		ConnectionType: req.ConnectionType,
		IP:             req.IP,
		Netmask:        req.Netmask,
		Gateway:        req.Gateway,
		DNS1:           req.DNS1,
		DNS2:           req.DNS2,
		Enabled:        req.Enabled,
	}
	
	return s.netConfigurator.UpdateWANConfig(config)
}

// GetLANConfig returns LAN configuration
func (s *NetworkService) GetLANConfig() (*LANConfig, error) {
	return s.netConfigurator.GetLANConfig()
}

// GetWirelessConfig returns wireless configuration
func (s *NetworkService) GetWirelessConfig() (*WirelessConfig, error) {
	return s.netConfigurator.GetWirelessConfig()
}

// GetRoutes returns routing table
func (s *NetworkService) GetRoutes() ([]Route, error) {
	routes := []Route{}
	
	db := core.GetDB()
	rows, err := db.Query("SELECT id, interface, destination, gateway, metric, enabled FROM network_routes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var route Route
		err := rows.Scan(&route.ID, &route.Interface, &route.Destination, &route.Gateway, &route.Metric, &route.Enabled)
		if err != nil {
			continue
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// CreateRoute creates a new static route
func (s *NetworkService) CreateRoute(req RouteRequest) error {
	db := core.GetDB()
	
	// Insert into database
	query := "INSERT INTO network_routes (interface, destination, gateway, metric, enabled) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, req.Interface, req.Destination, req.Gateway, req.Metric, req.Enabled)
	if err != nil {
		return err
	}

	// Apply route
	return s.netConfigurator.AddRoute(req.Interface, req.Destination, req.Gateway, req.Metric)
}

// GetFirewallConfig returns firewall configuration
func (s *NetworkService) GetFirewallConfig() (*FirewallConfig, error) {
	config := &FirewallConfig{
		Backend: "nftables", // Default to nftables
		Rules:   []FirewallRule{},
		Chains:  []FirewallChain{},
	}

	db := core.GetDB()
	
	// Get rules from database
	rows, err := db.Query("SELECT id, chain, rule_text, position, enabled FROM firewall_rules ORDER BY position")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rule FirewallRule
		err := rows.Scan(&rule.ID, &rule.Chain, &rule.RuleText, &rule.Position, &rule.Enabled)
		if err != nil {
			continue
		}
		config.Rules = append(config.Rules, rule)
	}

	// Get chain information
	config.Chains = s.getFirewallChains()

	return config, nil
}

// getFirewallChains gets firewall chain information
func (s *NetworkService) getFirewallChains() []FirewallChain {
	chains := []FirewallChain{
		{Name: "INPUT", Policy: "ACCEPT", Rules: 0},
		{Name: "OUTPUT", Policy: "ACCEPT", Rules: 0},
		{Name: "FORWARD", Policy: "ACCEPT", Rules: 0},
	}

	// Count rules per chain
	db := core.GetDB()
	rows, err := db.Query("SELECT chain, COUNT(*) FROM firewall_rules GROUP BY chain")
	if err != nil {
		return chains
	}
	defer rows.Close()

	chainMap := make(map[string]int)
	for rows.Next() {
		var chain string
		var count int
		if rows.Scan(&chain, &count) == nil {
			chainMap[chain] = count
		}
	}

	for i := range chains {
		if count, ok := chainMap[chains[i].Name]; ok {
			chains[i].Rules = count
		}
	}

	return chains
}
