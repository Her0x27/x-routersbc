package net

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DHCPLease represents a DHCP lease
type DHCPLease struct {
	MAC       string    `json:"mac"`
	IP        string    `json:"ip"`
	Hostname  string    `json:"hostname"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
}

// DHCPReservation represents a static DHCP reservation
type DHCPReservation struct {
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Enabled  bool   `json:"enabled"`
}

// DHCPRange represents a DHCP address range
type DHCPRange struct {
	Start     string `json:"start"`
	End       string `json:"end"`
	LeaseTime string `json:"lease_time"`
	Enabled   bool   `json:"enabled"`
}

// DHCPConfiguration represents DHCP server configuration
type DHCPConfiguration struct {
	Mode         string            `json:"mode"`         // server, relay, disabled
	Interface    string            `json:"interface"`
	Domain       string            `json:"domain"`
	Ranges       []DHCPRange       `json:"ranges"`
	Reservations []DHCPReservation `json:"reservations"`
	Options      map[string]string `json:"options"`
	LeaseTime    string            `json:"lease_time"`
	Authoritative bool             `json:"authoritative"`
	Enabled      bool              `json:"enabled"`
	RelayTarget  string            `json:"relay_target,omitempty"`
}

// DHCPConfigurator handles DHCP configuration
type DHCPConfigurator struct {
	configPath    string
	dhcpdConfPath string
	leasesPath    string
	dnsmasqPath   string
}

// NewDHCPConfigurator creates a new DHCP configurator
func NewDHCPConfigurator() *DHCPConfigurator {
	return &DHCPConfigurator{
		configPath:    "/etc/routersbc",
		dhcpdConfPath: "/etc/dhcp/dhcpd.conf",
		leasesPath:    "/var/lib/dhcp/dhcpd.leases",
		dnsmasqPath:   "/etc/dnsmasq.conf",
	}
}

// GetDHCPConfiguration gets current DHCP configuration
func (d *DHCPConfigurator) GetDHCPConfiguration() (*DHCPConfiguration, error) {
	config := &DHCPConfiguration{
		Mode:          "server",
		Interface:     "br0",
		Domain:        "local",
		LeaseTime:     "24h",
		Authoritative: true,
		Enabled:       true,
		Ranges:        []DHCPRange{},
		Reservations:  []DHCPReservation{},
		Options:       make(map[string]string),
	}

	// Try to load saved configuration
	configFile := filepath.Join(d.configPath, "dhcp.json")
	if data, err := ioutil.ReadFile(configFile); err == nil {
		if err := json.Unmarshal(data, config); err == nil {
			return config, nil
		}
	}

	// Detect current system DHCP configuration
	return d.detectCurrentDHCPConfig()
}

// detectCurrentDHCPConfig detects current DHCP configuration from system
func (d *DHCPConfigurator) detectCurrentDHCPConfig() (*DHCPConfiguration, error) {
	config := &DHCPConfiguration{
		Mode:          "disabled",
		Interface:     "br0",
		Domain:        "local",
		LeaseTime:     "24h",
		Authoritative: true,
		Enabled:       false,
		Ranges:        []DHCPRange{},
		Reservations:  []DHCPReservation{},
		Options:       make(map[string]string),
	}

	// Check if isc-dhcp-server is running
	if d.isDHCPDRunning() {
		config.Mode = "server"
		config.Enabled = true
		
		// Parse dhcpd.conf if it exists
		if _, err := os.Stat(d.dhcpdConfPath); err == nil {
			d.parseDHCPDConf(config)
		}
	}

	// Check if dnsmasq is handling DHCP
	if d.isDNSMasqDHCPRunning() {
		config.Mode = "server"
		config.Enabled = true
		d.parseDNSMasqDHCP(config)
	}

	return config, nil
}

// SetDHCPConfiguration applies DHCP configuration
func (d *DHCPConfigurator) SetDHCPConfiguration(config *DHCPConfiguration) error {
	// Save configuration
	if err := d.saveDHCPConfig(config); err != nil {
		return fmt.Errorf("failed to save DHCP config: %v", err)
	}

	// Apply configuration based on mode
	switch config.Mode {
	case "server":
		return d.configureDHCPServer(config)
	case "relay":
		return d.configureDHCPRelay(config)
	case "disabled":
		return d.disableDHCP()
	default:
		return fmt.Errorf("unsupported DHCP mode: %s", config.Mode)
	}
}

// configureDHCPServer configures DHCP server
func (d *DHCPConfigurator) configureDHCPServer(config *DHCPConfiguration) error {
	// Prefer dnsmasq for integrated DNS/DHCP
	if d.dnsmasqAvailable() {
		return d.configureDNSMasqDHCP(config)
	}
	
	// Fallback to isc-dhcp-server
	return d.configureISCDHCP(config)
}

// configureDNSMasqDHCP configures DHCP using dnsmasq
func (d *DHCPConfigurator) configureDNSMasqDHCP(config *DHCPConfiguration) error {
	// Read existing dnsmasq configuration
	var existingConfig string
	if data, err := ioutil.ReadFile(d.dnsmasqPath); err == nil {
		existingConfig = string(data)
	}

	var content strings.Builder
	
	// Keep existing non-DHCP configuration
	lines := strings.Split(existingConfig, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "dhcp-") && !strings.HasPrefix(line, "enable-tftp") {
			content.WriteString(line + "\n")
		}
	}

	if config.Enabled {
		// Add DHCP configuration
		content.WriteString("\n# DHCP Configuration - Generated by RouterSBC\n")
		
		// DHCP ranges
		for _, dhcpRange := range config.Ranges {
			if dhcpRange.Enabled {
				content.WriteString(fmt.Sprintf("dhcp-range=%s,%s,%s\n", 
					dhcpRange.Start, dhcpRange.End, dhcpRange.LeaseTime))
			}
		}

		// DHCP options
		if config.Domain != "" {
			content.WriteString(fmt.Sprintf("dhcp-option=option:domain-name,%s\n", config.Domain))
		}

		// DNS servers
		content.WriteString("dhcp-option=option:dns-server,0.0.0.0\n")

		// Static reservations
		for _, reservation := range config.Reservations {
			if reservation.Enabled {
				if reservation.Hostname != "" {
					content.WriteString(fmt.Sprintf("dhcp-host=%s,%s,%s\n", 
						reservation.MAC, reservation.IP, reservation.Hostname))
				} else {
					content.WriteString(fmt.Sprintf("dhcp-host=%s,%s\n", 
						reservation.MAC, reservation.IP))
				}
			}
		}

		// Interface binding
		if config.Interface != "" {
			content.WriteString(fmt.Sprintf("interface=%s\n", config.Interface))
		}

		// Authoritative
		if config.Authoritative {
			content.WriteString("dhcp-authoritative\n")
		}

		// Additional options
		for key, value := range config.Options {
			content.WriteString(fmt.Sprintf("dhcp-option=%s,%s\n", key, value))
		}
	}

	// Write configuration
	if err := ioutil.WriteFile(d.dnsmasqPath, []byte(content.String()), 0644); err != nil {
		return err
	}

	// Restart dnsmasq
	return d.restartDNSMasq()
}

// configureISCDHCP configures ISC DHCP server
func (d *DHCPConfigurator) configureISCDHCP(config *DHCPConfiguration) error {
	var content strings.Builder
	
	content.WriteString("# Generated by RouterSBC\n")
	content.WriteString("default-lease-time 86400;\n")
	content.WriteString("max-lease-time 172800;\n")
	
	if config.Authoritative {
		content.WriteString("authoritative;\n")
	}

	content.WriteString("\n")

	// Configure subnets based on interface
	interfaceIP, interfaceNetwork, err := d.getInterfaceNetwork(config.Interface)
	if err != nil {
		return err
	}

	content.WriteString(fmt.Sprintf("subnet %s netmask %s {\n", interfaceNetwork, "255.255.255.0"))
	
	// DHCP ranges
	for _, dhcpRange := range config.Ranges {
		if dhcpRange.Enabled {
			content.WriteString(fmt.Sprintf("    range %s %s;\n", dhcpRange.Start, dhcpRange.End))
		}
	}

	// Options
	content.WriteString(fmt.Sprintf("    option routers %s;\n", interfaceIP))
	content.WriteString(fmt.Sprintf("    option domain-name-servers %s;\n", interfaceIP))
	
	if config.Domain != "" {
		content.WriteString(fmt.Sprintf("    option domain-name \"%s\";\n", config.Domain))
	}

	content.WriteString("}\n\n")

	// Static reservations
	for _, reservation := range config.Reservations {
		if reservation.Enabled {
			content.WriteString(fmt.Sprintf("host %s {\n", reservation.Hostname))
			content.WriteString(fmt.Sprintf("    hardware ethernet %s;\n", reservation.MAC))
			content.WriteString(fmt.Sprintf("    fixed-address %s;\n", reservation.IP))
			content.WriteString("}\n\n")
		}
	}

	// Write configuration
	if err := ioutil.WriteFile(d.dhcpdConfPath, []byte(content.String()), 0644); err != nil {
		return err
	}

	// Configure interface
	defaultsFile := "/etc/default/isc-dhcp-server"
	defaultsContent := fmt.Sprintf("INTERFACESv4=\"%s\"\nINTERFACESv6=\"\"\n", config.Interface)
	if err := ioutil.WriteFile(defaultsFile, []byte(defaultsContent), 0644); err != nil {
		return err
	}

	// Restart DHCP server
	return d.restartDHCPD()
}

// configureDHCPRelay configures DHCP relay
func (d *DHCPConfigurator) configureDHCPRelay(config *DHCPConfiguration) error {
	// Stop DHCP server if running
	d.stopDHCPD()

	// Configure dhcp-relay
	cmd := exec.Command("systemctl", "enable", "isc-dhcp-relay")
	cmd.Run()

	// Configure relay target
	defaultsFile := "/etc/default/isc-dhcp-relay"
	content := fmt.Sprintf("SERVERS=\"%s\"\nINTERFACES=\"%s\"\nOPTIONS=\"\"\n", 
		config.RelayTarget, config.Interface)
	
	if err := ioutil.WriteFile(defaultsFile, []byte(content), 0644); err != nil {
		return err
	}

	// Start relay
	return exec.Command("systemctl", "start", "isc-dhcp-relay").Run()
}

// disableDHCP disables DHCP server
func (d *DHCPConfigurator) disableDHCP() error {
	// Stop and disable services
	services := []string{"isc-dhcp-server", "isc-dhcp-relay", "dnsmasq"}
	
	for _, service := range services {
		exec.Command("systemctl", "stop", service).Run()
		exec.Command("systemctl", "disable", service).Run()
	}

	return nil
}

// GetDHCPLeases returns current DHCP leases
func (d *DHCPConfigurator) GetDHCPLeases() ([]DHCPLease, error) {
	var leases []DHCPLease

	// Try dnsmasq leases first
	if dnsmasqLeases, err := d.getDNSMasqLeases(); err == nil && len(dnsmasqLeases) > 0 {
		return dnsmasqLeases, nil
	}

	// Try ISC DHCP leases
	return d.getISCDHCPLeases()
}

// getDNSMasqLeases gets leases from dnsmasq
func (d *DHCPConfigurator) getDNSMasqLeases() ([]DHCPLease, error) {
	leasesFile := "/var/lib/dhcp/dnsmasq.leases"
	
	file, err := os.Open(leasesFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var leases []DHCPLease
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 4 {
			timestamp, _ := strconv.ParseInt(parts[0], 10, 64)
			expiresAt := time.Unix(timestamp, 0)
			
			lease := DHCPLease{
				MAC:       parts[1],
				IP:        parts[2],
				Hostname:  parts[3],
				ExpiresAt: expiresAt,
				Active:    time.Now().Before(expiresAt),
			}
			leases = append(leases, lease)
		}
	}

	return leases, scanner.Err()
}

// getISCDHCPLeases gets leases from ISC DHCP server
func (d *DHCPConfigurator) getISCDHCPLeases() ([]DHCPLease, error) {
	file, err := os.Open(d.leasesPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var leases []DHCPLease
	var currentLease *DHCPLease
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.HasPrefix(line, "lease ") {
			// New lease
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentLease = &DHCPLease{
					IP: parts[1],
				}
			}
		} else if currentLease != nil {
			if strings.HasPrefix(line, "hardware ethernet ") {
				mac := strings.TrimPrefix(line, "hardware ethernet ")
				mac = strings.TrimSuffix(mac, ";")
				currentLease.MAC = mac
			} else if strings.HasPrefix(line, "client-hostname ") {
				hostname := strings.TrimPrefix(line, "client-hostname ")
				hostname = strings.Trim(hostname, "\";")
				currentLease.Hostname = hostname
			} else if strings.HasPrefix(line, "ends ") {
				// Parse lease end time
				endStr := strings.TrimPrefix(line, "ends ")
				endStr = strings.TrimSuffix(endStr, ";")
				if endTime, err := time.Parse("1 2006/01/02 15:04:05", endStr); err == nil {
					currentLease.ExpiresAt = endTime
					currentLease.Active = time.Now().Before(endTime)
				}
			} else if line == "}" {
				// End of lease
				if currentLease.MAC != "" {
					leases = append(leases, *currentLease)
				}
				currentLease = nil
			}
		}
	}

	return leases, scanner.Err()
}

// AddReservation adds a DHCP reservation
func (d *DHCPConfigurator) AddReservation(reservation DHCPReservation) error {
	config, err := d.GetDHCPConfiguration()
	if err != nil {
		return err
	}

	config.Reservations = append(config.Reservations, reservation)
	return d.SetDHCPConfiguration(config)
}

// RemoveReservation removes a DHCP reservation
func (d *DHCPConfigurator) RemoveReservation(mac string) error {
	config, err := d.GetDHCPConfiguration()
	if err != nil {
		return err
	}

	var filteredReservations []DHCPReservation
	for _, reservation := range config.Reservations {
		if reservation.MAC != mac {
			filteredReservations = append(filteredReservations, reservation)
		}
	}

	config.Reservations = filteredReservations
	return d.SetDHCPConfiguration(config)
}

// ReleaseIP releases a DHCP lease
func (d *DHCPConfigurator) ReleaseIP(ip string) error {
	// Try to find and remove the lease
	if d.isDNSMasqDHCPRunning() {
		// Signal dnsmasq to release the lease
		cmd := exec.Command("killall", "-HUP", "dnsmasq")
		return cmd.Run()
	}

	if d.isDHCPDRunning() {
		// For ISC DHCP, we need to edit the leases file
		return d.removeISCDHCPLease(ip)
	}

	return fmt.Errorf("no DHCP server running")
}

// Helper functions

func (d *DHCPConfigurator) saveDHCPConfig(config *DHCPConfiguration) error {
	if err := os.MkdirAll(d.configPath, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configFile := filepath.Join(d.configPath, "dhcp.json")
	return ioutil.WriteFile(configFile, data, 0644)
}

func (d *DHCPConfigurator) isDHCPDRunning() bool {
	cmd := exec.Command("pgrep", "dhcpd")
	return cmd.Run() == nil
}

func (d *DHCPConfigurator) isDNSMasqDHCPRunning() bool {
	if !d.isDNSMasqRunning() {
		return false
	}
	
	// Check if dnsmasq is configured for DHCP
	if data, err := ioutil.ReadFile(d.dnsmasqPath); err == nil {
		return strings.Contains(string(data), "dhcp-range")
	}
	
	return false
}

func (d *DHCPConfigurator) isDNSMasqRunning() bool {
	cmd := exec.Command("pgrep", "dnsmasq")
	return cmd.Run() == nil
}

func (d *DHCPConfigurator) dnsmasqAvailable() bool {
	_, err := exec.LookPath("dnsmasq")
	return err == nil
}

func (d *DHCPConfigurator) restartDNSMasq() error {
	commands := [][]string{
		{"systemctl", "restart", "dnsmasq"},
		{"service", "dnsmasq", "restart"},
	}

	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to restart dnsmasq")
}

func (d *DHCPConfigurator) restartDHCPD() error {
	commands := [][]string{
		{"systemctl", "restart", "isc-dhcp-server"},
		{"service", "isc-dhcp-server", "restart"},
	}

	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to restart DHCP server")
}

func (d *DHCPConfigurator) stopDHCPD() error {
	commands := [][]string{
		{"systemctl", "stop", "isc-dhcp-server"},
		{"service", "isc-dhcp-server", "stop"},
	}

	for _, cmd := range commands {
		exec.Command(cmd[0], cmd[1:]...).Run()
	}

	return nil
}

func (d *DHCPConfigurator) getInterfaceNetwork(interfaceName string) (string, string, error) {
	cmd := exec.Command("ip", "addr", "show", interfaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 2 {
				cidr := parts[1]
				ip, ipNet, err := net.ParseCIDR(cidr)
				if err != nil {
					return "", "", err
				}
				return ip.String(), ipNet.IP.String(), nil
			}
		}
	}

	return "", "", fmt.Errorf("no IP address found for interface %s", interfaceName)
}

func (d *DHCPConfigurator) parseDHCPDConf(config *DHCPConfiguration) error {
	file, err := os.Open(d.dhcpdConfPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.HasPrefix(line, "range ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				dhcpRange := DHCPRange{
					Start:     parts[1],
					End:       strings.TrimSuffix(parts[2], ";"),
					LeaseTime: config.LeaseTime,
					Enabled:   true,
				}
				config.Ranges = append(config.Ranges, dhcpRange)
			}
		}
	}

	return scanner.Err()
}

func (d *DHCPConfigurator) parseDNSMasqDHCP(config *DHCPConfiguration) error {
	file, err := os.Open(d.dnsmasqPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.HasPrefix(line, "dhcp-range=") {
			rangeStr := strings.TrimPrefix(line, "dhcp-range=")
			parts := strings.Split(rangeStr, ",")
			if len(parts) >= 2 {
				dhcpRange := DHCPRange{
					Start:   parts[0],
					End:     parts[1],
					Enabled: true,
				}
				if len(parts) >= 3 {
					dhcpRange.LeaseTime = parts[2]
				}
				config.Ranges = append(config.Ranges, dhcpRange)
			}
		}
	}

	return scanner.Err()
}

func (d *DHCPConfigurator) removeISCDHCPLease(ip string) error {
	// This would require parsing and modifying the leases file
	// For now, just restart the service to reload leases
	return d.restartDHCPD()
}
