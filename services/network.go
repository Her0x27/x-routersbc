package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Her0x27/x-routersbc/utils/configurators/net"
)

type NetworkService struct {
	db *sql.DB
}

type NetworkInterface struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Config  string `json:"config"`
}

type StaticRoute struct {
	ID          int    `json:"id"`
	Interface   string `json:"interface"`
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Metric      int    `json:"metric"`
	Enabled     bool   `json:"enabled"`
}

type FirewallRule struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Chain    string `json:"chain"`
	RuleText string `json:"rule_text"`
	Enabled  bool   `json:"enabled"`
	Position int    `json:"position"`
}

type WANConfiguration struct {
	Interface     string `json:"interface"`
	ConnectionType string `json:"connection_type"`
	Config        map[string]interface{} `json:"config"`
}

type LANConfiguration struct {
	DHCPEnabled   bool   `json:"dhcp_enabled"`
	DHCPMode      string `json:"dhcp_mode"`
	DNSMode       string `json:"dns_mode"`
	BridgeEnabled bool   `json:"bridge_enabled"`
	Config        map[string]interface{} `json:"config"`
}

type WirelessConfiguration struct {
	APEnabled   bool `json:"ap_enabled"`
	STAEnabled  bool `json:"sta_enabled"`
	APConfig    map[string]interface{} `json:"ap_config"`
	STAConfig   map[string]interface{} `json:"sta_config"`
}

type FirewallConfiguration struct {
	Backend string         `json:"backend"`
	Rules   []FirewallRule `json:"rules"`
	Chains  []string       `json:"chains"`
}

func NewNetworkService(db *sql.DB) *NetworkService {
	return &NetworkService{db: db}
}

func (s *NetworkService) GetNetworkInterfaces() ([]NetworkInterface, error) {
	// Get real network interfaces from the system
	realInterfaces, err := net.GetSystemInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get system interfaces: %v", err)
	}
	
	// Get configured interfaces from database
	rows, err := s.db.Query(`
		SELECT id, name, type, enabled, config 
		FROM network_interfaces 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var interfaces []NetworkInterface
	for rows.Next() {
		var iface NetworkInterface
		if err := rows.Scan(&iface.ID, &iface.Name, &iface.Type, &iface.Enabled, &iface.Config); err != nil {
			return nil, err
		}
		interfaces = append(interfaces, iface)
	}
	
	// Merge with real system interfaces
	return s.mergeWithSystemInterfaces(interfaces, realInterfaces), nil
}

func (s *NetworkService) mergeWithSystemInterfaces(dbInterfaces []NetworkInterface, sysInterfaces []net.SystemInterface) []NetworkInterface {
	// Create a map of database interfaces by name
	dbMap := make(map[string]NetworkInterface)
	for _, iface := range dbInterfaces {
		dbMap[iface.Name] = iface
	}
	
	var merged []NetworkInterface
	
	// Add all system interfaces, merging with database config if available
	for _, sysIface := range sysInterfaces {
		if dbIface, exists := dbMap[sysIface.Name]; exists {
			// Use database configuration
			merged = append(merged, dbIface)
		} else {
			// Create new interface from system data
			merged = append(merged, NetworkInterface{
				Name:    sysIface.Name,
				Type:    sysIface.Type,
				Enabled: sysIface.IsUp,
				Config:  fmt.Sprintf(`{"ip": "%s", "mac": "%s"}`, sysIface.IP, sysIface.MAC),
			})
		}
	}
	
	return merged
}

func (s *NetworkService) CreateNetworkInterface(iface *NetworkInterface) error {
	// Apply configuration to system
	if err := net.ApplyInterfaceConfiguration(iface.Name, iface.Type, iface.Config); err != nil {
		return fmt.Errorf("failed to apply interface configuration: %v", err)
	}
	
	// Save to database
	_, err := s.db.Exec(`
		INSERT INTO network_interfaces (name, type, enabled, config) 
		VALUES (?, ?, ?, ?)
	`, iface.Name, iface.Type, iface.Enabled, iface.Config)
	
	return err
}

func (s *NetworkService) UpdateNetworkInterface(iface *NetworkInterface) error {
	// Apply configuration to system
	if err := net.ApplyInterfaceConfiguration(iface.Name, iface.Type, iface.Config); err != nil {
		return fmt.Errorf("failed to apply interface configuration: %v", err)
	}
	
	// Update database
	_, err := s.db.Exec(`
		UPDATE network_interfaces 
		SET type = ?, enabled = ?, config = ? 
		WHERE id = ?
	`, iface.Type, iface.Enabled, iface.Config, iface.ID)
	
	return err
}

func (s *NetworkService) DeleteNetworkInterface(id int) error {
	// Get interface name for system cleanup
	var name string
	err := s.db.QueryRow("SELECT name FROM network_interfaces WHERE id = ?", id).Scan(&name)
	if err != nil {
		return err
	}
	
	// Remove from system
	if err := net.RemoveInterfaceConfiguration(name); err != nil {
		return fmt.Errorf("failed to remove interface configuration: %v", err)
	}
	
	// Delete from database
	_, err = s.db.Exec("DELETE FROM network_interfaces WHERE id = ?", id)
	return err
}

func (s *NetworkService) GetWANConfiguration() (*WANConfiguration, error) {
	config, err := net.GetWANConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to get WAN configuration: %v", err)
	}
	
	return &WANConfiguration{
		Interface:      config.Interface,
		ConnectionType: config.Type,
		Config:         config.Settings,
	}, nil
}

func (s *NetworkService) GetLANConfiguration() (*LANConfiguration, error) {
	dhcpConfig, err := net.GetDHCPConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to get DHCP configuration: %v", err)
	}
	
	dnsConfig, err := net.GetDNSConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS configuration: %v", err)
	}
	
	return &LANConfiguration{
		DHCPEnabled:   dhcpConfig.Enabled,
		DHCPMode:      dhcpConfig.Mode,
		DNSMode:       dnsConfig.Mode,
		BridgeEnabled: false, // TODO: Implement bridge detection
		Config: map[string]interface{}{
			"dhcp": dhcpConfig,
			"dns":  dnsConfig,
		},
	}, nil
}

func (s *NetworkService) GetWirelessConfiguration() (*WirelessConfiguration, error) {
	// Get wireless interface configuration
	apConfig, err := net.GetWirelessAPConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to get wireless AP configuration: %v", err)
	}
	
	staConfig, err := net.GetWirelessSTAConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to get wireless STA configuration: %v", err)
	}
	
	return &WirelessConfiguration{
		APEnabled:  apConfig.Enabled,
		STAEnabled: staConfig.Enabled,
		APConfig:   apConfig.Settings,
		STAConfig:  staConfig.Settings,
	}, nil
}

func (s *NetworkService) GetStaticRoutes() ([]StaticRoute, error) {
	// Get routes from database
	rows, err := s.db.Query(`
		SELECT id, interface, destination, gateway, metric, enabled 
		FROM static_routes 
		ORDER BY destination
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var routes []StaticRoute
	for rows.Next() {
		var route StaticRoute
		if err := rows.Scan(&route.ID, &route.Interface, &route.Destination, &route.Gateway, &route.Metric, &route.Enabled); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	
	// Merge with system routes
	systemRoutes, err := net.GetSystemRoutes()
	if err != nil {
		return routes, nil // Return database routes if system routes fail
	}
	
	return s.mergeWithSystemRoutes(routes, systemRoutes), nil
}

func (s *NetworkService) mergeWithSystemRoutes(dbRoutes []StaticRoute, sysRoutes []net.SystemRoute) []StaticRoute {
	// For now, return database routes
	// TODO: Implement proper merging logic
	return dbRoutes
}

func (s *NetworkService) CreateStaticRoute(route *StaticRoute) error {
	// Apply route to system
	if err := net.AddStaticRoute(route.Destination, route.Gateway, route.Interface, route.Metric); err != nil {
		return fmt.Errorf("failed to add static route: %v", err)
	}
	
	// Save to database
	_, err := s.db.Exec(`
		INSERT INTO static_routes (interface, destination, gateway, metric, enabled) 
		VALUES (?, ?, ?, ?, ?)
	`, route.Interface, route.Destination, route.Gateway, route.Metric, route.Enabled)
	
	return err
}

func (s *NetworkService) UpdateStaticRoute(route *StaticRoute) error {
	// Get old route for removal
	var oldDest, oldGateway, oldInterface string
	var oldMetric int
	err := s.db.QueryRow(`
		SELECT destination, gateway, interface, metric 
		FROM static_routes WHERE id = ?
	`, route.ID).Scan(&oldDest, &oldGateway, &oldInterface, &oldMetric)
	
	if err != nil {
		return err
	}
	
	// Remove old route
	net.RemoveStaticRoute(oldDest, oldGateway, oldInterface, oldMetric)
	
	// Add new route
	if err := net.AddStaticRoute(route.Destination, route.Gateway, route.Interface, route.Metric); err != nil {
		return fmt.Errorf("failed to update static route: %v", err)
	}
	
	// Update database
	_, err = s.db.Exec(`
		UPDATE static_routes 
		SET interface = ?, destination = ?, gateway = ?, metric = ?, enabled = ? 
		WHERE id = ?
	`, route.Interface, route.Destination, route.Gateway, route.Metric, route.Enabled, route.ID)
	
	return err
}

func (s *NetworkService) DeleteStaticRoute(id int) error {
	// Get route details for system removal
	var dest, gateway, iface string
	var metric int
	err := s.db.QueryRow(`
		SELECT destination, gateway, interface, metric 
		FROM static_routes WHERE id = ?
	`, id).Scan(&dest, &gateway, &iface, &metric)
	
	if err != nil {
		return err
	}
	
	// Remove from system
	if err := net.RemoveStaticRoute(dest, gateway, iface, metric); err != nil {
		return fmt.Errorf("failed to remove static route: %v", err)
	}
	
	// Delete from database
	_, err = s.db.Exec("DELETE FROM static_routes WHERE id = ?", id)
	return err
}

func (s *NetworkService) GetFirewallConfiguration() (*FirewallConfiguration, error) {
	// Determine firewall backend
	backend, err := net.GetFirewallBackend()
	if err != nil {
		return nil, fmt.Errorf("failed to determine firewall backend: %v", err)
	}
	
	// Get rules from database
	rows, err := s.db.Query(`
		SELECT id, name, chain, rule_text, enabled, position 
		FROM firewall_rules 
		ORDER BY position, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var rules []FirewallRule
	for rows.Next() {
		var rule FirewallRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Chain, &rule.RuleText, &rule.Enabled, &rule.Position); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	
	// Get available chains
	chains, err := net.GetFirewallChains(backend)
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall chains: %v", err)
	}
	
	return &FirewallConfiguration{
		Backend: backend,
		Rules:   rules,
		Chains:  chains,
	}, nil
}

func (s *NetworkService) CreateFirewallRule(rule *FirewallRule) error {
	// Apply rule to system
	if err := net.AddFirewallRule(rule.Chain, rule.RuleText, rule.Position); err != nil {
		return fmt.Errorf("failed to add firewall rule: %v", err)
	}
	
	// Save to database
	_, err := s.db.Exec(`
		INSERT INTO firewall_rules (name, chain, rule_text, enabled, position) 
		VALUES (?, ?, ?, ?, ?)
	`, rule.Name, rule.Chain, rule.RuleText, rule.Enabled, rule.Position)
	
	return err
}
