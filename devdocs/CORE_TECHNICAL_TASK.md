# RouterSBC CORE TECHNICAL TASK - Complete System Implementation

**Project**: RouterSBC Hardware-Driven Network Management System  
**Date**: 2025-06-09  
**Type**: Core Technical Implementation - Echo Framework  
**Framework**: Echo v4  

---

# ОБЯЗАТЕЛЬНАЯ АРХИТЕКТУРА ИНТЕРФЕЙСА

## Template-Based UI System (НЕ ИЗМЕНЯТЬ!)

### Базовые шаблоны (обязательные для всех модулей):
```
core/templates/base/
├── layout.html      # Main wrapper с CSS/JS includes
├── header.html      # Site header с system status
├── footer.html      # Footer с system information  
├── sidebar.html     # Navigation sidebar container
└── nav_menu.html    # DYNAMIC menu с module hooks
```

### Система хуков навигации (АВТОМАТИЧЕСКАЯ):
**НЕ ИЗМЕНЯЙТЕ nav_menu.html!** Система автоматически включает модули через:

1. **module.json регистрация:**
```json
{
  "name": "network",
  "display_name": "Network Settings",
  "menu_icon": "network-wired",
  "menu_order": 2,
  "menu_section": "configuration",
  "requires_hardware": ["ethernet"],
  "optional_hardware": ["wireless"]
}
```

2. **nav_menu.html автоматически загружает активные модули**
3. **Модули активируются только при наличии совместимого оборудования**

### Компоненты UI (ИСПОЛЬЗОВАТЬ ОБЯЗАТЕЛЬНО):
```
core/templates/components/
├── card.html        # {{template "card" .CardData}}
├── modal.html       # {{template "modal" .ModalData}}
├── tabs.html        # {{template "tabs" .TabsData}}
└── form.html        # {{template "form" .FormData}}
```

### Стандарт наследования шаблонов:
```html
{{define "title"}}Module Name{{end}}

{{define "content"}}
<div class="p-8">
    {{template "card" .SystemCard}}
    {{template "tabs" .ConfigTabs}}
</div>
{{end}}

{{define "scripts"}}
<script src="/static/js/module_specific.js"></script>
{{end}}

{{template "layout.html" .}}
```

---

# CORE SYSTEM IMPLEMENTATION

## 1. Echo Framework Server (main.go)

```go
package main

import (
    "context"
    "fmt"
    "html/template"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    
    "routersbc/core/config"
    "routersbc/core/hardware"
    "routersbc/core/modules"
    "routersbc/core/tasks"
    "routersbc/modules/dashboard"
    "routersbc/modules/login"
    "routersbc/modules/network"
    "routersbc/modules/wireless"
    "routersbc/modules/tasks"
)

type TemplateRenderer struct {
    templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

type RouterSBCServer struct {
    echo              *echo.Echo
    hardwareDetector  *hardware.HardwareDetector
    configManager     *config.ConfigurationManager
    taskManager       *tasks.TaskManager
    moduleLoader      *modules.ModuleLoader
    navigationHooks   []modules.NavigationHook
}

func main() {
    server := NewRouterSBCServer()
    
    // Initialize core systems
    if err := server.Initialize(); err != nil {
        panic(fmt.Sprintf("Failed to initialize server: %v", err))
    }
    
    // Start hardware monitoring
    go server.startHardwareMonitoring()
    
    // Start task processing
    go server.startTaskProcessing()
    
    // Start server
    port := os.Getenv("SERVER_PORT")
    if port == "" {
        port = "3000"
    }
    
    host := os.Getenv("SERVER_HOST")
    if host == "" {
        host = "0.0.0.0"
    }
    
    fmt.Printf("RouterSBC server starting on %s:%s\n", host, port)
    fmt.Println("Default login: admin/admin")
    
    server.echo.Logger.Fatal(server.echo.Start(fmt.Sprintf("%s:%s", host, port)))
}

func NewRouterSBCServer() *RouterSBCServer {
    e := echo.New()
    
    return &RouterSBCServer{
        echo:             e,
        hardwareDetector: hardware.NewHardwareDetector(),
        configManager:    config.NewConfigurationManager("./configs", "./configs/backups", "./temp"),
        taskManager:      tasks.NewTaskManager(),
        moduleLoader:     modules.NewModuleLoader(),
        navigationHooks:  make([]modules.NavigationHook, 0),
    }
}

func (s *RouterSBCServer) Initialize() error {
    // Configure Echo middleware
    s.echo.Use(middleware.Logger())
    s.echo.Use(middleware.Recover())
    s.echo.Use(middleware.CORS())
    s.echo.Use(middleware.Secure())
    
    // Initialize template system
    if err := s.initializeTemplates(); err != nil {
        return fmt.Errorf("failed to initialize templates: %v", err)
    }
    
    // Static files
    s.echo.Static("/static", "static")
    
    // Initial hardware scan
    if err := s.hardwareDetector.ScanAllHardware(); err != nil {
        return fmt.Errorf("failed to scan hardware: %v", err)
    }
    
    // Load and register modules
    if err := s.loadModules(); err != nil {
        return fmt.Errorf("failed to load modules: %v", err)
    }
    
    // Register core routes
    s.registerCoreRoutes()
    
    return nil
}

func (s *RouterSBCServer) initializeTemplates() error {
    // Load base templates
    baseTemplates := []string{
        "core/templates/base/layout.html",
        "core/templates/base/header.html",
        "core/templates/base/footer.html",
        "core/templates/base/sidebar.html",
        "core/templates/base/nav_menu.html",
    }
    
    // Load component templates
    componentTemplates := []string{
        "core/templates/components/card.html",
        "core/templates/components/modal.html",
        "core/templates/components/tabs.html",
        "core/templates/components/form.html",
    }
    
    allTemplates := append(baseTemplates, componentTemplates...)
    
    // Load module templates dynamically
    moduleTemplates, err := s.findModuleTemplates()
    if err != nil {
        return err
    }
    allTemplates = append(allTemplates, moduleTemplates...)
    
    tmpl, err := template.ParseFiles(allTemplates...)
    if err != nil {
        return err
    }
    
    s.echo.Renderer = &TemplateRenderer{
        templates: tmpl,
    }
    
    return nil
}

func (s *RouterSBCServer) findModuleTemplates() ([]string, error) {
    var templates []string
    
    err := filepath.Walk("modules", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if strings.HasSuffix(path, ".html") && strings.Contains(path, "templates") {
            templates = append(templates, path)
        }
        
        return nil
    })
    
    return templates, err
}

func (s *RouterSBCServer) loadModules() error {
    // Scan for module directories
    modulesList := []string{"login", "dashboard", "network", "wireless", "tasks", "firewall", "system", "services"}
    
    for _, moduleName := range modulesList {
        moduleDir := filepath.Join("modules", moduleName)
        
        // Check if module directory exists
        if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
            continue
        }
        
        // Load module configuration
        moduleConfig, err := s.moduleLoader.LoadModuleConfig(moduleDir)
        if err != nil {
            continue // Skip problematic modules
        }
        
        // Check hardware dependencies
        if !s.checkHardwareDependencies(moduleConfig.RequiresHW) {
            continue // Skip modules with unmet hardware dependencies
        }
        
        // Register navigation hook
        hook := modules.NavigationHook{
            ModuleName:  moduleConfig.Name,
            DisplayName: moduleConfig.DisplayName,
            Icon:        moduleConfig.MenuIcon,
            Order:       moduleConfig.MenuOrder,
            Section:     moduleConfig.MenuSection,
            Available:   true,
        }
        s.navigationHooks = append(s.navigationHooks, hook)
        
        // Register module routes
        if err := s.registerModuleRoutes(moduleConfig); err != nil {
            return fmt.Errorf("failed to register routes for module %s: %v", moduleName, err)
        }
    }
    
    return nil
}

func (s *RouterSBCServer) checkHardwareDependencies(requirements []string) bool {
    capabilities := s.hardwareDetector.GetCapabilities()
    
    for _, requirement := range requirements {
        switch requirement {
        case "ethernet":
            if len(s.hardwareDetector.GetEthernetInterfaces()) == 0 {
                return false
            }
        case "wireless":
            if len(s.hardwareDetector.GetWirelessAdapters()) == 0 {
                return false
            }
        case "multi_wan":
            if len(s.hardwareDetector.GetEthernetInterfaces()) < 2 {
                return false
            }
        default:
            if !capabilities.HasFeature(requirement) {
                return false
            }
        }
    }
    
    return true
}

func (s *RouterSBCServer) registerModuleRoutes(moduleConfig *modules.ModuleConfig) error {
    moduleDir := filepath.Join("modules", moduleConfig.Name)
    templatesDir := filepath.Join(moduleDir, "templates")
    
    // Find all template files
    err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !strings.HasSuffix(path, ".html") {
            return nil
        }
        
        // Generate route from template path
        relativePath := strings.TrimPrefix(path, templatesDir+"/")
        routePath := strings.TrimSuffix(relativePath, ".html")
        
        if routePath == "index" {
            routePath = ""
        }
        
        fullRoutePath := "/" + moduleConfig.Name
        if routePath != "" {
            fullRoutePath += "/" + routePath
        }
        
        // Register GET route
        s.echo.GET(fullRoutePath, s.createModuleHandler(moduleConfig.Name, relativePath))
        
        return nil
    })
    
    return err
}

func (s *RouterSBCServer) createModuleHandler(moduleName, templatePath string) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Prepare template data
        data := map[string]interface{}{
            "ModuleName":      moduleName,
            "NavigationHooks": s.navigationHooks,
            "HardwareInfo":    s.hardwareDetector.GetHardwareInfo(),
            "SystemStatus":    s.getSystemStatus(),
            "Tasks":           s.taskManager.GetTasks(),
        }
        
        // Add module-specific data
        switch moduleName {
        case "dashboard":
            data["SystemOverview"] = s.getDashboardData()
        case "network":
            data["NetworkConfig"] = s.getNetworkData()
        case "wireless":
            data["WirelessConfig"] = s.getWirelessData()
        case "tasks":
            data["TaskQueue"] = s.getTaskData()
        }
        
        templateName := strings.TrimSuffix(templatePath, ".html")
        return c.Render(http.StatusOK, templateName, data)
    }
}

func (s *RouterSBCServer) registerCoreRoutes() {
    // API routes
    api := s.echo.Group("/api")
    
    // Hardware API
    api.GET("/hardware/scan", s.handleHardwareScan)
    api.GET("/hardware/status", s.handleHardwareStatus)
    api.GET("/hardware/capabilities", s.handleHardwareCapabilities)
    
    // Configuration API
    api.POST("/config/:module", s.handleConfigSave)
    api.GET("/config/:module", s.handleConfigLoad)
    api.DELETE("/config/:module", s.handleConfigDelete)
    
    // Task API
    api.POST("/tasks", s.handleTaskCreate)
    api.GET("/tasks", s.handleTaskList)
    api.DELETE("/tasks/:id", s.handleTaskDelete)
    api.POST("/tasks/execute", s.handleTaskExecute)
    
    // System API
    api.GET("/system/status", s.handleSystemStatus)
    api.POST("/system/restart", s.handleSystemRestart)
}

func (s *RouterSBCServer) startHardwareMonitoring() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := s.hardwareDetector.ScanAllHardware(); err != nil {
                s.echo.Logger.Error("Hardware scan failed:", err)
            }
        }
    }
}

func (s *RouterSBCServer) startTaskProcessing() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := s.taskManager.ProcessPendingTasks(); err != nil {
                s.echo.Logger.Error("Task processing failed:", err)
            }
        }
    }
}
```

## 2. Hardware Detection System (core/hardware/)

### Enhanced Hardware Detector (detector.go)
```go
package hardware

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type HardwareDetector struct {
    interfaces       map[string]*NetworkInterface
    wirelessAdapters map[string]*WirelessAdapter
    systemInfo       *SystemInfo
    capabilities     *SystemCapabilities
    lastScanTime     time.Time
    scanInterval     time.Duration
}

type NetworkInterface struct {
    Name         string   `json:"name"`
    Type         string   `json:"type"`         // ethernet, wireless, loopback, bridge, vlan
    Status       string   `json:"status"`       // up, down, unknown
    MAC          string   `json:"mac"`
    IP           string   `json:"ip"`
    Netmask      string   `json:"netmask"`
    Gateway      string   `json:"gateway"`
    Speed        string   `json:"speed"`        // 1000Mb/s, 100Mb/s
    Duplex       string   `json:"duplex"`       // full, half
    MTU          int      `json:"mtu"`
    RXBytes      int64    `json:"rx_bytes"`
    TXBytes      int64    `json:"tx_bytes"`
    RXPackets    int64    `json:"rx_packets"`
    TXPackets    int64    `json:"tx_packets"`
    RXErrors     int64    `json:"rx_errors"`
    TXErrors     int64    `json:"tx_errors"`
    Carrier      bool     `json:"carrier"`
    OperState    string   `json:"oper_state"`
    AdminState   string   `json:"admin_state"`
    Driver       string   `json:"driver"`
    PCIAddress   string   `json:"pci_address"`
    Capabilities []string `json:"capabilities"`
}

type WirelessAdapter struct {
    Interface    string    `json:"interface"`     // wlan0, wlan1
    Driver       string    `json:"driver"`        // ath9k, rt2800usb
    Chipset      string    `json:"chipset"`       // Atheros AR9271
    Status       string    `json:"status"`        // up, down, connected
    Mode         string    `json:"mode"`          // managed, master, monitor
    SSID         string    `json:"ssid"`          // Connected network (if any)
    Frequency    float64   `json:"frequency"`     // Current frequency in MHz
    Channel      int       `json:"channel"`       // Current channel
    TXPower      int       `json:"tx_power"`      // TX power in dBm
    Quality      int       `json:"quality"`       // Signal quality %
    
    SupportedModes      []string  `json:"supported_modes"`      // managed, master, monitor, mesh
    SupportedFrequencies []float64 `json:"supported_frequencies"` // 2.4GHz, 5GHz, 6GHz
    SupportedChannels   []int     `json:"supported_channels"`   // Available channels per frequency
    SecuritySupport     []string  `json:"security_support"`     // WPA, WPA2, WPA3
    
    HardwareInfo struct {
        Vendor      string `json:"vendor"`       // Atheros
        Product     string `json:"product"`      // AR9271 802.11n
        USBInfo     string `json:"usb_info"`     // USB VID:PID if applicable
        PCIInfo     string `json:"pci_info"`     // PCI device info if applicable
    } `json:"hardware_info"`
}

type SystemInfo struct {
    Hostname     string            `json:"hostname"`
    Kernel       string            `json:"kernel"`
    Architecture string            `json:"architecture"`
    OS           string            `json:"os"`
    Uptime       time.Duration     `json:"uptime"`
    LoadAverage  []float64         `json:"load_average"`
    MemoryTotal  uint64            `json:"memory_total"`
    MemoryFree   uint64            `json:"memory_free"`
    DiskUsage    map[string]uint64 `json:"disk_usage"`
    CPUInfo      []CPUInfo         `json:"cpu_info"`
}

type CPUInfo struct {
    Model     string  `json:"model"`
    Cores     int     `json:"cores"`
    Frequency float64 `json:"frequency"`
    Usage     float64 `json:"usage"`
}

func NewHardwareDetector() *HardwareDetector {
    return &HardwareDetector{
        interfaces:       make(map[string]*NetworkInterface),
        wirelessAdapters: make(map[string]*WirelessAdapter),
        scanInterval:     30 * time.Second,
    }
}

func (hd *HardwareDetector) ScanAllHardware() error {
    // Scan network interfaces
    if err := hd.scanNetworkInterfaces(); err != nil {
        return fmt.Errorf("failed to scan network interfaces: %v", err)
    }
    
    // Scan wireless adapters with detailed capabilities
    if err := hd.scanWirelessAdapters(); err != nil {
        return fmt.Errorf("failed to scan wireless adapters: %v", err)
    }
    
    // Scan system information
    if err := hd.scanSystemInfo(); err != nil {
        return fmt.Errorf("failed to scan system info: %v", err)
    }
    
    // Analyze system capabilities
    hd.capabilities = hd.analyzeCapabilities()
    
    hd.lastScanTime = time.Now()
    return nil
}

func (hd *HardwareDetector) scanWirelessAdapters() error {
    // Get wireless interfaces from /proc/net/wireless
    wirelessInterfaces, err := hd.getWirelessInterfaces()
    if err != nil {
        return err
    }
    
    newAdapters := make(map[string]*WirelessAdapter)
    
    for _, ifaceName := range wirelessInterfaces {
        adapter := &WirelessAdapter{
            Interface: ifaceName,
        }
        
        // Get basic interface info
        if err := hd.readWirelessBasicInfo(adapter); err != nil {
            continue
        }
        
        // Get wireless capabilities using iw
        if err := hd.readWirelessCapabilities(adapter); err != nil {
            // Not critical, continue
        }
        
        // Get current wireless status
        if err := hd.readWirelessStatus(adapter); err != nil {
            // Not critical, continue
        }
        
        // Get hardware information
        if err := hd.readWirelessHardwareInfo(adapter); err != nil {
            // Not critical, continue
        }
        
        newAdapters[ifaceName] = adapter
    }
    
    hd.wirelessAdapters = newAdapters
    return nil
}

func (hd *HardwareDetector) getWirelessInterfaces() ([]string, error) {
    var interfaces []string
    
    // Read from /proc/net/wireless
    file, err := os.Open("/proc/net/wireless")
    if err != nil {
        return interfaces, err
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    lineNum := 0
    for scanner.Scan() {
        lineNum++
        if lineNum <= 2 { // Skip header lines
            continue
        }
        
        line := scanner.Text()
        fields := strings.Fields(line)
        if len(fields) > 0 {
            ifaceName := strings.TrimSuffix(fields[0], ":")
            interfaces = append(interfaces, ifaceName)
        }
    }
    
    return interfaces, scanner.Err()
}

func (hd *HardwareDetector) readWirelessCapabilities(adapter *WirelessAdapter) error {
    // Use 'iw' command to get detailed capabilities
    cmd := exec.Command("iw", "phy")
    output, err := cmd.Output()
    if err != nil {
        return err
    }
    
    lines := strings.Split(string(output), "\n")
    currentPhy := ""
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        // Detect PHY for this interface
        if strings.HasPrefix(line, "Wiphy ") {
            currentPhy = line
            continue
        }
        
        // Check if this PHY corresponds to our interface
        if strings.Contains(line, adapter.Interface) {
            // Parse capabilities for this PHY
            hd.parseWirelessPHYCapabilities(adapter, lines)
            break
        }
    }
    
    return nil
}

func (hd *HardwareDetector) parseWirelessPHYCapabilities(adapter *WirelessAdapter, phyOutput []string) {
    adapter.SupportedModes = []string{}
    adapter.SupportedFrequencies = []float64{}
    adapter.SupportedChannels = []int{}
    adapter.SecuritySupport = []string{}
    
    inSupportedModes := false
    inBands := false
    
    for _, line := range phyOutput {
        line = strings.TrimSpace(line)
        
        if strings.Contains(line, "Supported interface modes:") {
            inSupportedModes = true
            continue
        }
        
        if inSupportedModes && strings.HasPrefix(line, "*") {
            mode := strings.TrimSpace(strings.TrimPrefix(line, "*"))
            mode = strings.ToLower(mode)
            if mode == "ap" {
                mode = "master"
            }
            adapter.SupportedModes = append(adapter.SupportedModes, mode)
        }
        
        if strings.Contains(line, "Band") && strings.Contains(line, "GHz") {
            inBands = true
            
            // Extract frequency from band description
            re := regexp.MustCompile(`(\d+\.?\d*)\s*GHz`)
            matches := re.FindStringSubmatch(line)
            if len(matches) > 1 {
                if freq, err := strconv.ParseFloat(matches[1], 64); err == nil {
                    adapter.SupportedFrequencies = append(adapter.SupportedFrequencies, freq)
                }
            }
        }
        
        if inBands && strings.Contains(line, "MHz") && strings.Contains(line, "channel") {
            // Parse channel information
            re := regexp.MustCompile(`channel\s+(\d+)`)
            matches := re.FindStringSubmatch(line)
            if len(matches) > 1 {
                if channel, err := strconv.Atoi(matches[1]); err == nil {
                    adapter.SupportedChannels = append(adapter.SupportedChannels, channel)
                }
            }
        }
        
        // End of current section
        if strings.HasPrefix(line, "Wiphy ") && inSupportedModes {
            break
        }
    }
    
    // Default security support (most adapters support these)
    adapter.SecuritySupport = []string{"WPA", "WPA2"}
    
    // Check for WPA3 support (requires newer hardware)
    if hd.checkWPA3Support(adapter.Interface) {
        adapter.SecuritySupport = append(adapter.SecuritySupport, "WPA3")
    }
}

func (hd *HardwareDetector) checkWPA3Support(interfaceName string) bool {
    cmd := exec.Command("iw", interfaceName, "info")
    output, err := cmd.Output()
    if err != nil {
        return false
    }
    
    return strings.Contains(string(output), "SAE") || strings.Contains(string(output), "WPA3")
}

func (hd *HardwareDetector) GetEthernetInterfaces() []*NetworkInterface {
    var ethernet []*NetworkInterface
    for _, iface := range hd.interfaces {
        if iface.Type == "ethernet" {
            ethernet = append(ethernet, iface)
        }
    }
    return ethernet
}

func (hd *HardwareDetector) GetWirelessAdapters() []*WirelessAdapter {
    var adapters []*WirelessAdapter
    for _, adapter := range hd.wirelessAdapters {
        adapters = append(adapters, adapter)
    }
    return adapters
}

func (hd *HardwareDetector) GetCapabilities() *SystemCapabilities {
    return hd.capabilities
}

func (hd *HardwareDetector) GetHardwareInfo() map[string]interface{} {
    return map[string]interface{}{
        "interfaces":        hd.interfaces,
        "wireless_adapters": hd.wirelessAdapters,
        "system_info":       hd.systemInfo,
        "capabilities":      hd.capabilities,
        "last_scan":         hd.lastScanTime,
    }
}
```

## 3. Task Management System (core/tasks/)

### Task Manager with Configuration Buffer (manager.go)
```go
package tasks

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

type TaskManager struct {
    tasks              []Task
    pendingTasks       []Task
    completedTasks     []Task
    failedTasks        []Task
    dependencyResolver *DependencyResolver
    conflictDetector   *ConflictDetector
    configValidator    *ConfigValidator
    mutex              sync.RWMutex
}

type Task struct {
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`         // network, wireless, firewall, system
    Module       string                 `json:"module"`       // Source module
    Action       string                 `json:"action"`       // create, update, delete, apply
    Description  string                 `json:"description"`  // Human-readable description
    Config       map[string]interface{} `json:"config"`       // Configuration data
    Dependencies []string               `json:"dependencies"` // Required tasks/configurations
    Conflicts    []string               `json:"conflicts"`    // Conflicting tasks
    Priority     int                    `json:"priority"`     // Execution priority (1-10)
    CreatedAt    time.Time              `json:"created_at"`
    ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
    ExecutedAt   *time.Time             `json:"executed_at,omitempty"`
    CompletedAt  *time.Time             `json:"completed_at,omitempty"`
    Status       string                 `json:"status"`       // pending, executing, completed, failed, cancelled
    Error        string                 `json:"error,omitempty"`
    Result       map[string]interface{} `json:"result,omitempty"`
    
    // Hardware requirements
    RequiredHardware []string `json:"required_hardware"`
    RequiredCapabilities []string `json:"required_capabilities"`
}

type DependencyRule struct {
    Requires     []string `json:"requires"`      // Required configurations/tasks
    Conflicts    []string `json:"conflicts"`     // Conflicting configurations
    Recommends   []string `json:"recommends"`    // Recommended configurations
    Suggests     []string `json:"suggests"`      // Suggested configurations
    Hardware     []string `json:"hardware"`      // Required hardware
    Capabilities []string `json:"capabilities"`  // Required system capabilities
}

type ConfigDependency struct {
    Module       string       `json:"module"`
    Field        string       `json:"field"`
    Value        interface{}  `json:"value"`
    Condition    string       `json:"condition"`    // equals, not_equals, exists, not_exists
    Description  string       `json:"description"`
}

func NewTaskManager() *TaskManager {
    return &TaskManager{
        tasks:              make([]Task, 0),
        pendingTasks:       make([]Task, 0),
        completedTasks:     make([]Task, 0),
        failedTasks:        make([]Task, 0),
        dependencyResolver: NewDependencyResolver(),
        conflictDetector:   NewConflictDetector(),
        configValidator:    NewConfigValidator(),
    }
}

func (tm *TaskManager) AddTask(task Task) error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    // Generate task ID if not provided
    if task.ID == "" {
        task.ID = tm.generateTaskID()
    }
    
    // Set creation time
    task.CreatedAt = time.Now()
    task.Status = "pending"
    
    // Validate task configuration
    if err := tm.configValidator.ValidateTaskConfig(task); err != nil {
        return fmt.Errorf("task validation failed: %v", err)
    }
    
    // Analyze dependencies
    dependencies, err := tm.dependencyResolver.ResolveDependencies(task)
    if err != nil {
        return fmt.Errorf("dependency resolution failed: %v", err)
    }
    task.Dependencies = dependencies
    
    // Check for conflicts
    conflicts, err := tm.conflictDetector.DetectConflicts(task, tm.tasks)
    if err != nil {
        return fmt.Errorf("conflict detection failed: %v", err)
    }
    task.Conflicts = conflicts
    
    // Add to task lists
    tm.tasks = append(tm.tasks, task)
    tm.pendingTasks = append(tm.pendingTasks, task)
    
    return nil
}

func (tm *TaskManager) RemoveTask(taskID string) error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    // Remove from all task lists
    tm.tasks = tm.removeTaskFromSlice(tm.tasks, taskID)
    tm.pendingTasks = tm.removeTaskFromSlice(tm.pendingTasks, taskID)
    
    return nil
}

func (tm *TaskManager) ExecuteAllTasks() error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    if len(tm.pendingTasks) == 0 {
        return fmt.Errorf("no pending tasks to execute")
    }
    
    // Sort tasks by dependencies and priority
    orderedTasks, err := tm.dependencyResolver.OrderTasks(tm.pendingTasks)
    if err != nil {
        return fmt.Errorf("failed to order tasks: %v", err)
    }
    
    // Execute tasks in order
    for _, task := range orderedTasks {
        if err := tm.executeTask(&task); err != nil {
            task.Status = "failed"
            task.Error = err.Error()
            tm.failedTasks = append(tm.failedTasks, task)
            
            // Decide whether to continue or abort
            if task.Priority >= 8 { // Critical task failed
                return fmt.Errorf("critical task failed: %v", err)
            }
            
            continue
        }
        
        task.Status = "completed"
        task.CompletedAt = &[]time.Time{time.Now()}[0]
        tm.completedTasks = append(tm.completedTasks, task)
    }
    
    // Clear pending tasks
    tm.pendingTasks = make([]Task, 0)
    
    return nil
}

func (tm *TaskManager) executeTask(task *Task) error {
    task.Status = "executing"
    task.ExecutedAt = &[]time.Time{time.Now()}[0]
    
    switch task.Type {
    case "network":
        return tm.executeNetworkTask(task)
    case "wireless":
        return tm.executeWirelessTask(task)
    case "firewall":
        return tm.executeFirewallTask(task)
    case "system":
        return tm.executeSystemTask(task)
    default:
        return fmt.Errorf("unknown task type: %s", task.Type)
    }
}

func (tm *TaskManager) executeNetworkTask(task *Task) error {
    switch task.Action {
    case "configure_lan":
        return tm.configureNetworkLAN(task.Config)
    case "configure_wan":
        return tm.configureNetworkWAN(task.Config)
    case "configure_multiwan":
        return tm.configureMultiWAN(task.Config)
    case "configure_dhcp":
        return tm.configureDHCP(task.Config)
    case "configure_dns":
        return tm.configureDNS(task.Config)
    case "configure_vlan":
        return tm.configureVLAN(task.Config)
    case "configure_bridge":
        return tm.configureBridge(task.Config)
    default:
        return fmt.Errorf("unknown network action: %s", task.Action)
    }
}

func (tm *TaskManager) executeWirelessTask(task *Task) error {
    switch task.Action {
    case "create_ap":
        return tm.createWirelessAP(task.Config)
    case "connect_client":
        return tm.connectWirelessClient(task.Config)
    case "configure_security":
        return tm.configureWirelessSecurity(task.Config)
    case "set_channel":
        return tm.setWirelessChannel(task.Config)
    case "set_power":
        return tm.setWirelessPower(task.Config)
    default:
        return fmt.Errorf("unknown wireless action: %s", task.Action)
    }
}

func (tm *TaskManager) GetTasks() []Task {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    return append([]Task(nil), tm.tasks...)
}

func (tm *TaskManager) GetPendingTasks() []Task {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    return append([]Task(nil), tm.pendingTasks...)
}

func (tm *TaskManager) GetTasksByModule(module string) []Task {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    
    var moduleTasks []Task
    for _, task := range tm.tasks {
        if task.Module == module {
            moduleTasks = append(moduleTasks, task)
        }
    }
    return moduleTasks
}

func (tm *TaskManager) ProcessPendingTasks() error {
    // Auto-process tasks that are scheduled
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    now := time.Now()
    var readyTasks []Task
    
    for i, task := range tm.pendingTasks {
        if task.ScheduledAt != nil && task.ScheduledAt.Before(now) {
            readyTasks = append(readyTasks, task)
            // Remove from pending
            tm.pendingTasks = append(tm.pendingTasks[:i], tm.pendingTasks[i+1:]...)
        }
    }
    
    // Execute ready tasks
    for _, task := range readyTasks {
        if err := tm.executeTask(&task); err != nil {
            task.Status = "failed"
            task.Error = err.Error()
            tm.failedTasks = append(tm.failedTasks, task)
        } else {
            task.Status = "completed"
            task.CompletedAt = &[]time.Time{time.Now()}[0]
            tm.completedTasks = append(tm.completedTasks, task)
        }
    }
    
    return nil
}

func (tm *TaskManager) generateTaskID() string {
    return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

func (tm *TaskManager) removeTaskFromSlice(tasks []Task, taskID string) []Task {
    for i, task := range tasks {
        if task.ID == taskID {
            return append(tasks[:i], tasks[i+1:]...)
        }
    }
    return tasks
}
```

## 4. Development Standards

### Template Development Rules
1. **ОБЯЗАТЕЛЬНО использовать базовые шаблоны** - Никогда не создавать standalone HTML
2. **НИКОГДА не изменять nav_menu.html** - Только через module.json hooks
3. **ИСПОЛЬЗОВАТЬ систему компонентов** - card/modal/tabs/form компоненты
4. **Следовать DaisyUI классам** - Единообразие стилей
5. **Реализовать responsive дизайн** - Mobile-first подход

### Module Development Rules
1. **Декларация зависимостей оборудования** в module.json
2. **Реализация валидации конфигурации** обязательна
3. **Интеграция с системой задач** для всех изменений конфигурации
4. **Real-time обновления** для функций зависящих от оборудования
5. **Проверка зависимостей** перед применением конфигурации

### Configuration Dependency Examples
```json
{
  "wireless_wan": {
    "requires": ["wireless_client_connection"],
    "conflicts": ["wireless_ap_same_interface"],
    "hardware": ["wireless"],
    "description": "Wireless WAN requires active client connection"
  },
  "dhcp_server": {
    "conflicts": ["dhcp_client_same_interface"],
    "requires": ["static_ip_configuration"],
    "description": "DHCP server conflicts with DHCP client on same interface"
  },
  "bridge_configuration": {
    "requires": ["compatible_interfaces"],
    "hardware": ["ethernet", "wireless"],
    "description": "Bridge requires compatible network interfaces"
  }
}
```

This core technical implementation provides the foundation for a hardware-driven RouterSBC system with comprehensive network management capabilities, template-based architecture, and task buffering for configuration changes.