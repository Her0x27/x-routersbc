package net

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type DHCPConfig struct {
	Enabled    bool                   `json:"enabled"`
	Mode       string                 `json:"mode"` // server, relay, disabled
	Interface  string                 `json:"interface"`
	Range      DHCPRange              `json:"range"`
	Options    map[string]interface{} `json:"options"`
	StaticHosts []DHCPStaticHost      `json:"static_hosts"`
}

type DHCPRange struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	Subnet    string `json:"subnet"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
	DNS       []string `json:"dns"`
	LeaseTime string `json:"lease_time"`
}

type DHCPStaticHost struct {
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}

type DHCPLease struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
	Expires  string `json:"expires"`
}

const (
	dhcpdConfPath   = "/etc/dhcp/dhcpd.conf"
	dhcpdLeasesPath = "/var/lib/dhcp/dhcpd.leases"
	dnsmasqDHCPConf = "/etc/dnsmasq.d/dhcp.conf"
)

func GetDHCPConfiguration() (*DHCPConfig, error) {
	config := &DHCPConfig{
		Enabled:     false,
		Mode:        "disabled",
		Interface:   "",
		Range:       DHCPRange{},
		Options:     make(map[string]interface{}),
		StaticHosts: []DHCPStaticHost{},
	}

	// Check if ISC DHCP server is running
	if isISCDHCPRunning() {
		dhcpConfig, err := parseISCDHCPConfig()
		if err == nil {
			*config = *dhcpConfig
			config.Enabled = true
			config.Mode = "server"
		}
	} else if isDHCPRelayRunning() {
		relayConfig, err := parseDHCPRelayConfig()
		if err == nil {
			config.Enabled = true
			config.Mode = "relay"
			config.Options = relayConfig
		}
	} else if isDNSMasqDHCPRunning() {
		dnsmasqConfig, err := parseDNSMasqDHCPConfig()
		if err == nil {
			*config = *dnsmasqConfig
			config.Enabled = true
			config.Mode = "server"
		}
	}

	return config, nil
}

func isISCDHCPRunning() bool {
	err := exec.Command("pgrep", "dhcpd").Run()
	return err == nil
}

func isDHCPRelayRunning() bool {
	err := exec.Command("pgrep", "dhcrelay").Run()
	return err == nil
}

func isDNSMasqDHCPRunning() bool {
	// Check if dnsmasq is running with DHCP enabled
	if !isDNSMasqRunning() {
		return false
	}
	
	// Check if DHCP is enabled in dnsmasq config
	data, err := ioutil.ReadFile(dnsmasqConfPath)
	if err != nil {
		return false
	}
	
	return strings.Contains(string(data), "dhcp-range")
}

func parseISCDHCPConfig() (*DHCPConfig, error) {
	config := &DHCPConfig{
		Options:     make(map[string]interface{}),
		StaticHosts: []DHCPStaticHost{},
	}

	data, err := ioutil.ReadFile(dhcpdConfPath)
	if err != nil {
		return config, err
	}

	lines := strings.Split(string(data), "\n")
	inSubnet := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "subnet ") {
			inSubnet = true
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				config.Range.Subnet = parts[1]
				config.Range.Netmask = parts[3]
			}
		} else if inSubnet && strings.HasPrefix(line, "range ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				config.Range.Start = parts[1]
				config.Range.End = strings.TrimSuffix(parts[2], ";")
			}
		} else if strings.HasPrefix(line, "option routers ") {
			gateway := strings.TrimSuffix(strings.TrimPrefix(line, "option routers "), ";")
			config.Range.Gateway = gateway
		} else if strings.HasPrefix(line, "option domain-name-servers ") {
			dnsLine := strings.TrimSuffix(strings.TrimPrefix(line, "option domain-name-servers "), ";")
			dnsServers := strings.Split(strings.ReplaceAll(dnsLine, " ", ""), ",")
			config.Range.DNS = dnsServers
		} else if strings.HasPrefix(line, "default-lease-time ") {
			leaseTime := strings.TrimSuffix(strings.TrimPrefix(line, "default-lease-time "), ";")
			config.Range.LeaseTime = leaseTime + "s"
		}
	}

	return config, nil
}

func parseDHCPRelayConfig() (map[string]interface{}, error) {
	// Parse DHCP relay configuration
	config := make(map[string]interface{})
	
	// Try to get relay server from command line or config
	output, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return config, err
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "dhcrelay") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "-s" && i+1 < len(parts) {
					config["relay_server"] = parts[i+1]
					break
				}
			}
		}
	}
	
	return config, nil
}

func parseDNSMasqDHCPConfig() (*DHCPConfig, error) {
	config := &DHCPConfig{
		Options:     make(map[string]interface{}),
		StaticHosts: []DHCPStaticHost{},
	}

	// Parse main dnsmasq config
	data, err := ioutil.ReadFile(dnsmasqConfPath)
	if err != nil {
		return config, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "dhcp-range=") {
			rangeStr := strings.TrimPrefix(line, "dhcp-range=")
			parts := strings.Split(rangeStr, ",")
			if len(parts) >= 2 {
			  config.Range.Start = parts[0]
			  config.Range.End = parts[1]
			}
			if len(parts) >= 3 {
				config.Range.Netmask = parts[2]
			}
			if len(parts) >= 4 {
				config.Range.LeaseTime = parts[3]
			}
		} else if strings.HasPrefix(line, "dhcp-option=3,") {
			config.Range.Gateway = strings.TrimPrefix(line, "dhcp-option=3,")
		} else if strings.HasPrefix(line, "dhcp-option=6,") {
			dnsStr := strings.TrimPrefix(line, "dhcp-option=6,")
			config.Range.DNS = strings.Split(dnsStr, ",")
		} else if strings.HasPrefix(line, "dhcp-host=") {
			hostStr := strings.TrimPrefix(line, "dhcp-host=")
			parts := strings.Split(hostStr, ",")
			if len(parts) >= 2 {
				host := DHCPStaticHost{
					MAC: parts[0],
					IP:  parts[1],
				}
				if len(parts) >= 3 {
					host.Hostname = parts[2]
				}
				config.StaticHosts = append(config.StaticHosts, host)
			}
		}
	}

	return config, nil
}

func SetDHCPConfiguration(config *DHCPConfig) error {
	if !config.Enabled {
		return disableDHCP()
	}

	switch config.Mode {
	case "server":
		return configureDHCPServer(config)
	case "relay":
		return configureDHCPRelay(config)
	default:
		return fmt.Errorf("unsupported DHCP mode: %s", config.Mode)
	}
}

func disableDHCP() error {
	// Stop all DHCP services
	exec.Command("systemctl", "stop", "isc-dhcp-server").Run()
	exec.Command("systemctl", "disable", "isc-dhcp-server").Run()
	exec.Command("systemctl", "stop", "dhcrelay").Run()
	exec.Command("systemctl", "disable", "dhcrelay").Run()
	
	// Remove DHCP from dnsmasq if running
	if isDNSMasqRunning() {
		return removeDNSMasqDHCP()
	}
	
	return nil
}

func configureDHCPServer(config *DHCPConfig) error {
	// Prefer dnsmasq for DHCP server if available
	if isDNSMasqRunning() || isDNSMasqInstalled() {
		return configureDNSMasqDHCP(config)
	}
	
	// Fallback to ISC DHCP server
	return configureISCDHCPServer(config)
}

func isDNSMasqInstalled() bool {
	_, err := exec.LookPath("dnsmasq")
	return err == nil
}

func configureDNSMasqDHCP(config *DHCPConfig) error {
	var lines []string
	lines = append(lines, "# DHCP Configuration generated by router-sbc")
	
	// DHCP range
	rangeStr := fmt.Sprintf("dhcp-range=%s,%s", config.Range.Start, config.Range.End)
	if config.Range.Netmask != "" {
		rangeStr += "," + config.Range.Netmask
	}
	if config.Range.LeaseTime != "" {
		rangeStr += "," + config.Range.LeaseTime
	}
	lines = append(lines, rangeStr)
	
	// Gateway
	if config.Range.Gateway != "" {
		lines = append(lines, fmt.Sprintf("dhcp-option=3,%s", config.Range.Gateway))
	}
	
	// DNS servers
	if len(config.Range.DNS) > 0 {
		dnsStr := strings.Join(config.Range.DNS, ",")
		lines = append(lines, fmt.Sprintf("dhcp-option=6,%s", dnsStr))
	}
	
	// Static hosts
	for _, host := range config.StaticHosts {
		hostStr := fmt.Sprintf("dhcp-host=%s,%s", host.MAC, host.IP)
		if host.Hostname != "" {
			hostStr += "," + host.Hostname
		}
		lines = append(lines, hostStr)
	}
	
	content := strings.Join(lines, "\n") + "\n"
	if err := ioutil.WriteFile(dnsmasqDHCPConf, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write dnsmasq DHCP config: %v", err)
	}
	
	return restartDNSMasq()
}

func removeDNSMasqDHCP() error {
	if err := os.Remove(dnsmasqDHCPConf); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove dnsmasq DHCP config: %v", err)
	}
	
	return restartDNSMasq()
}

func configureISCDHCPServer(config *DHCPConfig) error {
	var lines []string
	lines = append(lines, "# DHCP Configuration generated by router-sbc")
	lines = append(lines, "authoritative;")
	lines = append(lines, "")
	
	// Global options
	if config.Range.LeaseTime != "" {
		leaseTime := strings.TrimSuffix(config.Range.LeaseTime, "s")
		lines = append(lines, fmt.Sprintf("default-lease-time %s;", leaseTime))
		lines = append(lines, fmt.Sprintf("max-lease-time %s;", leaseTime))
	}
	
	lines = append(lines, "")
	
	// Subnet declaration
	lines = append(lines, fmt.Sprintf("subnet %s netmask %s {", config.Range.Subnet, config.Range.Netmask))
	lines = append(lines, fmt.Sprintf("  range %s %s;", config.Range.Start, config.Range.End))
	
	if config.Range.Gateway != "" {
		lines = append(lines, fmt.Sprintf("  option routers %s;", config.Range.Gateway))
	}
	
	if len(config.Range.DNS) > 0 {
		dnsStr := strings.Join(config.Range.DNS, ", ")
		lines = append(lines, fmt.Sprintf("  option domain-name-servers %s;", dnsStr))
	}
	
	lines = append(lines, "}")
	lines = append(lines, "")
	
	// Static hosts
	for _, host := range config.StaticHosts {
		lines = append(lines, fmt.Sprintf("host %s {", host.Hostname))
		lines = append(lines, fmt.Sprintf("  hardware ethernet %s;", host.MAC))
		lines = append(lines, fmt.Sprintf("  fixed-address %s;", host.IP))
		lines = append(lines, "}")
		lines = append(lines, "")
	}
	
	content := strings.Join(lines, "\n")
	if err := ioutil.WriteFile(dhcpdConfPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write DHCP config: %v", err)
	}
	
	// Enable and start ISC DHCP server
	exec.Command("systemctl", "enable", "isc-dhcp-server").Run()
	return exec.Command("systemctl", "restart", "isc-dhcp-server").Run()
}

func configureDHCPRelay(config *DHCPConfig) error {
	relayServer, ok := config.Options["relay_server"].(string)
	if !ok || relayServer == "" {
		return fmt.Errorf("relay server not specified")
	}
	
	// Configure dhcrelay service
	serviceConfig := fmt.Sprintf(`[Unit]
Description=DHCP Relay Agent
After=network.target

[Service]
Type=forking
ExecStart=/usr/sbin/dhcrelay -s %s %s
PIDFile=/var/run/dhcrelay.pid

[Install]
WantedBy=multi-user.target
`, relayServer, config.Interface)
	
	servicePath := "/etc/systemd/system/dhcrelay.service"
	if err := ioutil.WriteFile(servicePath, []byte(serviceConfig), 0644); err != nil {
		return fmt.Errorf("failed to write dhcrelay service config: %v", err)
	}
	
	// Reload systemd and start service
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "dhcrelay").Run()
	return exec.Command("systemctl", "restart", "dhcrelay").Run()
}

func GetDHCPLeases() ([]DHCPLease, error) {
	var leases []DHCPLease
	
	// Try to get leases from ISC DHCP server first
	if iscLeases := getISCDHCPLeases(); len(iscLeases) > 0 {
		leases = iscLeases
	} else if dnsmasqLeases := getDNSMasqLeases(); len(dnsmasqLeases) > 0 {
		leases = dnsmasqLeases
	}
	
	return leases, nil
}

func getISCDHCPLeases() []DHCPLease {
	var leases []DHCPLease
	
	data, err := ioutil.ReadFile(dhcpdLeasesPath)
	if err != nil {
		return leases
	}
	
	// Parse ISC DHCP lease file
	content := string(data)
	leaseBlocks := strings.Split(content, "lease ")
	
	for _, block := range leaseBlocks[1:] { // Skip first empty element
		lines := strings.Split(block, "\n")
		if len(lines) == 0 {
			continue
		}
		
		lease := DHCPLease{}
		
		// Get IP from first line
		ip := strings.TrimSpace(strings.Split(lines[0], " ")[0])
		lease.IP = ip
		
		// Parse lease details
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "hardware ethernet ") {
				lease.MAC = strings.TrimSuffix(strings.TrimPrefix(line, "hardware ethernet "), ";")
			} else if strings.HasPrefix(line, "client-hostname ") {
				hostname := strings.TrimPrefix(line, "client-hostname ")
				lease.Hostname = strings.Trim(strings.TrimSuffix(hostname, ";"), `"`)
			} else if strings.HasPrefix(line, "ends ") {
				lease.Expires = strings.TrimSuffix(strings.TrimPrefix(line, "ends "), ";")
			}
		}
		
		if lease.IP != "" && lease.MAC != "" {
			leases = append(leases, lease)
		}
	}
	
	return leases
}

func getDNSMasqLeases() []DHCPLease {
	var leases []DHCPLease
	
	// dnsmasq typically stores leases in /var/lib/dhcp/dnsmasq.leases
	leasesPath := "/var/lib/dhcp/dnsmasq.leases"
	data, err := ioutil.ReadFile(leasesPath)
	if err != nil {
		return leases
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			lease := DHCPLease{
				Expires:  parts[0],
				MAC:      parts[1],
				IP:       parts[2],
				Hostname: parts[3],
			}
			leases = append(leases, lease)
		}
	}
	
	return leases
}

func AddDHCPStaticHost(host DHCPStaticHost) error {
	config, err := GetDHCPConfiguration()
	if err != nil {
		return err
	}
	
	// Check if host already exists
	for i, existingHost := range config.StaticHosts {
		if existingHost.MAC == host.MAC {
			config.StaticHosts[i] = host // Update existing
			return SetDHCPConfiguration(config)
		}
	}
	
	// Add new host
	config.StaticHosts = append(config.StaticHosts, host)
	return SetDHCPConfiguration(config)
}

func RemoveDHCPStaticHost(mac string) error {
	config, err := GetDHCPConfiguration()
	if err != nil {
		return err
	}
	
	// Remove host with matching MAC
	var newHosts []DHCPStaticHost
	for _, host := range config.StaticHosts {
		if host.MAC != mac {
			newHosts = append(newHosts, host)
		}
	}
	
	config.StaticHosts = newHosts
	return SetDHCPConfiguration(config)
}

func GetDHCPStatus() (map[string]interface{}, error) {
	config, err := GetDHCPConfiguration()
	if err != nil {
		return nil, err
	}
	
	status := map[string]interface{}{
		"enabled":      config.Enabled,
		"mode":         config.Mode,
		"interface":    config.Interface,
		"range_start":  config.Range.Start,
		"range_end":    config.Range.End,
		"gateway":      config.Range.Gateway,
		"dns_servers":  config.Range.DNS,
		"static_hosts": len(config.StaticHosts),
		"service_status": getDHCPServiceStatus(config.Mode),
	}
	
	// Get active leases
	leases, err := GetDHCPLeases()
	if err == nil {
		status["active_leases"] = len(leases)
	}
	
	return status, nil
}

func getDHCPServiceStatus(mode string) string {
	switch mode {
	case "server":
		if isISCDHCPRunning() {
			return "running (ISC DHCP)"
		} else if isDNSMasqDHCPRunning() {
			return "running (dnsmasq)"
		}
		return "stopped"
	case "relay":
		if isDHCPRelayRunning() {
			return "running (relay)"
		}
		return "stopped"
	default:
		return "disabled"
	}
}
