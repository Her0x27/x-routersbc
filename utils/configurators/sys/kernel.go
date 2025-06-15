package sys

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type KernelParameter struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type KernelModule struct {
	Name         string   `json:"name"`
	Size         int      `json:"size"`
	UsedBy       []string `json:"used_by"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
}

const (
	sysctlConfPath = "/etc/sysctl.conf"
	sysctlDPath    = "/etc/sysctl.d/"
	procSysPath    = "/proc/sys/"
)

func GetKernelParameters() ([]KernelParameter, error) {
	var parameters []KernelParameter

	// Read current sysctl values
	output, err := exec.Command("sysctl", "-a").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kernel parameters: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, " = ", 2)
		if len(parts) == 2 {
			param := KernelParameter{
				Name:     strings.TrimSpace(parts[0]),
				Value:    strings.TrimSpace(parts[1]),
				Category: getParameterCategory(parts[0]),
			}
			parameters = append(parameters, param)
		}
	}

	return parameters, nil
}

func getParameterCategory(name string) string {
	switch {
	case strings.HasPrefix(name, "net."):
		return "network"
	case strings.HasPrefix(name, "vm."):
		return "memory"
	case strings.HasPrefix(name, "fs."):
		return "filesystem"
	case strings.HasPrefix(name, "kernel."):
		return "kernel"
	case strings.HasPrefix(name, "user."):
		return "user"
	default:
		return "other"
	}
}

func GetKernelParameter(name string) (string, error) {
	output, err := exec.Command("sysctl", "-n", name).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get parameter %s: %v", name, err)
	}

	return strings.TrimSpace(string(output)), nil
}

func SetKernelParameter(name, value string, persistent bool) error {
	// Set current value
	if err := exec.Command("sysctl", "-w", fmt.Sprintf("%s=%s", name, value)).Run(); err != nil {
		return fmt.Errorf("failed to set parameter %s: %v", name, err)
	}

	// Make persistent if requested
	if persistent {
		return setKernelParameterPersistent(name, value)
	}

	return nil
}

func setKernelParameterPersistent(name, value string) error {
	configFile := fmt.Sprintf("%s99-router-sbc.conf", sysctlDPath)
	
	// Create sysctl.d directory if it doesn't exist
	if err := os.MkdirAll(sysctlDPath, 0755); err != nil {
		return fmt.Errorf("failed to create sysctl.d directory: %v", err)
	}

	// Read existing configuration
	var existingLines []string
	if data, err := ioutil.ReadFile(configFile); err == nil {
		existingLines = strings.Split(string(data), "\n")
	}

	// Update or add the parameter
	paramLine := fmt.Sprintf("%s = %s", name, value)
	found := false

	for i, line := range existingLines {
		if strings.HasPrefix(strings.TrimSpace(line), name+"=") || 
		   strings.HasPrefix(strings.TrimSpace(line), name+" =") {
			existingLines[i] = paramLine
			found = true
			break
		}
	}

	if !found {
		existingLines = append(existingLines, paramLine)
	}

	// Write back to file
	content := strings.Join(existingLines, "\n")
	return ioutil.WriteFile(configFile, []byte(content), 0644)
}

func RemoveKernelParameter(name string) error {
	configFile := fmt.Sprintf("%s99-router-sbc.conf", sysctlDPath)
	
	// Read existing configuration
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil // File doesn't exist, nothing to remove
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string

	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), name+"=") && 
		   !strings.HasPrefix(strings.TrimSpace(line), name+" =") {
			newLines = append(newLines, line)
		}
	}

	// Write back to file
	content := strings.Join(newLines, "\n")
	return ioutil.WriteFile(configFile, []byte(content), 0644)
}

func GetLoadedKernelModules() ([]KernelModule, error) {
	var modules []KernelModule

	data, err := ioutil.ReadFile("/proc/modules")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/modules: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 4 {
			module := KernelModule{
				Name:   parts[0],
				Status: "loaded",
			}

			// Parse size
			if size, err := strconv.Atoi(parts[1]); err == nil {
				module.Size = size
			}

			// Parse dependencies
			if parts[3] != "-" {
				module.Dependencies = strings.Split(parts[3], ",")
			}

			modules = append(modules, module)
		}
	}

	return modules, nil
}

func LoadKernelModule(name string, parameters map[string]string) error {
	args := []string{name}

	// Add parameters
	for key, value := range parameters {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}

	if err := exec.Command("modprobe", args...).Run(); err != nil {
		return fmt.Errorf("failed to load module %s: %v", name, err)
	}

	return nil
}

func UnloadKernelModule(name string) error {
	if err := exec.Command("modprobe", "-r", name).Run(); err != nil {
		return fmt.Errorf("failed to unload module %s: %v", name, err)
	}

	return nil
}

func IsKernelModuleLoaded(name string) (bool, error) {
	modules, err := GetLoadedKernelModules()
	if err != nil {
		return false, err
	}

	for _, module := range modules {
		if module.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func GetAvailableKernelModules() ([]string, error) {
	output, err := exec.Command("find", "/lib/modules", "-name", "*.ko", "-type", "f").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to find kernel modules: %v", err)
	}

	var modules []string
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Extract module name from path
		parts := strings.Split(line, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			moduleName := strings.TrimSuffix(filename, ".ko")
			modules = append(modules, moduleName)
		}
	}

	return modules, nil
}

func GetKernelVersion() (string, error) {
	output, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get kernel version: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func GetKernelCommandLine() (string, error) {
	data, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return "", fmt.Errorf("failed to read kernel command line: %v", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func EnableIPForwarding() error {
	return SetKernelParameter("net.ipv4.ip_forward", "1", true)
}

func DisableIPForwarding() error {
	return SetKernelParameter("net.ipv4.ip_forward", "0", true)
}

func IsIPForwardingEnabled() (bool, error) {
	value, err := GetKernelParameter("net.ipv4.ip_forward")
	if err != nil {
		return false, err
	}

	return value == "1", nil
}

func EnableIPv6() error {
	return SetKernelParameter("net.ipv6.conf.all.disable_ipv6", "0", true)
}

func DisableIPv6() error {
	return SetKernelParameter("net.ipv6.conf.all.disable_ipv6", "1", true)
}

func IsIPv6Enabled() (bool, error) {
	value, err := GetKernelParameter("net.ipv6.conf.all.disable_ipv6")
	if err != nil {
		return false, err
	}

	return value == "0", nil
}

func OptimizeNetworkPerformance() error {
	optimizations := map[string]string{
		"net.core.rmem_max":         "134217728",
		"net.core.wmem_max":         "134217728",
		"net.ipv4.tcp_rmem":         "4096 87380 134217728",
		"net.ipv4.tcp_wmem":         "4096 65536 134217728",
		"net.core.netdev_max_backlog": "5000",
		"net.ipv4.tcp_congestion_control": "bbr",
	}

	for param, value := range optimizations {
		if err := SetKernelParameter(param, value, true); err != nil {
			return fmt.Errorf("failed to set %s: %v", param, err)
		}
	}

	return nil
}

func OptimizeMemoryManagement() error {
	optimizations := map[string]string{
		"vm.swappiness":       "10",
		"vm.dirty_ratio":      "15",
		"vm.dirty_background_ratio": "5",
		"vm.vfs_cache_pressure": "50",
	}

	for param, value := range optimizations {
		if err := SetKernelParameter(param, value, true); err != nil {
			return fmt.Errorf("failed to set %s: %v", param, err)
		}
	}

	return nil
}

func GetKernelLimits() (map[string]string, error) {
	limits := make(map[string]string)

	// Get file descriptor limits
	data, err := ioutil.ReadFile("/proc/sys/fs/file-max")
	if err == nil {
		limits["max_files"] = strings.TrimSpace(string(data))
	}

	// Get process limits
	data, err = ioutil.ReadFile("/proc/sys/kernel/pid_max")
	if err == nil {
		limits["max_processes"] = strings.TrimSpace(string(data))
	}

	// Get memory limits
	data, err = ioutil.ReadFile("/proc/meminfo")
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					limits["total_memory"] = parts[1] + " kB"
				}
				break
			}
		}
	}

	return limits, nil
}

func TuneForRouter() error {
	routerSettings := map[string]string{
		// Network forwarding
		"net.ipv4.ip_forward":                   "1",
		"net.ipv6.conf.all.forwarding":         "1",
		
		// Network security
		"net.ipv4.conf.all.rp_filter":          "1",
		"net.ipv4.conf.default.rp_filter":      "1",
		"net.ipv4.icmp_echo_ignore_broadcasts":  "1",
		"net.ipv4.icmp_ignore_bogus_error_responses": "1",
		"net.ipv4.conf.all.log_martians":       "1",
		"net.ipv4.conf.all.accept_source_route": "0",
		"net.ipv4.conf.default.accept_source_route": "0",
		"net.ipv4.conf.all.accept_redirects":    "0",
		"net.ipv4.conf.default.accept_redirects": "0",
		"net.ipv4.conf.all.send_redirects":      "0",
		
		// Connection tracking
		"net.netfilter.nf_conntrack_max":        "65536",
		"net.netfilter.nf_conntrack_tcp_timeout_established": "7200",
		
		// Buffer sizes
		"net.core.rmem_default":     "262144",
		"net.core.rmem_max":         "16777216",
		"net.core.wmem_default":     "262144",
		"net.core.wmem_max":         "16777216",
		"net.core.netdev_max_backlog": "2500",
	}

	for param, value := range routerSettings {
		if err := SetKernelParameter(param, value, true); err != nil {
			// Log error but continue with other parameters
			fmt.Printf("Warning: failed to set %s: %v\n", param, err)
		}
	}

	return nil
}
