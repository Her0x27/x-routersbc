package net

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	
	"gopkg.in/yaml.v2"
)

type NetplanConfig struct {
	Network struct {
		Version   int                    `yaml:"version"`
		Renderer  string                 `yaml:"renderer,omitempty"`
		Ethernets map[string]EthernetConfig `yaml:"ethernets,omitempty"`
		Wifis     map[string]WifiConfig     `yaml:"wifis,omitempty"`
		Vlans     map[string]VlanConfig     `yaml:"vlans,omitempty"`
		Bridges   map[string]BridgeConfig   `yaml:"bridges,omitempty"`
	} `yaml:"network"`
}

type EthernetConfig struct {
	DHCP4     bool     `yaml:"dhcp4,omitempty"`
	DHCP6     bool     `yaml:"dhcp6,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
	Gateway4  string   `yaml:"gateway4,omitempty"`
	Gateway6  string   `yaml:"gateway6,omitempty"`
	Nameservers struct {
		Addresses []string `yaml:"addresses,omitempty"`
		Search    []string `yaml:"search,omitempty"`
	} `yaml:"nameservers,omitempty"`
}

type WifiConfig struct {
	DHCP4       bool     `yaml:"dhcp4,omitempty"`
	DHCP6       bool     `yaml:"dhcp6,omitempty"`
	Addresses   []string `yaml:"addresses,omitempty"`
	Gateway4    string   `yaml:"gateway4,omitempty"`
	AccessPoints map[string]AccessPoint `yaml:"access-points,omitempty"`
}

type AccessPoint struct {
	Password string `yaml:"password,omitempty"`
}

type VlanConfig struct {
	ID   int    `yaml:"id"`
	Link string `yaml:"link"`
	DHCP4 bool  `yaml:"dhcp4,omitempty"`
	DHCP6 bool  `yaml:"dhcp6,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
}

type BridgeConfig struct {
	DHCP4     bool     `yaml:"dhcp4,omitempty"`
	DHCP6     bool     `yaml:"dhcp6,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
	Interfaces []string `yaml:"interfaces,omitempty"`
}

const netplanConfigPath = "/etc/netplan/01-router-sbc.yaml"

func IsNetplanAvailable() bool {
	_, err := exec.LookPath("netplan")
	return err == nil
}

func GetNetplanConfig() (*NetplanConfig, error) {
	if !IsNetplanAvailable() {
		return nil, fmt.Errorf("netplan is not available on this system")
	}
	
	data, err := ioutil.ReadFile(netplanConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &NetplanConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read netplan config: %v", err)
	}
	
	var config NetplanConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse netplan config: %v", err)
	}
	
	return &config, nil
}

func SaveNetplanConfig(config *NetplanConfig) error {
	if !IsNetplanAvailable() {
		return fmt.Errorf("netplan is not available on this system")
	}
	
	// Set default values
	config.Network.Version = 2
	if config.Network.Renderer == "" {
		config.Network.Renderer = "networkd"
	}
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal netplan config: %v", err)
	}
	
	// Write config file
	if err := ioutil.WriteFile(netplanConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write netplan config: %v", err)
	}
	
	// Apply configuration
	return ApplyNetplanConfig()
}

func ApplyNetplanConfig() error {
	if !IsNetplanAvailable() {
		return fmt.Errorf("netplan is not available on this system")
	}
	
	// Test configuration first
	if err := exec.Command("netplan", "try", "--timeout=10").Run(); err != nil {
		return fmt.Errorf("netplan configuration test failed: %v", err)
	}
	
	// Apply configuration
	if err := exec.Command("netplan", "apply").Run(); err != nil {
		return fmt.Errorf("failed to apply netplan configuration: %v", err)
	}
	
	return nil
}

func AddEthernetInterface(name string, config EthernetConfig) error {
	netplanConfig, err := GetNetplanConfig()
	if err != nil {
		return err
	}
	
	if netplanConfig.Network.Ethernets == nil {
		netplanConfig.Network.Ethernets = make(map[string]EthernetConfig)
	}
	
	netplanConfig.Network.Ethernets[name] = config
	
	return SaveNetplanConfig(netplanConfig)
}

func AddWifiInterface(name string, config WifiConfig) error {
	netplanConfig, err := GetNetplanConfig()
	if err != nil {
		return err
	}
	
	if netplanConfig.Network.Wifis == nil {
		netplanConfig.Network.Wifis = make(map[string]WifiConfig)
	}
	
	netplanConfig.Network.Wifis[name] = config
	
	return SaveNetplanConfig(netplanConfig)
}

func AddVlanInterface(name string, config VlanConfig) error {
	netplanConfig, err := GetNetplanConfig()
	if err != nil {
		return err
	}
	
	if netplanConfig.Network.Vlans == nil {
		netplanConfig.Network.Vlans = make(map[string]VlanConfig)
	}
	
	netplanConfig.Network.Vlans[name] = config
	
	return SaveNetplanConfig(netplanConfig)
}

func AddBridgeInterface(name string, config BridgeConfig) error {
	netplanConfig, err := GetNetplanConfig()
	if err != nil {
		return err
	}
	
	if netplanConfig.Network.Bridges == nil {
		netplanConfig.Network.Bridges = make(map[string]BridgeConfig)
	}
	
	netplanConfig.Network.Bridges[name] = config
	
	return SaveNetplanConfig(netplanConfig)
}

func RemoveInterface(name string) error {
	netplanConfig, err := GetNetplanConfig()
	if err != nil {
		return err
	}
	
	// Remove from all possible interface types
	delete(netplanConfig.Network.Ethernets, name)
	delete(netplanConfig.Network.Wifis, name)
	delete(netplanConfig.Network.Vlans, name)
	delete(netplanConfig.Network.Bridges, name)
	
	return SaveNetplanConfig(netplanConfig)
}
