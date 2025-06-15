package sys

import (
	"fmt"
	"os/exec"
	"strings"
)

// ServiceManager handles system service management
type ServiceManager struct {
	systemType string // "systemd", "sysv", "upstart"
}

// ServiceStatus represents the status of a system service
type ServiceStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`    // active, inactive, failed
	Enabled   bool   `json:"enabled"`   // auto-start enabled
	Running   bool   `json:"running"`   // currently running
	PID       string `json:"pid"`       // process ID if running
	Uptime    string `json:"uptime"`    // how long it's been running
	Memory    string `json:"memory"`    // memory usage
	CPU       string `json:"cpu"`       // CPU usage
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	sm := &ServiceManager{}
	sm.detectSystemType()
	return sm
}

// detectSystemType detects the init system type
func (sm *ServiceManager) detectSystemType() {
	// Check for systemd
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "--version")
		if err := cmd.Run(); err == nil {
			sm.systemType = "systemd"
			return
		}
	}
	
	// Check for upstart
	if _, err := exec.LookPath("initctl"); err == nil {
		sm.systemType = "upstart"
		return
	}
	
	// Default to SysV
	sm.systemType = "sysv"
}

// GetSystemType returns the detected init system type
func (sm *ServiceManager) GetSystemType() string {
	return sm.systemType
}

// StartService starts a system service
func (sm *ServiceManager) StartService(serviceName string) error {
	switch sm.systemType {
	case "systemd":
		cmd := exec.Command("systemctl", "start", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service %s: %v", serviceName, err)
		}
	case "upstart":
		cmd := exec.Command("start", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service %s: %v", serviceName, err)
		}
	case "sysv":
		cmd := exec.Command("service", serviceName, "start")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service %s: %v", serviceName, err)
		}
	default:
		return fmt.Errorf("unsupported init system: %s", sm.systemType)
	}
	
	return nil
}

// StopService stops a system service
func (sm *ServiceManager) StopService(serviceName string) error {
	switch sm.systemType {
	case "systemd":
		cmd := exec.Command("systemctl", "stop", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service %s: %v", serviceName, err)
		}
	case "upstart":
		cmd := exec.Command("stop", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service %s: %v", serviceName, err)
		}
	case "sysv":
		cmd := exec.Command("service", serviceName, "stop")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service %s: %v", serviceName, err)
		}
	default:
		return fmt.Errorf("unsupported init system: %s", sm.systemType)
	}
	
	return nil
}

// RestartService restarts a system service
func (sm *ServiceManager) RestartService(serviceName string) error {
	switch sm.systemType {
	case "systemd":
		cmd := exec.Command("systemctl", "restart", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart service %s: %v", serviceName, err)
		}
	case "upstart":
		cmd := exec.Command("restart", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart service %s: %v", serviceName, err)
		}
	case "sysv":
		cmd := exec.Command("service", serviceName, "restart")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart service %s: %v", serviceName, err)
		}
	default:
		return fmt.Errorf("unsupported init system: %s", sm.systemType)
	}
	
	return nil
}

// ReloadService reloads a system service configuration
func (sm *ServiceManager) ReloadService(serviceName string) error {
	switch sm.systemType {
	case "systemd":
		cmd := exec.Command("systemctl", "reload", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reload service %s: %v", serviceName, err)
		}
	case "upstart":
		cmd := exec.Command("reload", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reload service %s: %v", serviceName, err)
		}
	case "sysv":
		cmd := exec.Command("service", serviceName, "reload")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reload service %s: %v", serviceName, err)
		}
	default:
		return fmt.Errorf("unsupported init system: %s", sm.systemType)
	}
	
	return nil
}

// EnableService enables a service to start automatically
func (sm *ServiceManager) EnableService(serviceName string) error {
	switch sm.systemType {
	case "systemd":
		cmd := exec.Command("systemctl", "enable", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to enable service %s: %v", serviceName, err)
		}
	case "upstart":
		// Upstart services are enabled by default if they have start on conditions
		return nil
	case "sysv":
		cmd := exec.Command("update-rc.d", serviceName, "enable")
		if err := cmd.Run(); err != nil {
			// Try alternative method
			cmd = exec.Command("chkconfig", serviceName, "on")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to enable service %s: %v", serviceName, err)
			}
		}
	default:
		return fmt.Errorf("unsupported init system: %s", sm.systemType)
	}
	
	return nil
}

// DisableService disables a service from starting automatically
func (sm *ServiceManager) DisableService(serviceName string) error {
	switch sm.systemType {
	case "systemd":
		cmd := exec.Command("systemctl", "disable", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to disable service %s: %v", serviceName, err)
		}
	case "upstart":
		// Upstart services can be disabled by modifying their configuration
		return fmt.Errorf("upstart service disable not implemented")
	case "sysv":
		cmd := exec.Command("update-rc.d", serviceName, "disable")
		if err := cmd.Run(); err != nil {
			// Try alternative method
			cmd = exec.Command("chkconfig", serviceName, "off")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to disable service %s: %v", serviceName, err)
			}
		}
	default:
		return fmt.Errorf("unsupported init system: %s", sm.systemType)
	}
	
	return nil
}

// GetServiceStatus gets the status of a system service
func (sm *ServiceManager) GetServiceStatus(serviceName string) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name: serviceName,
	}
	
	switch sm.systemType {
	case "systemd":
		return sm.getSystemdStatus(serviceName)
	case "upstart":
		return sm.getUpstartStatus(serviceName)
	case "sysv":
		return sm.getSysVStatus(serviceName)
	default:
		return nil, fmt.Errorf("unsupported init system: %s", sm.systemType)
	}
}

// getSystemdStatus gets service status from systemd
func (sm *ServiceManager) getSystemdStatus(serviceName string) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name: serviceName,
	}
	
	// Get service status
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err == nil {
		status.Status = strings.TrimSpace(string(output))
		status.Running = status.Status == "active"
	} else {
		status.Status = "inactive"
		status.Running = false
	}
	
	// Check if enabled
	cmd = exec.Command("systemctl", "is-enabled", serviceName)
	output, err = cmd.Output()
	if err == nil {
		enabled := strings.TrimSpace(string(output))
		status.Enabled = enabled == "enabled"
	}
	
	// Get detailed status
	cmd = exec.Command("systemctl", "status", serviceName, "--no-pager", "-l")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Main PID:") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					status.PID = parts[2]
				}
			}
			if strings.Contains(line, "Active:") {
				if strings.Contains(line, "since") {
					parts := strings.Split(line, "since")
					if len(parts) > 1 {
						status.Uptime = strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}
	
	return status, nil
}

// getUpstartStatus gets service status from upstart
func (sm *ServiceManager) getUpstartStatus(serviceName string) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name: serviceName,
	}
	
	cmd := exec.Command("status", serviceName)
	output, err := cmd.Output()
	if err != nil {
		status.Status = "inactive"
		status.Running = false
		return status, nil
	}
	
	outputStr := strings.TrimSpace(string(output))
	if strings.Contains(outputStr, "start/running") {
		status.Status = "active"
		status.Running = true
		
		// Extract PID
		if parts := strings.Split(outputStr, "process "); len(parts) > 1 {
			pidPart := strings.Fields(parts[1])
			if len(pidPart) > 0 {
				status.PID = pidPart[0]
			}
		}
	} else {
		status.Status = "inactive"
		status.Running = false
	}
	
	// Upstart services are typically enabled if the config file exists
	status.Enabled = true
	
	return status, nil
}

// getSysVStatus gets service status from SysV init
func (sm *ServiceManager) getSysVStatus(serviceName string) (*ServiceStatus, error) {
	status := &ServiceStatus{
		Name: serviceName,
	}
	
	cmd := exec.Command("service", serviceName, "status")
	output, err := cmd.Output()
	outputStr := strings.ToLower(strings.TrimSpace(string(output)))
	
	if err == nil && (strings.Contains(outputStr, "running") || strings.Contains(outputStr, "active")) {
		status.Status = "active"
		status.Running = true
	} else {
		status.Status = "inactive"
		status.Running = false
	}
	
	// Check if enabled using chkconfig or update-rc.d
	cmd = exec.Command("chkconfig", "--list", serviceName)
	if err := cmd.Run(); err == nil {
		status.Enabled = true
	} else {
		// Try update-rc.d
		cmd = exec.Command("update-rc.d", "-n", serviceName, "remove")
		if err := cmd.Run(); err != nil {
			status.Enabled = true // Service exists in rc.d
		}
	}
	
	return status, nil
}

// ListServices lists all available services
func (sm *ServiceManager) ListServices() ([]ServiceStatus, error) {
	var services []ServiceStatus
	
	switch sm.systemType {
	case "systemd":
		cmd := exec.Command("systemctl", "list-units", "--type=service", "--no-pager", "--no-legend")
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				serviceName := strings.TrimSuffix(fields[0], ".service")
				status := ServiceStatus{
					Name:    serviceName,
					Status:  fields[2],
					Running: fields[2] == "active",
				}
				services = append(services, status)
			}
		}
		
	case "upstart":
		cmd := exec.Command("initctl", "list")
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				status := ServiceStatus{
					Name:    parts[0],
					Running: strings.Contains(line, "start/running"),
					Enabled: true,
				}
				if status.Running {
					status.Status = "active"
				} else {
					status.Status = "inactive"
				}
				services = append(services, status)
			}
		}
		
	case "sysv":
		cmd := exec.Command("service", "--status-all")
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			// SysV service --status-all format: [ + ] servicename
			if strings.Contains(line, "]") {
				parts := strings.Split(line, "]")
				if len(parts) >= 2 {
					serviceName := strings.TrimSpace(parts[1])
					status := ServiceStatus{
						Name:    serviceName,
						Running: strings.Contains(line, "+"),
						Enabled: strings.Contains(line, "+"),
					}
					if status.Running {
						status.Status = "active"
					} else {
						status.Status = "inactive"
					}
					services = append(services, status)
				}
			}
		}
	}
	
	return services, nil
}
