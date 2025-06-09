# RouterSBC Project Structure

## ВАЖНОЕ ЗАМЕЧАНИЕ ОБ АРХИТЕКТУРЕ ИНТЕРФЕЙСА

### Обязательная система базовых шаблонов
Все модули ДОЛЖНЫ использовать базовые шаблоны для обеспечения единообразия интерфейса:

**Базовые компоненты (обязательные):**
- `core/templates/base/layout.html` - основной layout с подключением CSS/JS
- `core/templates/base/header.html` - заголовок сайта с навигацией
- `core/templates/base/footer.html` - подвал с системной информацией
- `core/templates/base/sidebar.html` - боковая панель навигации
- `core/templates/base/nav_menu.html` - динамическое меню

### Система хуков навигации
**НЕ ИЗМЕНЯЙТЕ БАЗОВЫЕ ШАБЛОНЫ!** Используйте систему хуков:

1. **Регистрация модуля через module.json:**
```json
{
  "name": "network",
  "display_name": "Network Settings",
  "menu_icon": "network-wired",
  "menu_order": 2,
  "requires_hardware": ["ethernet"],
  "menu_section": "configuration"
}
```

2. **nav_menu.html автоматически включает активные модули** на основе:
   - Обнаруженного оборудования
   - Конфигурации модуля
   - Системных зависимостей

3. **Компоненты UI (используйте готовые):**
   - `{{template "card" .CardData}}` - карточки
   - `{{template "modal" .ModalData}}` - модальные окна
   - `{{template "tabs" .TabsData}}` - табы
   - `{{template "form" .FormData}}` - формы

---

## Структура проекта

```
routersbc/
├── main.go                      # Entry point, Echo server initialization
├── go.mod                       # Go module dependencies
├── go.sum                       # Dependency checksums
├── config.json                  # System configuration
│
├── core/                        # Core system components
│   ├── config/                  # Configuration management
│   │   ├── config.go           # Configuration structures
│   │   ├── manager.go          # Configuration manager
│   │   └── validator.go        # Configuration validation
│   │
│   ├── hardware/               # Hardware detection and management
│   │   ├── detector.go         # Hardware detection implementation
│   │   └── capabilities.go     # System capabilities detection
│   │
│   ├── middleware/             # Echo middleware
│   │   └── auth.go            # Authentication middleware
│   │
│   ├── modules/               # Module management system
│   │   └── base.go           # Base module interface and loader
│   │
│   ├── tasks/                 # Task management system
│   │   └── manager.go        # Task queue and execution
│   │
│   ├── templates/             # Base template system (НЕ ИЗМЕНЯТЬ!)
│   │   ├── base/             # Core layout templates
│   │   │   ├── layout.html   # Main layout wrapper
│   │   │   ├── header.html   # Common header component
│   │   │   ├── footer.html   # Common footer component
│   │   │   ├── sidebar.html  # Navigation sidebar
│   │   │   └── nav_menu.html # Dynamic navigation menu (HOOKS!)
│   │   └── components/       # Reusable UI components
│   │       ├── modal.html    # Modal dialog component
│   │       ├── card.html     # Card component
│   │       ├── tabs.html     # Tab navigation component
│   │       └── form.html     # Form components
│   │
│   └── system_integration.go  # System integration utilities
│
├── modules/                    # Feature modules (auto-loaded)
│   │
│   ├── login/                 # Authentication module
│   │   ├── handlers.go       # Login/logout handlers
│   │   ├── module.json       # Module metadata
│   │   └── templates/        # Login-specific templates
│   │       └── index.html    # Login page
│   │
│   ├── dashboard/            # Dashboard - Панель управления
│   │   ├── handlers.go      # Dashboard handlers
│   │   ├── module.json      # Module metadata and menu hooks
│   │   └── templates/       # Dashboard templates
│   │       └── index.html   # Main dashboard with system overview
│   │
│   ├── network/             # Network Settings - Настройки сети
│   │   ├── handlers.go     # Main network handlers
│   │   ├── hardware_info.go # 1. Информация о сетевом оборудовании
│   │   ├── lan_handler.go  # 2. Настройки LAN (DHCP/DNS/VLAN/Bridge)
│   │   ├── wan_handler.go  # 3. Настройки WAN (Interface/Connection/MAC)
│   │   ├── multiwan_handler.go # 3.1. Multi-WAN/Load Balancing
│   │   ├── dependency_validator.go # Configuration dependencies
│   │   ├── module.json     # Module metadata and menu hooks
│   │   └── templates/      # Network configuration templates
│   │       ├── index.html  # Network overview with hardware status
│   │       ├── hardware.html # 1. Network hardware information
│   │       ├── lan.html    # 2. LAN settings interface
│   │       │               #    - DHCP (PROXY|RELAY|SERVER)
│   │       │               #    - DNS (AUTO|PROXY|IP SERVERS)
│   │       │               #    - VLAN/Bridge (Ethernet + Wireless)
│   │       ├── wan.html    # 3. WAN configuration interface
│   │       │               #    - Interface selection (eth0, wlan0)
│   │       │               #    - Connection type (DHCP/Static/PPPoE)
│   │       │               #    - MAC address modification
│   │       ├── multiwan.html # 3.1. Multi-WAN configuration
│   │       │               #    - WAN selection (multiple required)
│   │       │               #    - Mode (Failover/Load Balance)
│   │       │               #    - Priority and traffic distribution
│   │       └── wireless_client.html # Wireless client for WAN dependency
│   │
│   ├── wireless/           # Wireless Settings - Настройки беспроводной сети
│   │   ├── handlers.go    # Main wireless handlers
│   │   ├── adapter_manager.go # Wireless adapter management
│   │   ├── scanner.go     # WiFi scanning and discovery
│   │   ├── mode_handler.go # Wireless modes (AP/Managed/Monitor)
│   │   ├── security_handler.go # Security configuration (WPA/WPA2/WPA3)
│   │   ├── frequency_manager.go # Frequency and channel management
│   │   ├── module.json    # Module metadata (requires wireless hardware)
│   │   └── templates/     # Wireless configuration templates
│   │       ├── index.html # 4. Wireless overview with adapter status
│   │       ├── adapters.html # Interface/adapter display and management
│   │       ├── create_network.html # Network/connection creation wizard
│   │       │               #    - Interface/adapter selection
│   │       │               #    - Mode selection (AP/Managed/Monitor)
│   │       │               #    - Frequency (2.4/5/6 GHz - hardware dependent)
│   │       │               #    - Channel selection (1,2,3... - frequency dependent)
│   │       │               #    - Security (WPA/WPA2/WPA3)
│   │       ├── ap_config.html # Access Point configuration
│   │       ├── client_config.html # Client mode configuration
│   │       └── frequency_config.html # Advanced frequency/channel settings
│   │
│   ├── tasks/             # Task Manager - Диспетчер задач (Configuration Buffer)
│   │   ├── handlers.go   # Task management handlers
│   │   ├── task_executor.go # Task execution engine
│   │   ├── config_buffer.go # Configuration change buffer
│   │   ├── dependency_resolver.go # Task dependency resolution
│   │   ├── conflict_detector.go # Configuration conflict detection
│   │   ├── module.json   # Module metadata
│   │   └── templates/    # Task management templates
│   │       ├── index.html # Task queue with pending changes
│   │       ├── pending.html # Pending configuration changes review
│   │       ├── conflicts.html # Configuration conflicts resolution
│   │       └── history.html # Configuration change history
│   │
│   ├── firewall/          # Firewall configuration
│   │   ├── handlers.go   # Firewall handlers
│   │   ├── module.json   # Module metadata
│   │   └── templates/    # Firewall templates
│   │       └── index.html # Firewall rules management
│   │
│   ├── system/           # System management
│   │   ├── handlers.go  # System handlers
│   │   ├── module.json  # Module metadata
│   │   └── templates/   # System templates
│   │       └── index.html # System status and settings
│   │
│   └── services/         # Service management
│       ├── handlers.go  # Service handlers
│       ├── module.json  # Module metadata
│       └── templates/   # Service templates
│           └── index.html # Service status and control
│
├── static/               # Static assets
│   ├── css/             # Stylesheets
│   │   ├── vendor/     # Third-party CSS
│   │   │   ├── tailwind.min.css
│   │   │   ├── flyonui.min.css
│   │   │   └── routersbc.css
│   │   ├── adaptive_ui.css # Responsive design
│   │   └── routersbc.css # Custom styles
│   │
│   └── js/             # JavaScript files
│       ├── vendor/    # Third-party JS
│       ├── routersbc.js # Main application JS
│       ├── hardware_detection.js # Hardware detection
│       ├── real_time_validation.js # Real-time form validation
│       ├── dependency_checker.js # Configuration dependencies
│       └── task_manager.js # Task management
│
├── configs/            # Configuration storage
│   ├── network.json   # Network configurations
│   ├── wireless.json  # Wireless configurations
│   ├── firewall.json  # Firewall rules
│   └── backups/       # Configuration backups
│
└── devdocs/           # Development documentation
    ├── CORE_TECHNICAL_TASK.md # Core implementation guide
    ├── GENERAL_MODULES_TASKS.md # Module implementation guide
    ├── DEPLOYMENT_GUIDE.md # Deployment instructions
    ├── DEV_PLAN.md     # Development roadmap
    └── TECHNICAL_TASK.md # Technical specifications
```

## Архитектура модулей

### 1. Автоматическая генерация маршрутов
- `modules/network/templates/index.html` → маршрут `/network`
- `modules/wireless/templates/adapters.html` → маршрут `/wireless/adapters`
- Echo автоматически регистрирует маршруты на основе структуры templates

### 2. Структура модуля (стандарт)
```
modules/{module_name}/
├── module.json          # Конфигурация и зависимости от оборудования
├── handlers.go          # Echo handlers для маршрутов
├── {feature}_handler.go # Специализированные обработчики
└── templates/           # HTML шаблоны модуля
    ├── index.html      # Главная страница модуля
    └── {feature}.html  # Функциональные страницы
```

### 3. module.json (обязательная структура)
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
      "requires": ["wireless_client_connection"]
    }
  }
}
```

## Система зависимостей конфигурации

### 1. Аппаратные зависимости
- **WiFi модуль**: активен только при обнаружении беспроводных адаптеров
- **Multi-WAN**: требует несколько сетевых интерфейсов
- **Частоты беспроводной сети**: зависят от возможностей адаптера

### 2. Конфигурационные зависимости
- **Беспроводной WAN**: требует активного подключения в режиме клиента
- **DHCP сервер**: конфликтует с DHCP клиентом на том же интерфейсе
- **Bridge конфигурация**: требует совместимых интерфейсов

### 3. Диспетчер задач (Configuration Buffer)
- Все изменения настроек добавляются в очередь задач
- Пользователь может просмотреть, удалить или применить изменения
- Автоматическое разрешение зависимостей между задачами
- Обнаружение конфликтов конфигурации

## Template Architecture Features

### 1. Echo Framework Integration
- Main server uses Echo for routing and middleware
- All modules use Echo handlers for request processing
- Automatic route registration from template structure

### 2. Hardware-Driven Module Loading
- Modules activate based on detected hardware capabilities
- nav_menu.html dynamically includes available modules
- No interface elements for unavailable hardware

### 3. Configuration Dependencies
- Real-time validation of configuration changes
- Dependency conflicts highlighted to user
- Hardware capability checking before configuration

### 4. Task Management System
- Configuration buffer prevents immediate application of changes
- Selective application of configuration changes
- Rollback capability with configuration backups
- Dependency resolution and conflict detection
