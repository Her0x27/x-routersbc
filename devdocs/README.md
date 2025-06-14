### Полное RESTful API Design 
###        API Architecture и Versioning:
```go
// internal/api/versioning.go
type APIVersioning struct {
    CurrentVersion string
    SupportedVersions []string
    DeprecationPolicy DeprecationPolicy
}

type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Meta      *APIMeta    `json:"meta,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
    Version   string      `json:"version"`
}

type APIMeta struct {
    Page       int `json:"page,omitempty"`
    PerPage    int `json:"per_page,omitempty"`
    Total      int `json:"total,omitempty"`
    TotalPages int `json:"total_pages,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
    Field   string `json:"field,omitempty"`
}
```

### Authentication & Authorization API:
```go
// /api/v1/auth/
type AuthAPI struct {
    service AuthService
    limiter RateLimiter
}

// POST /api/v1/auth/login
type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
    Remember bool   `json:"remember"`
}

type LoginResponse struct {
    Token        string    `json:"token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
    User         UserInfo  `json:"user"`
}

// POST /api/v1/auth/refresh
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

// POST /api/v1/auth/logout
// DELETE /api/v1/auth/sessions/:id

// GET /api/v1/auth/sessions - List active sessions
type SessionInfo struct {
    ID        string    `json:"id"`
    UserAgent string    `json:"user_agent"`
    IPAddress string    `json:"ip_address"`
    CreatedAt time.Time `json:"created_at"`
    LastSeen  time.Time `json:"last_seen"`
    Current   bool      `json:"current"`
}
```

### System Management API:
```go
// /api/v1/system/
type SystemAPI struct {
    service SystemService
    auth    AuthMiddleware
}

// GET /api/v1/system/info
type SystemInfo struct {
    Hostname     string            `json:"hostname"`
    Uptime       time.Duration     `json:"uptime"`
    LoadAverage  []float64         `json:"load_average"`
    CPUInfo      CPUInfo           `json:"cpu"`
    MemoryInfo   MemoryInfo        `json:"memory"`
    StorageInfo  []StorageInfo     `json:"storage"`
    NetworkInfo  []NetworkInfo     `json:"network"`
    KernelInfo   KernelInfo        `json:"kernel"`
    Hardware     HardwareInfo      `json:"hardware"`
    Services     []ServiceStatus   `json:"services"`
}

// GET /api/v1/system/status
type SystemStatus struct {
    Status      string             `json:"status"` // healthy, warning, critical
    Alerts      []Alert            `json:"alerts"`
    Services    map[string]string  `json:"services"`
    Resources   ResourceUsage      `json:"resources"`
    LastCheck   time.Time          `json:"last_check"`
}

// POST /api/v1/system/reboot
type RebootRequest struct {
    Delay   int    `json:"delay"` // seconds
    Message string `json:"message,omitempty"`
    Force   bool   `json:"force"`
}

// POST /api/v1/system/shutdown
type ShutdownRequest struct {
    Delay   int    `json:"delay"`
    Message string `json:"message,omitempty"`
}

// GET /api/v1/system/logs
type LogsRequest struct {
    Service   string    `query:"service"`
    Level     string    `query:"level"`
    Since     time.Time `query:"since"`
    Until     time.Time `query:"until"`
    Lines     int       `query:"lines"`
    Follow    bool      `query:"follow"`
}

// GET /api/v1/system/processes
type ProcessInfo struct {
    PID         int     `json:"pid"`
    Name        string  `json:"name"`
    Command     string  `json:"command"`
    User        string  `json:"user"`
    CPUPercent  float64 `json:"cpu_percent"`
    MemoryMB    int64   `json:"memory_mb"`
    Status      string  `json:"status"`
    StartTime   time.Time `json:"start_time"`
}

// POST /api/v1/system/processes/:pid/kill
type KillProcessRequest struct {
    Signal string `json:"signal"` // TERM, KILL, HUP, etc.
}

// GET /api/v1/system/services
// POST /api/v1/system/services/:name/start
// POST /api/v1/system/services/:name/stop
// POST /api/v1/system/services/:name/restart
// POST /api/v1/system/services/:name/reload

```

### Network Management API:
```go
// /api/v1/network/
type NetworkAPI struct {
    service NetworkService
    auth    AuthMiddleware
}

// GET /api/v1/network/interfaces
type NetworkInterface struct {
    Name         string            `json:"name"`
    Type         string            `json:"type"` // ethernet, wireless, bridge, vlan
    Status       string            `json:"status"` // up, down, unknown
    MAC          string            `json:"mac"`
    MTU          int               `json:"mtu"`
    Speed        int64             `json:"speed"` // Mbps
    Duplex       string            `json:"duplex"`
    Addresses    []IPAddress       `json:"addresses"`
    Statistics   InterfaceStats    `json:"statistics"`
    Configuration InterfaceConfig  `json:"configuration"`
}

type InterfaceStats struct {
    RxBytes   uint64 `json:"rx_bytes"`
    TxBytes   uint64 `json:"tx_bytes"`
    RxPackets uint64 `json:"rx_packets"`
    TxPackets uint64 `json:"tx_packets"`
    RxErrors  uint64 `json:"rx_errors"`
    TxErrors  uint64 `json:"tx_errors"`
    RxDropped uint64 `json:"rx_dropped"`
    TxDropped uint64 `json:"tx_dropped"`
}

// POST /api/v1/network/interfaces
type CreateInterfaceRequest struct {
    Name   string                 `json:"name" validate:"required"`
    Type   string                 `json:"type" validate:"required"`
    Config map[string]interface{} `json:"config"`
}

// PUT /api/v1/network/interfaces/:name
type UpdateInterfaceRequest struct {
    Config map[string]interface{} `json:"config"`
    Apply  bool                   `json:"apply"` // Apply immediately or just save
}

// POST /api/v1/network/interfaces/:name/up
// POST /api/v1/network/interfaces/:name/down
// DELETE /api/v1/network/interfaces/:name

// GET /api/v1/network/routing
type RoutingTable struct {
    Routes []Route `json:"routes"`
}

type Route struct {
    Destination string `json:"destination"`
    Gateway     string `json:"gateway"`
    Interface   string `json:"interface"`
    Metric      int    `json:"metric"`
    Protocol    string `json:"protocol"`
    Scope       string `json:"scope"`
    Type        string `json:"type"`
}

// POST /api/v1/network/routing/routes
type CreateRouteRequest struct {
    Destination string `json:"destination" validate:"required"`
    Gateway     string `json:"gateway"`
    Interface   string `json:"interface"`
    Metric      int    `json:"metric"`
}

// DELETE /api/v1/network/routing/routes/:id

// GET /api/v1/network/dhcp
type DHCPConfig struct {
    Enabled     bool          `json:"enabled"`
    Interface   string        `json:"interface"`
    Range       DHCPRange     `json:"range"`
    Options     DHCPOptions   `json:"options"`
    Reservations []DHCPReservation `json:"reservations"`
    Leases      []DHCPLease   `json:"leases"`
}

type DHCPLease struct {
    IP        string    `json:"ip"`
    MAC       string    `json:"mac"`
    Hostname  string    `json:"hostname"`
    ExpiresAt time.Time `json:"expires_at"`
    State     string    `json:"state"`
}

// POST /api/v1/network/dhcp/reservations
type DHCPReservation struct {
    MAC      string `json:"mac" validate:"required"`
    IP       string `json:"ip" validate:"required"`
    Hostname string `json:"hostname"`
}

// GET /api/v1/network/dns
type DNSConfig struct {
    Mode        string      `json:"mode"` // proxy, server, forwarder
    Forwarders  []string    `json:"forwarders"`
    LocalZones  []DNSZone   `json:"local_zones"`
    Security    DNSSecurity `json:"security"`
    Cache       DNSCache    `json:"cache"`
}

// POST /api/v1/network/dns/zones
type DNSZone struct {
    Name    string      `json:"name"`
    Type    string      `json:"type"` // forward, reverse
    Records []DNSRecord `json:"records"`
}

type DNSRecord struct {
    Name  string `json:"name"`
    Type  string `json:"type"` // A, AAAA, CNAME, MX, TXT, etc.
    Value string `json:"value"`
    TTL   int    `json:"ttl"`
}

```

### Wireless Management API:
```go
// /api/v1/wireless/
type WirelessAPI struct {
    service WirelessService
    auth    AuthMiddleware
}

// GET /api/v1/wireless/adapters
type WirelessAdapter struct {
    Name         string              `json:"name"`
    Driver       string              `json:"driver"`
    Chipset      string              `json:"chipset"`
    Modes        []string            `json:"supported_modes"`
    CurrentMode  string              `json:"current_mode"`
    Frequencies  []int               `json:"supported_frequencies"`
    Encryption   []string            `json:"supported_encryption"`
    Status       string              `json:"status"`
    Temperature  float64             `json:"temperature,omitempty"`
}

// GET /api/v1/wireless/scan
type WirelessScanRequest struct {
    Interface string `query:"interface"`
    Active    bool   `query:"active"`
}

type WirelessNetwork struct {
    SSID         string    `json:"ssid"`
    BSSID        string    `json:"bssid"`
    Channel      int       `json:"channel"`
    Frequency    int       `json:"frequency"`
    Signal       int       `json:"signal"` // dBm
    Quality      int       `json:"quality"` // percentage
    Encryption   []string  `json:"encryption"`
    Mode         string    `json:"mode"`
    LastSeen     time.Time `json:"last_seen"`
}

// POST /api/v1/wireless/connect
type WirelessConnectRequest struct {
    Interface string `json:"interface" validate:"required"`
    SSID      string `json:"ssid" validate:"required"`
    Password  string `json:"password,omitempty"`
    Security  string `json:"security"` // WPA2, WPA3, WEP, OPEN
    Hidden    bool   `json:"hidden"`
}

// GET /api/v1/wireless/access-points
type AccessPoint struct {
    Interface    string        `json:"interface"`
    SSID         string        `json:"ssid"`
    Channel      int           `json:"channel"`
    Security     APSecurity    `json:"security"`
    Clients      []APClient    `json:"clients"`
    Statistics   APStatistics  `json:"statistics"`
    Configuration APConfig     `json:"configuration"`
}

type APClient struct {
    MAC          string    `json:"mac"`
    IP           string    `json:"ip,omitempty"`
    Hostname     string    `json:"hostname,omitempty"`
    Signal       int       `json:"signal"`
    ConnectedAt  time.Time `json:"connected_at"`
    RxBytes      uint64    `json:"rx_bytes"`
    TxBytes      uint64    `json:"tx_bytes"`
}

// POST /api/v1/wireless/access-points
type CreateAPRequest struct {
    Interface string   `json:"interface" validate:"required"`
    SSID      string   `json:"ssid" validate:"required"`
    Channel   int      `json:"channel"`
    Security  APSecurity `json:"security"`
    Hidden    bool     `json:"hidden"`
}

// POST /api/v1/wireless/access-points/:id/kick-client
type KickClientRequest struct {
    MAC    string `json:"mac" validate:"required"`
    Reason string `json:"reason,omitempty"`
}

```
### Firewall Management API:
```go
// /api/v1/firewall/
type FirewallAPI struct {
    service FirewallService
    auth    AuthMiddleware
}

// GET /api/v1/firewall/status
type FirewallStatus struct {
    Enabled     bool              `json:"enabled"`
    Backend     string            `json:"backend"` // iptables, nftables
    DefaultPolicy map[string]string `json:"default_policy"`
    RuleCount   map[string]int    `json:"rule_count"`
    Statistics  FirewallStats     `json:"statistics"`
}

// GET /api/v1/firewall/rules
type FirewallRule struct {
    ID          string            `json:"id"`
    Chain       string            `json:"chain"`
    Table       string            `json:"table"`
    Position    int               `json:"position"`
    Action      string            `json:"action"` // ACCEPT, DROP, REJECT, LOG
    Protocol    string            `json:"protocol"`
    Source      string            `json:"source"`
    Destination string            `json:"destination"`
    Port        string            `json:"port"`
    Interface   string            `json:"interface"`
    State       string            `json:"state"`
    Comment     string            `json:"comment"`
    Enabled     bool              `json:"enabled"`
    Statistics  RuleStatistics    `json:"statistics"`
}

// POST /api/v1/firewall/rules
type CreateFirewallRuleRequest struct {
    Chain       string `json:"chain" validate:"required"`
    Action      string `json:"action" validate:"required"`
    Protocol    string `json:"protocol"`
    Source      string `json:"source"`
    Destination string `json:"destination"`
    Port        string `json:"port"`
    Interface   string `json:"interface"`
    Comment     string `json:"comment"`
    Position    int    `json:"position"`
}

// PUT /api/v1/firewall/rules/:id
// DELETE /api/v1/firewall/rules/:id
// POST /api/v1/firewall/rules/:id/enable
// POST /api/v1/firewall/rules/:id/disable

// GET /api/v1/firewall/zones
type FirewallZone struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Interfaces  []string `json:"interfaces"`
    Services    []string `json:"services"`
    Ports       []string `json:"ports"`
    Protocols   []string `json:"protocols"`
    Sources     []string `json:"sources"`
    Target      string   `json:"target"` // default, ACCEPT, REJECT, DROP
    Masquerade  bool     `json:"masquerade"`
    ForwardPorts []ForwardPort `json:"forward_ports"`
}

type ForwardPort struct {
    Port     string `json:"port"`
    Protocol string `json:"protocol"`
    ToPort   string `json:"to_port"`
    ToAddr   string `json:"to_addr"`
}

// POST /api/v1/firewall/zones
type CreateZoneRequest struct {
    Name        string   `json:"name" validate:"required"`
    Description string   `json:"description"`
    Target      string   `json:"target"`
}

// GET /api/v1/firewall/services
type FirewallService struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Ports       []string `json:"ports"`
    Protocols   []string `json:"protocols"`
    Modules     []string `json:"modules"`
    Destinations []string `json:"destinations"`
}

// GET /api/v1/firewall/logs
type FirewallLogRequest struct {
    Action    string    `query:"action"`
    Protocol  string    `query:"protocol"`
    Source    string    `query:"source"`
    Since     time.Time `query:"since"`
    Until     time.Time `query:"until"`
    Limit     int       `query:"limit"`
}

type FirewallLogEntry struct {
    Timestamp   time.Time `json:"timestamp"`
    Action      string    `json:"action"`
    Protocol    string    `json:"protocol"`
    Source      string    `json:"source"`
    Destination string    `json:"destination"`
    Port        int       `json:"port"`
    Interface   string    `json:"interface"`
    Rule        string    `json:"rule"`
}
```

#### **Monitoring & Metrics API:**
```go
// /api/v1/monitoring/
type MonitoringAPI struct {
    service MonitoringService
    auth    AuthMiddleware
}

// GET /api/v1/monitoring/metrics
type MetricsRequest struct {
    Metrics   []string  `query:"metrics"`
    Start     time.Time `query:"start"`
    End       time.Time `query:"end"`
    Step      string    `query:"step"` // 1m, 5m, 1h, etc.
    Format    string    `query:"format"` // json, prometheus
}

type MetricsResponse struct {
    Metrics []MetricSeries `json:"metrics"`
}

type MetricSeries struct {
    Name   string            `json:"name"`
    Labels map[string]string `json:"labels"`
    Values []MetricValue     `json:"values"`
}

type MetricValue struct {
    Timestamp time.Time `json:"timestamp"`
    Value     float64   `json:"value"`
}

// GET /api/v1/monitoring/alerts
type AlertsRequest struct {
    Status    string    `query:"status"` // active, resolved, all
    Severity  string    `query:"severity"`
    Since     time.Time `query:"since"`
    Labels    string    `query:"labels"`
}

type Alert struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Status      string            `json:"status"`
    Severity    string            `json:"severity"`
    Description string            `json:"description"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    StartsAt    time.Time         `json:"starts_at"`
    EndsAt      *time.Time        `json:"ends_at,omitempty"`
    UpdatedAt   time.Time         `json:"updated_at"`
    Fingerprint string            `json:"fingerprint"`
}

// POST /api/v1/monitoring/alerts/:id/acknowledge
type AcknowledgeAlertRequest struct {
    Comment string `json:"comment,omitempty"`
}

// POST /api/v1/monitoring/alerts/:id/resolve
type ResolveAlertRequest struct {
    Comment string `json:"comment,omitempty"`
}

// GET /api/v1/monitoring/health
type HealthCheck struct {
    Status    string                 `json:"status"` // healthy, degraded, unhealthy
    Checks    map[string]CheckResult `json:"checks"`
    Timestamp time.Time              `json:"timestamp"`
    Duration  time.Duration          `json:"duration"`
}

type CheckResult struct {
    Status  string        `json:"status"`
    Message string        `json:"message,omitempty"`
    Duration time.Duration `json:"duration"`
    Details interface{}   `json:"details,omitempty"`
}

// GET /api/v1/monitoring/performance
type PerformanceMetrics struct {
    CPU        CPUMetrics        `json:"cpu"`
    Memory     MemoryMetrics     `json:"memory"`
    Storage    []StorageMetrics  `json:"storage"`
    Network    []NetworkMetrics  `json:"network"`
    Load       LoadMetrics       `json:"load"`
    Temperature []TempSensor     `json:"temperature"`
}

type CPUMetrics struct {
    Usage      float64   `json:"usage_percent"`
    LoadAvg    []float64 `json:"load_average"`
    Cores      int       `json:"cores"`
    Frequency  float64   `json:"frequency_mhz"`
    Temperature float64  `json:"temperature_celsius"`
}
```

#### **Device Management API:**
```go
// /api/v1/devices/
type DeviceAPI struct {
    service DeviceService
    auth    AuthMiddleware
}

// GET /api/v1/devices
type DevicesRequest struct {
    Type     string `query:"type"`
    Status   string `query:"status"`
    Driver   string `query:"driver"`
    Connected bool  `query:"connected"`
}

type Device struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Type         string            `json:"type"`
    Vendor       string            `json:"vendor"`
    Product      string            `json:"product"`
    SerialNumber string            `json:"serial_number"`
    BusInfo      string            `json:"bus_info"`
    Driver       DriverInfo        `json:"driver"`
    Status       string            `json:"status"`
    Connected    bool              `json:"connected"`
    Properties   map[string]string `json:"properties"`
    Capabilities []string          `json:"capabilities"`
    LastSeen     time.Time         `json:"last_seen"`
    Power        PowerInfo         `json:"power,omitempty"`
}

type DriverInfo struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Status      string `json:"status"`
    Module      string `json:"module"`
    Parameters  map[string]string `json:"parameters"`
    AutoLoad    bool   `json:"auto_load"`
}

// POST /api/v1/devices/:id/enable
// POST /api/v1/devices/:id/disable
// POST /api/v1/devices/:id/reset

// GET /api/v1/devices/drivers
type DriverPackage struct {
    Name         string       `json:"name"`
    Version      string       `json:"version"`
    Description  string       `json:"description"`
    Status       string       `json:"status"` // installed, available, incompatible
    SupportedHW  []HardwareID `json:"supported_hardware"`
    Dependencies []string     `json:"dependencies"`
    Size         int64        `json:"size"`
    License      string       `json:"license"`
}

// POST /api/v1/devices/drivers/:name/install
type InstallDriverRequest struct {
    Version string `json:"version,omitempty"`
    Force   bool   `json:"force"`
}

// DELETE /api/v1/devices/drivers/:name

// GET /api/v1/devices/usb
type USBDevice struct {
    BusNumber    int    `json:"bus_number"`
    DeviceNumber int    `json:"device_number"`
    VendorID     string `json:"vendor_id"`
    ProductID    string `json:"product_id"`
    Vendor       string `json:"vendor"`
    Product      string `json:"product"`
    Speed        string `json:"speed"`
    Class        string `json:"class"`
    Power        int    `json:"power_ma"`
}

// POST /api/v1/devices/scan
type ScanDevicesRequest struct {
    Type  string `json:"type,omitempty"` // usb, pci, network
    Force bool   `json:"force"`
}
```

#### **Backup & Restore API:**
```go
// /api/v1/backup/
type BackupAPI struct {
    service BackupService
    auth    AuthMiddleware
}

// GET /api/v1/backup/jobs
type BackupJob struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Type        string            `json:"type"` // full, incremental, differential
    Schedule    string            `json:"schedule"` // cron expression
    Enabled     bool              `json:"enabled"`
    Components  []string          `json:"components"`
    Destination BackupDestination `json:"destination"`
    Retention   RetentionPolicy   `json:"retention"`
    LastRun     *time.Time        `json:"last_run,omitempty"`
    NextRun     *time.Time        `json:"next_run,omitempty"`
    Status      string            `json:"status"`
}

type BackupDestination struct {
    Type   string                 `json:"type"` // local, ftp, sftp, s3, nfs
    Config map[string]interface{} `json:"config"`
}

// POST /api/v1/backup/jobs
type CreateBackupJobRequest struct {
    Name        string            `json:"name" validate:"required"`
    Type        string            `json:"type" validate:"required"`
    Schedule    string            `json:"schedule"`
    Components  []string          `json:"components" validate:"required"`
    Destination BackupDestination `json:"destination" validate:"required"`
    Retention   RetentionPolicy   `json:"retention"`
}

// POST /api/v1/backup/jobs/:id/run
type RunBackupRequest struct {
    Type    string `json:"type,omitempty"` // override job type
    Comment string `json:"comment,omitempty"`
}

// GET /api/v1/backup/archives
type BackupArchive struct {
    ID          string    `json:"id"`
    JobID       string    `json:"job_id"`
    JobName     string    `json:"job_name"`
    Type        string    `json:"type"`
    Size        int64     `json:"size"`
    Compressed  bool      `json:"compressed"`
    Encrypted   bool      `json:"encrypted"`
    Components  []string  `json:"components"`
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`
    Checksum    string    `json:"checksum"`
    Status      string    `json:"status"`
}

// POST /api/v1/backup/restore
type RestoreRequest struct {
    ArchiveID   string   `json:"archive_id" validate:"required"`
    Components  []string `json:"components,omitempty"`
    Destination string   `json:"destination,omitempty"`
    Overwrite   bool     `json:"overwrite"`
    PreviewOnly bool     `json:"preview_only"`
}

type RestorePreview struct {
    Files       []RestoreFile `json:"files"`
    Conflicts   []string      `json:"conflicts"`
    Warnings    []string      `json:"warnings"`
    EstimatedTime time.Duration `json:"estimated_time"`
}

type RestoreFile struct {
    Path         string    `json:"path"`
    Size         int64     `json:"size"`
    ModTime      time.Time `json:"mod_time"`
    Permissions  string    `json:"permissions"`
    Owner        string    `json:"owner"`
    Group        string    `json:"group"`
    WillOverwrite bool     `json:"will_overwrite"`
}

// GET /api/v1/backup/status
type BackupStatus struct {
    RunningJobs   []RunningBackup `json:"running_jobs"`
    QueuedJobs    []QueuedBackup  `json:"queued_jobs"`
    RecentJobs    []RecentBackup  `json:"recent_jobs"`
    StorageUsage  StorageUsage    `json:"storage_usage"`
}

type RunningBackup struct {
    JobID       string        `json:"job_id"`
    JobName     string        `json:"job_name"`
    StartedAt   time.Time     `json:"started_at"`
    Progress    int           `json:"progress_percent"`
    CurrentFile string        `json:"current_file"`
    Speed       int64         `json:"speed_bytes_per_sec"`
    ETA         time.Duration `json:"eta"`
}
```

#### **Configuration Management API:**
```go
// /api/v1/config/
type ConfigAPI struct {
    service ConfigService
    auth    AuthMiddleware
}

// GET /api/v1/config/modules
type ConfigModule struct {
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Version     string            `json:"version"`
    Schema      ConfigSchema      `json:"schema"`
    Current     map[string]interface{} `json:"current"`
    Default     map[string]interface{} `json:"default"`
    Modified    bool              `json:"modified"`
    LastChanged time.Time         `json:"last_changed"`
}

type ConfigSchema struct {
    Properties map[string]PropertySchema `json:"properties"`
    Required   []string                  `json:"required"`
}

type PropertySchema struct {
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Default     interface{} `json:"default,omitempty"`
    Enum        []string    `json:"enum,omitempty"`
    Minimum     *float64    `json:"minimum,omitempty"`
    Maximum     *float64    `json:"maximum,omitempty"`
    Pattern     string      `json:"pattern,omitempty"`
    Format      string      `json:"format,omitempty"`
}

// GET /api/v1/config/modules/:name
// PUT /api/v1/config/modules/:name
type UpdateConfigRequest struct {
    Config  map[string]interface{} `json:"config" validate:"required"`
    Apply   bool                   `json:"apply"` // Apply immediately
    Comment string                 `json:"comment,omitempty"`
}

// POST /api/v1/config/modules/:name/reset
type ResetConfigRequest struct {
    Keys []string `json:"keys,omitempty"` // Reset specific keys, empty = reset all
}

// GET /api/v1/config/templates
type ConfigTemplate struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Category    string                 `json:"category"`
    Variables   []TemplateVariable     `json:"variables"`
    Config      map[string]interface{} `json:"config"`
    Tags        []string               `json:"tags"`
    Author      string                 `json:"author"`
    Version     string                 `json:"version"`
}

type TemplateVariable struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Description string
```
