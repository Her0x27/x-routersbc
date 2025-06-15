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

// SystemService handles system operations
type SystemService struct {
	serviceManager *sys.ServiceManager
}

// DeviceInformation represents device hardware information
type DeviceInformation struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Storage StorageInfo `json:"storage"`
	Network NetworkInfo `json:"network"`
	Video   VideoInfo   `json:"video"`
	Audio   AudioInfo   `json:"audio"`
	IO      IOInfo      `json:"io"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Model string `json:"model"`
	Cores int    `json:"cores"`
	Arch  string `json:"architecture"`
}

// MemoryInfo represents memory information
type MemoryInfo struct {
	Total     uint64 `json:"total"`
	Available uint64 `json:"available"`
	Used      uint64 `json:"used"`
	Free      uint64 `json:"free"`
}

// StorageInfo represents storage information
type StorageInfo struct {
	Devices []StorageDevice `json:"devices"`
}

// StorageDevice represents a storage device
type StorageDevice struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // ssd, sdcard, emmc, hdd
	Size       uint64 `json:"size"`
	Used       uint64 `json:"used"`
	Available  uint64 `json:"available"`
	MountPoint string `json:"mount_point"`
}

// NetworkInfo represents network hardware information
type NetworkInfo struct {
	Ethernet  []string `json:"ethernet"`
	WiFi      []string `json:"wifi"`
	Bluetooth []string `json:"bluetooth"`
}

// VideoInfo represents video hardware information
type VideoInfo struct {
	Devices []VideoDevice `json:"devices"`
}

// VideoDevice represents a video device
type VideoDevice struct {
	Name        string `json:"name"`
	Driver      string `json:"driver"`
	Description string `json:"description"`
}

// AudioInfo represents audio hardware information
type AudioInfo struct {
	Devices []AudioDevice `json:"devices"`
}

// AudioDevice represents an audio device
type AudioDevice struct {
	Name        string `json:"name"`
	Driver      string `json:"driver"`
	Description string `json:"description"`
}

// IOInfo represents I/O information
type IOInfo struct {
	USBPorts int      `json:"usb_ports"`
	GPIO     []string `json:"gpio"`
	SPI      []string `json:"spi"`
	UART     []string `json:"uart"`
	I2C      []string `json:"i2c"`
}

// USBDevice represents a USB device
type USBDevice struct {
	ID          string `json:"id"`
	VendorID    string `json:"vendor_id"`
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // wifi, ethernet, bluetooth, modem, storage, webcam
	Driver      string `json:"driver"`
	Status      string `json:"status"` // connected, disconnected, error
	Description string `json:"description"`
}

// SystemStatus represents current system status
type SystemStatus struct {
	Uptime       string  `json:"uptime"`
	LoadAverage  string  `json:"load_average"`
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	Temperature  float64 `json:"temperature"`
	NetworkRX    uint64  `json:"network_rx"`
	NetworkTX    uint64  `json:"network_tx"`
	ProcessCount int     `json:"process_count"`
}

// NewSystemService creates a new system service
func NewSystemService() *SystemService {
	return &SystemService{
		serviceManager: sys.NewServiceManager(),
	}
}

// GetDeviceInformation gets device hardware information
func (s *SystemService) GetDeviceInformation() (*DeviceInformation, error) {
	info := &DeviceInformation{}
	
	// Get CPU information
	cpuInfo, err := s.getCPUInformation()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %v", err)
	}
	info.CPU = *cpuInfo
	
	// Get memory information
	memInfo, err := s.getMemoryInformation()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}
	info.Memory = *memInfo
	
	// Get storage information
	storageInfo, err := s.getStorageInformation()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage info: %v", err)
	}
	info.Storage = *storageInfo
	
	// Get network hardware information
	networkInfo, err := s.getNetworkInformation()
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %v", err)
	}
	info.Network = *networkInfo
	
	// Get video information
	videoInfo, err := s.getVideoInformation()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %v", err)
	}
	info.Video = *videoInfo
	
	// Get audio information
	audioInfo, err := s.getAudioInformation()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %v", err)
	}
	info.Audio = *audioInfo
	
	// Get I/O information
	ioInfo, err := s.getIOInformation()
	if err != nil {
		return nil, fmt.Errorf("failed to get I/O info: %v", err)
	}
	info.IO = *ioInfo
	
	return info, nil
}

// getCPUInformation gets CPU information
func (s *SystemService) getCPUInformation() (*CPUInfo, error) {
	info := &CPUInfo{
		Cores: runtime.NumCPU(),
		Arch:  runtime.GOARCH,
	}
	
	// Try to get CPU model from /proc/cpuinfo
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					info.Model = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}
	
	if info.Model == "" {
		info.Model = "Unknown CPU"
	}
	
	return info, nil
}

// getMemoryInformation gets memory information
func (s *SystemService) getMemoryInformation() (*MemoryInfo, error) {
	info := &MemoryInfo{}
	
	// Read /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		
		// Convert from KB to bytes
		value *= 1024
		
		switch key {
		case "MemTotal":
			info.Total = value
		case "MemAvailable":
			info.Available = value
		case "MemFree":
			info.Free = value
		}
	}
	
	info.Used = info.Total - info.Available
	
	return info, nil
}

// getStorageInformation gets storage information
func (s *SystemService) getStorageInformation() (*StorageInfo, error) {
	info := &StorageInfo{
		Devices: []StorageDevice{},
	}
	
	// Execute df command to get disk usage
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return info, nil // Return empty if df fails
	}
	
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" { // Skip header
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		
		device := StorageDevice{
			Name:       fields[0],
			MountPoint: fields[5],
		}
		
		// Parse sizes (remove units like G, M, K)
		sizeStr := strings.TrimRight(fields[1], "GMKT")
		if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
			multiplier := uint64(1)
			if strings.HasSuffix(fields[1], "G") {
				multiplier = 1024 * 1024 * 1024
			} else if strings.HasSuffix(fields[1], "M") {
				multiplier = 1024 * 1024
			} else if strings.HasSuffix(fields[1], "K") {
				multiplier = 1024
			}
			device.Size = uint64(size * float64(multiplier))
		}
		
		usedStr := strings.TrimRight(fields[2], "GMKT")
		if used, err := strconv.ParseFloat(usedStr, 64); err == nil {
			multiplier := uint64(1)
			if strings.HasSuffix(fields[2], "G") {
				multiplier = 1024 * 1024 * 1024
			} else if strings.HasSuffix(fields[2], "M") {
				multiplier = 1024 * 1024
			} else if strings.HasSuffix(fields[2], "K") {
				multiplier = 1024
			}
			device.Used = uint64(used * float64(multiplier))
		}
		
		availStr := strings.TrimRight(fields[3], "GMKT")
		if avail, err := strconv.ParseFloat(availStr, 64); err == nil {
			multiplier := uint64(1)
			if strings.HasSuffix(fields[3], "G") {
				multiplier = 1024 * 1024 * 1024
			} else if strings.HasSuffix(fields[3], "M") {
				multiplier = 1024 * 1024
			} else if strings.HasSuffix(fields[3], "K") {
				multiplier = 1024
			}
			device.Available = uint64(avail * float64(multiplier))
		}
		
		// Determine storage type based on device name
		if strings.Contains(device.Name, "mmc") {
			device.Type = "emmc"
		} else if strings.Contains(device.Name, "sd") {
			device.Type = "sdcard"
		} else if strings.Contains(device.Name, "nvme") {
			device.Type = "ssd"
		} else {
			device.Type = "hdd"
		}
		
		info.Devices = append(info.Devices, device)
	}
	
	return info, nil
}

// getNetworkInformation gets network hardware information
func (s *SystemService) getNetworkInformation() (*NetworkInfo, error) {
	info := &NetworkInfo{
		Ethernet:  []string{},
		WiFi:      []string{},
		Bluetooth: []string{},
	}
	
	// List network interfaces
	cmd := exec.Command("ls", "/sys/class/net")
	output, err := cmd.Output()
	if err != nil {
		return info, nil
	}
	
	interfaces := strings.Fields(string(output))
	for _, iface := range interfaces {
		if strings.HasPrefix(iface, "eth") {
			info.Ethernet = append(info.Ethernet, iface)
		} else if strings.HasPrefix(iface, "wlan") {
			info.WiFi = append(info.WiFi, iface)
		}
	}
	
	// Check for Bluetooth
	if _, err := os.Stat("/sys/class/bluetooth"); err == nil {
		cmd := exec.Command("ls", "/sys/class/bluetooth")
		if output, err := cmd.Output(); err == nil {
			devices := strings.Fields(string(output))
			info.Bluetooth = devices
		}
	}
	
	return info, nil
}

// getVideoInformation gets video hardware information
func (s *SystemService) getVideoInformation() (*VideoInfo, error) {
	info := &VideoInfo{
		Devices: []VideoDevice{},
	}
	
	// Check for video devices in /dev
	cmd := exec.Command("ls", "/dev/video*")
	if output, err := cmd.Output(); err == nil {
		devices := strings.Fields(string(output))
		for _, device := range devices {
			videoDevice := VideoDevice{
				Name:        device,
				Driver:      "unknown",
				Description: "Video device",
			}
			info.Devices = append(info.Devices, videoDevice)
		}
	}
	
	return info, nil
}

// getAudioInformation gets audio hardware information
func (s *SystemService) getAudioInformation() (*AudioInfo, error) {
	info := &AudioInfo{
		Devices: []AudioDevice{},
	}
	
	// Check for ALSA devices
	if _, err := os.Stat("/proc/asound/cards"); err == nil {
		data, err := os.ReadFile("/proc/asound/cards")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					audioDevice := AudioDevice{
						Name:        line,
						Driver:      "ALSA",
						Description: "Audio device",
					}
					info.Devices = append(info.Devices, audioDevice)
				}
			}
		}
	}
	
	return info, nil
}

// getIOInformation gets I/O information
func (s *SystemService) getIOInformation() (*IOInfo, error) {
	info := &IOInfo{
		GPIO: []string{},
		SPI:  []string{},
		UART: []string{},
		I2C:  []string{},
	}
	
	// Count USB ports by checking /sys/bus/usb/devices
	if entries, err := os.ReadDir("/sys/bus/usb/devices"); err == nil {
		info.USBPorts = len(entries)
	}
	
	// Check for GPIO
	if entries, err := os.ReadDir("/sys/class/gpio"); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "gpio") {
				info.GPIO = append(info.GPIO, entry.Name())
			}
		}
	}
	
	// Check for SPI
	if entries, err := os.ReadDir("/dev"); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "spi") {
				info.SPI = append(info.SPI, entry.Name())
			}
		}
	}
	
	// Check for UART/Serial
	if entries, err := os.ReadDir("/dev"); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "ttyS") || strings.HasPrefix(entry.Name(), "ttyUSB") {
				info.UART = append(info.UART, entry.Name())
			}
		}
	}
	
	// Check for I2C
	if entries, err := os.ReadDir("/dev"); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "i2c-") {
				info.I2C = append(info.I2C, entry.Name())
			}
		}
	}
	
	return info, nil
}

// GetUSBDevices gets connected USB devices
func (s *SystemService) GetUSBDevices() ([]USBDevice, error) {
	var devices []USBDevice
	
	// Execute lsusb command
	cmd := exec.Command("lsusb")
	output, err := cmd.Output()
	if err != nil {
		return devices, nil // Return empty list if lsusb fails
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Parse lsusb output
		// Format: Bus 001 Device 002: ID 1d6b:0002 Linux Foundation 2.0 root hub
		parts := strings.Split(line, " ")
		if len(parts) < 6 {
			continue
		}
		
		idPart := ""
		namePart := ""
		for i, part := range parts {
			if strings.HasPrefix(part, "ID") {
				idPart = strings.TrimPrefix(part, "ID")
				namePart = strings.Join(parts[i+1:], " ")
				break
			}
		}
		
		if idPart == "" {
			continue
		}
		
		// Split vendor:product ID
		idParts := strings.Split(idPart, ":")
		if len(idParts) != 2 {
			continue
		}
		
		device := USBDevice{
			ID:        idPart,
			VendorID:  idParts[0],
			ProductID: idParts[1],
			Name:      namePart,
			Status:    "connected",
			Driver:    "unknown",
		}
		
		// Determine device type based on name/vendor
		deviceName := strings.ToLower(device.Name)
		if strings.Contains(deviceName, "wifi") || strings.Contains(deviceName, "wireless") {
			device.Type = "wifi"
		} else if strings.Contains(deviceName, "ethernet") || strings.Contains(deviceName, "network") {
			device.Type = "ethernet"
		} else if strings.Contains(deviceName, "bluetooth") {
			device.Type = "bluetooth"
		} else if strings.Contains(deviceName, "modem") || strings.Contains(deviceName, "3g") || strings.Contains(deviceName, "lte") {
			device.Type = "modem"
		} else if strings.Contains(deviceName, "storage") || strings.Contains(deviceName, "mass") {
			device.Type = "storage"
		} else if strings.Contains(deviceName, "camera") || strings.Contains(deviceName, "webcam") {
			device.Type = "webcam"
		} else {
			device.Type = "unknown"
		}
		
		devices = append(devices, device)
	}
	
	return devices, nil
}

// GetSystemStatus gets current system status
func (s *SystemService) GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{}
	
	// Get uptime
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			if seconds, err := strconv.ParseFloat(fields[0], 64); err == nil {
				duration := time.Duration(seconds) * time.Second
				status.Uptime = duration.String()
			}
		}
	}
	
	// Get load average
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			status.LoadAverage = fmt.Sprintf("%s %s %s", fields[0], fields[1], fields[2])
		}
	}
	
	// Get memory usage
	if memInfo, err := s.getMemoryInformation(); err == nil {
		if memInfo.Total > 0 {
			status.MemoryUsage = float64(memInfo.Used) / float64(memInfo.Total) * 100
		}
	}
	
	// Get disk usage (root filesystem)
	if storageInfo, err := s.getStorageInformation(); err == nil {
		for _, device := range storageInfo.Devices {
			if device.MountPoint == "/" && device.Size > 0 {
				status.DiskUsage = float64(device.Used) / float64(device.Size) * 100
				break
			}
		}
	}
	
	// Get CPU temperature (if available)
	if data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp"); err == nil {
		if temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
			status.Temperature = temp / 1000 // Convert from millidegrees
		}
	}
	
	// Get network statistics
	if data, err := os.ReadFile("/proc/net/dev"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.Contains(line, ":") {
				fields := strings.Fields(line)
				if len(fields) >= 10 {
					// Sum up RX and TX bytes for all interfaces
					if rx, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						status.NetworkRX += rx
					}
					if tx, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
						status.NetworkTX += tx
					}
				}
			}
		}
	}
	
	// Get process count
	if entries, err := os.ReadDir("/proc"); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				if _, err := strconv.Atoi(entry.Name()); err == nil {
					status.ProcessCount++
				}
			}
		}
	}
	
	return status, nil
}

// CreateBackup creates a system backup
func (s *SystemService) CreateBackup() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("/tmp/routersbc_backup_%s.tar.gz", timestamp)
	
	// Create backup of configuration files
	cmd := exec.Command("tar", "-czf", backupPath,
		"routersbc.sqlitedb",
		"/etc/network/",
		"/etc/netplan/",
		"/etc/hostapd/",
		"/etc/dhcp/",
	)
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create backup: %v", err)
	}
	
	return backupPath, nil
}

// RestoreBackup restores from a backup file
func (s *SystemService) RestoreBackup(file *multipart.FileHeader) error {
	// Save uploaded file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()
	
	backupPath := "/tmp/restore_backup.tar.gz"
	dst, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer dst.Close()
	
	// Copy file contents
	if _, err := dst.ReadFrom(src); err != nil {
		return fmt.Errorf("failed to save backup file: %v", err)
	}
	dst.Close()
	
	// Extract backup
	cmd := exec.Command("tar", "-xzf", backupPath, "-C", "/")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract backup: %v", err)
	}
	
	// Clean up
	os.Remove(backupPath)
	
	return nil
}
