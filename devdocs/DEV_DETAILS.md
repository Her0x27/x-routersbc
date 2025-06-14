## **ДЕТАЛИЗИРОВАННАЯ СПЕЦИФИКАЦИЯ ПРОЕКТА**

### **1. СТРУКТУРА ФАЙЛОВ И ДИРЕКТОРИЙ:**

```bash
project-root/
├── .replit                          # Конфигурация Replit
├── .backup/                         # Автоматические backup'ы
├── workflows/                       # CI/CD конфигурации
│   ├── debug.yml
│   └── release.yml
├── dev/                            # Документация разработки
│   ├── RULES.md
│   ├── DEV_TASKS.md
│   ├── CHANGELOG.md
│   └── reports/
├── scripts/                        # Утилиты и валидаторы
│   ├── code_validator.sh
│   ├── url_checker.sh
│   ├── backup_manager.sh
│   └── startup_reminder.sh
├── cmd/                           # Точки входа приложения
│   └── server/
│       └── main.go
├── internal/                      # Внутренняя логика
│   ├── config/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── services/
│   └── utils/
├── pkg/                          # Публичные пакеты
├── templates/                    # HTML шаблоны
│   ├── base/
│   ├── components/
│   └── modules/
├── static/                       # Статические файлы
│   ├── css/
│   ├── js/
│   └── assets/
└── tests/                        # Тесты
```

### **2. ДЕТАЛИЗАЦИЯ BACKEND АРХИТЕКТУРЫ:**

#### **2.1 Основной сервер (cmd/server/main.go):**
```go:cmd/server/main.go
// Инициализация Echo сервера с HTTP/2
// Подключение middleware (логирование, CORS, аутентификация)
// Регистрация маршрутов
// Настройка WebSocket соединений
// Graceful shutdown
```

#### **2.2 Структура handlers:**
```bash
internal/handlers/
├── dashboard.go          # GET /dashboard
├── system/
│   ├── settings.go       # GET/POST /system/settings
│   ├── hardware.go       # GET /system/hardware
│   └── devices.go        # GET /system/devices_detected
├── network/
│   ├── overview.go       # GET /network/
│   ├── wan.go           # GET/POST /network/wans
│   ├── lan.go           # GET/POST /network/lan
│   ├── firewall.go      # GET/POST /network/firewall
│   └── wireless.go      # GET/POST /network/wireless
└── api/                 # REST API endpoints
    ├── v1/
    └── websocket.go
```

#### **2.3 Модели данных (internal/models/):**
```go:internal/models/network.go
type NetworkInterface struct {
    Name        string    `json:"name" db:"name"`
    Type        string    `json:"type" db:"type"` // wan, lan, wireless
    Status      string    `json:"status" db:"status"`
    IPAddress   string    `json:"ip_address" db:"ip_address"`
    Netmask     string    `json:"netmask" db:"netmask"`
    Gateway     string    `json:"gateway" db:"gateway"`
    DNSServers  []string  `json:"dns_servers" db:"dns_servers"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type DHCPConfig struct {
    Mode         string `json:"mode" db:"mode"` // RELAY|SERVER|OFF
    StartIP      string `json:"start_ip" db:"start_ip"`
    EndIP        string `json:"end_ip" db:"end_ip"`
    LeaseTime    int    `json:"lease_time" db:"lease_time"`
    DefaultGW    string `json:"default_gw" db:"default_gw"`
}
```

### **3. ДЕТАЛИЗАЦИЯ FRONTEND ШАБЛОНОВ:**

#### **3.1 Базовые шаблоны (templates/base/):**
```html:templates/base/layout.html
<!DOCTYPE html>
<html lang="ru">
<head>
    {{template "head" .}}
</head>
<body class="sidebar-mini layout-fixed">
    <div class="wrapper">
        {{template "header" .}}
        {{template "sidebar" .}}
        
        <div class="content-wrapper">
            {{template "pageloader" .}}
            <div class="content">
                {{block "content" .}}{{end}}
            </div>
        </div>
        
        {{template "footer" .}}
    </div>
    {{template "scripts" .}}
</body>
</html>
```

#### **3.2 Компоненты (templates/components/):**
```html:templates/components/form.html
{{define "input-text"}}
<div class="form-group">
    <label for="{{.ID}}">{{.Label}}</label>
    <input type="text" 
           class="form-control {{if .Error}}is-invalid{{end}}" 
           id="{{.ID}}" 
           name="{{.Name}}" 
           value="{{.Value}}"
           {{if .Required}}required{{end}}
           {{if .Placeholder}}placeholder="{{.Placeholder}}"{{end}}>
    {{if .Error}}
    <div class="invalid-feedback">{{.Error}}</div>
    {{end}}
    {{if .Help}}
    <small class="form-text text-muted">{{.Help}}</small>
    {{end}}
</div>
{{end}}
```

### **4. СИСТЕМА КОНТРОЛЯ КАЧЕСТВА:**

#### **4.1 Валидатор кода (scripts/code_validator.sh):**
```bash:scripts/code_validator.sh
#!/bin/bash
# Проверка на упрощенный код и игнорирование ошибок
# Поиск паттернов: _, err := и отсутствие проверки err
# Проверка на TODO/FIXME комментарии
# Валидация форматирования (gofmt, golint)
# Проверка покрытия тестами
# Анализ цикломатической сложности
```

#### **4.2 Проверка URL (scripts/url_checker.sh):**
```bash:scripts/url_checker.sh
#!/bin/bash
# Парсинг всех внутренних URL из шаблонов и кода
# Проверка доступности каждого endpoint'а
# Валидация HTTP статус кодов
# Проверка содержимого ответов на корректность
# Генерация отчета о недоступных/некорректных URL
```

### **5. ДЕТАЛИЗАЦИЯ МОДУЛЕЙ:**

#### **5.1 Dashboard Module:**
```go:internal/handlers/dashboard.go
// GET /dashboard
func (h *DashboardHandler) Index(c echo.Context) error {
    data := struct {
        Title           string
        NetworkMap      *NetworkTopology
        WANInterfaces   []NetworkInterface
        WirelessNetworks []WirelessNetwork
        DHCPClients     []DHCPClient
        SystemStats     *SystemStatistics
    }{
        Title: "Панель управления",
        // ... заполнение данных из сервисов
    }
    
    return c.Render(http.StatusOK, "content_dashboard", data)
}
```

```html:templates/dashboard/index.html
{{define "content_dashboard"}}
<div class="content-header">
    <h1>{{.Title}}</h1>
</div>

<div class="row">
    <!-- Карта сети -->
    <div class="col-md-8">
        <div class="card">
            <div class="card-header">
                <h3 class="card-title">Топология сети</h3>
            </div>
            <div class="card-body">
                <div id="network-map" style="height: 400px;"></div>
            </div>
        </div>
    </div>
    
    <!-- Статистика -->
    <div class="col-md-4">
        {{template "wan-status-widget" .WANInterfaces}}
        {{template "wireless-status-widget" .WirelessNetworks}}
        {{template "dhcp-clients-widget" .DHCPClients}}
    </div>
</div>

<script>
// Инициализация jsVectorMap
document.addEventListener('DOMContentLoaded', function() {
    const networkMap = new jsVectorMap({
        selector: '#network-map',
        map: 'custom_network_topology',
        backgroundColor: 'transparent',
        // ... конфигурация карты
    });
});
</script>
{{end}}
```

#### **5.2 System Settings Module:**
```go:internal/handlers/system/settings.go
type SettingsData struct {
    PasswordForm    *PasswordChangeForm
    NTPConfig       *NTPConfiguration
    BackupList      []BackupFile
    SystemInfo      *SystemInformation
    ValidationErrors map[string]string
}

func (h *SettingsHandler) Index(c echo.Context) error {
    // Получение текущих настроек системы
    // Подготовка данных для шаблона
    // Обработка flash сообщений
}

func (h *SettingsHandler) ChangePassword(c echo.Context) error {
    // Валидация формы смены пароля
    // Проверка текущего пароля
    // Хеширование нового пароля
    // Сохранение в базу данных
    // Логирование действия
}
```

### **6. СИСТЕМА BACKUP И ВЕРСИОНИРОВАНИЯ:**

#### **6.1 Менеджер backup'ов (scripts/backup_manager.sh):**
```bash:scripts/backup_manager.sh
#!/bin/bash

create_backup() {
    local file_path="$1"
    local reason="$2"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_dir=".backup/$(dirname "$file_path")"
    local backup_file="$backup_dir/$(basename "$file_path")_${timestamp}.bak"
    
    mkdir -p "$backup_dir"
    cp "$file_path" "$backup_file"
    
    # Создание метаданных backup'а
    cat > "${backup_file}.meta" << EOF
original_file: $file_path
backup_time: $(date -Iseconds)
reason: $reason
user: $(whoami)
git_commit: $(git rev-parse HEAD 2>/dev/null || echo "N/A")
EOF
    
    echo "Backup created: $backup_file"
}
```

### **7. СИСТЕМА НАПОМИНАНИЙ:**

#### **7.1 Startup reminder (scripts/startup_reminder.sh):**
```bash:scripts/startup_reminder.sh
#!/bin/bash

show_development_rules() {
    echo "=================================="
    echo "🚨 ПРАВИЛА РАЗРАБОТКИ 🚨"
    echo "=================================="
    echo "1. Перед изменением кода создавай backup"
    echo "2. Пиши подробные комментарии"
    echo "3. Запрещены фиктивные данные"
    echo "4. Обязательная обработка ошибок"
    echo "5. Ведение подробных reports"
    echo "=================================="
    echo "Документация: dev/RULES.md"
    echo "Задачи: dev/DEV_TASKS.md"
    echo "=================================="
    
    read -p "Нажми Enter для продолжения..." -r
}

# Интеграция в main.go
```

### **8. КОНФИГУРАЦИЯ REPLIT:**

```yaml:.replit
language = "go"
modules = ["go-1.23"]

[nix]
channel = "stable-24.05"

[deployment]
run = ["sh", "-c", "scripts/startup_reminder.sh && go run cmd/server/main.go"]
deploymentTarget = "cloudrun"

[[ports]]
localPort = 8080
externalPort = 80

[env]
GO_ENV = "development"
LOG_LEVEL = "debug"
```

### **9. ДОПОЛНИТЕЛЬНЫЕ ТРЕБОВАНИЯ К ТЕСТИРОВАНИЮ:**

```go:tests/handlers/dashboard_test.go
func TestDashboardHandler_Index(t *testing.T) {
    // Тестирование рендеринга страницы
    // Проверка корректности данных
    // Валидация HTML структуры
    // Тестирование WebSocket соединений
}
```

### **10. ЛОГИРОВАНИЕ И МОНИТОРИНГ:**

```go:internal/middleware/logging.go
// Структурированное логирование всех запросов
// Метрики производительности
// Отслеживание ошибок
// Аудит изменений конфигурации
```

Эта детализация обеспечивает полное понимание архитектуры и требований к проекту, включая все аспекты разработки, тестирования и поддержки качества кода.
