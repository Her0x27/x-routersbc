package utils

import (
        "fmt"
        "io/ioutil"
        "mime/multipart"
        "os"
        "os/exec"
        "path/filepath"
        "strconv"
        "strings"
        "time"
)

type CPUInfo struct {
        Model        string `json:"model"`
        Cores        int    `json:"cores"`
        Architecture string `json:"architecture"`
        Speed        string `json:"speed"`
        Temperature  string `json:"temperature"`
}

type MemoryInfo struct {
        Total     uint64 `json:"total"`
        Available uint64 `json:"available"`
        Used      uint64 `json:"used"`
        Type      string `json:"type"`
        Cached    uint64 `json:"cached"`
        Buffers   uint64 `json:"buffers"`
}

type StorageDevice struct {
        Name        string `json:"name"`
        Type        string `json:"type"`
        Size        uint64 `json:"size"`
        Available   uint64 `json:"available"`
        Used        uint64 `json:"used"`
        Filesystem  string `json:"filesystem"`
        MountPoint  string `json:"mount_point"`
        Temperature string `json:"temperature"`
}

type NetworkDevice struct {
        Name     string `json:"name"`
        Type     string `json:"type"`
        Driver   string `json:"driver"`
        Status   string `json:"status"`
        MAC      string `json:"mac"`
        Speed    string `json:"speed"`
        Duplex   string `json:"duplex"`
}

type VideoDevice struct {
        Name   string `json:"name"`
        Driver string `json:"driver"`
        Status string `json:"status"`
        Vendor string `json:"vendor"`
        Model  string `json:"model"`
}

type AudioDevice struct {
        Name   string `json:"name"`
        Driver string `json:"driver"`
        Status string `json:"status"`
        Type   string `json:"type"`
}

type USBDevice struct {
        Port      int    `json:"port"`
        Vendor    string `json:"vendor"`
        Product   string `json:"product"`
        VendorID  string `json:"vendor_id"`
        ProductID string `json:"product_id"`
        Driver    string `json:"driver"`
        Status    string `json:"status"`
        Class     string `json:"class"`
}

type IOInfo struct {
        USB  USBPortInfo `json:"usb"`
        GPIO GPIOInfo    `json:"gpio"`
        SPI  bool        `json:"spi"`
        UART bool        `json:"uart"`
        I2C  bool        `json:"i2c"`
}

type USBPortInfo struct {
        Ports   int         `json:"ports"`
        Devices []USBDevice `json:"devices"`
}

type GPIOInfo struct {
        Available bool `json:"available"`
        Pins      int  `json:"pins"`
        Controller string `json:"controller"`
}

func GetCPUInfo() (*CPUInfo, error) {
        cpuInfo := &CPUInfo{}

        // Read /proc/cpuinfo
        data, err := ioutil.ReadFile("/proc/cpuinfo")
        if err != nil {
                return nil, fmt.Errorf("failed to read /proc/cpuinfo: %v", err)
        }

        lines := strings.Split(string(data), "\n")
        cores := 0

        for _, line := range lines {
                if strings.Contains(line, ":") {
                        parts := strings.SplitN(line, ":", 2)
                        if len(parts) == 2 {
                                key := strings.TrimSpace(parts[0])
                                value := strings.TrimSpace(parts[1])

                                switch key {
                                case "model name":
                                        if cpuInfo.Model == "" {
                                                cpuInfo.Model = value
                                        }
                                case "processor":
                                        cores++
                                case "cpu MHz":
                                        if cpuInfo.Speed == "" {
                                                cpuInfo.Speed = value + " MHz"
                                        }
                                }
                        }
                }
        }

        cpuInfo.Cores = cores

        // Get architecture
        output, err := exec.Command("uname", "-m").Output()
        if err == nil {
                cpuInfo.Architecture = strings.TrimSpace(string(output))
        }

        // Try to get CPU temperature
        cpuInfo.Temperature = getCPUTemperature()

        return cpuInfo, nil
}

func getCPUTemperature() string {
        // Try different thermal zones
        thermalPaths := []string{
                "/sys/class/thermal/thermal_zone0/temp",
                "/sys/class/thermal/thermal_zone1/temp",
                "/sys/devices/virtual/thermal/thermal_zone0/temp",
        }

        for _, path := range thermalPaths {
                if data, err := ioutil.ReadFile(path); err == nil {
                        if temp, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
                                return fmt.Sprintf("%.1f°C", float64(temp)/1000.0)
                        }
                }
        }

        return "N/A"
}

func GetMemoryInfo() (*MemoryInfo, error) {
        memInfo := &MemoryInfo{}

        data, err := ioutil.ReadFile("/proc/meminfo")
        if err != nil {
                return nil, fmt.Errorf("failed to read /proc/meminfo: %v", err)
        }

        lines := strings.Split(string(data), "\n")
        for _, line := range lines {
                if strings.Contains(line, ":") {
                        parts := strings.Fields(line)
                        if len(parts) >= 2 {
                                key := strings.TrimSuffix(parts[0], ":")
                                value := parseMemoryValue(parts[1])

                                switch key {
                                case "MemTotal":
                                        memInfo.Total = value
                                case "MemAvailable":
                                        memInfo.Available = value
                                case "Cached":
                                        memInfo.Cached = value
                                case "Buffers":
                                        memInfo.Buffers = value
                                }
                        }
                }
        }

        memInfo.Used = memInfo.Total - memInfo.Available
        memInfo.Type = getMemoryType()

        return memInfo, nil
}

func parseMemoryValue(value string) uint64 {
        // Parse memory value in kB
        if val, err := strconv.ParseUint(value, 10, 64); err == nil {
                return val * 1024 // Convert to bytes
        }
        return 0
}

func getMemoryType() string {
        // Try to determine memory type from DMI
        output, err := exec.Command("dmidecode", "-t", "memory").Output()
        if err == nil {
                content := string(output)
                if strings.Contains(content, "DDR4") {
                        return "DDR4"
                } else if strings.Contains(content, "DDR3") {
                        return "DDR3"
                } else if strings.Contains(content, "DDR2") {
                        return "DDR2"
                }
        }
        return "Unknown"
}

func GetStorageInfo() ([]StorageDevice, error) {
        var devices []StorageDevice

        // Get block devices
        output, err := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,AVAIL,USED,FSTYPE,MOUNTPOINT,TYPE").Output()
        if err != nil {
                return devices, fmt.Errorf("failed to get block devices: %v", err)
        }

        // Parse lsblk JSON output would go here
        // For now, use a simpler approach with df

        output, err = exec.Command("df", "-h").Output()
        if err != nil {
                return devices, fmt.Errorf("failed to get filesystem info: %v", err)
        }

        lines := strings.Split(string(output), "\n")
        for i, line := range lines {
                if i == 0 || strings.TrimSpace(line) == "" {
                        continue // Skip header
                }

                fields := strings.Fields(line)
                if len(fields) >= 6 {
                        device := StorageDevice{
                                Name:       fields[0],
                                Size:       parseSize(fields[1]),
                                Used:       parseSize(fields[2]),
                                Available:  parseSize(fields[3]),
                                MountPoint: fields[5],
                                Type:       getStorageType(fields[0]),
                        }

                        devices = append(devices, device)
                }
        }

        return devices, nil
}

func parseSize(sizeStr string) uint64 {
        // Parse size strings like "1.5G", "500M", etc.
        if len(sizeStr) == 0 {
                return 0
        }

        suffix := sizeStr[len(sizeStr)-1]
        valueStr := sizeStr[:len(sizeStr)-1]

        value, err := strconv.ParseFloat(valueStr, 64)
        if err != nil {
                return 0
        }

        switch suffix {
        case 'K':
                return uint64(value * 1024)
        case 'M':
                return uint64(value * 1024 * 1024)
        case 'G':
                return uint64(value * 1024 * 1024 * 1024)
        case 'T':
                return uint64(value * 1024 * 1024 * 1024 * 1024)
        default:
                // Assume bytes if no suffix
                return uint64(value)
        }
}

func getStorageType(device string) string {
        // Determine storage type based on device name
        switch {
        case strings.HasPrefix(device, "/dev/sda"), strings.HasPrefix(device, "/dev/sdb"):
                return "HDD"
        case strings.HasPrefix(device, "/dev/nvme"):
                return "NVMe SSD"
        case strings.HasPrefix(device, "/dev/mmcblk"):
                return "eMMC/SD"
        case strings.Contains(device, "loop"):
                return "Loop"
        default:
                return "Unknown"
        }
}

func GetNetworkDevices() ([]NetworkDevice, error) {
        var devices []NetworkDevice

        // Get network interfaces
        interfaces, err := filepath.Glob("/sys/class/net/*")
        if err != nil {
                return devices, fmt.Errorf("failed to list network interfaces: %v", err)
        }

        for _, iface := range interfaces {
                name := filepath.Base(iface)
                
                // Skip loopback
                if name == "lo" {
                        continue
                }

                device := NetworkDevice{
                        Name: name,
                        Type: getNetworkDeviceType(name),
                }

                // Get MAC address
                if mac, err := ioutil.ReadFile(filepath.Join(iface, "address")); err == nil {
                        device.MAC = strings.TrimSpace(string(mac))
                }

                // Get operational state
                if state, err := ioutil.ReadFile(filepath.Join(iface, "operstate")); err == nil {
                        device.Status = strings.TrimSpace(string(state))
                }

                // Get speed (if available)
                if speed, err := ioutil.ReadFile(filepath.Join(iface, "speed")); err == nil {
                        device.Speed = strings.TrimSpace(string(speed)) + " Mbps"
                }

                // Get driver name
                driverPath := filepath.Join(iface, "device/driver")
                if link, err := os.Readlink(driverPath); err == nil {
                        device.Driver = filepath.Base(link)
                }

                devices = append(devices, device)
        }

        return devices, nil
}

func getNetworkDeviceType(name string) string {
        switch {
        case strings.HasPrefix(name, "eth"):
                return "ethernet"
        case strings.HasPrefix(name, "wlan"), strings.HasPrefix(name, "wlp"):
                return "wifi"
        case strings.HasPrefix(name, "bt"):
                return "bluetooth"
        case strings.HasPrefix(name, "usb"):
                return "usb"
        default:
                return "unknown"
        }
}

func GetVideoDevices() ([]VideoDevice, error) {
        var devices []VideoDevice

        // Check for video devices in /dev
        videoDevices, err := filepath.Glob("/dev/video*")
        if err == nil {
                for _, device := range videoDevices {
                        videoDevice := VideoDevice{
                                Name:   filepath.Base(device),
                                Status: "available",
                        }
                        devices = append(devices, videoDevice)
                }
        }

        return devices, nil
}

func GetAudioDevices() ([]AudioDevice, error) {
        var devices []AudioDevice

        // Check ALSA devices
        output, err := exec.Command("aplay", "-l").Output()
        if err == nil {
                lines := strings.Split(string(output), "\n")
                for _, line := range lines {
                        if strings.Contains(line, "card") {
                                device := AudioDevice{
                                        Name:   line,
                                        Status: "available",
                                        Type:   "alsa",
                                        Driver: "alsa",
                                }
                                devices = append(devices, device)
                        }
                }
        }

        return devices, nil
}

func GetIOInfo() (*IOInfo, error) {
        ioInfo := &IOInfo{}

        // Get USB information
        usbDevices, err := GetUSBDevices()
        if err == nil {
                ioInfo.USB = USBPortInfo{
                        Ports:   getUSBPortCount(),
                        Devices: usbDevices,
                }
        }

        // Check for GPIO
        ioInfo.GPIO = getGPIOInfo()

        // Check for SPI
        ioInfo.SPI = checkSPIAvailable()

        // Check for UART
        ioInfo.UART = checkUARTAvailable()

        // Check for I2C
        ioInfo.I2C = checkI2CAvailable()

        return ioInfo, nil
}

func GetUSBDevices() ([]USBDevice, error) {
        var devices []USBDevice

        output, err := exec.Command("lsusb").Output()
        if err != nil {
                return devices, fmt.Errorf("failed to get USB devices: %v", err)
        }

        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
                if strings.TrimSpace(line) == "" {
                        continue
                }

                // Parse lsusb output
                if strings.Contains(line, ":") {
                        parts := strings.Fields(line)
                        if len(parts) >= 6 {
                                device := USBDevice{
                                        VendorID:  strings.Split(parts[5], ":")[0],
                                        ProductID: strings.Split(parts[5], ":")[1],
                                        Vendor:    "",
                                        Product:   strings.Join(parts[6:], " "),
                                        Status:    "connected",
                                }

                                // Parse bus and device numbers
                                if strings.HasPrefix(parts[1], "Bus") {
                                        device.Port, _ = strconv.Atoi(strings.TrimPrefix(parts[1], "Bus"))
                                }

                                devices = append(devices, device)
                        }
                }
        }

        return devices, nil
}

func getUSBPortCount() int {
        // Count USB hubs/controllers
        output, err := exec.Command("find", "/sys/bus/usb/devices", "-name", "usb*", "-type", "l").Output()
        if err != nil {
                return 0
        }

        return len(strings.Split(strings.TrimSpace(string(output)), "\n"))
}

func getGPIOInfo() GPIOInfo {
        gpio := GPIOInfo{Available: false}

        // Check for GPIO chips
        chips, err := filepath.Glob("/dev/gpiochip*")
        if err == nil && len(chips) > 0 {
                gpio.Available = true
                
                // Try to get pin count from first chip
                if len(chips) > 0 {
                        // This would require more complex parsing
                        gpio.Pins = 40 // Default for Raspberry Pi
                        gpio.Controller = "gpiochip0"
                }
        }

        return gpio
}

func checkSPIAvailable() bool {
        // Check for SPI devices
        _, err := os.Stat("/dev/spidev0.0")
        return err == nil
}

func checkUARTAvailable() bool {
        // Check for UART devices
        devices, err := filepath.Glob("/dev/ttyS*")
        return err == nil && len(devices) > 0
}

func checkI2CAvailable() bool {
        // Check for I2C devices
        devices, err := filepath.Glob("/dev/i2c-*")
        return err == nil && len(devices) > 0
}

// Device type detection functions
func IsWiFiDevice(device USBDevice) bool {
        // Check vendor/product IDs for known WiFi devices
        wifiVendors := []string{"0bda", "148f", "0cf3", "0e8d"}
        for _, vendor := range wifiVendors {
                if strings.EqualFold(device.VendorID, vendor) {
                        return true
                }
        }
        return strings.Contains(strings.ToLower(device.Product), "wifi") ||
                   strings.Contains(strings.ToLower(device.Product), "wireless")
}

func IsEthernetDevice(device USBDevice) bool {
        return strings.Contains(strings.ToLower(device.Product), "ethernet") ||
                   strings.Contains(strings.ToLower(device.Product), "network")
}

func IsBluetoothDevice(device USBDevice) bool {
        return strings.Contains(strings.ToLower(device.Product), "bluetooth")
}

func IsModemDevice(device USBDevice) bool {
        return strings.Contains(strings.ToLower(device.Product), "modem") ||
                   strings.Contains(strings.ToLower(device.Product), "3g") ||
                   strings.Contains(strings.ToLower(device.Product), "lte")
}

func IsStorageDevice(device USBDevice) bool {
        return strings.Contains(strings.ToLower(device.Product), "storage") ||
                   strings.Contains(strings.ToLower(device.Product), "disk") ||
                   strings.Contains(strings.ToLower(device.Product), "flash")
}

func IsWebcamDevice(device USBDevice) bool {
        return strings.Contains(strings.ToLower(device.Product), "camera") ||
                   strings.Contains(strings.ToLower(device.Product), "webcam")
}

// Driver status functions
func GetWiFiDriverStatus(device USBDevice) string {
        // Check if WiFi driver is loaded
        output, err := exec.Command("lsmod").Output()
        if err != nil {
                return "unknown"
        }
        
        // Common WiFi driver modules
        wifiModules := []string{"rtl8188eu", "rtl8192cu", "mt7601u", "ath9k_htc"}
        for _, module := range wifiModules {
                if strings.Contains(string(output), module) {
                        return "loaded"
                }
        }
        
        return "not_loaded"
}

func GetWiFiDriverSuggestion(device USBDevice) string {
        return "Install appropriate WiFi driver package for your device"
}

func GetEthernetDriverStatus(device USBDevice) string {
        return "loaded" // Ethernet drivers are usually loaded automatically
}

func GetEthernetDriverSuggestion(device USBDevice) string {
        return "Ethernet drivers are typically loaded automatically"
}

func GetBluetoothDriverStatus(device USBDevice) string {
        output, err := exec.Command("lsmod").Output()
        if err != nil {
                return "unknown"
        }
        
        if strings.Contains(string(output), "bluetooth") {
                return "loaded"
        }
        
        return "not_loaded"
}

func GetBluetoothDriverSuggestion(device USBDevice) string {
        return "Install bluetooth package: sudo apt install bluetooth bluez"
}

func GetModemDriverStatus(device USBDevice) string {
        output, err := exec.Command("lsmod").Output()
        if err != nil {
                return "unknown"
        }
        
        modemModules := []string{"usb_wwan", "option", "qmi_wwan"}
        for _, module := range modemModules {
                if strings.Contains(string(output), module) {
                        return "loaded"
                }
        }
        
        return "not_loaded"
}

func GetModemDriverSuggestion(device USBDevice) string {
        return "Install modem manager: sudo apt install modemmanager"
}

func GetStorageDriverStatus(device USBDevice) string {
        return "loaded" // Storage drivers are usually loaded automatically
}

func GetStorageDriverSuggestion(device USBDevice) string {
        return "Storage drivers are typically loaded automatically"
}

func GetWebcamDriverStatus(device USBDevice) string {
        output, err := exec.Command("lsmod").Output()
        if err != nil {
                return "unknown"
        }
        
        if strings.Contains(string(output), "uvcvideo") {
                return "loaded"
        }
        
        return "not_loaded"
}

func GetWebcamDriverSuggestion(device USBDevice) string {
        return "Most webcams work with UVC driver, which should be included in kernel"
}

func GetDeviceStatus(deviceType string) (map[string]interface{}, error) {
        status := make(map[string]interface{})
        
        switch deviceType {
        case "USB_WIFI":
                status["driver_available"] = checkDriverAvailable("wifi")
                status["interfaces"] = getWiFiInterfaces()
        case "USB_ETHERNET":
                status["driver_available"] = checkDriverAvailable("ethernet")
                status["interfaces"] = getEthernetInterfaces()
        case "USB_BLUETOOTH":
                status["driver_available"] = checkDriverAvailable("bluetooth")
                status["service_status"] = getBluetoothServiceStatus()
        }
        
        return status, nil
}

func checkDriverAvailable(deviceType string) bool {
        // Simplified driver check
        return true
}

func getWiFiInterfaces() []string {
        output, err := exec.Command("iw", "dev").Output()
        if err != nil {
                return []string{}
        }
        
        var interfaces []string
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
                if strings.Contains(line, "Interface") {
                        parts := strings.Fields(line)
                        if len(parts) >= 2 {
                                interfaces = append(interfaces, parts[1])
                        }
                }
        }
        
        return interfaces
}

func getEthernetInterfaces() []string {
        output, err := exec.Command("ip", "link", "show").Output()
        if err != nil {
                return []string{}
        }
        
        var interfaces []string
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
                if strings.Contains(line, "eth") {
                        parts := strings.Split(line, ":")
                        if len(parts) >= 2 {
                                iface := strings.TrimSpace(parts[1])
                                interfaces = append(interfaces, iface)
                        }
                }
        }
        
        return interfaces
}

func getBluetoothServiceStatus() string {
        err := exec.Command("systemctl", "is-active", "bluetooth").Run()
        if err == nil {
                return "active"
        }
        return "inactive"
}

func InstallDrivers(deviceType string) error {
        switch deviceType {
        case "USB_WIFI":
                return exec.Command("apt", "update").Run()
        case "USB_BLUETOOTH":
                return exec.Command("apt", "install", "-y", "bluetooth", "bluez").Run()
        default:
                return fmt.Errorf("unsupported device type: %s", deviceType)
        }
}

func CreateSystemBackup() (string, error) {
        timestamp := time.Now().Format("20060102_150405")
        backupPath := fmt.Sprintf("/tmp/router_backup_%s.tar.gz", timestamp)
        
        // Create backup of important configuration files
        paths := []string{
                "/etc/network/",
                "/etc/systemd/network/",
                "/etc/netplan/",
                "/etc/dnsmasq.conf",
                "/etc/hostapd/",
                "routersbc.sqlitedb",
        }
        
        args := []string{"czf", backupPath}
        for _, path := range paths {
                if _, err := os.Stat(path); err == nil {
                        args = append(args, path)
                }
        }
        
        if err := exec.Command("tar", args...).Run(); err != nil {
                return "", fmt.Errorf("failed to create backup: %v", err)
        }
        
        return backupPath, nil
}

func RestoreSystemBackup(fileHeader *multipart.FileHeader) error {
        // Save uploaded file
        tempPath := "/tmp/restore_backup.tar.gz"
        
        file, err := fileHeader.Open()
        if err != nil {
                return fmt.Errorf("failed to open backup file: %v", err)
        }
        defer file.Close()
        
        data, err := ioutil.ReadAll(file)
        if err != nil {
                return fmt.Errorf("failed to read backup file: %v", err)
        }
        
        if err := ioutil.WriteFile(tempPath, data, 0644); err != nil {
                return fmt.Errorf("failed to save backup file: %v", err)
        }
        
        // Extract backup
        if err := exec.Command("tar", "xzf", tempPath, "-C", "/").Run(); err != nil {
                return fmt.Errorf("failed to extract backup: %v", err)
        }
        
        // Clean up
        os.Remove(tempPath)
        
        return nil
}
