package net

import (
	"fmt"
	"os/exec"
	"strings"
)

type FirewallBackend string

const (
	BackendNFTables FirewallBackend = "nftables"
	BackendIPTables FirewallBackend = "iptables"
)

type FirewallChain struct {
	Name     string `json:"name"`
	Table    string `json:"table"`
	Type     string `json:"type"`
	Hook     string `json:"hook"`
	Priority int    `json:"priority"`
	Policy   string `json:"policy"`
}

type FirewallNATRule struct {
	ID          int    `json:"id"`
	Chain       string `json:"chain"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Interface   string `json:"interface"`
	Action      string `json:"action"`
	Target      string `json:"target"`
	Enabled     bool   `json:"enabled"`
}

type FirewallStatus struct {
	Backend     string            `json:"backend"`
	Status      string            `json:"status"`
	RuleCount   int               `json:"rule_count"`
	Chains      []FirewallChain   `json:"chains"`
	NATRules    []FirewallNATRule `json:"nat_rules"`
}

func GetFirewallBackend() (string, error) {
	// Check if nftables is available and preferred
	if isNFTablesAvailable() {
		// Check if nftables is actually being used
		output, err := exec.Command("nft", "list", "tables").Output()
		if err == nil && strings.TrimSpace(string(output)) != "" {
			return string(BackendNFTables), nil
		}
	}

	// Check if iptables is available
	if isIPTablesAvailable() {
		return string(BackendIPTables), nil
	}

	return "", fmt.Errorf("no firewall backend available")
}

func isNFTablesAvailable() bool {
	_, err := exec.LookPath("nft")
	return err == nil
}

func isIPTablesAvailable() bool {
	_, err := exec.LookPath("iptables")
	return err == nil
}

func GetFirewallChains(backend string) ([]string, error) {
	switch FirewallBackend(backend) {
	case BackendNFTables:
		return getNFTablesChains()
	case BackendIPTables:
		return getIPTablesChains()
	default:
		return nil, fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func getNFTablesChains() ([]string, error) {
	output, err := exec.Command("nft", "list", "chains").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list nftables chains: %v", err)
	}

	var chains []string
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "chain ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				chainName := parts[1]
				chains = append(chains, chainName)
			}
		}
	}

	return chains, nil
}

func getIPTablesChains() ([]string, error) {
	var chains []string
	
	// Get chains from filter table
	output, err := exec.Command("iptables", "-t", "filter", "-L").Output()
	if err == nil {
		filterChains := parseIPTablesChains(string(output))
		for _, chain := range filterChains {
			chains = append(chains, "filter:"+chain)
		}
	}

	// Get chains from nat table
	output, err = exec.Command("iptables", "-t", "nat", "-L").Output()
	if err == nil {
		natChains := parseIPTablesChains(string(output))
		for _, chain := range natChains {
			chains = append(chains, "nat:"+chain)
		}
	}

	// Get chains from mangle table
	output, err = exec.Command("iptables", "-t", "mangle", "-L").Output()
	if err == nil {
		mangleChains := parseIPTablesChains(string(output))
		for _, chain := range mangleChains {
			chains = append(chains, "mangle:"+chain)
		}
	}

	return chains, nil
}

func parseIPTablesChains(output string) []string {
	var chains []string
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "Chain ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				chainName := parts[1]
				chains = append(chains, chainName)
			}
		}
	}

	return chains
}

func AddFirewallRule(chain, ruleText string, position int) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return addNFTablesRule(chain, ruleText, position)
	case BackendIPTables:
		return addIPTablesRule(chain, ruleText, position)
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func addNFTablesRule(chain, ruleText string, position int) error {
	// Parse chain to get table and chain name
	parts := strings.Split(chain, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid chain format, expected table:chain")
	}

	table := parts[0]
	chainName := parts[1]

	args := []string{"add", "rule", table, chainName}
	
	if position > 0 {
		args = append(args, "position", fmt.Sprintf("%d", position))
	}

	// Add rule text
	ruleArgs := strings.Fields(ruleText)
	args = append(args, ruleArgs...)

	return exec.Command("nft", args...).Run()
}

func addIPTablesRule(chain, ruleText string, position int) error {
	// Parse chain to get table and chain name
	parts := strings.Split(chain, ":")
	table := "filter"
	chainName := chain

	if len(parts) == 2 {
		table = parts[0]
		chainName = parts[1]
	}

	args := []string{"-t", table}
	
	if position > 0 {
		args = append(args, "-I", chainName, fmt.Sprintf("%d", position))
	} else {
		args = append(args, "-A", chainName)
	}

	// Add rule text
	ruleArgs := strings.Fields(ruleText)
	args = append(args, ruleArgs...)

	return exec.Command("iptables", args...).Run()
}

func RemoveFirewallRule(chain, ruleText string) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return removeNFTablesRule(chain, ruleText)
	case BackendIPTables:
		return removeIPTablesRule(chain, ruleText)
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func removeNFTablesRule(chain, ruleText string) error {
	// For nftables, we need to find the rule handle first
	parts := strings.Split(chain, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid chain format, expected table:chain")
	}

	table := parts[0]
	chainName := parts[1]

	// List rules with handles to find the rule to delete
	output, err := exec.Command("nft", "-a", "list", "chain", table, chainName).Output()
	if err != nil {
		return fmt.Errorf("failed to list rules: %v", err)
	}

	// Find rule handle (this is a simplified implementation)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ruleText) && strings.Contains(line, "# handle") {
			// Extract handle number
			parts := strings.Split(line, "# handle ")
			if len(parts) == 2 {
				handle := strings.TrimSpace(parts[1])
				return exec.Command("nft", "delete", "rule", table, chainName, "handle", handle).Run()
			}
		}
	}

	return fmt.Errorf("rule not found")
}

func removeIPTablesRule(chain, ruleText string) error {
	parts := strings.Split(chain, ":")
	table := "filter"
	chainName := chain

	if len(parts) == 2 {
		table = parts[0]
		chainName = parts[1]
	}

	args := []string{"-t", table, "-D", chainName}
	ruleArgs := strings.Fields(ruleText)
	args = append(args, ruleArgs...)

	return exec.Command("iptables", args...).Run()
}

func FlushFirewallChain(chain string) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return flushNFTablesChain(chain)
	case BackendIPTables:
		return flushIPTablesChain(chain)
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func flushNFTablesChain(chain string) error {
	parts := strings.Split(chain, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid chain format, expected table:chain")
	}

	table := parts[0]
	chainName := parts[1]

	return exec.Command("nft", "flush", "chain", table, chainName).Run()
}

func flushIPTablesChain(chain string) error {
	parts := strings.Split(chain, ":")
	table := "filter"
	chainName := chain

	if len(parts) == 2 {
		table = parts[0]
		chainName = parts[1]
	}

	return exec.Command("iptables", "-t", table, "-F", chainName).Run()
}

func CreateFirewallChain(name, table, chainType string) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return createNFTablesChain(name, table, chainType)
	case BackendIPTables:
		return createIPTablesChain(name, table)
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func createNFTablesChain(name, table, chainType string) error {
	args := []string{"add", "chain", table, name}
	
	if chainType != "" {
		args = append(args, "{", "type", chainType, "hook", "input", "priority", "0", ";", "}")
	}

	return exec.Command("nft", args...).Run()
}

func createIPTablesChain(name, table string) error {
	return exec.Command("iptables", "-t", table, "-N", name).Run()
}

func DeleteFirewallChain(name, table string) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return exec.Command("nft", "delete", "chain", table, name).Run()
	case BackendIPTables:
		return exec.Command("iptables", "-t", table, "-X", name).Run()
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func SetFirewallPolicy(chain, table, policy string) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return exec.Command("nft", "add", "chain", table, chain, "{", "policy", policy, ";", "}").Run()
	case BackendIPTables:
		return exec.Command("iptables", "-t", table, "-P", chain, policy).Run()
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func EnableFirewall() error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return exec.Command("systemctl", "enable", "nftables").Run()
	case BackendIPTables:
		return exec.Command("systemctl", "enable", "iptables").Run()
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func DisableFirewall() error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		// Flush all rules
		exec.Command("nft", "flush", "ruleset").Run()
		return exec.Command("systemctl", "disable", "nftables").Run()
	case BackendIPTables:
		// Set all policies to ACCEPT and flush
		exec.Command("iptables", "-P", "INPUT", "ACCEPT").Run()
		exec.Command("iptables", "-P", "FORWARD", "ACCEPT").Run()
		exec.Command("iptables", "-P", "OUTPUT", "ACCEPT").Run()
		exec.Command("iptables", "-F").Run()
		exec.Command("iptables", "-X").Run()
		return exec.Command("systemctl", "disable", "iptables").Run()
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func GetFirewallStatus() (*FirewallStatus, error) {
	backend, err := GetFirewallBackend()
	if err != nil {
		return nil, err
	}

	status := &FirewallStatus{
		Backend: backend,
		Status:  "unknown",
	}

	// Get chains
	chains, err := GetFirewallChains(backend)
	if err == nil {
		for _, chain := range chains {
			status.Chains = append(status.Chains, FirewallChain{Name: chain})
		}
	}

	// Count rules
	switch FirewallBackend(backend) {
	case BackendNFTables:
		status.RuleCount, status.Status = getNFTablesStatus()
	case BackendIPTables:
		status.RuleCount, status.Status = getIPTablesStatus()
	}

	return status, nil
}

func getNFTablesStatus() (int, string) {
	output, err := exec.Command("nft", "list", "ruleset").Output()
	if err != nil {
		return 0, "error"
	}

	ruleCount := strings.Count(string(output), "\n")
	
	if ruleCount > 0 {
		return ruleCount, "active"
	}
	
	return 0, "inactive"
}

func getIPTablesStatus() (int, string) {
	output, err := exec.Command("iptables-save").Output()
	if err != nil {
		return 0, "error"
	}

	lines := strings.Split(string(output), "\n")
	ruleCount := 0
	
	for _, line := range lines {
		if strings.HasPrefix(line, "-A ") {
			ruleCount++
		}
	}

	if ruleCount > 0 {
		return ruleCount, "active"
	}
	
	return 0, "inactive"
}

func SaveFirewallConfiguration() error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return exec.Command("nft", "list", "ruleset").Run()
	case BackendIPTables:
		return exec.Command("iptables-save").Run()
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func RestoreFirewallConfiguration(configPath string) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return exec.Command("nft", "-f", configPath).Run()
	case BackendIPTables:
		return exec.Command("iptables-restore", configPath).Run()
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func AddNATRule(rule FirewallNATRule) error {
	backend, err := GetFirewallBackend()
	if err != nil {
		return err
	}

	switch FirewallBackend(backend) {
	case BackendNFTables:
		return addNFTablesNATRule(rule)
	case BackendIPTables:
		return addIPTablesNATRule(rule)
	default:
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
}

func addNFTablesNATRule(rule FirewallNATRule) error {
	var args []string
	args = append(args, "add", "rule", "ip", "nat", rule.Chain)

	if rule.Source != "" {
		args = append(args, "ip", "saddr", rule.Source)
	}

	if rule.Destination != "" {
		args = append(args, "ip", "daddr", rule.Destination)
	}

	if rule.Interface != "" {
		args = append(args, "oif", rule.Interface)
	}

	switch rule.Action {
	case "masquerade":
		args = append(args, "masquerade")
	case "snat":
		args = append(args, "snat", "to", rule.Target)
	case "dnat":
		args = append(args, "dnat", "to", rule.Target)
	}

	return exec.Command("nft", args...).Run()
}

func addIPTablesNATRule(rule FirewallNATRule) error {
	var args []string
	args = append(args, "-t", "nat", "-A", rule.Chain)

	if rule.Source != "" {
		args = append(args, "-s", rule.Source)
	}

	if rule.Destination != "" {
		args = append(args, "-d", rule.Destination)
	}

	if rule.Interface != "" {
		args = append(args, "-o", rule.Interface)
	}

	switch rule.Action {
	case "masquerade":
		args = append(args, "-j", "MASQUERADE")
	case "snat":
		args = append(args, "-j", "SNAT", "--to-source", rule.Target)
	case "dnat":
		args = append(args, "-j", "DNAT", "--to-destination", rule.Target)
	}

	return exec.Command("iptables", args...).Run()
}
