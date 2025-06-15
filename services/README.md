
# Services Module - Бизнес-логика

## Сетевые сервисы (network.go)

### ОБЯЗАТЕЛЬНАЯ обработка команд системы
```go
func (s *NetworkService) GetInterfaces() ([]NetworkInterface, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "ip", "addr", "show")
    output, err := cmd.Output()
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            return nil, fmt.Errorf("ip command failed with exit code %d: %s", 
                exitErr.ExitCode(), string(exitErr.Stderr))
        }
        return nil, fmt.Errorf("failed to execute ip command: %w", err)
    }
    
    interfaces, err := s.parseInterfaceOutput(string(output))
    if err != nil {
        return nil, fmt.Errorf("failed to parse interface output: %w", err)
    }
    
    return interfaces, nil
}

// ОБЯЗАТЕЛЬНАЯ валидация перед системными вызовами
func (s *NetworkService) UpdateInterface(name string, config InterfaceConfig) error {
    // Валидация имени интерфейса
    if !s.validateInterfaceName(name) {
        return errors.New("invalid interface name")
    }
    
    // Валидация IP адреса
    if config.IPAddress != "" {
        if net.ParseIP(config.IPAddress) == nil {
            return errors.New("invalid IP address")
        }
    }
    
    // Проверка существования интерфейса
    exists, err := s.interfaceExists(name)
    if err != nil {
        return fmt.Errorf("failed to check interface existence: %w", err)
    }
    if !exists {
        return fmt.Errorf("interface %s does not exist", name)
    }
    
    // Выполнение команд с валидацией
    commands := s.buildInterfaceCommands(name, config)
    for _, cmd := range commands {
        if err := s.executeSecureCommand(cmd); err != nil {
            return fmt.Errorf("failed to execute command %v: %w", cmd, err)
        }
    }
    
    return nil
}

// Безопасное выполнение системных команд
func (s *NetworkService) executeSecureCommand(args []string) error {
    if len(args) == 0 {
        return errors.New("empty command")
    }
    
    // Валидация команды
    if !s.isAllowedCommand(args[0]) {
        return fmt.Errorf("command not allowed: %s", args[0])
    }
    
    // Экранирование аргументов
    for i, arg := range args {
        args[i] = s.sanitizeArgument(arg)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, args[0], args[1:]...)
    
    // Захват stderr для диагностики
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
    }
    
    return nil
}
```

## Системные сервисы (system.go)

### ОБЯЗАТЕЛЬНАЯ обработка USB устройств
```go
func (s *SystemService) GetUSBDevices() ([]USBDevice, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, "lsusb")
    output, err := cmd.Output()
    if err != nil {
        // НЕ игнорировать ошибку, а обработать
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            if exitErr.ExitCode() == 127 {
                return nil, errors.New("lsusb command not found - install usbutils package")
            }
            return nil, fmt.Errorf("lsusb failed: %s", string(exitErr.Stderr))
        }
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, errors.New("lsusb command timed out")
        }
        return nil, fmt.Errorf("failed to execute lsusb: %w", err)
    }
    
    devices, err := s.parseUSBOutput(string(output))
    if err != nil {
        return nil, fmt.Errorf("failed to parse USB output: %w", err)
    }
    
    return devices, nil
}

// ОБЯЗАТЕЛЬНАЯ проверка драйверов
func (s *SystemService) getDeviceDriver(vendorID, productID string) (string, error) {
    // Проверка в /sys/bus/usb/devices/
    pattern := fmt.Sprintf("/sys/bus/usb/devices/*")
    matches, err := filepath.Glob(pattern)
    if err != nil {
        return "unknown", fmt.Errorf("failed to glob USB devices: %w", err)
    }
    
    for _, devicePath := range matches {
        driver, err := s.readDeviceDriver(devicePath, vendorID, productID)
        if err != nil {
            // Логирование, но продолжение поиска
            log.Printf("Error reading driver for %s: %v", devicePath, err)
            continue
        }
        if driver != "" {
            return driver, nil
        }
    }
    
    return "unknown", nil
}

func (s *SystemService) readDeviceDriver(devicePath, vendorID, productID string) (string, error) {
    // Чтение vendor ID
    vendorPath := filepath.Join(devicePath, "idVendor")
    vendorData, err := os.ReadFile(vendorPath)
    if err != nil {
        return "", err
    }
    
    // Чтение product ID
    productPath := filepath.Join(devicePath, "idProduct")
    productData, err := os.ReadFile(productPath)
    if err != nil {
        return "", err
    }
    
    // Сравнение ID
    if strings.TrimSpace(string(vendorData)) == vendorID && 
       strings.TrimSpace(string(productData)) == productID {
        
        // Чтение драйвера
        driverPath := filepath.Join(devicePath, "driver", "driver")
        if _, err := os.Stat(driverPath); err == nil {
            driverData, err := os.ReadFile(driverPath)
            if err != nil {
                return "", err
            }
            return strings.TrimSpace(string(driverData)), nil
        }
    }
    
    return "", nil
}
```

## ЗАПРЕЩЕННЫЕ практики:
- ❌ `cmd.Output()` без обработки ошибок
- ❌ Системные команды без таймаутов
- ❌ Игнорирование exit кодов
- ❌ Отсутствие валидации входных данных
- ❌ Неэкранированные аргументы команд

## ОБЯЗАТЕЛЬНЫЕ практики:
- ✅ context.WithTimeout для всех команд
- ✅ Валидация всех входных параметров
- ✅ Проверка существования файлов/интерфейсов
- ✅ Обработка всех типов ошибок (ExitError, timeout, etc.)
- ✅ Экранирование аргументов команд
