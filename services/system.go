package services

import (
	"database/sql"
	"fmt"
	"mime/multipart"

	"github.com/Her0x27/x-routersbc/utils"
	"github.com/Her0x27/x-routersbc/utils/configurators/sys"
)

type SystemService struct {
	db *sql.DB
}

type DeviceInformation struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Storage StorageInfo `json:"storage"`
	Network NetworkInfo `json:"network"`
	Video   VideoInfo   `json:"video"`
	Audio   AudioInfo   `json:"audio"`
	IO      IOInfo      `json:"io"`
}

type CPUInfo struct {
	Model  string `json:"model"`
	Cores  int    `json:"cores"`
	Arch   string `json:"architecture"`
	Speed  string `json:"speed"`
}

type MemoryInfo struct {
	Total     uint64 `json:"total"`
	Available uint64 `json:"available"`
	Used      uint64 `json:"used"`
	Type      string `json:"type"`
}

type StorageInfo struct {
	Devices []StorageDevice `json:"devices"`
}

type StorageDevice struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Size       uint64 `json:"size"`
	Available  uint64 `json:"available"`
	MountPoint string `json:"mount_point"`
}

type NetworkInfo struct {
	Ethernet  []NetworkDevice `json:"ethernet"`
	WiFi      []NetworkDevice `json:"wifi"`
	Bluetooth []NetworkDevice `json:"bluetooth"`
}

type NetworkDevice struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Status string `json:"status"`
	MAC    string `json:"mac"`
}

type VideoInfo struct {
	Supported bool              `json:"supported"`
	Devices   []VideoDevice     `json:"devices"`
}

type VideoDevice struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Status string `json:"status"`
}

type AudioInfo struct {
	Supported bool              `json:"supported"`
	Devices   []AudioDevice     `json:"devices"`
}

type AudioDevice struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Status string `json:"status"`
}

type IOInfo struct {
	USB  USBInfo  `json:"usb"`
	GPIO GPIOInfo `json:"gpio"`
	SPI  bool     `json:"spi"`
	UART bool     `json:"uart"`
	I2C  bool     `json:"i2c"`
}

type USBInfo struct {
	Ports   int         `json:"ports"`
	Devices []USBDevice `json:"devices"`
}

type USBDevice struct {
	Port     int    `json:"port"`
	Vendor   string `json:"vendor"`
	Product  string `json:"product"`
	VendorID string `json:"vendor_id"`
	ProductID string `json:"product_id"`
}

type GPIOInfo struct {
	Available bool `json:"available"`
	Pins      int  `json:"pins"`
}

type PortableDevice struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Driver      string `json:"driver"`
	DriverStatus string `json:"driver_status"`
	Suggestion  string `json:"suggestion"`
}

type SystemTime struct {
	CurrentTime string   `json:"current_time"`
	Timezone    string   `json:"timezone"`
	NTPEnabled  bool     `json:"ntp_enabled"`
	NTPServers  []string `json:"ntp_servers"`
}

func NewSystemService(db *sql.DB) *SystemService {
	return &SystemService{db: db}
}

func (s *SystemService) GetDeviceInformation() (*DeviceInformation, error) {
	deviceInfo := &DeviceInformation{}
	
	// Get CPU information
	cpuInfo, err := utils.GetCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %v", err)
	}
	deviceInfo.CPU = CPUInfo{
		Model: cpuInfo.Model,
		Cores: cpuInfo.Cores,
		Arch:  cpuInfo.Architecture,
		Speed: cpuInfo.Speed,
	}
	
	// Get memory information
	memInfo, err := utils.GetMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}
	deviceInfo.Memory = MemoryInfo{
		Total:     memInfo.Total,
		Available: memInfo.Available,
		Used:      memInfo.Used,
		Type:      memInfo.Type,
	}
	
	// Get storage information
	storageDevices, err := utils.GetStorageInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage info: %v", err)
	}
	
	var devices []StorageDevice
	for _, dev := range storageDevices {
		devices = append(devices, StorageDevice{
			Name:       dev.Name,
			Type:       dev.Type,
			Size:       dev.Size,
			Available:  dev.Available,
			MountPoint: dev.MountPoint,
		})
	}
	deviceInfo.Storage = StorageInfo{Devices: devices}
	
	// Get network device information
	networkDevices, err := utils.GetNetworkDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get network devices: %v", err)
	}
	
	var ethernet, wifi, bluetooth []NetworkDevice
	for _, dev := range networkDevices {
		netDev := NetworkDevice{
			Name:   dev.Name,
			Driver: dev.Driver,
			Status: dev.Status,
			MAC:    dev.MAC,
		}
		
		switch dev.Type {
		case "ethernet":
			ethernet = append(ethernet, netDev)
		case "wifi":
			wifi = append(wifi, netDev)
		case "bluetooth":
			bluetooth = append(bluetooth, netDev)
		}
	}
	
	deviceInfo.Network = NetworkInfo{
		Ethernet:  ethernet,
		WiFi:      wifi,
		Bluetooth: bluetooth,
	}
	
	// Get video information
	videoDevices, err := utils.GetVideoDevices()
	if err == nil {
		var vDevices []VideoDevice
		for _, dev := range videoDevices {
			vDevices = append(vDevices, VideoDevice{
				Name:   dev.Name,
				Driver: dev.Driver,
				Status: dev.Status,
			})
		}
		deviceInfo.Video = VideoInfo{
			Supported: len(vDevices) > 0,
			Devices:   vDevices,
		}
	}
	
	// Get audio information
	audioDevices, err := utils.GetAudioDevices()
	if err == nil {
		var aDevices []AudioDevice
		for _, dev := range audioDevices {
			aDevices = append(aDevices, AudioDevice{
				Name:   dev.Name,
				Driver: dev.Driver,
				Status: dev.Status,
			})
		}
		deviceInfo.Audio = AudioInfo{
			Supported: len(aDevices) > 0,
			Devices:   aDevices,
		}
	}
	
	// Get I/O information
	ioInfo, err := utils.GetIOInfo()
	if err == nil {
		var usbDevices []USBDevice
		for _, dev := range ioInfo.USB.Devices {
			usbDevices = append(usbDevices, USBDevice{
				Port:      dev.Port,
				Vendor:    dev.Vendor,
				Product:   dev.Product,
				VendorID:  dev.VendorID,
				ProductID: dev.ProductID,
			})
		}
		
		deviceInfo.IO = IOInfo{
			USB: USBInfo{
				Ports:   ioInfo.USB.Ports,
				Devices: usbDevices,
			},
			GPIO: GPIOInfo{
				Available: ioInfo.GPIO.Available,
				Pins:      ioInfo.GPIO.Pins,
			},
			SPI:  ioInfo.SPI,
			UART: ioInfo.UART,
			I2C:  ioInfo.I2C,
		}
	}
	
	return deviceInfo, nil
}

func (s *SystemService) GetPortableDevices() ([]PortableDevice, error) {
	devices, err := utils.GetUSBDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get USB devices: %v", err)
	}
	
	var portableDevices []PortableDevice
	
	for _, device := range devices {
		portableDevice := PortableDevice{
			Name:   device.Product,
			Status: device.Status,
			Driver: device.Driver,
		}
		
		// Determine device type and driver status
		switch {
		case utils.IsWiFiDevice(device):
			portableDevice.Type = "USB_WIFI"
			portableDevice.DriverStatus = utils.GetWiFiDriverStatus(device)
			portableDevice.Suggestion = utils.GetWiFiDriverSuggestion(device)
			
		case utils.IsEthernetDevice(device):
			portableDevice.Type = "USB_ETHERNET"
			portableDevice.DriverStatus = utils.GetEthernetDriverStatus(device)
			portableDevice.Suggestion = utils.GetEthernetDriverSuggestion(device)
			
		case utils.IsBluetoothDevice(device):
			portableDevice.Type = "USB_BLUETOOTH"
			portableDevice.DriverStatus = utils.GetBluetoothDriverStatus(device)
			portableDevice.Suggestion = utils.GetBluetoothDriverSuggestion(device)
			
		case utils.IsModemDevice(device):
			portableDevice.Type = "USB_3G_LTE_MODEM"
			portableDevice.DriverStatus = utils.GetModemDriverStatus(device)
			portableDevice.Suggestion = utils.GetModemDriverSuggestion(device)
			
		case utils.IsStorageDevice(device):
			portableDevice.Type = "USB_STORAGE"
			portableDevice.DriverStatus = utils.GetStorageDriverStatus(device)
			portableDevice.Suggestion = utils.GetStorageDriverSuggestion(device)
			
		case utils.IsWebcamDevice(device):
			portableDevice.Type = "USB_WEBCAM"
			portableDevice.DriverStatus = utils.GetWebcamDriverStatus(device)
			portableDevice.Suggestion = utils.GetWebcamDriverSuggestion(device)
			
		default:
			continue // Skip unknown devices
		}
		
		portableDevices = append(portableDevices, portableDevice)
	}
	
	return portableDevices, nil
}

func (s *SystemService) GetSystemTime() (*SystemTime, error) {
	timeInfo, err := sys.GetSystemTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get system time: %v", err)
	}
	
	return &SystemTime{
		CurrentTime: timeInfo.CurrentTime,
		Timezone:    timeInfo.Timezone,
		NTPEnabled:  timeInfo.NTPEnabled,
		NTPServers:  timeInfo.NTPServers,
	}, nil
}

func (s *SystemService) SetSystemTime(timezone string, ntpEnabled bool, ntpServers []string) error {
	return sys.SetSystemTime(timezone, ntpEnabled, ntpServers)
}

func (s *SystemService) CreateBackup() (string, error) {
	return utils.CreateSystemBackup()
}

func (s *SystemService) RestoreBackup(file *multipart.FileHeader) error {
	return utils.RestoreSystemBackup(file)
}

func (s *SystemService) GetPortableDeviceStatus(deviceType string) (map[string]interface{}, error) {
	return utils.GetDeviceStatus(deviceType)
}

func (s *SystemService) InstallDrivers(deviceType string) error {
	return utils.InstallDrivers(deviceType)
}
