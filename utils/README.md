
# Utils Module - Вспомогательные инструменты

## Конфигураторы сети (configurators/net/)

### КРИТИЧЕСКИ ВАЖНАЯ безопасность

```go
// interfaces.go - ОБЯЗАТЕЛЬНАЯ валидация
func UpdateInterface(name, ipAddress, netmask string) error {
    // ОБЯЗАТЕЛЬНАЯ валидация имени интерфейса
    if !isValidInterfaceName(name) {
        return fmt.Errorf("invalid interface name: %s", name)
    }
    
    // ОБЯЗАТЕЛЬНАЯ валидация IP
    if ipAddress != "" && net.ParseIP(ipAddress) == nil {
        return fmt.Errorf("invalid IP address: %s", ipAddress)
    }
    
    // ОБЯЗАТЕЛЬНАЯ проверка существования
    if !interfaceExists(name) {
        return fmt.Errorf("interface %s does not exist", name)
    }
    
    // НИКОГДА не использовать прямую конкатенацию команд
    cmd := exec.Command("ip", "addr", "add", 
        fmt.Sprintf("%s/%s", ipAddress, netmask), 
        "dev", name)
    
    // ОБЯЗАТЕЛЬНАЯ обработка ошибок
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to update interface %s: %w, stderr: %s", 
            name, err, stderr.String())
    }
    
    return nil
}

// ОБЯЗАТЕЛЬНАЯ валидация имени интерфейса
func isValidInterfaceName(name string) bool {
    if len(name) == 0 || len(name) > 15 {
        return false
    }
    
    // Только разрешенные символы
    matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9-]*$`, name)
    return matched
}

// ОБЯЗАТЕЛЬНАЯ проверка существования
func interfaceExists(name string) bool {
    _, err := net.InterfaceByName(name)
    return err == nil
}
```

### Firewall конфигурация - КРИТИЧЕСКИЕ ошибки
```go
// firewall.go - ОБЯЗАТЕЛЬНАЯ валидация правил
func AddFirewallRule(rule FirewallRule) error {
    // ОБЯЗАТЕЛЬНАЯ валидация всех полей
    if err := validateFirewallRule(rule); err != nil {
        return fmt.Errorf("invalid firewall rule: %w", err)
    }
    
    // ТОЛЬКО безопасные команды iptables
    args := buildSafeIPTablesArgs(rule)
    
    cmd := exec.Command("iptables", args...)
    
    // ОБЯЗАТЕЛЬНАЯ обработка ошибок
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to add firewall rule: %w, stderr: %s", 
            err, stderr.String())
    }
    
    return nil
}

// КРИТИЧЕСКИ ВАЖНАЯ валидация правил
func validateFirewallRule(rule FirewallRule) error {
    // Валидация действия
    allowedActions := []string{"ACCEPT", "DROP", "REJECT"}
    if !contains(allowedActions, rule.Action) {
        return fmt.Errorf("invalid action: %s", rule.Action)
    }
    
    // Валидация протокола
    allowedProtocols := []string{"tcp", "udp", "icmp", "all"}
    if !contains(allowedProtocols, rule.Protocol) {
        return fmt.Errorf("invalid protocol: %s", rule.Protocol)
    }
    
    // Валидация портов
    if rule.Port != "" {
        port, err := strconv.Atoi(rule.Port)
        if err != nil || port < 1 || port > 65535 {
            return fmt.Errorf("invalid port: %s", rule.Port)
        }
    }
    
    // Валидация IP адресов
    if rule.Source != "" && rule.Source != "0.0.0.0/0" {
        if _, _, err := net.ParseCIDR(rule.Source); err != nil {
            if net.ParseIP(rule.Source) == nil {
                return fmt.Errorf("invalid source IP: %s", rule.Source)
            }
        }
    }
    
    return nil
}

// БЕЗОПАСНОЕ построение аргументов iptables
func buildSafeIPTablesArgs(rule FirewallRule) []string {
    args := []string{"-A", "INPUT"}
    
    if rule.Protocol != "all" {
        args = append(args, "-p", rule.Protocol)
    }
    
    if rule.Source != "" {
        args = append(args, "-s", rule.Source)
    }
    
    if rule.Port != "" && rule.Protocol != "icmp" {
        args = append(args, "--dport", rule.Port)
    }
    
    args = append(args, "-j", rule.Action)
    
    return args
}
```

### Netplan конфигурация - ОБЯЗАТЕЛЬНАЯ валидация
```go
// netplan.go - БЕЗОПАСНАЯ генерация конфигурации
func GenerateNetplanConfig(interfaces []NetworkInterface) error {
    // ОБЯЗАТЕЛЬНАЯ валидация входных данных
    for _, iface := range interfaces {
        if err := validateNetworkInterface(iface); err != nil {
            return fmt.Errorf("invalid interface %s: %w", iface.Name, err)
        }
    }
    
    // Генерация безопасной конфигурации
    config := NetplanConfig{
        Network: NetworkConfig{
            Version:   2,
            Renderer:  "networkd",
            Ethernets: make(map[string]EthernetConfig),
        },
    }
    
    for _, iface := range interfaces {
        config.Network.Ethernets[iface.Name] = EthernetConfig{
            DHCP4:     iface.DHCP,
            Addresses: []string{fmt.Sprintf("%s/%d", iface.IP, iface.Prefix)},
            Gateway4:  iface.Gateway,
            Nameservers: NameserversConfig{
                Addresses: iface.DNS,
            },
        }
    }
    
    // БЕЗОПАСНАЯ запись файла
    return writeNetplanConfigSecurely(config)
}

func writeNetplanConfigSecurely(config NetplanConfig) error {
    // Временный файл с правильными правами
    tmpFile, err := os.CreateTemp("/tmp", "netplan-*.yaml")
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    defer os.Remove(tmpFile.Name())
    
    // Установка правильных прав доступа
    if err := tmpFile.Chmod(0600); err != nil {
        return fmt.Errorf("failed to set file permissions: %w", err)
    }
    
    // Маршалинг в YAML
    data, err := yaml.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    // Запись данных
    if _, err := tmpFile.Write(data); err != nil {
        return fmt.Errorf("failed to write config: %w", err)
    }
    
    if err := tmpFile.Close(); err != nil {
        return fmt.Errorf("failed to close temp file: %w", err)
    }
    
    // Атомарное перемещение файла
    if err := os.Rename(tmpFile.Name(), "/etc/netplan/99-routersbc.yaml"); err != nil {
        return fmt.Errorf("failed to move config file: %w", err)
    }
    
    return nil
}
```

## Системные конфигураторы (configurators/sys/)

### ОБЯЗАТЕЛЬНАЯ безопасность сервисов
```go
// services.go - БЕЗОПАСНОЕ управление systemd
func ManageService(name, action string) error {
    // ОБЯЗАТЕЛЬНАЯ валидация имени сервиса
    if !isValidServiceName(name) {
        return fmt.Errorf("invalid service name: %s", name)
    }
    
    // ТОЛЬКО разрешенные действия
    allowedActions := []string{"start", "stop", "restart", "enable", "disable", "status"}
    if !contains(allowedActions, action) {
        return fmt.Errorf("invalid action: %s", action)
    }
    
    // БЕЗОПАСНАЯ команда systemctl
    cmd := exec.Command("systemctl", action, name)
    
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to %s service %s: %w, stderr: %s", 
            action, name, err, stderr.String())
    }
    
    return nil
}

// ОБЯЗАТЕЛЬНАЯ валидация имени сервиса
func isValidServiceName(name string) bool {
    if len(name) == 0 || len(name) > 64 {
        return false
    }
    
    // Только разрешенные символы для systemd
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9\-_.]*[a-zA-Z0-9]$`, name)
    return matched
}
```

## Хеширование паролей (hash.go)

### ОБЯЗАТЕЛЬНАЯ криптографическая безопасность
```go
// ИСПОЛЬЗОВАТЬ ТОЛЬКО bcrypt или argon2
func HashPassword(password string) (string, error) {
    // МИНИМАЛЬНАЯ сложность bcrypt
    const minCost = 12
    
    // Валидация пароля
    if len(password) < 8 {
        return "", errors.New("password too short")
    }
    
    if len(password) > 128 {
        return "", errors.New("password too long")
    }
    
    hash, err := bcrypt.GenerateFromPassword([]byte(password), minCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    
    return string(hash), nil
}

func VerifyPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

## ЗАПРЕЩЕННЫЕ практики:
- ❌ Команды через shell без валидации
- ❌ Прямая конкатенация в exec.Command
- ❌ Файлы с неправильными правами доступа
- ❌ MD5/SHA1 для паролей
- ❌ Неэкранированные пользовательские данные в командах

## ОБЯЗАТЕЛЬНЫЕ практики:
- ✅ Валидация ВСЕХ пользовательских данных
- ✅ Использование exec.Command с отдельными аргументами
- ✅ Проверка существования файлов/интерфейсов
- ✅ Правильные права доступа (600/644)
- ✅ bcrypt для хеширования паролей
