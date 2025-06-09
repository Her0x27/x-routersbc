# RouterSBC GENERAL MODULES TASKS - Complete Network Management Implementation

**Project**: RouterSBC Hardware-Driven Network Management System  
**Date**: 2025-06-09  
**Type**: Module Implementation Guide - Echo Framework  
**Framework**: Echo v4  

---

# МОДУЛЬНАЯ АРХИТЕКТУРА СЕТЕВОГО УПРАВЛЕНИЯ

## Module Structure Standard (ОБЯЗАТЕЛЬНО)

### Каждый модуль ДОЛЖЕН содержать:
```
modules/{module_name}/
├── module.json          # Конфигурация и зависимости от оборудования
├── handlers.go          # Echo handlers для маршрутов
├── {feature}_handler.go # Специализированные обработчики
└── templates/           # HTML шаблоны модуля
    ├── index.html      # Главная страница модуля
    └── {feature}.html  # Функциональные страницы
```

### module.json Structure (СТАНДАРТ):
```json
{
  "name": "network",
  "display_name": "Network Settings",
  "menu_icon": "network-wired",
  "menu_order": 2,
  "menu_section": "configuration",
  "requires_hardware": ["ethernet"],
  "optional_hardware": ["wireless"],
  "dependencies": {
    "wireless_wan": {
      "requires": ["wireless_client_connection"],
      "conflicts": ["wireless_ap_same_interface"],
      "description": "Wireless WAN requires active client connection"
    }
  }
}
```

---

# 1. DASHBOARD MODULE (modules/dashboard/)

## Dashboard Implementation

### module.json
```json
{
  "name": "dashboard",
  "display_name": "Dashboard",
  "menu_icon": "dashboard",
  "menu_order": 1,
  "menu_section": "overview",
  "requires_hardware": [],
  "dependencies": {}
}
```

### handlers.go
```go
package dashboard

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "routersbc/core/hardware"
    "routersbc/core/tasks"
)

type DashboardHandler struct {
    hardwareDetector *hardware.HardwareDetector
    taskManager      *tasks.TaskManager
}

func NewDashboardHandler(hd *hardware.HardwareDetector, tm *tasks.TaskManager) *DashboardHandler {
    return &DashboardHandler{
        hardwareDetector: hd,
        taskManager:      tm,
    }
}

func (dh *DashboardHandler) GetSystemOverview(c echo.Context) error {
    data := map[string]interface{}{
        "system_status":     dh.getSystemStatus(),
        "hardware_summary":  dh.getHardwareSummary(),
        "network_status":    dh.getNetworkStatus(),
        "task_summary":      dh.getTaskSummary(),
        "performance_metrics": dh.getPerformanceMetrics(),
    }
    
    return c.Render(http.StatusOK, "dashboard/index", data)
}

func (dh *DashboardHandler) getSystemStatus() map[string]interface{} {
    systemInfo := dh.hardwareDetector.GetSystemInfo()
    
    return map[string]interface{}{
        "status":      "operational",
        "uptime":      systemInfo.Uptime,
        "load_avg":    systemInfo.LoadAverage,
        "memory_usage": float64(systemInfo.MemoryFree) / float64(systemInfo.MemoryTotal) * 100,
        "cpu_usage":   dh.calculateCPUUsage(systemInfo.CPUInfo),
    }
}

func (dh *DashboardHandler) getHardwareSummary() map[string]interface{} {
    ethernetCount := len(dh.hardwareDetector.GetEthernetInterfaces())
    wirelessCount := len(dh.hardwareDetector.GetWirelessAdapters())
    
    return map[string]interface{}{
        "ethernet_interfaces": ethernetCount,
        "wireless_adapters":   wirelessCount,
        "multi_wan_capable":   ethernetCount >= 2,
        "wireless_capable":    wirelessCount > 0,
    }
}

func (dh *DashboardHandler) getNetworkStatus() map[string]interface{} {
    interfaces := dh.hardwareDetector.GetEthernetInterfaces()
    activeInterfaces := 0
    totalTraffic := int64(0)
    
    for _, iface := range interfaces {
        if iface.Status == "up" {
            activeInterfaces++
        }
        totalTraffic += iface.RXBytes + iface.TXBytes
    }
    
    return map[string]interface{}{
        "active_interfaces": activeInterfaces,
        "total_interfaces":  len(interfaces),
        "total_traffic":     totalTraffic,
        "connectivity":      dh.checkConnectivity(),
    }
}

func (dh *DashboardHandler) getTaskSummary() map[string]interface{} {
    pendingTasks := dh.taskManager.GetPendingTasks()
    
    return map[string]interface{}{
        "pending_tasks":   len(pendingTasks),
        "recent_tasks":    dh.taskManager.GetRecentTasks(5),
        "critical_tasks":  dh.taskManager.GetCriticalTasks(),
    }
}
```

### templates/index.html
```html
{{define "title"}}RouterSBC Dashboard{{end}}

{{define "content"}}
<div class="p-8">
    <!-- System Status Overview -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {{template "card" .SystemStatusCard}}
        {{template "card" .HardwareCard}}
        {{template "card" .NetworkCard}}
        {{template "card" .TasksCard}}
    </div>
    
    <!-- Hardware Information -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
        <div class="card bg-base-100 shadow-xl">
            <div class="card-body">
                <h2 class="card-title">Network Interfaces</h2>
                <div class="overflow-x-auto">
                    <table class="table table-sm">
                        <thead>
                            <tr>
                                <th>Interface</th>
                                <th>Type</th>
                                <th>Status</th>
                                <th>Speed</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range .HardwareInfo.interfaces}}
                            <tr>
                                <td>{{.Name}}</td>
                                <td>{{.Type}}</td>
                                <td>
                                    <div class="badge {{if eq .Status "up"}}badge-success{{else}}badge-error{{end}}">
                                        {{.Status}}
                                    </div>
                                </td>
                                <td>{{.Speed}}</td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
        
        <div class="card bg-base-100 shadow-xl">
            <div class="card-body">
                <h2 class="card-title">Wireless Adapters</h2>
                {{if .HardwareInfo.wireless_adapters}}
                <div class="space-y-4">
                    {{range .HardwareInfo.wireless_adapters}}
                    <div class="flex justify-between items-center p-4 border rounded-lg">
                        <div>
                            <div class="font-semibold">{{.Interface}}</div>
                            <div class="text-sm text-base-content/60">{{.Driver}} - {{.Mode}}</div>
                        </div>
                        <div class="badge {{if eq .Status "up"}}badge-success{{else}}badge-error{{end}}">
                            {{.Status}}
                        </div>
                    </div>
                    {{end}}
                </div>
                {{else}}
                <div class="text-center text-base-content/60 py-8">
                    No wireless adapters detected
                </div>
                {{end}}
            </div>
        </div>
    </div>
    
    <!-- Recent Tasks -->
    {{if .Tasks}}
    <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
            <div class="flex justify-between items-center">
                <h2 class="card-title">Recent Configuration Changes</h2>
                <a href="/tasks" class="btn btn-primary btn-sm">View All Tasks</a>
            </div>
            <div class="overflow-x-auto">
                <table class="table table-sm">
                    <thead>
                        <tr>
                            <th>Description</th>
                            <th>Module</th>
                            <th>Status</th>
                            <th>Created</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Tasks}}
                        <tr>
                            <td>{{.Description}}</td>
                            <td>{{.Module}}</td>
                            <td>
                                <div class="badge 
                                    {{if eq .Status "completed"}}badge-success
                                    {{else if eq .Status "failed"}}badge-error
                                    {{else if eq .Status "pending"}}badge-warning
                                    {{else}}badge-info{{end}}">
                                    {{.Status}}
                                </div>
                            </td>
                            <td>{{.CreatedAt.Format "15:04 02/01"}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    {{end}}
</div>
{{end}}

{{define "scripts"}}
<script src="/static/js/routersbc.js"></script>
<script>
// Real-time updates
setInterval(async () => {
    try {
        const response = await fetch('/api/system/status');
        const data = await response.json();
        updateSystemStatus(data);
    } catch (error) {
        console.error('Failed to update system status:', error);
    }
}, 30000);

function updateSystemStatus(data) {
    // Update system status indicators
    document.querySelector('#system-status').textContent = data.status;
    document.querySelector('#uptime').textContent = data.uptime;
    document.querySelector('#memory-usage').textContent = data.memory_usage + '%';
}
</script>
{{end}}

{{template "layout.html" .}}
```

---

# 2. NETWORK SETTINGS MODULE (modules/network/)

## Network Module Implementation

### module.json
```json
{
  "name": "network",
  "display_name": "Network Settings",
  "menu_icon": "network-wired",
  "menu_order": 2,
  "menu_section": "configuration",
  "requires_hardware": ["ethernet"],
  "optional_hardware": ["wireless"],
  "dependencies": {
    "wireless_wan": {
      "requires": ["wireless_client_connection"],
      "conflicts": ["wireless_ap_same_interface"],
      "description": "Wireless WAN requires active client connection"
    },
    "dhcp_server": {
      "conflicts": ["dhcp_client_same_interface"],
      "requires": ["static_ip_configuration"],
      "description": "DHCP server conflicts with DHCP client on same interface"
    },
    "multi_wan": {
      "requires": ["multiple_wan_interfaces"],
      "hardware": ["ethernet"],
      "description": "Multi-WAN requires multiple network interfaces"
    }
  }
}
```

### handlers.go
```go
package network

import (
    "fmt"
    "net/http"
    "github.com/labstack/echo/v4"
    "routersbc/core/hardware"
    "routersbc/core/tasks"
    "routersbc/core/config"
)

type NetworkHandler struct {
    hardwareDetector *hardware.HardwareDetector
    taskManager      *tasks.TaskManager
    configManager    *config.ConfigurationManager
    validator        *DependencyValidator
}

func NewNetworkHandler(hd *hardware.HardwareDetector, tm *tasks.TaskManager, cm *config.ConfigurationManager) *NetworkHandler {
    return &NetworkHandler{
        hardwareDetector: hd,
        taskManager:      tm,
        configManager:    cm,
        validator:        NewDependencyValidator(),
    }
}

// 1. Hardware Information Display
func (nh *NetworkHandler) GetHardwareInfo(c echo.Context) error {
    interfaces := nh.hardwareDetector.GetEthernetInterfaces()
    capabilities := nh.hardwareDetector.GetCapabilities()
    
    data := map[string]interface{}{
        "interfaces":    interfaces,
        "capabilities":  capabilities,
        "scan_time":     nh.hardwareDetector.GetLastScanTime(),
    }
    
    return c.Render(http.StatusOK, "network/hardware", data)
}

// 2. LAN Configuration
func (nh *NetworkHandler) GetLANConfig(c echo.Context) error {
    currentConfig := nh.configManager.GetLANConfig()
    
    data := map[string]interface{}{
        "lan_config":       currentConfig,
        "available_interfaces": nh.getAvailableLANInterfaces(),
        "dhcp_modes":       []string{"PROXY", "RELAY", "SERVER"},
        "dns_modes":        []string{"AUTO", "PROXY", "IP_SERVERS"},
    }
    
    return c.Render(http.StatusOK, "network/lan", data)
}

func (nh *NetworkHandler) UpdateLANConfig(c echo.Context) error {
    var config LANConfig
    if err := c.Bind(&config); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    // Validate configuration
    if err := nh.validator.ValidateLANConfig(config); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    // Create task for configuration change
    task := tasks.Task{
        Type:        "network",
        Module:      "network",
        Action:      "configure_lan",
        Description: fmt.Sprintf("Configure LAN: DHCP %s, DNS %s", config.DHCP.Mode, config.DNS.Mode),
        Config:      map[string]interface{}{"lan": config},
        Priority:    5,
    }
    
    if err := nh.taskManager.AddTask(task); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, map[string]string{"status": "task_created"})
}

// 3. WAN Configuration
func (nh *NetworkHandler) GetWANConfig(c echo.Context) error {
    currentConfig := nh.configManager.GetWANConfig()
    
    data := map[string]interface{}{
        "wan_configs":         currentConfig,
        "available_interfaces": nh.getAvailableWANInterfaces(),
        "connection_types":    []string{"DHCP", "STATIC", "PPPOE"},
        "wireless_clients":    nh.getWirelessClientConnections(),
    }
    
    return c.Render(http.StatusOK, "network/wan", data)
}

func (nh *NetworkHandler) UpdateWANConfig(c echo.Context) error {
    var config WANConfig
    if err := c.Bind(&config); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    // Check wireless WAN dependency
    if isWirelessInterface(config.Interface) {
        if !nh.hasWirelessClientConnection(config.Interface) {
            return c.JSON(http.StatusBadRequest, map[string]string{
                "error": "Wireless WAN requires active client connection",
            })
        }
    }
    
    // Validate configuration
    if err := nh.validator.ValidateWANConfig(config); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    // Create task for configuration change
    task := tasks.Task{
        Type:        "network",
        Module:      "network",
        Action:      "configure_wan",
        Description: fmt.Sprintf("Configure WAN %s: %s connection", config.Interface, config.Type),
        Config:      map[string]interface{}{"wan": config},
        Priority:    7,
    }
    
    if err := nh.taskManager.AddTask(task); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, map[string]string{"status": "task_created"})
}

// 3.1. Multi-WAN Configuration
func (nh *NetworkHandler) GetMultiWANConfig(c echo.Context) error {
    currentConfig := nh.configManager.GetMultiWANConfig()
    availableWANs := nh.getConfiguredWANInterfaces()
    
    if len(availableWANs) < 2 {
        return c.Render(http.StatusOK, "network/multiwan_unavailable", map[string]interface{}{
            "message": "Multi-WAN requires at least 2 configured WAN connections",
        })
    }
    
    data := map[string]interface{}{
        "multiwan_config": currentConfig,
        "available_wans":  availableWANs,
        "modes":          []string{"Failover", "LoadBalance"},
    }
    
    return c.Render(http.StatusOK, "network/multiwan", data)
}

func (nh *NetworkHandler) UpdateMultiWANConfig(c echo.Context) error {
    var config MultiWANConfig
    if err := c.Bind(&config); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    // Validate Multi-WAN requirements
    if len(config.WANs) < 2 {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Multi-WAN requires at least 2 WAN connections",
        })
    }
    
    // Validate all selected WANs are configured
    for _, wan := range config.WANs {
        if !nh.isWANConfigured(wan) {
            return c.JSON(http.StatusBadRequest, map[string]string{
                "error": fmt.Sprintf("WAN %s is not configured", wan),
            })
        }
    }
    
    // Create task for Multi-WAN configuration
    task := tasks.Task{
        Type:        "network",
        Module:      "network",
        Action:      "configure_multiwan",
        Description: fmt.Sprintf("Configure Multi-WAN: %s mode with %d connections", config.Mode, len(config.WANs)),
        Config:      map[string]interface{}{"multiwan": config},
        Priority:    8,
    }
    
    if err := nh.taskManager.AddTask(task); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, map[string]string{"status": "task_created"})
}

// Helper functions
func (nh *NetworkHandler) getAvailableLANInterfaces() []NetworkInterface {
    interfaces := nh.hardwareDetector.GetEthernetInterfaces()
    var available []NetworkInterface
    
    for _, iface := range interfaces {
        if !nh.isInterfaceUsedAsWAN(iface.Name) {
            available = append(available, *iface)
        }
    }
    
    return available
}

func (nh *NetworkHandler) getAvailableWANInterfaces() []interface{} {
    var available []interface{}
    
    // Add ethernet interfaces
    ethInterfaces := nh.hardwareDetector.GetEthernetInterfaces()
    for _, iface := range ethInterfaces {
        available = append(available, map[string]interface{}{
            "name": iface.Name,
            "type": "ethernet",
            "status": iface.Status,
        })
    }
    
    // Add wireless interfaces with client connections
    wirelessConnections := nh.getWirelessClientConnections()
    for _, conn := range wirelessConnections {
        available = append(available, map[string]interface{}{
            "name": conn.Interface,
            "type": "wireless",
            "status": "connected",
            "ssid": conn.SSID,
        })
    }
    
    return available
}
```

### templates/index.html (Network Overview)
```html
{{define "title"}}Network Settings{{end}}

{{define "content"}}
<div class="p-8">
    <!-- Network Overview -->
    <div class="mb-8">
        <h1 class="text-3xl font-bold text-base-content mb-2">Network Settings</h1>
        <p class="text-base-content/60">Configure network interfaces and connections</p>
    </div>
    
    <!-- Navigation Tabs -->
    {{template "tabs" .NetworkTabs}}
    
    <div class="tab-content">
        <!-- Hardware Information Tab -->
        <div id="hardware-tab" class="tab-pane active">
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <!-- Ethernet Interfaces -->
                <div class="card bg-base-100 shadow-xl">
                    <div class="card-body">
                        <div class="flex justify-between items-center">
                            <h2 class="card-title">Ethernet Interfaces</h2>
                            <button class="btn btn-sm btn-outline" onclick="rescanHardware()">
                                Rescan
                            </button>
                        </div>
                        
                        {{if .HardwareInfo.interfaces}}
                        <div class="overflow-x-auto">
                            <table class="table table-sm">
                                <thead>
                                    <tr>
                                        <th>Interface</th>
                                        <th>Status</th>
                                        <th>Speed</th>
                                        <th>MAC</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {{range .HardwareInfo.interfaces}}
                                    {{if eq .Type "ethernet"}}
                                    <tr>
                                        <td>{{.Name}}</td>
                                        <td>
                                            <div class="badge {{if eq .Status "up"}}badge-success{{else}}badge-error{{end}}">
                                                {{.Status}}
                                            </div>
                                        </td>
                                        <td>{{.Speed}}</td>
                                        <td class="font-mono text-sm">{{.MAC}}</td>
                                    </tr>
                                    {{end}}
                                    {{end}}
                                </tbody>
                            </table>
                        </div>
                        {{else}}
                        <div class="text-center text-base-content/60 py-8">
                            No ethernet interfaces detected
                        </div>
                        {{end}}
                    </div>
                </div>
                
                <!-- System Capabilities -->
                <div class="card bg-base-100 shadow-xl">
                    <div class="card-body">
                        <h2 class="card-title">Network Capabilities</h2>
                        
                        <div class="space-y-4">
                            <div class="flex justify-between items-center">
                                <span>Multi-WAN Support</span>
                                <div class="badge {{if .HardwareInfo.capabilities.MultiWANCapable}}badge-success{{else}}badge-error{{end}}">
                                    {{if .HardwareInfo.capabilities.MultiWANCapable}}Available{{else}}Unavailable{{end}}
                                </div>
                            </div>
                            
                            <div class="flex justify-between items-center">
                                <span>VLAN Support</span>
                                <div class="badge {{if .HardwareInfo.capabilities.VLANSupport}}badge-success{{else}}badge-error{{end}}">
                                    {{if .HardwareInfo.capabilities.VLANSupport}}Supported{{else}}Not Supported{{end}}
                                </div>
                            </div>
                            
                            <div class="flex justify-between items-center">
                                <span>Bridge Support</span>
                                <div class="badge {{if .HardwareInfo.capabilities.BridgeSupport}}badge-success{{else}}badge-error{{end}}">
                                    {{if .HardwareInfo.capabilities.BridgeSupport}}Supported{{else}}Not Supported{{end}}
                                </div>
                            </div>
                            
                            <div class="flex justify-between items-center">
                                <span>Traffic Control</span>
                                <div class="badge {{if .HardwareInfo.capabilities.TrafficControlSupport}}badge-success{{else}}badge-error{{end}}">
                                    {{if .HardwareInfo.capabilities.TrafficControlSupport}}Available{{else}}Unavailable{{end}}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
{{end}}

{{define "scripts"}}
<script src="/static/js/hardware_detection.js"></script>
<script src="/static/js/dependency_checker.js"></script>
<script src="/static/js/real_time_validation.js"></script>
<script>
// Initialize hardware detection
const hardwareDetector = new HardwareDetector();
hardwareDetector.init();

// Initialize dependency checker
const dependencyChecker = new DependencyChecker();
dependencyChecker.init();

// Initialize real-time validation
const validator = new RealTimeValidator();
validator.init();

async function rescanHardware() {
    await hardwareDetector.rescanHardware();
    location.reload();
}
</script>
{{end}}

{{template "layout.html" .}}
```

---

# 3. WIRELESS MODULE (modules/wireless/)

### module.json
```json
{
  "name": "wireless",
  "display_name": "Wireless Settings",
  "menu_icon": "wifi",
  "menu_order": 3,
  "menu_section": "configuration",
  "requires_hardware": ["wireless"],
  "dependencies": {
    "ap_mode": {
      "conflicts": ["client_mode_same_interface"],
      "description": "AP mode conflicts with client mode on same interface"
    },
    "monitor_mode": {
      "conflicts": ["ap_mode_same_interface", "client_mode_same_interface"],
      "description": "Monitor mode conflicts with other modes on same interface"
    }
  }
}
```

### handlers.go
```go
package wireless

import (
    "fmt"
    "net/http"
    "github.com/labstack/echo/v4"
    "routersbc/core/hardware"
    "routersbc/core/tasks"
)

type WirelessHandler struct {
    hardwareDetector *hardware.HardwareDetector
    taskManager      *tasks.TaskManager
    scanner          *WirelessScanner
    validator        *WirelessValidator
}

// 4. Wireless Adapter Management
func (wh *WirelessHandler) GetAdapters(c echo.Context) error {
    adapters := wh.hardwareDetector.GetWirelessAdapters()
    
    data := map[string]interface{}{
        "adapters":        adapters,
        "supported_modes": []string{"managed", "master", "monitor", "mesh"},
        "scan_time":       wh.hardwareDetector.GetLastScanTime(),
    }
    
    return c.Render(http.StatusOK, "wireless/adapters", data)
}

// Network Creation Wizard
func (wh *WirelessHandler) GetCreateNetwork(c echo.Context) error {
    adapters := wh.hardwareDetector.GetWirelessAdapters()
    
    data := map[string]interface{}{
        "adapters":          adapters,
        "available_modes":   wh.getAvailableModes(),
        "security_options":  []string{"open", "WPA", "WPA2", "WPA3"},
        "frequency_bands":   wh.getAvailableFrequencies(),
    }
    
    return c.Render(http.StatusOK, "wireless/create_network", data)
}

func (wh *WirelessHandler) CreateNetwork(c echo.Context) error {
    var config WirelessNetworkConfig
    if err := c.Bind(&config); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    
    // Validate adapter capabilities
    adapter := wh.getAdapterByInterface(config.Interface)
    if adapter == nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid wireless interface",
        })
    }
    
    // Check mode support
    if !wh.supportsMode(adapter, config.Mode) {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": fmt.Sprintf("Interface %s does not support %s mode", config.Interface, config.Mode),
        })
    }
    
    // Check frequency support
    if !wh.supportsFrequency(adapter, config.Frequency) {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": fmt.Sprintf("Interface %s does not support %s frequency", config.Interface, config.Frequency),
        })
    }
    
    // Check channel availability
    if !wh.isChannelAvailable(adapter, config.Frequency, config.Channel) {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": fmt.Sprintf("Channel %d not available for %s on %s", config.Channel, config.Frequency, config.Interface),
        })
    }
    
    // Check security support
    if !wh.supportsSecurity(adapter, config.Security) {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": fmt.Sprintf("Interface %s does not support %s security", config.Interface, config.Security),
        })
    }
    
    // Create task based on mode
    var task tasks.Task
    switch config.Mode {
    case "master": // AP mode
        task = tasks.Task{
            Type:        "wireless",
            Module:      "wireless",
            Action:      "create_ap",
            Description: fmt.Sprintf("Create AP %s on %s (%s, Channel %d)", config.SSID, config.Interface, config.Security, config.Channel),
            Config:      map[string]interface{}{"wireless_ap": config},
            Priority:    6,
        }
    case "managed": // Client mode
        task = tasks.Task{
            Type:        "wireless",
            Module:      "wireless",
            Action:      "connect_client",
            Description: fmt.Sprintf("Connect to %s on %s", config.SSID, config.Interface),
            Config:      map[string]interface{}{"wireless_client": config},
            Priority:    6,
        }
    case "monitor": // Monitor mode
        task = tasks.Task{
            Type:        "wireless",
            Module:      "wireless",
            Action:      "enable_monitor",
            Description: fmt.Sprintf("Enable monitor mode on %s (Channel %d)", config.Interface, config.Channel),
            Config:      map[string]interface{}{"wireless_monitor": config},
            Priority:    4,
        }
    default:
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Unsupported wireless mode",
        })
    }
    
    if err := wh.taskManager.AddTask(task); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    return c.JSON(http.StatusOK, map[string]string{"status": "task_created"})
}

// Helper functions for hardware validation
func (wh *WirelessHandler) getAvailableModes() map[string][]string {
    adapters := wh.hardwareDetector.GetWirelessAdapters()
    modes := make(map[string][]string)
    
    for _, adapter := range adapters {
        modes[adapter.Interface] = adapter.SupportedModes
    }
    
    return modes
}

func (wh *WirelessHandler) getAvailableFrequencies() map[string][]float64 {
    adapters := wh.hardwareDetector.GetWirelessAdapters()
    frequencies := make(map[string][]float64)
    
    for _, adapter := range adapters {
        frequencies[adapter.Interface] = adapter.SupportedFrequencies
    }
    
    return frequencies
}

func (wh *WirelessHandler) getAvailableChannels() map[string]map[string][]int {
    adapters := wh.hardwareDetector.GetWirelessAdapters()
    channels := make(map[string]map[string][]int)
    
    for _, adapter := range adapters {
        channels[adapter.Interface] = make(map[string][]int)
        
        // Map channels to frequency bands
        for _, freq := range adapter.SupportedFrequencies {
            freqStr := fmt.Sprintf("%.1f", freq)
            if freq < 3.0 { // 2.4GHz
                channels[adapter.Interface]["2.4"] = wh.get24GHzChannels(adapter.SupportedChannels)
            } else if freq < 6.0 { // 5GHz
                channels[adapter.Interface]["5"] = wh.get5GHzChannels(adapter.SupportedChannels)
            } else { // 6GHz
                channels[adapter.Interface]["6"] = wh.get6GHzChannels(adapter.SupportedChannels)
            }
        }
    }
    
    return channels
}

func (wh *WirelessHandler) supportsMode(adapter *hardware.WirelessAdapter, mode string) bool {
    for _, supportedMode := range adapter.SupportedModes {
        if supportedMode == mode {
            return true
        }
    }
    return false
}

func (wh *WirelessHandler) supportsFrequency(adapter *hardware.WirelessAdapter, frequency string) bool {
    freqFloat := parseFrequency(frequency)
    
    for _, supportedFreq := range adapter.SupportedFrequencies {
        if isFrequencyInBand(freqFloat, supportedFreq) {
            return true
        }
    }
    return false
}

func (wh *WirelessHandler) isChannelAvailable(adapter *hardware.WirelessAdapter, frequency string, channel int) bool {
    // Check if channel is supported by adapter
    for _, supportedChannel := range adapter.SupportedChannels {
        if supportedChannel == channel {
            // Additional check: is channel valid for frequency band
            return wh.isChannelValidForFrequency(channel, frequency)
        }
    }
    return false
}

func (wh *WirelessHandler) supportsSecurity(adapter *hardware.WirelessAdapter, security string) bool {
    if security == "open" {
        return true // All adapters support open networks
    }
    
    for _, supportedSecurity := range adapter.SecuritySupport {
        if supportedSecurity == security {
            return true
        }
    }
    return false
}
```

---

# 4. TASK MANAGEMENT MODULE (modules/tasks/)

### module.json
```json
{
  "name": "tasks",
  "display_name": "Task Manager",
  "menu_icon": "list-checks",
  "menu_order": 9,
  "menu_section": "management",
  "requires_hardware": [],
  "dependencies": {}
}
```

### templates/index.html (Task Manager Interface)
```html
{{define "title"}}Task Manager{{end}}

{{define "content"}}
<div class="p-8">
    <!-- Task Manager Header -->
    <div class="flex justify-between items-center mb-8">
        <div>
            <h1 class="text-3xl font-bold text-base-content">Task Manager</h1>
            <p class="text-base-content/60">Configuration change buffer and execution</p>
        </div>
        
        <div class="flex gap-4">
            <button class="btn btn-primary" onclick="executeAllTasks()" {{if not .PendingTasks}}disabled{{end}}>
                Apply All Changes ({{len .PendingTasks}})
            </button>
            <button class="btn btn-outline btn-error" onclick="clearAllTasks()" {{if not .PendingTasks}}disabled{{end}}>
                Clear All
            </button>
        </div>
    </div>
    
    <!-- Pending Tasks -->
    {{if .PendingTasks}}
    <div class="card bg-base-100 shadow-xl mb-8">
        <div class="card-body">
            <h2 class="card-title text-warning">Pending Configuration Changes</h2>
            <p class="text-base-content/60 mb-4">
                Review and manage pending configuration changes before applying them to the system.
            </p>
            
            <div class="overflow-x-auto">
                <table class="table table-sm">
                    <thead>
                        <tr>
                            <th>
                                <input type="checkbox" class="checkbox checkbox-sm" onchange="toggleAllTasks(this)">
                            </th>
                            <th>Description</th>
                            <th>Module</th>
                            <th>Priority</th>
                            <th>Dependencies</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .PendingTasks}}
                        <tr class="hover">
                            <td>
                                <input type="checkbox" class="checkbox checkbox-sm task-checkbox" value="{{.ID}}">
                            </td>
                            <td>
                                <div class="font-medium">{{.Description}}</div>
                                <div class="text-sm text-base-content/60">{{.Action}}</div>
                            </td>
                            <td>
                                <div class="badge badge-outline">{{.Module}}</div>
                            </td>
                            <td>
                                <div class="badge 
                                    {{if ge .Priority 8}}badge-error
                                    {{else if ge .Priority 6}}badge-warning  
                                    {{else}}badge-info{{end}}">
                                    {{.Priority}}
                                </div>
                            </td>
                            <td>
                                {{if .Dependencies}}
                                <div class="tooltip" data-tip="{{range .Dependencies}}{{.}}, {{end}}">
                                    <div class="badge badge-outline badge-sm">{{len .Dependencies}} deps</div>
                                </div>
                                {{else}}
                                <span class="text-base-content/40">None</span>
                                {{end}}
                            </td>
                            <td>
                                <div class="flex gap-2">
                                    <button class="btn btn-xs btn-outline" onclick="viewTaskDetails('{{.ID}}')">
                                        View
                                    </button>
                                    <button class="btn btn-xs btn-error" onclick="removeTask('{{.ID}}')">
                                        Remove
                                    </button>
                                </div>
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            
            <!-- Dependency Warnings -->
            {{if .DependencyWarnings}}
            <div class="alert alert-warning mt-4">
                <svg class="w-6 h-6 stroke-current" fill="none" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
                </svg>
                <div>
                    <h3 class="font-bold">Configuration Dependencies Detected</h3>
                    <div class="text-sm">
                        {{range .DependencyWarnings}}
                        <div>• {{.}}</div>
                        {{end}}
                    </div>
                </div>
            </div>
            {{end}}
            
            <!-- Conflict Warnings -->
            {{if .ConflictWarnings}}
            <div class="alert alert-error mt-4">
                <svg class="w-6 h-6 stroke-current" fill="none" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                <div>
                    <h3 class="font-bold">Configuration Conflicts Detected</h3>
                    <div class="text-sm">
                        {{range .ConflictWarnings}}
                        <div>• {{.}}</div>
                        {{end}}
                    </div>
                </div>
            </div>
            {{end}}
        </div>
    </div>
    {{else}}
    <div class="card bg-base-100 shadow-xl mb-8">
        <div class="card-body text-center py-12">
            <div class="text-6xl text-base-content/20 mb-4">📋</div>
            <h2 class="text-xl font-semibold text-base-content/60">No Pending Tasks</h2>
            <p class="text-base-content/40">Configuration changes will appear here before being applied to the system</p>
        </div>
    </div>
    {{end}}
    
    <!-- Recent Task History -->
    {{if .RecentTasks}}
    <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
            <h2 class="card-title">Recent Configuration Changes</h2>
            
            <div class="overflow-x-auto">
                <table class="table table-sm">
                    <thead>
                        <tr>
                            <th>Description</th>
                            <th>Module</th>
                            <th>Status</th>
                            <th>Executed</th>
                            <th>Result</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .RecentTasks}}
                        <tr>
                            <td>{{.Description}}</td>
                            <td>
                                <div class="badge badge-outline">{{.Module}}</div>
                            </td>
                            <td>
                                <div class="badge 
                                    {{if eq .Status "completed"}}badge-success
                                    {{else if eq .Status "failed"}}badge-error
                                    {{else if eq .Status "executing"}}badge-warning
                                    {{else}}badge-info{{end}}">
                                    {{.Status}}
                                </div>
                            </td>
                            <td>
                                {{if .ExecutedAt}}{{.ExecutedAt.Format "15:04 02/01/06"}}{{else}}-{{end}}
                            </td>
                            <td>
                                {{if eq .Status "failed"}}
                                <div class="text-error text-sm">{{.Error}}</div>
                                {{else if eq .Status "completed"}}
                                <div class="text-success text-sm">Success</div>
                                {{else}}
                                -
                                {{end}}
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
    {{end}}
</div>

<!-- Task Details Modal -->
{{template "modal" .TaskDetailsModal}}
{{end}}

{{define "scripts"}}
<script src="/static/js/task_manager.js"></script>
<script>
// Initialize task manager
const taskManager = new TaskManager();
taskManager.init();

async function executeAllTasks() {
    if (!confirm('Apply all pending configuration changes?')) return;
    
    try {
        const response = await fetch('/api/tasks/execute', {
            method: 'POST'
        });
        
        if (response.ok) {
            showSuccess('All tasks executed successfully');
            setTimeout(() => location.reload(), 2000);
        } else {
            const error = await response.json();
            showError('Failed to execute tasks: ' + error.message);
        }
    } catch (error) {
        showError('Network error: ' + error.message);
    }
}

async function removeTask(taskId) {
    if (!confirm('Remove this configuration change?')) return;
    
    try {
        const response = await fetch(`/api/tasks/${taskId}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            showSuccess('Task removed');
            location.reload();
        } else {
            const error = await response.json();
            showError('Failed to remove task: ' + error.message);
        }
    } catch (error) {
        showError('Network error: ' + error.message);
    }
}

function toggleAllTasks(checkbox) {
    const taskCheckboxes = document.querySelectorAll('.task-checkbox');
    taskCheckboxes.forEach(cb => cb.checked = checkbox.checked);
}

function viewTaskDetails(taskId) {
    // Implementation for viewing task details in modal
    taskManager.showTaskDetails(taskId);
}

async function clearAllTasks() {
    if (!confirm('Clear all pending configuration changes?')) return;
    
    try {
        const response = await fetch('/api/tasks', {
            method: 'DELETE'
        });
        
        if (response.ok) {
            showSuccess('All tasks cleared');
            location.reload();
        } else {
            const error = await response.json();
            showError('Failed to clear tasks: ' + error.message);
        }
    } catch (error) {
        showError('Network error: ' + error.message);
    }
}

function showSuccess(message) {
    // Implementation for success notification
    console.log('Success:', message);
}

function showError(message) {
    // Implementation for error notification
    console.error('Error:', message);
}
</script>
{{end}}

{{template "layout.html" .}}
```

---

This comprehensive module implementation provides the complete RouterSBC network management system with hardware-driven interface generation, configuration dependency validation, and task-based configuration buffering. Each module follows the template-based architecture with navigation hooks and integrates seamlessly with the core hardware detection and task management systems.