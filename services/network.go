package services

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/Her0x27/x-routersbc/utils/configurators/net"
)

// NetworkService handles network operations
type NetworkService struct {
	interfaceConfigurator *net.InterfaceConfigurator
	firewallConfigurator  *net.FirewallConfigurator
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Enabled   bool   `json:"enabled"`
	IPAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
	MAC       string `json:"mac"`
	Status    string `json:"status"`
}

// WANConfiguration represents WAN settings
type WANConfiguration struct {
	Interface string `json:"interface"`
	Type      string `json:"type"` // wire/wireless
	DHCP      bool   `json:"dhcp"`
	IPAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
	DNS1      string `json:"dns1"`
	DNS2      string `json:"dns2"`
}

// LANConfiguration represents LAN settings
type LANConfiguration struct {
	Interface   string `json:"interface"`
	IPAddress   string `json:"ip_address"`
	Netmask     string `json:"netmask"`
	DHCP        string `json:"dhcp"` // relay/server/disabled
	DHCPStart   string `json:"dhcp_start"`
	DHCPEnd     string `json:"dhcp_end"`
	LeaseTime   string `json:"lease_time"`
	DNS         string `json:"dns"` // direct/proxy/forward/server
	DNSServers  []string `json:"dns_servers"`
	LocalZones  []DNSZone `json:"local_zones"`
}

// WirelessConfiguration represents wireless settings
type WirelessConfiguration struct {
	Interfaces []WirelessInterface `json:"interfaces"`
}

// WirelessInterface represents a wireless interface
type WirelessInterface struct {
	Name     string `json:"name"`
	Mode     string `json:"mode"` // AP/STA/ADHOC/MONITOR
	SSID     string `json:"ssid"`
	Security string `json:"security"`
	Password string `json:"password"`
	Channel  int    `json:"channel"`
	Enabled  bool   `json:"enabled"`
}

// StaticRoute represents a static route
type StaticRoute struct {
	ID          int    `json:"id"`
	Interface   string `json:"interface"`
	Destination string `json:"destination"`
	Netmask     string `json:"netmask"`
	Gateway     string `json:"gateway"`
	Metric      int    `json:"metric"`
}

// FirewallConfiguration represents firewall settings
type FirewallConfiguration struct {
	Backend string         `json:"backend"` // nftables/iptables
	Rules   []FirewallRule `json:"rules"`
	Chains  []FirewallChain `json:"chains"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Chain       string `json:"chain"`
	Action      string `json:"action"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Port        string `json:"port"`
	Enabled     bool   `json:"enabled"`
}

// FirewallChain represents a firewall chain
type FirewallChain struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
	Table  string `json:"table"`
}

// DNSZone represents a DNS zone
type DNSZone struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

// InterfaceRequest represents a request to create/update an interface
type InterfaceRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Enabled   bool   `json:"enabled"`
	IPAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
}

// NewNetworkService creates a new network service
func NewNetworkService() *NetworkService {
	return &NetworkService{
		interfaceConfigurator: net.NewInterfaceConfigurator(),
		firewallConfigurator:  net.NewFirewallConfigurator(),
	}
}

// GetNetworkInterfaces gets all network interfaces
func (s *NetworkService) GetNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}
	
	var result []NetworkInterface
	for _, iface := range interfaces {
		netInterface := NetworkInterface{
			Name:   iface.Name,
			MAC:    iface.HardwareAddr.String(),
			Status: "down",
		}
		
		// Get interface status
		if iface.Flags&net.FlagUp != 0 {
			netInterface.Status = "up"
		}
		
		// Get IP addresses
		addrs, err := iface.Addrs()
		if err == nil && len(addrs) > 0 {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						netInterface.IPAddress = ipnet.IP.String()
						netInterface.Netmask = ipnet.Mask.String()
						break
					}
				}
			}
		}
		
		// Determine interface type
		netInterface.Type = s.getInterfaceType(iface.Name)
		netInterface.Enabled = netInterface.Status == "up"
		
		result = append(result, netInterface)
	}
	
	return result, nil
}

// getInterfaceType determines the type of interface
func (s *NetworkService) getInterfaceType(name string) string {
	if strings.HasPrefix(name, "eth") {
		return "ethernet"
	} else if strings.HasPrefix(name, "wlan") {
		return "wireless"
	} else if strings.HasPrefix(name, "lo") {
		return "loopback"
	} else if strings.HasPrefix(name, "br") {
		return "bridge"
	} else if strings.Contains(name, "vlan") {
		return "vlan"
	} else if strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "tap") {
		return "vpn"
	}
	return "unknown"
}

// GetWANConfiguration gets WAN configuration
func (s *NetworkService) GetWANConfiguration() (*WANConfiguration, error) {
	// Read current WAN configuration from system
	config := &WANConfiguration{
		Interface: "eth0", // Default WAN interface
		Type:      "wire",
		DHCP:      true,
	}
	
	// Try to get actual configuration from /etc/network/interfaces or netplan
	return config, nil
}

// GetLANConfiguration gets LAN configuration
func (s *NetworkService) GetLANConfiguration() (*LANConfiguration, error) {
	config := &LANConfiguration{
		Interface: "br0",
		IPAddress: "192.168.1.1",
		Netmask:   "255.255.255.0",
		DHCP:      "server",
		DHCPStart: "192.168.1.100",
		DHCPEnd:   "192.168.1.200",
		LeaseTime: "24h",
		DNS:       "server",
		DNSServers: []string{"8.8.8.8", "8.8.4.4"},
		LocalZones: []DNSZone{
			{Domain: "router.local", Type: "A", Value: "192.168.1.1"},
		},
	}
	
	return config, nil
}

// GetWirelessConfiguration gets wireless configuration
func (s *NetworkService) GetWirelessConfiguration() (*WirelessConfiguration, error) {
	config := &WirelessConfiguration{
		Interfaces: []WirelessInterface{},
	}
	
	// Get wireless interfaces
	interfaces, err := s.GetNetworkInterfaces()
	if err != nil {
		return nil, err
	}
	
	for _, iface := range interfaces {
		if iface.Type == "wireless" {
			wifiInterface := WirelessInterface{
				Name:     iface.Name,
				Mode:     "AP",
				SSID:     "RouterSBC",
				Security: "WPA2",
				Channel:  6,
				Enabled:  iface.Enabled,
			}
			config.Interfaces = append(config.Interfaces, wifiInterface)
		}
	}
	
	return config, nil
}

// GetStaticRoutes gets static routes
func (s *NetworkService) GetStaticRoutes() ([]StaticRoute, error) {
	var routes []StaticRoute
	
	// Execute 'route -n' or 'ip route' to get current routes
	cmd := exec.Command("ip", "route", "show")
	output, err := cmd.Output()
	if err != nil {
		return routes, fmt.Errorf("failed to get routes: %v", err)
	}
	
	// Parse route output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	id := 1
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		
		if len(fields) >= 3 && !strings.Contains(line, "dev lo") {
			route := StaticRoute{
				ID:          id,
				Destination: fields[0],
				Gateway:     "",
				Interface:   "",
				Metric:      0,
			}
			
			// Parse route fields
			for i, field := range fields {
				switch field {
				case "via":
					if i+1 < len(fields) {
						route.Gateway = fields[i+1]
					}
				case "dev":
					if i+1 < len(fields) {
						route.Interface = fields[i+1]
					}
				case "metric":
					if i+1 < len(fields) {
						fmt.Sscanf(fields[i+1], "%d", &route.Metric)
					}
				}
			}
			
			routes = append(routes, route)
			id++
		}
	}
	
	return routes, nil
}

// GetFirewallConfiguration gets firewall configuration
func (s *NetworkService) GetFirewallConfiguration() (*FirewallConfiguration, error) {
	config := &FirewallConfiguration{
		Backend: s.getFirewallBackend(),
		Rules:   []FirewallRule{},
		Chains:  []FirewallChain{},
	}
	
	// Get firewall rules based on backend
	if config.Backend == "nftables" {
		return s.getNFTablesConfiguration()
	} else {
		return s.getIPTablesConfiguration()
	}
}

// getFirewallBackend determines which firewall backend is in use
func (s *NetworkService) getFirewallBackend() string {
	// Check if nftables is available and in use
	if _, err := exec.LookPath("nft"); err == nil {
		cmd := exec.Command("nft", "list", "tables")
		if err := cmd.Run(); err == nil {
			return "nftables"
		}
	}
	
	// Default to iptables
	return "iptables"
}

// getNFTablesConfiguration gets nftables configuration
func (s *NetworkService) getNFTablesConfiguration() (*FirewallConfiguration, error) {
	config := &FirewallConfiguration{
		Backend: "nftables",
		Rules:   []FirewallRule{},
		Chains:  []FirewallChain{},
	}
	
	// Execute nft list ruleset
	cmd := exec.Command("nft", "list", "ruleset")
	output, err := cmd.Output()
	if err != nil {
		return config, nil // Return empty config if nftables not configured
	}
	
	// Parse nftables output (simplified)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "chain ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				config.Chains = append(config.Chains, FirewallChain{
					Name:   parts[1],
					Policy: "accept",
					Table:  "filter",
				})
			}
		}
	}
	
	return config, nil
}

// getIPTablesConfiguration gets iptables configuration
func (s *NetworkService) getIPTablesConfiguration() (*FirewallConfiguration, error) {
	config := &FirewallConfiguration{
		Backend: "iptables",
		Rules:   []FirewallRule{},
		Chains:  []FirewallChain{},
	}
	
	// Execute iptables -L
	cmd := exec.Command("iptables", "-L", "-n")
	output, err := cmd.Output()
	if err != nil {
		return config, nil // Return empty config if iptables not available
	}
	
	// Parse iptables output (simplified)
	lines := strings.Split(string(output), "\n")
	ruleID := 1
	currentChain := ""
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chain ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				currentChain = parts[1]
				policy := strings.Trim(parts[2], "()")
				config.Chains = append(config.Chains, FirewallChain{
					Name:   currentChain,
					Policy: policy,
					Table:  "filter",
				})
			}
		} else if len(line) > 0 && !strings.HasPrefix(line, "target") && currentChain != "" {
			// Parse rule line
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				rule := FirewallRule{
					ID:      ruleID,
					Name:    fmt.Sprintf("Rule %d", ruleID),
					Chain:   currentChain,
					Action:  fields[0],
					Enabled: true,
				}
				
				if len(fields) > 1 {
					rule.Protocol = fields[1]
				}
				if len(fields) > 2 {
					rule.Source = fields[2]
				}
				if len(fields) > 3 {
					rule.Destination = fields[3]
				}
				
				config.Rules = append(config.Rules, rule)
				ruleID++
			}
		}
	}
	
	return config, nil
}

// SaveInterface saves or updates a network interface
func (s *NetworkService) SaveInterface(req *InterfaceRequest) error {
	return s.interfaceConfigurator.SaveInterface(req.Name, req.Type, req.IPAddress, req.Netmask, req.Gateway, req.Enabled)
}

// DeleteInterface deletes a network interface
func (s *NetworkService) DeleteInterface(name string) error {
	return s.interfaceConfigurator.DeleteInterface(name)
}
