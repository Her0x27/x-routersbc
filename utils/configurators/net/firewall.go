package net

import (
	"fmt"
	"os/exec"
	"strings"
)

// FirewallConfigurator handles firewall configuration
type FirewallConfigurator struct {
	backend string // "nftables" or "iptables"
}

// FirewallRuleConfig represents a firewall rule configuration
type FirewallRuleConfig struct {
	Chain       string
	Action      string // ACCEPT, DROP, REJECT
	Protocol    string // tcp, udp, icmp, all
	Source      string
	Destination string
	Port        string
	Comment     string
}

// NewFirewallConfigurator creates a new firewall configurator
func NewFirewallConfigurator() *FirewallConfigurator {
	fc := &FirewallConfigurator{}
	fc.detectBackend()
	return fc
}

// detectBackend detects which firewall backend to use
func (fc *FirewallConfigurator) detectBackend() {
	// Check if nftables is available and has rules
	if _, err := exec.LookPath("nft"); err == nil {
		cmd := exec.Command("nft", "list", "tables")
		if err := cmd.Run(); err == nil {
			fc.backend = "nftables"
			return
		}
	}
	
	// Default to iptables
	fc.backend = "iptables"
}

// GetBackend returns the current firewall backend
func (fc *FirewallConfigurator) GetBackend() string {
	return fc.backend
}

// SetBackend sets the firewall backend
func (fc *FirewallConfigurator) SetBackend(backend string) error {
	if backend != "nftables" && backend != "iptables" {
		return fmt.Errorf("unsupported firewall backend: %s", backend)
	}
	
	fc.backend = backend
	return nil
}

// AddRule adds a firewall rule
func (fc *FirewallConfigurator) AddRule(rule *FirewallRuleConfig) error {
	if fc.backend == "nftables" {
		return fc.addNFTablesRule(rule)
	} else {
		return fc.addIPTablesRule(rule)
	}
}

// RemoveRule removes a firewall rule
func (fc *FirewallConfigurator) RemoveRule(rule *FirewallRuleConfig) error {
	if fc.backend == "nftables" {
		return fc.removeNFTablesRule(rule)
	} else {
		return fc.removeIPTablesRule(rule)
	}
}

// FlushRules removes all firewall rules
func (fc *FirewallConfigurator) FlushRules() error {
	if fc.backend == "nftables" {
		return fc.flushNFTables()
	} else {
		return fc.flushIPTables()
	}
}

// addNFTablesRule adds a rule using nftables
func (fc *FirewallConfigurator) addNFTablesRule(rule *FirewallRuleConfig) error {
	// Create table and chain if they don't exist
	cmd := exec.Command("nft", "add", "table", "ip", "filter")
	cmd.Run() // Ignore error if table already exists
	
	chainCmd := fmt.Sprintf("add chain ip filter %s { type filter hook input priority 0; }", rule.Chain)
	cmd = exec.Command("nft", strings.Fields(chainCmd)...)
	cmd.Run() // Ignore error if chain already exists
	
	// Build rule command
	ruleCmd := fmt.Sprintf("add rule ip filter %s", rule.Chain)
	
	if rule.Protocol != "" && rule.Protocol != "all" {
		ruleCmd += fmt.Sprintf(" ip protocol %s", rule.Protocol)
	}
	
	if rule.Source != "" {
		ruleCmd += fmt.Sprintf(" ip saddr %s", rule.Source)
	}
	
	if rule.Destination != "" {
		ruleCmd += fmt.Sprintf(" ip daddr %s", rule.Destination)
	}
	
	if rule.Port != "" && (rule.Protocol == "tcp" || rule.Protocol == "udp") {
		ruleCmd += fmt.Sprintf(" %s dport %s", rule.Protocol, rule.Port)
	}
	
	ruleCmd += fmt.Sprintf(" %s", strings.ToLower(rule.Action))
	
	if rule.Comment != "" {
		ruleCmd += fmt.Sprintf(" comment \"%s\"", rule.Comment)
	}
	
	// Execute rule
	cmd = exec.Command("nft", strings.Fields(ruleCmd)...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add nftables rule: %v", err)
	}
	
	return nil
}

// addIPTablesRule adds a rule using iptables
func (fc *FirewallConfigurator) addIPTablesRule(rule *FirewallRuleConfig) error {
	args := []string{"-A", rule.Chain}
	
	if rule.Protocol != "" && rule.Protocol != "all" {
		args = append(args, "-p", rule.Protocol)
	}
	
	if rule.Source != "" {
		args = append(args, "-s", rule.Source)
	}
	
	if rule.Destination != "" {
		args = append(args, "-d", rule.Destination)
	}
	
	if rule.Port != "" && (rule.Protocol == "tcp" || rule.Protocol == "udp") {
		args = append(args, "--dport", rule.Port)
	}
	
	args = append(args, "-j", strings.ToUpper(rule.Action))
	
	if rule.Comment != "" {
		args = append(args, "-m", "comment", "--comment", rule.Comment)
	}
	
	cmd := exec.Command("iptables", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add iptables rule: %v", err)
	}
	
	return nil
}

// removeNFTablesRule removes a rule using nftables
func (fc *FirewallConfigurator) removeNFTablesRule(rule *FirewallRuleConfig) error {
	// This is simplified - in practice, you'd need to identify the rule handle
	// For now, we'll flush and recreate rules without this one
	return fmt.Errorf("nftables rule removal not implemented - use flush and recreate")
}

// removeIPTablesRule removes a rule using iptables
func (fc *FirewallConfigurator) removeIPTablesRule(rule *FirewallRuleConfig) error {
	args := []string{"-D", rule.Chain}
	
	if rule.Protocol != "" && rule.Protocol != "all" {
		args = append(args, "-p", rule.Protocol)
	}
	
	if rule.Source != "" {
		args = append(args, "-s", rule.Source)
	}
	
	if rule.Destination != "" {
		args = append(args, "-d", rule.Destination)
	}
	
	if rule.Port != "" && (rule.Protocol == "tcp" || rule.Protocol == "udp") {
		args = append(args, "--dport", rule.Port)
	}
	
	args = append(args, "-j", strings.ToUpper(rule.Action))
	
	cmd := exec.Command("iptables", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove iptables rule: %v", err)
	}
	
	return nil
}

// flushNFTables flushes all nftables rules
func (fc *FirewallConfigurator) flushNFTables() error {
	cmd := exec.Command("nft", "flush", "ruleset")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to flush nftables: %v", err)
	}
	return nil
}

// flushIPTables flushes all iptables rules
func (fc *FirewallConfigurator) flushIPTables() error {
	tables := []string{"filter", "nat", "mangle"}
	
	for _, table := range tables {
		cmd := exec.Command("iptables", "-t", table, "-F")
		cmd.Run() // Ignore errors for missing tables
		
		cmd = exec.Command("iptables", "-t", table, "-X")
		cmd.Run() // Ignore errors
	}
	
	return nil
}

// SaveConfiguration saves the current firewall configuration
func (fc *FirewallConfigurator) SaveConfiguration() error {
	if fc.backend == "nftables" {
		// Save nftables configuration
		cmd := exec.Command("sh", "-c", "nft list ruleset > /etc/nftables.conf")
		return cmd.Run()
	} else {
		// Save iptables configuration
		cmd := exec.Command("iptables-save")
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		
		// Write to rules file (location varies by distribution)
		rulesFiles := []string{
			"/etc/iptables/rules.v4",
			"/etc/iptables.rules",
			"/etc/sysconfig/iptables",
		}
		
		for _, file := range rulesFiles {
			if err := os.WriteFile(file, output, 0644); err == nil {
				return nil
			}
		}
		
		return fmt.Errorf("could not save iptables rules to any known location")
	}
}

// LoadConfiguration loads firewall configuration
func (fc *FirewallConfigurator) LoadConfiguration() error {
	if fc.backend == "nftables" {
		// Load nftables configuration
		cmd := exec.Command("nft", "-f", "/etc/nftables.conf")
		return cmd.Run()
	} else {
		// Load iptables configuration
		rulesFiles := []string{
			"/etc/iptables/rules.v4",
			"/etc/iptables.rules",
			"/etc/sysconfig/iptables",
		}
		
		for _, file := range rulesFiles {
			if _, err := os.Stat(file); err == nil {
				cmd := exec.Command("iptables-restore", file)
				return cmd.Run()
			}
		}
		
		return fmt.Errorf("no iptables rules file found")
	}
}
