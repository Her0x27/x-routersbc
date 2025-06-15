package net

import (
        "fmt"
        "net"
        "os"
        "os/exec"
        "strings"
)

type SystemInterface struct {
        Name string
        Type string
        MAC  string
        IP   string
        IsUp bool
}

type InterfaceConfig struct {
        Name      string
        Type      string
        Settings  map[string]interface{}
}

func GetSystemInterfaces() ([]SystemInterface, error) {
        interfaces, err := net.Interfaces()
        if err != nil {
                return nil, fmt.Errorf("failed to get network interfaces: %v", err)
        }
        
        var systemInterfaces []SystemInterface
        
        for _, iface := range interfaces {
                // Skip loopback interface
                if iface.Flags&net.FlagLoopback != 0 {
                        continue
                }
                
                sysIface := SystemInterface{
                        Name: iface.Name,
                        MAC:  iface.HardwareAddr.String(),
                        IsUp: iface.Flags&net.FlagUp != 0,
                }
                
                // Determine interface type
                sysIface.Type = determineInterfaceType(iface.Name)
                
                // Get IP address
                addrs, err := iface.Addrs()
                if err == nil && len(addrs) > 0 {
                        for _, addr := range addrs {
                                if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                                        if ipnet.IP.To4() != nil {
                                                sysIface.IP = ipnet.IP.String()
                                                break
                                        }
                                }
                        }
                }
                
                systemInterfaces = append(systemInterfaces, sysIface)
        }
        
        return systemInterfaces, nil
}

func determineInterfaceType(name string) string {
        switch {
        case strings.HasPrefix(name, "eth"):
                return "ethernet"
        case strings.HasPrefix(name, "wlan"):
                return "wifi"
        case strings.HasPrefix(name, "bt"):
                return "bluetooth"
        case strings.HasPrefix(name, "br"):
                return "bridge"
        case strings.Contains(name, "."):
                return "vlan"
        case strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "tap"):
                return "vpn"
        case strings.HasPrefix(name, "wg"):
                return "wireguard"
        case strings.HasPrefix(name, "ppp"):
                return "ppp"
        default:
                return "unknown"
        }
}

func ApplyInterfaceConfiguration(name, ifaceType, config string) error {
        // Try netplan first if available
        if IsNetplanAvailable() {
                return applyNetplanInterfaceConfig(name, ifaceType, config)
        }
        
        // Fallback to traditional interfaces file
        return applyTraditionalInterfaceConfig(name, ifaceType, config)
}

func applyNetplanInterfaceConfig(name, ifaceType, config string) error {
        // Parse config JSON and convert to appropriate netplan config
        // This is a simplified implementation
        switch ifaceType {
        case "ethernet":
                return AddEthernetInterface(name, EthernetConfig{
                        DHCP4: true, // Default to DHCP, should parse from config
                })
        case "wifi":
                return AddWifiInterface(name, WifiConfig{
                        DHCP4: true, // Default to DHCP, should parse from config
                })
        case "vlan":
                // Parse VLAN ID from config
                return AddVlanInterface(name, VlanConfig{
                        ID:   1, // Should parse from config
                        Link: "eth0", // Should parse from config
                        DHCP4: true,
                })
        case "bridge":
                return AddBridgeInterface(name, BridgeConfig{
                        DHCP4: true,
                        Interfaces: []string{}, // Should parse from config
                })
        }
        
        return fmt.Errorf("unsupported interface type: %s", ifaceType)
}

func applyTraditionalInterfaceConfig(name, ifaceType, config string) error {
        // Implementation for /etc/network/interfaces
        // This is a fallback for systems without netplan
        
        configLines := []string{
                fmt.Sprintf("auto %s", name),
                fmt.Sprintf("iface %s inet dhcp", name), // Default to DHCP
        }
        
        // Read existing interfaces file
        interfacesPath := "/etc/network/interfaces"
        existingContent := ""
        
        if data, err := os.ReadFile(interfacesPath); err == nil {
                existingContent = string(data)
        }
        
        // Remove existing configuration for this interface
        existingContent = removeInterfaceFromConfig(existingContent, name)
        
        // Add new configuration
        newConfig := existingContent + "\n" + strings.Join(configLines, "\n") + "\n"
        
        // Write updated configuration
        if err := os.WriteFile(interfacesPath, []byte(newConfig), 0644); err != nil {
                return fmt.Errorf("failed to write interfaces config: %v", err)
        }
        
        // Restart networking
        return restartNetworking()
}

func removeInterfaceFromConfig(content, interfaceName string) string {
        lines := strings.Split(content, "\n")
        var result []string
        skipLines := false
        
        for _, line := range lines {
                trimmed := strings.TrimSpace(line)
                
                // Check if this line starts configuration for our interface
                if strings.HasPrefix(trimmed, "auto "+interfaceName) ||
                   strings.HasPrefix(trimmed, "iface "+interfaceName) {
                        skipLines = true
                        continue
                }
                
                // Check if this line starts configuration for another interface
                if strings.HasPrefix(trimmed, "auto ") ||
                   strings.HasPrefix(trimmed, "iface ") {
                        skipLines = false
                }
                
                if !skipLines {
                        result = append(result, line)
                }
        }
        
        return strings.Join(result, "\n")
}

func RemoveInterfaceConfiguration(name string) error {
        // Try netplan first
        if IsNetplanAvailable() {
                return RemoveInterface(name)
        }
        
        // Fallback to traditional interfaces
        interfacesPath := "/etc/network/interfaces"
        
        data, err := os.ReadFile(interfacesPath)
        if err != nil {
                return fmt.Errorf("failed to read interfaces config: %v", err)
        }
        
        content := removeInterfaceFromConfig(string(data), name)
        
        if err := os.WriteFile(interfacesPath, []byte(content), 0644); err != nil {
                return fmt.Errorf("failed to write interfaces config: %v", err)
        }
        
        return restartNetworking()
}

func restartNetworking() error {
        // Try systemd first
        if err := exec.Command("systemctl", "restart", "networking").Run(); err == nil {
                return nil
        }
        
        // Fallback to traditional service command
        if err := exec.Command("service", "networking", "restart").Run(); err == nil {
                return nil
        }
        
        // Try ifup/ifdown
        return exec.Command("ifdown", "-a").Run()
}

func BringInterfaceUp(name string) error {
        return exec.Command("ip", "link", "set", name, "up").Run()
}

func BringInterfaceDown(name string) error {
        return exec.Command("ip", "link", "set", name, "down").Run()
}

func SetInterfaceIP(name, ip, netmask string) error {
        // Remove existing IP addresses
        exec.Command("ip", "addr", "flush", "dev", name).Run()
        
        // Add new IP address
        cidr := fmt.Sprintf("%s/%s", ip, netmask)
        return exec.Command("ip", "addr", "add", cidr, "dev", name).Run()
}

func EnableDHCP(name string) error {
        // This would typically be handled by the network configuration files
        // For immediate effect, we can try dhclient
        return exec.Command("dhclient", name).Run()
}

func DisableDHCP(name string) error {
        // Kill dhclient for this interface
        return exec.Command("pkill", "-f", "dhclient.*"+name).Run()
}
