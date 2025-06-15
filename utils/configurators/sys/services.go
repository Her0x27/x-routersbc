package sys

import (
        "fmt"
        "os/exec"
        "strings"
        "time"
)

type ServiceStatus struct {
        Name        string    `json:"name"`
        Status      string    `json:"status"`
        Enabled     bool      `json:"enabled"`
        Active      bool      `json:"active"`
        Running     bool      `json:"running"`
        PID         int       `json:"pid"`
        Uptime      string    `json:"uptime"`
        Memory      string    `json:"memory"`
        CPU         string    `json:"cpu"`
        Description string    `json:"description"`
        LoadState   string    `json:"load_state"`
        SubState    string    `json:"sub_state"`
}

type SystemTime struct {
        CurrentTime string   `json:"current_time"`
        Timezone    string   `json:"timezone"`
        NTPEnabled  bool     `json:"ntp_enabled"`
        NTPServers  []string `json:"ntp_servers"`
}

func GetServiceStatus(serviceName string) (*ServiceStatus, error) {
        // Check if systemd is available
        if isSystemdAvailable() {
                return getSystemdServiceStatus(serviceName)
        }

        // Fallback to SysV init
        return getSysVServiceStatus(serviceName)
}

func isSystemdAvailable() bool {
        _, err := exec.LookPath("systemctl")
        return err == nil
}

func getSystemdServiceStatus(serviceName string) (*ServiceStatus, error) {
        service := &ServiceStatus{Name: serviceName}

        // Get service status
        output, err := exec.Command("systemctl", "show", serviceName, "--no-page").Output()
        if err != nil {
                return service, fmt.Errorf("failed to get service status: %v", err)
        }

        // Parse systemctl show output
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
                if strings.Contains(line, "=") {
                        parts := strings.SplitN(line, "=", 2)
                        if len(parts) == 2 {
                                key := strings.TrimSpace(parts[0])
                                value := strings.TrimSpace(parts[1])

                                switch key {
                                case "ActiveState":
                                        service.Status = value
                                        service.Active = value == "active"
                                case "SubState":
                                        service.SubState = value
                                        service.Running = value == "running"
                                case "LoadState":
                                        service.LoadState = value
                                case "UnitFileState":
                                        service.Enabled = value == "enabled"
                                case "MainPID":
                                        if value != "0" && value != "" {
                                                if pid := parseInt(value); pid > 0 {
                                                        service.PID = pid
                                                }
                                        }
                                case "Description":
                                        service.Description = value
                                case "ActiveEnterTimestamp":
                                        if value != "" {
                                                service.Uptime = calculateUptime(value)
                                        }
                                }
                        }
                }
        }

        // Get memory and CPU usage if service is running
        if service.Running && service.PID > 0 {
                service.Memory, service.CPU = getProcessResources(service.PID)
        }

        return service, nil
}

func getSysVServiceStatus(serviceName string) (*ServiceStatus, error) {
        service := &ServiceStatus{Name: serviceName}

        // Try service command
        err := exec.Command("service", serviceName, "status").Run()
        if err == nil {
                service.Status = "active"
                service.Active = true
                service.Running = true
        } else {
                service.Status = "inactive"
                service.Active = false
                service.Running = false
        }

        // Check if service is enabled (this varies by distribution)
        err = exec.Command("chkconfig", "--list", serviceName).Run()
        if err == nil {
                service.Enabled = true
        }

        return service, nil
}

func parseInt(s string) int {
        var result int
        fmt.Sscanf(s, "%d", &result)
        return result
}

func calculateUptime(timestamp string) string {
        // Parse systemd timestamp format
        // This is a simplified implementation
        return "unknown"
}

func getProcessResources(pid int) (string, string) {
        // Get memory usage
        memory := "0"
        cpu := "0"

        // Use ps to get process info
        output, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "rss,pcpu", "--no-headers").Output()
        if err == nil {
                fields := strings.Fields(string(output))
                if len(fields) >= 2 {
                        memory = fields[0] + " KB"
                        cpu = fields[1] + "%"
                }
        }

        return memory, cpu
}

func StartService(serviceName string) error {
        if isSystemdAvailable() {
                return exec.Command("systemctl", "start", serviceName).Run()
        }

        // Fallback to service command
        return exec.Command("service", serviceName, "start").Run()
}

func StopService(serviceName string) error {
        if isSystemdAvailable() {
                return exec.Command("systemctl", "stop", serviceName).Run()
        }

        // Fallback to service command
        return exec.Command("service", serviceName, "stop").Run()
}

func RestartService(serviceName string) error {
        if isSystemdAvailable() {
                return exec.Command("systemctl", "restart", serviceName).Run()
        }

        // Fallback to service command
        return exec.Command("service", serviceName, "restart").Run()
}

func ReloadService(serviceName string) error {
        if isSystemdAvailable() {
                return exec.Command("systemctl", "reload", serviceName).Run()
        }

        // Fallback to service command
        return exec.Command("service", serviceName, "reload").Run()
}

func EnableService(serviceName string) error {
        if isSystemdAvailable() {
                return exec.Command("systemctl", "enable", serviceName).Run()
        }

        // Fallback to chkconfig or update-rc.d
        if err := exec.Command("chkconfig", serviceName, "on").Run(); err == nil {
                return nil
        }

        return exec.Command("update-rc.d", serviceName, "enable").Run()
}

func DisableService(serviceName string) error {
        if isSystemdAvailable() {
                return exec.Command("systemctl", "disable", serviceName).Run()
        }

        // Fallback to chkconfig or update-rc.d
        if err := exec.Command("chkconfig", serviceName, "off").Run(); err == nil {
                return nil
        }

        return exec.Command("update-rc.d", serviceName, "disable").Run()
}

func GetAllServices() ([]ServiceStatus, error) {
        if isSystemdAvailable() {
                return getSystemdServices()
        }

        return getSysVServices()
}

func getSystemdServices() ([]ServiceStatus, error) {
        var services []ServiceStatus

        output, err := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-page", "--no-legend").Output()
        if err != nil {
                return services, fmt.Errorf("failed to list services: %v", err)
        }

        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
                if strings.TrimSpace(line) == "" {
                        continue
                }

                fields := strings.Fields(line)
                if len(fields) >= 4 {
                        serviceName := fields[0]
                        // Remove .service suffix
                        serviceName = strings.TrimSuffix(serviceName, ".service")

                        service := ServiceStatus{
                                Name:      serviceName,
                                LoadState: fields[1],
                                Status:    fields[2],
                                SubState:  fields[3],
                                Active:    fields[2] == "active",
                                Running:   fields[3] == "running",
                        }

                        if len(fields) > 4 {
                                service.Description = strings.Join(fields[4:], " ")
                        }

                        services = append(services, service)
                }
        }

        return services, nil
}

func getSysVServices() ([]ServiceStatus, error) {
        var services []ServiceStatus

        // This is a simplified implementation for SysV systems
        // In practice, you'd need to check /etc/init.d/ and other locations

        output, err := exec.Command("service", "--status-all").Output()
        if err != nil {
                return services, fmt.Errorf("failed to list services: %v", err)
        }

        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
                if strings.TrimSpace(line) == "" {
                        continue
                }

                // Parse service status output
                if strings.Contains(line, "[") {
                        parts := strings.Fields(line)
                        if len(parts) >= 2 {
                                status := "inactive"
                                if strings.Contains(line, "[+]") {
                                        status = "active"
                                }

                                serviceName := parts[len(parts)-1]
                                service := ServiceStatus{
                                        Name:    serviceName,
                                        Status:  status,
                                        Active:  status == "active",
                                        Running: status == "active",
                                }

                                services = append(services, service)
                        }
                }
        }

        return services, nil
}

func GetSystemTime() (*SystemTime, error) {
        timeInfo := &SystemTime{}

        // Get current time
        output, err := exec.Command("date", "+%Y-%m-%d %H:%M:%S %Z").Output()
        if err == nil {
                timeInfo.CurrentTime = strings.TrimSpace(string(output))
        }

        // Get timezone
        output, err = exec.Command("timedatectl", "show", "--property=Timezone", "--value").Output()
        if err == nil {
                timeInfo.Timezone = strings.TrimSpace(string(output))
        } else {
                // Fallback to reading timezone file
                output, err = exec.Command("cat", "/etc/timezone").Output()
                if err == nil {
                        timeInfo.Timezone = strings.TrimSpace(string(output))
                }
        }

        // Check if NTP is enabled
        err = exec.Command("timedatectl", "show", "--property=NTP", "--value").Run()
        if err == nil {
                output, err = exec.Command("timedatectl", "show", "--property=NTP", "--value").Output()
                if err == nil {
                        timeInfo.NTPEnabled = strings.TrimSpace(string(output)) == "yes"
                }
        }

        // Get NTP servers (if chrony is used)
        if output, err := exec.Command("chronyc", "sources").Output(); err == nil {
                timeInfo.NTPServers = parseNTPServers(string(output))
        } else if output, err := exec.Command("ntpq", "-p").Output(); err == nil {
                timeInfo.NTPServers = parseNTPServersFromNtpq(string(output))
        }

        return timeInfo, nil
}

func parseNTPServers(output string) []string {
        var servers []string
        lines := strings.Split(output, "\n")

        for _, line := range lines {
                if strings.HasPrefix(line, "^") || strings.HasPrefix(line, "=") {
                        fields := strings.Fields(line)
                        if len(fields) >= 2 {
                                server := strings.TrimPrefix(fields[1], "^")
                                server = strings.TrimPrefix(server, "=")
                                servers = append(servers, server)
                        }
                }
        }

        return servers
}

func parseNTPServersFromNtpq(output string) []string {
        var servers []string
        lines := strings.Split(output, "\n")

        for _, line := range lines {
                if strings.HasPrefix(line, "*") || strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
                        fields := strings.Fields(line)
                        if len(fields) >= 1 {
                                server := strings.TrimPrefix(fields[0], "*")
                                server = strings.TrimPrefix(server, "+")
                                server = strings.TrimPrefix(server, "-")
                                servers = append(servers, server)
                        }
                }
        }

        return servers
}

func SetSystemTime(timezone string, ntpEnabled bool, ntpServers []string) error {
        // Set timezone
        if timezone != "" {
                if err := exec.Command("timedatectl", "set-timezone", timezone).Run(); err != nil {
                        return fmt.Errorf("failed to set timezone: %v", err)
                }
        }

        // Enable/disable NTP
        ntpSetting := "false"
        if ntpEnabled {
                ntpSetting = "true"
        }
        
        if err := exec.Command("timedatectl", "set-ntp", ntpSetting).Run(); err != nil {
                return fmt.Errorf("failed to set NTP: %v", err)
        }

        // Configure NTP servers if provided
        if len(ntpServers) > 0 {
                if err := configureNTPServers(ntpServers); err != nil {
                        return fmt.Errorf("failed to configure NTP servers: %v", err)
                }
        }

        return nil
}

func configureNTPServers(servers []string) error {
        // Try to configure chrony first
        if err := configureChronyServers(servers); err == nil {
                return RestartService("chrony")
        }

        // Fallback to ntpd
        if err := configureNtpdServers(servers); err == nil {
                return RestartService("ntp")
        }

        return fmt.Errorf("no NTP service available")
}

func configureChronyServers(servers []string) error {
        configPath := "/etc/chrony/chrony.conf"
        
        // This is a simplified implementation
        // In practice, you'd need to properly parse and update the config file
        var config []string
        config = append(config, "# NTP servers configured by router-sbc")
        
        for _, server := range servers {
                config = append(config, fmt.Sprintf("server %s iburst", server))
        }
        
        // Add other default configuration
        config = append(config, "driftfile /var/lib/chrony/drift")
        config = append(config, "makestep 1.0 3")
        config = append(config, "rtcsync")
        
        configContent := strings.Join(config, "\n")
        return writeFile(configPath, configContent)
}

func configureNtpdServers(servers []string) error {
        configPath := "/etc/ntp.conf"
        
        var config []string
        config = append(config, "# NTP servers configured by router-sbc")
        
        for _, server := range servers {
                config = append(config, fmt.Sprintf("server %s", server))
        }
        
        // Add other default configuration
        config = append(config, "driftfile /var/lib/ntp/drift")
        config = append(config, "restrict default kod notrap nomodify nopeer noquery")
        config = append(config, "restrict 127.0.0.1")
        
        configContent := strings.Join(config, "\n")
        return writeFile(configPath, configContent)
}

func writeFile(path, content string) error {
        return exec.Command("sh", "-c", fmt.Sprintf("echo '%s' > %s", content, path)).Run()
}

func GetSystemLoad() (map[string]interface{}, error) {
        load := make(map[string]interface{})

        // Get load average
        output, err := exec.Command("cat", "/proc/loadavg").Output()
        if err == nil {
                parts := strings.Fields(string(output))
                if len(parts) >= 3 {
                        load["load_1m"] = parts[0]
                        load["load_5m"] = parts[1]
                        load["load_15m"] = parts[2]
                }
        }

        // Get uptime
        output, err = exec.Command("uptime", "-s").Output()
        if err == nil {
                load["boot_time"] = strings.TrimSpace(string(output))
        }

        // Get process count
        output, err = exec.Command("ps", "aux").Output()
        if err == nil {
                processCount := len(strings.Split(string(output), "\n")) - 1
                load["process_count"] = processCount
        }

        return load, nil
}

func GetServiceLogs(serviceName string, lines int) ([]string, error) {
        if isSystemdAvailable() {
                output, err := exec.Command("journalctl", "-u", serviceName, "-n", fmt.Sprintf("%d", lines), "--no-pager").Output()
                if err != nil {
                        return nil, fmt.Errorf("failed to get service logs: %v", err)
                }
                return strings.Split(string(output), "\n"), nil
        }

        // Fallback to syslog
        output, err := exec.Command("tail", "-n", fmt.Sprintf("%d", lines), "/var/log/syslog").Output()
        if err != nil {
                return nil, fmt.Errorf("failed to get logs: %v", err)
        }

        return strings.Split(string(output), "\n"), nil
}

func ScheduleServiceRestart(serviceName string, delay time.Duration) error {
        // Schedule a service restart using at command
        command := fmt.Sprintf("systemctl restart %s", serviceName)
        atTime := time.Now().Add(delay).Format("15:04 Jan 2")
        
        cmd := exec.Command("at", atTime)
        cmd.Stdin = strings.NewReader(command)
        
        return cmd.Run()
}
