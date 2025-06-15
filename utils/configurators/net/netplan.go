package net

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// NetplanConfig represents a netplan configuration
type NetplanConfig struct {
	Network NetplanNetwork `yaml:"network"`
}

// NetplanNetwork represents the network section
type NetplanNetwork struct {
	Version   int                              `yaml:"version"`
	Renderer  string                          `yaml:"renderer,omitempty"`
	Ethernets map[string]NetplanEthernet      `yaml:"ethernets,omitempty"`
	Wifis     map[string]NetplanWifi          `yaml:"wifis,omitempty"`
	Bridges   map[string]NetplanBridge        `yaml:"bridges,omitempty"`
	VLANs     map[string]NetplanVLAN          `yaml:"vlans,omitempty"`
}

// NetplanEthernet represents ethernet configuration
type NetplanEthernet struct {
	DHCP4     bool                   `yaml:"dhcp4,omitempty"`
	DHCP6     bool                   `yaml:"dhcp6,omitempty"`
	Addresses []string               `yaml:"addresses,omitempty"`
	Gateway4  string                 `yaml:"gateway4,omitempty"`
	Gateway6  string                 `yaml:"gateway6,omitempty"`
	Nameservers NetplanNameservers   `yaml:"nameservers,omitempty"`
	Routes    []NetplanRoute         `yaml:"routes,omitempty"`
}

// NetplanWifi represents wifi configuration
type NetplanWifi struct {
	DHCP4       bool                   `yaml:"dhcp4,omitempty"`
	DHCP6       bool                   `yaml:"dhcp6,omitempty"`
	Addresses   []string               `yaml:"addresses,omitempty"`
	Gateway4    string                 `yaml:"gateway4,omitempty"`
	AccessPoints map[string]NetplanAP  `yaml:"access-points,omitempty"`
	Nameservers NetplanNameservers     `yaml:"nameservers,omitempty"`
}

// NetplanAP represents access point configuration
type NetplanAP struct {
	Password string `yaml:"password,omitempty"`
}

// NetplanBridge represents bridge configuration
type NetplanBridge struct {
	DHCP4      bool                   `yaml:"dhcp4,omitempty"`
	Addresses  []string               `yaml:"addresses,omitempty"`
	Gateway4   string                 `yaml:"gateway4,omitempty"`
	Interfaces []string               `yaml:"interfaces,omitempty"`
	Nameservers NetplanNameservers    `yaml:"nameservers,omitempty"`
}

// NetplanVLAN represents VLAN configuration
type NetplanVLAN struct {
	ID        int                    `yaml:"id"`
	Link      string                 `yaml:"link"`
	DHCP4     bool                   `yaml:"dhcp4,omitempty"`
	Addresses []string               `yaml:"addresses,omitempty"`
	Gateway4  string                 `yaml:"gateway4,omitempty"`
	Nameservers NetplanNameservers   `yaml:"nameservers,omitempty"`
}

// NetplanNameservers represents DNS configuration
type NetplanNameservers struct {
	Search    []string `yaml:"search,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
}

// NetplanRoute represents a route
type NetplanRoute struct {
	To     string `yaml:"to"`
	Via    string `yaml:"via"`
	Metric int    `yaml:"metric,omitempty"`
}

// NetplanConfigurator handles netplan-based network configuration
type NetplanConfigurator struct {
	configPath string
}

// NewNetplanConfigurator creates a new netplan configurator
func NewNetplanConfigurator() *NetplanConfigurator {
	return &NetplanConfigurator{
		configPath: "/etc/netplan",
	}
}

// IsAvailable checks if netplan is available on the system
func (n *NetplanConfigurator) IsAvailable() bool {
	if _, err := exec.LookPath("netplan"); err != nil {
		return false
	}
	
	if _, err := os.Stat(n.configPath); os.IsNotExist(err) {
		return false
	}
	
	return true
}

// GetConfiguration reads current netplan configuration
func (n *NetplanConfigurator) GetConfiguration() (*NetplanConfig, error) {
	files, err := filepath.Glob(filepath.Join(n.configPath, "*.yaml"))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		// Return default configuration
		return &NetplanConfig{
			Network: NetplanNetwork{
				Version:  2,
				Renderer: "networkd",
			},
		}, nil
	}

	// Read the first configuration file
	data, err := ioutil.ReadFile(files[0])
	if err != nil {
		return nil, err
	}

	var config NetplanConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// WriteConfiguration writes netplan configuration
func (n *NetplanConfigurator) WriteConfiguration(config *NetplanConfig) error {
	configFile := filepath.Join(n.configPath, "01-routersbc.yaml")
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		return err
	}

	// Apply configuration
	return n.Apply()
}

// Apply applies netplan configuration
func (n *NetplanConfigurator) Apply() error {
	cmd := exec.Command("netplan", "apply")
	return cmd.Run()
}

// ConfigureInterface configures a network interface using netplan
func (n *NetplanConfigurator) ConfigureInterface(name, interfaceType, configJSON string) error {
	config, err := n.GetConfiguration()
	if err != nil {
		return err
	}

	if config.Network.Ethernets == nil {
		config.Network.Ethernets = make(map[string]NetplanEthernet)
	}

	// Parse configuration based on type
	switch interfaceType {
	case "ethernet":
		ethernet := NetplanEthernet{
			DHCP4: true, // Default to DHCP
		}
		
		// TODO: Parse configJSON to set specific configuration
		config.Network.Ethernets[name] = ethernet

	case "wifi":
		if config.Network.Wifis == nil {
			config.Network.Wifis = make(map[string]NetplanWifi)
		}
		
		wifi := NetplanWifi{
			DHCP4: true,
		}
		
		// TODO: Parse configJSON to set WiFi configuration
		config.Network.Wifis[name] = wifi

	case "bridge":
		if config.Network.Bridges == nil {
			config.Network.Bridges = make(map[string]NetplanBridge)
		}
		
		bridge := NetplanBridge{
			DHCP4: true,
		}
		
		// TODO: Parse configJSON to set bridge configuration
		config.Network.Bridges[name] = bridge

	case "vlan":
		if config.Network.VLANs == nil {
			config.Network.VLANs = make(map[string]NetplanVLAN)
		}
		
		vlan := NetplanVLAN{
			ID:    100, // Default VLAN ID
			Link:  "eth0", // Default link
			DHCP4: true,
		}
		
		// TODO: Parse configJSON to set VLAN configuration
		config.Network.VLANs[name] = vlan
	}

	return n.WriteConfiguration(config)
}

// RemoveInterface removes an interface from netplan configuration
func (n *NetplanConfigurator) RemoveInterface(name string) error {
	config, err := n.GetConfiguration()
	if err != nil {
		return err
	}

	// Remove from all possible sections
	if config.Network.Ethernets != nil {
		delete(config.Network.Ethernets, name)
	}
	if config.Network.Wifis != nil {
		delete(config.Network.Wifis, name)
	}
	if config.Network.Bridges != nil {
		delete(config.Network.Bridges, name)
	}
	if config.Network.VLANs != nil {
		delete(config.Network.VLANs, name)
	}

	return n.WriteConfiguration(config)
}

// SetStaticIP sets static IP configuration for an interface
func (n *NetplanConfigurator) SetStaticIP(interfaceName, ip, netmask, gateway string, dnsServers []string) error {
	config, err := n.GetConfiguration()
	if err != nil {
		return err
	}

	if config.Network.Ethernets == nil {
		config.Network.Ethernets = make(map[string]NetplanEthernet)
	}

	// Calculate CIDR notation
	cidr := fmt.Sprintf("%s/%s", ip, netmask)
	
	ethernet := NetplanEthernet{
		DHCP4:     false,
		Addresses: []string{cidr},
		Gateway4:  gateway,
		Nameservers: NetplanNameservers{
			Addresses: dnsServers,
		},
	}

	config.Network.Ethernets[interfaceName] = ethernet
	return n.WriteConfiguration(config)
}

// EnableDHCP enables DHCP for an interface
func (n *NetplanConfigurator) EnableDHCP(interfaceName string) error {
	config, err := n.GetConfiguration()
	if err != nil {
		return err
	}

	if config.Network.Ethernets == nil {
		config.Network.Ethernets = make(map[string]NetplanEthernet)
	}

	ethernet := NetplanEthernet{
		DHCP4: true,
	}

	config.Network.Ethernets[interfaceName] = ethernet
	return n.WriteConfiguration(config)
}

// AddRoute adds a static route
func (n *NetplanConfigurator) AddRoute(interfaceName, destination, gateway string, metric int) error {
	config, err := n.GetConfiguration()
	if err != nil {
		return err
	}

	if config.Network.Ethernets == nil {
		config.Network.Ethernets = make(map[string]NetplanEthernet)
	}

	ethernet, exists := config.Network.Ethernets[interfaceName]
	if !exists {
		ethernet = NetplanEthernet{DHCP4: true}
	}

	route := NetplanRoute{
		To:     destination,
		Via:    gateway,
		Metric: metric,
	}

	ethernet.Routes = append(ethernet.Routes, route)
	config.Network.Ethernets[interfaceName] = ethernet

	return n.WriteConfiguration(config)
}
