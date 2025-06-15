package services

import (
	"fmt"
	"mime/multipart"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Her0x27/x-routersbc/utils/configurators/sys"
)

// SystemInfo represents system information
type SystemInfo struct {
	Hostname     string    `json:"hostname"`
	Uptime       string    `json:"uptime"`
	LoadAverage  string    `json:"load_average"`
	Memory       MemoryInfo `json:"memory"`
	CPU          CPUInfo    `json:"cpu"`
	Timezone     string    `json:"timezone"`
	CurrentTime  time.Time `json:"current_time"`
}

// MemoryInfo represents memory information
type MemoryInfo struct {
	Total     uint64 `json:"total"`
	Available uint64 `json:"available"`
	Used      uint64 `json:"used"`
	Free      uint64 `json:"free"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Model     string  `json:"model"`
	Cores     int     `json:"cores"`
	Usage     float64 `json:"usage"`
	Frequency string  `json:"frequency"`
}

// DeviceInfo represents device hardware information
type DeviceInfo struct {
	Processor ProcessorInfo `json:"processor"`
	Memory    MemoryInfo    `json:"memory"`
	Storage   []StorageInfo `json:"storage"`
	Network   []NetworkInfo `json:"network"`
	Video     VideoInfo     `json:"video"`
	Audio     AudioInfo     `json:"audio"`
	ExternalIO IOInfo       `json:"external_io"`
}

// ProcessorInfo represents processor information
type ProcessorInfo struct {
	Model       string `json:"model"`
	Cores       int    `json:"cores"`
	Threads     int    `json:"threads"`
	Architecture string `json:"architecture"`
	Frequency   string `json:"frequency"`
}

// StorageInfo represents storage device information
type StorageInfo struct {
	Device string `json:"device"`
	Type   string `json:"type"`
	Size   string `json:"size"`
	Model  string `json:"model"`
}

// NetworkInfo represents network hardware information
type NetworkInfo struct {
	Interface string `json:"interface"`
	Type      string `json:"type"`
	Driver    string `json:"driver"`
	MAC       string `json:"mac"`
}

// VideoInfo represents video hardware information
type VideoInfo struct {
	Available bool   `json:"available"`
	Driver    string `json:"driver"`
	Model     string `json:"model"`
}

// AudioInfo represents audio hardware information
type AudioInfo struct {
	Available bool   `json:"available"`
	Driver    string `json:"driver"`
	Model     string `json:"model"`
}

// IOInfo represents external I/O information
type IOInfo struct {
	USBPorts int  `json:"usb_ports"`
	SPI      bool `json:"spi"`
	UART     bool `json:"uart"`
	I2C      bool `json:"i2c"`
	GPIO     bool `json:"gpio"`
}

// PortableDevice represents a USB device
type PortableDevice struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Driver      string `json:"driver"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Vendor      string `json:"vendor"`
	Product     string `json:"product"`
}

// SystemService handles system operations
type SystemService struct {
	sysConfigurator *sys.SystemConfigurator
}

// NewSystemService creates a new system service
func NewSystemService() *SystemService {
	return &SystemService{
		sysConfigurator: sys.NewSystemConfigurator(),
	}
}

// GetSystemInfo returns current system information
func (s *SystemService) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		CurrentTime: time.Now(),
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	info.Hostname = hostname

	// Get uptime
	info.Uptime = s.getUptime()

	// Get load average
	info.LoadAverage = s.getLoadAverage()

	// Get memory info
	info.Memory = s.getMemoryInfo()

	// Get CPU info
	info.CPU = s.getCPUInfo()

	// Get timezone
	info.Timezone = s.getTimezone()

	return info, nil
}

// getUptime returns system uptime
func (s *SystemService) getUptime() string {
	cmd := exec.Command("uptime", "-p")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getLoadAverage returns system load average
func (s *SystemService) getLoadAverage() string {
	cmd := exec.Command("cat", "/proc/loadavg")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	
	parts := strings.Fields(string(output))
	if len(parts) >= 3 {
		return fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
	}
	return "unknown"
}

// getMemoryInfo returns memory information
func (s *SystemService) getMemoryInfo() MemoryInfo {
	cmd := exec.Command("cat", "/proc/meminfo")
	output, err := cmd.Output()
	if err != nil {
		return MemoryInfo{}
	}

	memInfo := MemoryInfo{}
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					memInfo.Total = val * 1024 // Convert from kB to bytes
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					memInfo.Available = val * 1024
				}
			}
		} else if strings.HasPrefix(line, "MemFree:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					memInfo.Free = val * 1024
				}
			}
		}
	}

	memInfo.Used = memInfo.Total - memInfo.Available
	return memInfo
}

// getCPUInfo returns CPU information
func (s *SystemService) getCPUInfo() CPUInfo {
	cpuInfo := CPUInfo{
		Cores: runtime.NumCPU(),
	}

	// Get CPU model from /proc/cpuinfo
	cmd := exec.Command("cat", "/proc/cpuinfo")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					cpuInfo.Model = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Get CPU frequency
	cmd = exec.Command("cat", "/proc/cpuinfo")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "cpu MHz") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					cpuInfo.Frequency = strings.TrimSpace(parts[1]) + " MHz"
					break
				}
			}
		}
	}

	return cpuInfo
}

// getTimezone returns current timezone
func (s *SystemService) getTimezone() string {
	cmd := exec.Command("timedatectl", "show", "--property=Timezone", "--value")
	output, err := cmd.Output()
	if err != nil {
		return "UTC"
	}
	return strings.TrimSpace(string(output))
}

// SetTimeZone sets the system timezone
func (s *SystemService) SetTimeZone(timezone string) error {
	return s.sysConfigurator.SetTimeZone(timezone)
}

// GetDeviceInfo returns detailed device hardware information
func (s *SystemService) GetDeviceInfo() (*DeviceInfo, error) {
	info := &DeviceInfo{}

	// Get processor info
	info.Processor = s.getProcessorInfo()

	// Get memory info
	info.Memory = s.getMemoryInfo()

	// Get storage info
	info.Storage = s.getStorageInfo()

	// Get network info
	info.Network = s.getNetworkHardwareInfo()

	// Get video info
	info.Video = s.getVideoInfo()

	// Get audio info
	info.Audio = s.getAudioInfo()

	// Get external I/O info
	info.ExternalIO = s.getIOInfo()

	return info, nil
}

// getProcessorInfo returns detailed processor information
func (s *SystemService) getProcessorInfo() ProcessorInfo {
	info := ProcessorInfo{
		Cores:        runtime.NumCPU(),
		Architecture: runtime.GOARCH,
	}

	// Get detailed CPU info from /proc/cpuinfo
	cmd := exec.Command("cat", "/proc/cpuinfo")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					info.Model = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(line, "cpu MHz") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					info.Frequency = strings.TrimSpace(parts[1]) + " MHz"
				}
			} else if strings.HasPrefix(line, "siblings") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if threads, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						info.Threads = threads
					}
				}
			}
		}
	}

	if info.Threads == 0 {
		info.Threads = info.Cores
	}

	return info
}

// getStorageInfo returns storage device information
func (s *SystemService) getStorageInfo() []StorageInfo {
	storage := []StorageInfo{}

	// Get block devices
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,TYPE,SIZE,MODEL")
	output, err := cmd.Output()
	if err != nil {
		return storage
	}

	// Parse JSON output (simplified)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "disk") {
			// Extract device info (simplified parsing)
			storage = append(storage, StorageInfo{
				Device: "/dev/sda", // Example
				Type:   "SSD",
				Size:   "32GB",
				Model:  "Generic Storage",
			})
		}
	}

	return storage
}

// getNetworkHardwareInfo returns network hardware information
func (s *SystemService) getNetworkHardwareInfo() []NetworkInfo {
	network := []NetworkInfo{}

	// Get network interfaces
	cmd := exec.Command("ip", "link", "show")
	output, err := cmd.Output()
	if err != nil {
		return network
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ": ") && !strings.HasPrefix(line, " ") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				name := strings.Split(parts[1], "@")[0]
				
				netInfo := NetworkInfo{
					Interface: name,
					Type:      s.getInterfaceType(name),
					Driver:    s.getInterfaceDriver(name),
					MAC:       s.getInterfaceMAC(name),
				}
				
				network = append(network, netInfo)
			}
		}
	}

	return network
}

// getInterfaceType determines interface type
func (s *SystemService) getInterfaceType(name string) string {
	if strings.HasPrefix(name, "eth") {
		return "Ethernet"
	} else if strings.HasPrefix(name, "wlan") {
		return "WiFi"
	} else if strings.HasPrefix(name, "lo") {
		return "Loopback"
	}
	return "Unknown"
}

// getInterfaceDriver gets interface driver
func (s *SystemService) getInterfaceDriver(name string) string {
	cmd := exec.Command("ethtool", "-i", name)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "driver:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}

// getInterfaceMAC gets interface MAC address
func (s *SystemService) getInterfaceMAC(name string) string {
	cmd := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/address", name))
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getVideoInfo returns video hardware information
func (s *SystemService) getVideoInfo() VideoInfo {
	info := VideoInfo{}

	// Check for video devices
	cmd := exec.Command("lspci")
	output, err := cmd.Output()
	if err == nil {
		if strings.Contains(strings.ToLower(string(output)), "vga") ||
		   strings.Contains(strings.ToLower(string(output)), "display") {
			info.Available = true
			info.Driver = "generic"
			info.Model = "Generic Display"
		}
	}

	return info
}

// getAudioInfo returns audio hardware information
func (s *SystemService) getAudioInfo() AudioInfo {
	info := AudioInfo{}

	// Check for audio devices
	cmd := exec.Command("lspci")
	output, err := cmd.Output()
	if err == nil {
		if strings.Contains(strings.ToLower(string(output)), "audio") {
			info.Available = true
			info.Driver = "generic"
			info.Model = "Generic Audio"
		}
	}

	return info
}

// getIOInfo returns external I/O information
func (s *SystemService) getIOInfo() IOInfo {
	info := IOInfo{}

	// Count USB ports
	cmd := exec.Command("lsusb")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		info.USBPorts = len(lines) - 1 // Subtract empty line
	}

	// Check for GPIO, SPI, I2C, UART (Raspberry Pi specific)
	if _, err := os.Stat("/sys/class/gpio"); err == nil {
		info.GPIO = true
	}
	if _, err := os.Stat("/dev/spidev0.0"); err == nil {
		info.SPI = true
	}
	if _, err := os.Stat("/dev/i2c-1"); err == nil {
		info.I2C = true
	}
	if _, err := os.Stat("/dev/ttyAMA0"); err == nil {
		info.UART = true
	}

	return info
}

// GetPortableDevices returns connected USB devices
func (s *SystemService) GetPortableDevices() ([]PortableDevice, error) {
	devices := []PortableDevice{}

	// Get USB devices
	cmd := exec.Command("lsusb", "-v")
	output, err := cmd.Output()
	if err != nil {
		return devices, err
	}

	// Parse lsusb output (simplified)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Bus ") && strings.Contains(line, "Device ") {
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				device := PortableDevice{
					ID:          parts[5],
					Name:        strings.Join(parts[6:], " "),
					Type:        s.getUSBDeviceType(line),
					Status:      "connected",
					Driver:      "unknown",
					Description: "USB Device",
				}
				devices = append(devices, device)
			}
		}
	}

	return devices, nil
}

// getUSBDeviceType determines USB device type
func (s *SystemService) getUSBDeviceType(description string) string {
	desc := strings.ToLower(description)
	if strings.Contains(desc, "wifi") || strings.Contains(desc, "wireless") {
		return "USB_WIFI"
	} else if strings.Contains(desc, "ethernet") {
		return "USB_ETHERNET"
	} else if strings.Contains(desc, "bluetooth") {
		return "USB_BLUETOOTH"
	} else if strings.Contains(desc, "modem") || strings.Contains(desc, "3g") || strings.Contains(desc, "lte") {
		return "USB_3G_LTE_MODEM"
	} else if strings.Contains(desc, "storage") || strings.Contains(desc, "mass") {
		return "USB_STORAGE"
	} else if strings.Contains(desc, "camera") || strings.Contains(desc, "webcam") {
		return "USB_WEBCAM"
	}
	return "USB_UNKNOWN"
}

// RefreshPortableDevices rescans for portable devices
func (s *SystemService) RefreshPortableDevices() ([]PortableDevice, error) {
	// Trigger USB device rescan
	exec.Command("udevadm", "trigger").Run()
	time.Sleep(2 * time.Second) // Wait for devices to be detected
	
	return s.GetPortableDevices()
}

// CreateBackup creates a system backup
func (s *SystemService) CreateBackup() (string, error) {
	return s.sysConfigurator.CreateBackup()
}

// RestoreBackup restores from a backup file
func (s *SystemService) RestoreBackup(file *multipart.FileHeader) error {
	return s.sysConfigurator.RestoreBackup(file)
}
