# X-RouterSBC

> Программный роутер и SBC (Single Board Computer) контроллер на базе Go 1.23 + Echo Framework.

## Структура проекта

```
github.com/Her0x27/x-routersbc/
├── core/           # Веб-сервер (HTTP/2, WebSocket, Auth, Auto init/load)
├── handlers/       # HTTP обработчики запросов
├── services/       # Бизнес-логика и сервисы
├── routes/         # Маршрутизация URL
├── templates/      # HTML шаблоны (Tabler.io)
├── static/         # Статические файлы (CSS, JS, изображения)
├── utils/          # Вспомогательные инструменты и конфигураторы
└── scripts/        # Скрипты автоматизации
```

## Технические характеристики

- **Go версия**: 1.23+
- **Web Framework**: Echo Framework
- **База данных**: SQLite (routersbc.sqlitedb)
- **Порт**: 5000 (HTTP только)
- **Шаблоны**: Tabler.io
- **Аутентификация**: Сессии (24ч), по умолчанию sbc:sbc

## Функциональность

### Network Management
- Управление интерфейсами (Physical, VLAN, VPN)
- Настройка WAN/LAN
- WiFi управление (AP/STA/Monitor/AdHoc)
- Маршрутизация и Firewall (NFTables/IPTables)

### System Management
- Управление пользователями и паролями
- Мониторинг устройств
- Резервное копирование/восстановление
- Управление портативными устройствами

## Запуск

```bash
go mod tidy
go run main.go
```

Веб-интерфейс доступен по адресу: http://localhost:5000

## Безопасность

⚠️ **ВАЖНО**: После первого входа обязательно смените пароль по умолчанию!

# Core - Веб-сервер

Основной модуль веб-сервера с поддержкой HTTP/2, WebSocket, аутентификации и автоматической загрузки модулей.

## Функциональность

### Веб-сервер
- **HTTP/2** поддержка
- **WebSocket** для real-time обновлений
- Порт: **5000** (только HTTP)
- Автоматическая инициализация и загрузка handlers/routes/modules

### Аутентификация
- Сессионная аутентификация
- Хранение сессий в SQLite (routersbc.sqlitedb)
- Время жизни сессии: **24 часа**
- Аккаунт по умолчанию: `sbc:sbc` (хэшированный пароль)
- Алерт для смены пароля после первого входа

### База данных
- **SQLite**: routersbc.sqlitedb
- Хранение сессий пользователей
- Конфигурация системы

### Обработка ошибок
- Централизованная обработка ошибок/исключений
- Страницы ошибок: 404, 501
- Шаблоны: `templates/{404,501}.html`

## Структура файлов

```
core/
├── server.go      # Основной веб-сервер
├── auth.go        # Система аутентификации
├── database.go    # Работа с SQLite
├── websocket.go   # WebSocket обработчики
├── middleware.go  # Middleware для Echo
└── errors.go      # Централизованная обработка ошибок
```

## Требования

- Без использования фиктивных данных
- Обязательная обработка всех ошибок
- Только реальные данные в шаблонах

# Handlers - HTTP обработчики

HTTP обработчики для всех маршрутов приложения.

## Структура

### Network Handlers
```
handlers/
├── network_interfaces.go  # Управление сетевыми интерфейсами
├── network_wan.go         # WAN настройки
├── network_lan.go         # LAN и DHCP настройки
├── network_wireless.go    # WiFi управление
├── network_routing.go     # Маршрутизация и UPnP
├── network_firewall.go    # Firewall (NFTables/IPTables)
└── network_index.go       # Главная страница Network
```

### System Handlers
```
handlers/
├── system_general.go      # Общие настройки системы
├── system_devices.go      # Информация об устройствах
├── system_portable.go     # Портативные устройства
└── system_index.go        # Главная страница System
```

### Auth Handlers
```
handlers/
├── auth_login.go          # Авторизация
├── auth_logout.go         # Выход
└── auth_password.go       # Смена пароля
```

## Функциональность обработчиков

### Network
- **Interfaces**: Physical (eth, wlan, bt), VLAN (vlan, br), VPN (wg, awg, ppp, pptp, tun)
- **WAN**: General WAN, Multi WAN Load balancing, WAN Settings
- **LAN**: DHCP (Relay/Server/Disabled), DNS, Bridge Settings
- **Wireless**: AP, STA/MANAGED, ADHOC/IBSS, MONITOR
- **Routing**: Static Routes, UPnP IGD & PCP
- **Firewall**: General Rules, Chains, Rule Templates

### System
- Смена пароля
- NTPD (часовой пояс/дата время)
- Резервное копирование/восстановление
- Мониторинг устройств
- Управление портативными устройствами

## Модальные окна

Обработчики для модальных окон:
- STUN Server | UPnP
- Static Route
- Firewall Rules/Chains
- Network Interface создание/изменение
- DNS Local Zones/Resolvers/Routing
- WiFi Station/AP соединения

## Принципы

- Только реальные данные (без симуляции)
- Обязательная обработка ошибок
- WebSocket обновления интерфейса

# Services - Бизнес-логика

Сервисный слой содержит всю бизнес-логику приложения.

## Структура

### Network Services
```
services/
├── interface_service.go   # Управление сетевыми интерфейсами
├── wan_service.go         # WAN конфигурация
├── lan_service.go         # LAN и DHCP сервисы
├── wireless_service.go    # WiFi управление
├── routing_service.go     # Маршрутизация
├── firewall_service.go    # Firewall логика
└── dns_service.go         # DNS сервисы
```

### System Services
```
services/
├── system_service.go      # Системные операции
├── device_service.go      # Мониторинг устройств
├── backup_service.go      # Резервное копирование
├── user_service.go        # Управление пользователями
└── portable_service.go    # Портативные устройства
```

## Функциональность

### Network Services
- **Interface Management**: Создание/удаление/настройка интерфейсов
- **WAN Configuration**: Multi-WAN, Load balancing
- **LAN Services**: DHCP Server/Relay, Bridge настройки
- **Wireless**: AP/STA режимы, мониторинг
- **Routing**: Static routes, UPnP IGD, PCP
- **Firewall**: NFTables/IPTables управление
- **DNS**: Local zones, resolvers, geo routing

### System Services
- **Device Monitoring**: CPU, RAM, Storage, Network адаптеры
- **Backup/Restore**: Конфигурация системы
- **User Management**: Аутентификация, смена паролей
- **Portable Devices**: USB WiFi/Ethernet/Bluetooth/Modems/Storage/Webcam

## Интеграция с Utils

Сервисы используют конфигураторы из `/utils/configurators/`:
- Network конфигурация (netplan, interfaces)
- System конфигурация (kernel, services)
- Firewall правила

## Принципы

- Работа только с реальными данными
- Полная обработка ошибок
- Интеграция с системными утилитами
- Поддержка различных ОС (Armbian, Debian, Ubuntu)

# Routes - Маршрутизация

URL маршруты для всех страниц и API endpoints.

## Структура

```
routes/
├── network_routes.go      # Network маршруты
├── system_routes.go       # System маршруты
├── auth_routes.go         # Аутентификация
├── api_routes.go          # REST API v1
└── static_routes.go       # Статические файлы
```

## Маршруты

### Network Routes (`/network`)
```
GET  /network                    # Главная страница
GET  /network/interfaces         # Управление интерфейсами
GET  /network/wan               # WAN настройки
GET  /network/lan               # LAN настройки
GET  /network/wireless          # WiFi управление
GET  /network/routing           # Маршрутизация
GET  /network/firewall          # Firewall
```

### System Routes (`/system`)
```
GET  /system                    # Главная страница
GET  /system/general            # Общие настройки
GET  /system/about-devices      # Информация об устройствах
GET  /system/portable-devices   # Портативные устройства
```

### Auth Routes
```
GET  /login                     # Страница входа
POST /login                     # Обработка входа
GET  /logout                    # Выход
GET  /change-password           # Смена пароля
POST /change-password           # Обработка смены пароля
```

### API Routes (`/api/v1`)
```
# Network API
GET    /api/v1/interfaces       # Список интерфейсов
POST   /api/v1/interfaces       # Создание интерфейса
PUT    /api/v1/interfaces/:id   # Обновление интерфейса
DELETE /api/v1/interfaces/:id   # Удаление интерфейса

# System API
GET    /api/v1/system/info      # Информация о системе
GET    /api/v1/devices          # Список устройств
POST   /api/v1/backup           # Создание резервной копии
POST   /api/v1/restore          # Восстановление
```

### Error Routes
```
GET /404                        # Страница не найдена
GET /501                        # Не реализовано
```

## WebSocket Routes
```
GET /ws                         # WebSocket соединение
```

## Middleware

- **Auth Middleware**: Проверка аутентификации
- **CORS Middleware**: Cross-Origin запросы
- **Logger Middleware**: Логирование запросов
- **Error Middleware**: Обработка ошибок

## Принципы

- RESTful API дизайн
- Централизованная обработка ошибок
- WebSocket для real-time обновлений

# Templates - HTML шаблоны

HTML5 шаблоны на базе Tabler.io для веб-интерфейса.

## Структура

### Network Templates
```
templates/network/
├── index.html              # Главная страница Network
├── interfaces.html         # Управление интерфейсами
├── wan.html               # WAN настройки
├── lan.html               # LAN настройки
├── wireless.html          # WiFi управление
├── routing.html           # Маршрутизация
├── firewall.html          # Firewall (основной)
├── firewall_new.html      # NFTables интерфейс
├── firewall_classic.html  # IPTables интерфейс
└── modal.html             # Модальные окна
```

### System Templates
```
templates/system/
├── index.html             # Главная страница System
├── about-devices.html     # Информация об устройствах
└── portable-devices.html  # Портативные устройства
```

### Base Templates
```
templates/
├── base.html              # Базовый шаблон (Tabler.io)
├── login.html             # Страница входа
├── 404.html               # Страница не найдена
├── 501.html               # Не реализовано
└── change-password.html   # Смена пароля
```

## Функциональность шаблонов

### Network Templates

#### Interfaces (`interfaces.html`)
**Tabs:**
- **Physical**: eth, wlan, bt интерфейсы
- **VLAN**: vlan, br интерфейсы  
- **VPN**: wg, awg, ppp, pptp, tun соединения

#### WAN (`wan.html`)
**Tabs:**
- **General WAN**: Wire/Wireless выбор
- **Multi WAN**: Load balancing
- **WAN Settings**: Конфигурация

#### LAN (`lan.html`)
**Tabs:**
- **DHCP**: Relay/Server/Disabled
- **DNS**: Direct ISP/Proxy/Forward/Server, DoT/DoH, Local DNS Zones
- **Bridge Settings**: Настройки моста

#### Wireless (`wireless.html`)
**Tabs:**
- **AP**: Access Point режим
- **STA/MANAGED**: Station режим
- **ADHOC/IBSS**: Ad-hoc сети
- **MONITOR**: Мониторинг режим

#### Routing (`routing.html`)
**Tabs:**
- **Static Route**: interface, ip, mask, gateway, metric
- **UPnP IGD & PCP**: Port mapping, STUN, Traffic Shaping

#### Firewall (`firewall.html`)
Динамическая загрузка:
- `firewall_new.html` для NFTables
- `firewall_classic.html` для IPTables

**Tabs:**
- **General Rules**: Основные правила
- **Chains**: NAT, OUT/IN, FORWARD
- **Rule Templates**: allow/disallow/forwarding/redirect
- **Firewall Settings**: Core backend [nftables|iptables]

### System Templates

#### General (`system/index.html`)
- Смена пароля
- NTPD (часовой пояс/дата время)
- Создание резервной копии
- Восстановление из резервной копии

#### About Device (`system/about-devices.html`)
- Процессор (количество ядер)
- Оперативная память
- Хранилище (SSD, SD Card, eMMC)
- Сеть (Ethernet, WiFi, Bluetooth)
- Видео (если поддерживается)
- Аудио (если поддерживается)
- External I/O (USB Ports, SPI, UART, I2C, GPIO)

#### Portable Devices (`system/portable-devices.html`)
- Обнаружение устройств
- Информация о железе
- Статус драйверов
- Статус работы устройств
- Подсказки по установке драйверов
- **Поддерживаемые устройства**:
  - USB WiFi
  - USB Ethernet
  - USB Bluetooth
  - USB 3G/LTE Modem
  - USB Storage
  - USB Webcam

## Модальные окна (`modal.html`)

- **STUN Server | UPnP**: Добавление/изменение
- **Static Route**: interface, ip, mask, gateway, metric
- **Firewall Rules**: Добавление/изменение правил
- **Firewall Chains**: Управление цепочками
- **Network Interface**: Создание интерфейсов
  - Bridge / VLAN / VWLAN / HevSocks5Tunnel
  - **HevSocks5Tunnel**: Address, Port, Auth, UDP, buffer sizes, timeouts
- **DNS Local Zones**: root zone .local, router.local A 192.168.55.1
- **DNS Resolvers**: UDP/TCP/TLS/HTTPS resolvers
- **DNS Routing**: domain/geozone routing rules
- **WiFi Station**: Подключение к сетям
- **WiFi Access Point**: Создание точки доступа

## Технические требования

- **Фреймворк**: Tabler.io (единая основа)
- **Стандарты**: HTML5, CSS3, JavaScript
- **Данные**: Только реальные данные (без статических/фиктивных)
- **Обновления**: WebSocket для real-time интерфейса
- **Адаптивность**: Responsive дизайн

## Принципы

- Централизованная обработка ошибок
- Модульная структура шаблонов
- Переиспользование компонентов
- Accessibility поддержка

# Static - Статические файлы

Статические ресурсы для веб-интерфейса.

## Структура

```
static/
├── css/
│   ├── tabler.min.css         # Tabler.io основные стили
│   ├── tabler-icons.min.css   # Иконки Tabler
│   └── custom.css             # Кастомные стили
├── js/
│   ├── tabler.min.js          # Tabler.io JavaScript
│   ├── websocket.js           # WebSocket клиент
│   ├── network.js             # Network модуль JS
│   ├── system.js              # System модуль JS
│   └── modal.js               # Модальные окна
├── img/
│   ├── logo.png               # Логотип приложения
│   ├── icons/                 # Иконки устройств
│   └── backgrounds/           # Фоновые изображения
└── fonts/
    └── tabler-icons/          # Шрифты иконок
```

## CSS Файлы

### Tabler.io Framework
- **tabler.min.css**: Основные стили фреймворка
- **tabler-icons.min.css**: Набор иконок

### Кастомные стили
- **custom.css**: Дополнительные стили для специфичных компонентов

## JavaScript Модули

### Core JavaScript
- **tabler.min.js**: Основной функционал Tabler.io
- **websocket.js**: WebSocket соединение и обновления

### Модульные скрипты
- **network.js**: Функции для Network страниц
- **system.js**: Функции для System страниц  
- **modal.js**: Управление модальными окнами

## WebSocket Integration

```javascript
// websocket.js функциональность
- Real-time обновления интерфейсов
- Статус устройств
- Уведомления об изменениях
- Прогресс операций
```

## Изображения

### Иконки устройств
```
img/icons/
├── ethernet.svg
├── wifi.svg
├── bluetooth.svg
├── usb.svg
├── storage.svg
└── modem.svg
```

### Системные изображения
- Логотип приложения
- Фоновые изображения для страниц ошибок

## Принципы

- Минификация всех ресурсов
- Оптимизация изображений
- Кэширование статических файлов
- Responsive изображения

# Utils - Вспомогательные инструменты

Утилиты и конфигураторы для управления системой.

## Структура

```
utils/
├── configurators/
│   ├── net/
│   │   ├── netplan.go         # Netplan конфигурация
│   │   ├── interfaces.go      # /etc/network/interfaces
│   │   ├── wan.go             # WAN настройки
│   │   ├── dns.go             # DNS конфигурация
│   │   ├── dhcp.go            # DHCP сервер
│   │   ├── routing.go         # Маршрутизация
│   │   └── firewall.go        # Firewall правила
│   └── sys/
│       ├── kernel.go          # Kernel параметры
│       └── services.go        # Systemd сервисы
├── helpers/
│   ├── validation.go          # Валидация данных
│   ├── conversion.go          # Конвертация типов
│   └── filesystem.go          # Файловые операции
└── constants/
    ├── network.go             # Сетевые константы
    └── system.go              # Системные константы
```

## Network Configurators

### Netplan (`net/netplan.go`)
```go
// Поддержка ОС: Ubuntu 18+, современный Debian
- Генерация YAML конфигураций
- Управление интерфейсами
- WiFi настройки
- VLAN конфигурация
```

### Interfaces (`net/interfaces.go`)
```go
// Поддержка ОС: Debian, старые Ubuntu, Armbian
- Управление /etc/network/interfaces
- Статические/DHCP настройки
- Bridge конфигурация
- VLAN интерфейсы
```

### WAN Management (`net/wan.go`)
```go
- Multi-WAN конфигурация
- Load balancing
- Failover настройки
- Метрики маршрутов
```

### DNS Configuration (`net/dns.go`)
```go
- /etc/resolv.conf управление
- DNS серверы
- Local DNS zones
- DoT/DoH конфигурация
- Geo DNS routing
```

### DHCP Server (`net/dhcp.go`)
```go
- ISC DHCP Server конфигурация
- Статические резервации
- DHCP Relay настройки
- Lease управление
```

### Routing (`net/routing.go`)
```go
- Статические маршруты
- Policy routing
- UPnP IGD настройки
- PCP конфигурация
```

### Firewall (`net/firewall.go`)
```go
- NFTables правила (приоритет)
- IPTables правила (legacy)
- NAT конфигурация
- Port forwarding
- Traffic shaping
```

## System Configurators

### Kernel Settings (`sys/kernel.go`)
```go
- /proc/sys параметры
- sysctl конфигурация
- Network stack настройки
- Security параметры
```

### Service Management (`sys/services.go`)
```go
// Systemd операции
- start/stop/restart/reload
- enable/disable сервисов
- Статус мониторинг
- Dependency управление
```

## Helper Functions

### Validation (`helpers/validation.go`)
```go
- IP адрес валидация
- MAC адрес проверка
- Network range валидация
- Port range проверка
```

### Conversion (`helpers/conversion.go`)
```go
- IP/CIDR конвертация
- MAC адрес форматирование
- Bandwidth единицы
- Time duration parsing
```

### Filesystem (`helpers/filesystem.go`)
```go
- Безопасное чтение/запись файлов
- Backup/restore операции
- Права доступа управление
- Atomic file operations
```

## Поддерживаемые ОС

- **Armbian**: ARM-based SBC
- **Debian**: 10+ (Buster, Bullseye, Bookworm)
- **Ubuntu**: 18.04+ (Server/Desktop)

## Принципы

- Автоматическое определение ОС
- Graceful fallback между конфигураторами
- Atomic конфигурационные изменения
- Rollback при ошибках
- Полное логирование операций

# Scripts - Скрипты автоматизации

Скрипты для автоматизации развертывания, сборки и управления.

## Структура

```
scripts/
├── build/
│   ├── build.sh               # Основная сборка
│   ├── cross-compile.sh       # Кросс-компиляция
│   └── release.sh             # Подготовка релиза
├── deploy/
│   ├── install.sh             # Установка на целевую систему
│   ├── update.sh              # Обновление системы
│   └── uninstall.sh           # Удаление
├── dev/
│   ├── setup-dev.sh           # Настройка dev окружения
│   ├── run-tests.sh           # Запуск тестов
│   └── lint.sh                # Code linting
└── system/
    ├── backup-config.sh       # Резервное копирование
    ├── restore-config.sh      # Восстановление
    └── reset-password.sh      # Сброс пароля
```

## Build Scripts

### Main Build (`build/build.sh`)
```bash
#!/bin/bash
# Основная сборка приложения
- Go mod download
- Static assets embedding
- Binary compilation
- Version tagging
```

### Cross Compilation (`build/cross-compile.sh`)
```bash
#!/bin/bash
# Кросс-компиляция для различных архитектур
TARGETS:
- linux/amd64    # x86_64 системы
- linux/arm64    # ARM64 (Raspberry Pi 4, etc.)
- linux/arm      # ARM32 (Raspberry Pi 3, etc.)
- linux/mips     # MIPS роутеры
- linux/mipsle   # MIPS Little Endian
```

### Release Preparation (`build/release.sh`)
```bash
#!/bin/bash
# Подготовка релизных пакетов
- Binary compilation для всех архитектур
- Packaging (tar.gz, deb, rpm)
- Checksums генерация
- Release notes
```

## Deployment Scripts

### Installation (`deploy/install.sh`)
```bash
#!/bin/bash
# Установка на целевую систему
ПОДДЕРЖИВАЕМЫЕ ОС:
- Debian 10+ (Buster, Bullseye, Bookworm)
- Ubuntu 18.04+ (Server/Desktop)
- Armbian (ARM SBC)

ФУНКЦИИ:
- Dependency проверка и установка
- User/group создание
- Service файлы установка
- Database инициализация
- Firewall правила
- Auto-start настройка
```

### System Update (`deploy/update.sh`)
```bash
#!/bin/bash
# Обновление установленной системы
- Backup текущей конфигурации
- Binary замена
- Database migration
- Service restart
- Rollback при ошибках
```

### Uninstallation (`deploy/uninstall.sh`)
```bash
#!/bin/bash
# Полное удаление системы
- Service остановка и удаление
- User/group удаление
- Configuration cleanup
- Database backup (опционально)
- Log files cleanup
```

## Development Scripts

### Dev Environment (`dev/setup-dev.sh`)
```bash
#!/bin/bash
# Настройка development окружения
- Go 1.23+ установка
- Development dependencies
- Git hooks настройка
- IDE конфигурация
- Test database setup
```

### Testing (`dev/run-tests.sh`)
```bash
#!/bin/bash
# Запуск всех тестов
- Unit tests
- Integration tests
- Network configuration tests
- Database tests
- Coverage reports
```

### Code Quality (`dev/lint.sh`)
```bash
#!/bin/bash
# Code quality проверки
- go fmt
- go vet
- golangci-lint
- Security scanning
- Dependency vulnerability check
```

## System Management Scripts

### Configuration Backup (`system/backup-config.sh`)
```bash
#!/bin/bash
# Резервное копирование конфигурации
BACKUP INCLUDES:
- SQLite database
- Network configurations
- System settings
- User data
- Logs (последние 7 дней)

OUTPUT: timestamped tar.gz archive
```

### Configuration Restore (`system/restore-config.sh`)
```bash
#!/bin/bash
# Восстановление из резервной копии
- Validation backup archive
- Service остановка
- Configuration restore
- Database restore
- Service restart
- Verification checks
```

### Password Reset (`system/reset-password.sh`)
```bash
#!/bin/bash
# Сброс пароля администратора
- Service остановка
- Database password hash reset
- Default sbc:sbc restore
- Service restart
- Security warning log
```

## Использование

### Сборка проекта
```bash
./scripts/build/build.sh
```

### Установка на Debian/Ubuntu
```bash
sudo ./scripts/deploy/install.sh
```

### Development setup
```bash
./scripts/dev/setup-dev.sh
```

### Создание резервной копии
```bash
sudo ./scripts/system/backup-config.sh
```
