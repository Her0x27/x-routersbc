# X-RouterSBC
> Advanced Software-Based Controller (SBC) router solution built with Go 1.23 and Echo Framework.

## Project Structure

```
github.com/Her0x27/x-routersbc/
├── core/           # Web server core (HTTP/2, WebSocket, Auth, Auto-loading)
├── handlers/       # Request handlers
├── services/       # Business logic services
├── routes/         # URL routing definitions
├── templates/      # HTML templates (Tabler.io based)
├── static/         # Static assets (CSS, JS, images)
├── utils/          # Utility tools and helpers
└── scripts/        # Build and deployment scripts
```

## Features

- **Web Server**: HTTP/2 support on port 5000
- **Authentication**: Built-in auth system (admin:sbc/sbc)
- **Database**: SQLite (routersbc.sqlitedb)
- **Real-time Updates**: WebSocket support
- **REST API**: v1 API endpoints
- **Auto-loading**: Automatic handlers/routes/modules initialization

## Network Management

- Interface Management (Physical, VLAN, VPN)
- WAN Configuration (Single/Multi-WAN, Load Balancing)
- LAN Services (DHCP, DNS, Bridge)
- Wireless Management (AP, STA, ADHOC, MONITOR)
- Advanced Routing (Static Routes, UPnP IGD & PCP)
- Firewall (NFTables/IPTables support)

## Requirements

- Go 1.23+
- Linux-based OS (Armbian, Debian, Ubuntu)
- Network management privileges

## Quick Start

```bash
go mod tidy
go run main.go
```

Access the web interface at: http://localhost:5000

## Default Credentials

- Username: `sbc`
- Password: `sbc`

# Core Module
> Web server core module providing HTTP/2, WebSocket, authentication, and auto-initialization functionality.

## Features

- **HTTP/2 Server**: High-performance web server on port 5000
- **WebSocket Support**: Real-time communication for UI updates
- **Authentication System**: Built-in auth with SQLite backend
- **Auto-loading**: Automatic discovery and loading of:
  - Handlers
  - Routes
  - Modules
- **REST API v1**: RESTful API endpoints
- **Database**: SQLite integration (routersbc.sqlitedb)

## Components

- `server.go` - Main HTTP/2 server implementation
- `auth.go` - Authentication middleware and handlers
- `websocket.go` - WebSocket connection management
- `database.go` - SQLite database operations
- `loader.go` - Auto-loading system for handlers/routes
- `middleware.go` - Common middleware functions

## Configuration

```go
type ServerConfig struct {
    Port     int    `json:"port"`
    Host     string `json:"host"`
    Database string `json:"database"`
    Debug    bool   `json:"debug"`
}
```

## Default Settings

- Port: 5000
- Protocol: HTTP (no HTTPS support)
- Database: routersbc.sqlitedb
- Default Admin: sbc:sbc

# Handlers Module
> Request handlers for processing HTTP requests and WebSocket connections.

## Structure

- `network.go` - Network management handlers
- `interfaces.go` - Network interface operations
- `wan.go` - WAN configuration handlers
- `lan.go` - LAN services handlers
- `wireless.go` - Wireless management handlers
- `routing.go` - Routing configuration handlers
- `firewall.go` - Firewall management handlers
- `auth.go` - Authentication handlers
- `websocket.go` - WebSocket message handlers
- `api.go` - REST API v1 handlers

## Handler Types

### Page Handlers
- Render HTML templates
- Handle form submissions
- Manage user sessions

### API Handlers
- REST API endpoints
- JSON request/response
- CRUD operations

### WebSocket Handlers
- Real-time updates
- Configuration changes
- Status monitoring

## Modal Operations

Handlers support modal operations for:
- UPnP STUN Server management
- Static Route configuration
- Firewall Rule management
- Network Interface creation
- DNS Zone management
- WiFi connection setup

# Services Module
> Business logic services for network management and system operations.

## Network Services

- `interface_service.go` - Network interface management
- `wan_service.go` - WAN configuration service
- `lan_service.go` - LAN services management
- `wireless_service.go` - Wireless operations
- `routing_service.go` - Routing table management
- `firewall_service.go` - Firewall rule management
- `dns_service.go` - DNS configuration service
- `dhcp_service.go` - DHCP server management

## System Services

- `config_service.go` - Configuration management
- `monitor_service.go` - System monitoring
- `update_service.go` - Real-time updates via WebSocket

## Service Features

### Configuration Management
- Load/Save network configurations
- Validate configuration changes
- Apply configurations to system

### Real-time Monitoring
- Interface status monitoring
- Traffic statistics
- System resource usage

### Integration
- Works with utils/configurators
- Supports multiple OS distributions
- WebSocket notifications for changes

# Routes Module
> URL routing definitions for the web application.

## Route Files

- `network.go` - Network management routes
- `api.go` - REST API v1 routes
- `auth.go` - Authentication routes
- `static.go` - Static file serving
- `websocket.go` - WebSocket endpoints

## Route Groups

### Web Interface Routes
```
/                    - Dashboard
/network             - Network overview
/network/interfaces  - Interface management
/network/wan         - WAN configuration
/network/lan         - LAN services
/network/wireless    - Wireless management
/network/routing     - Routing configuration
/network/firewall    - Firewall management
```

### API Routes
```
/api/v1/interfaces   - Interface CRUD
/api/v1/wan          - WAN configuration API
/api/v1/lan          - LAN services API
/api/v1/wireless     - Wireless API
/api/v1/routing      - Routing API
/api/v1/firewall     - Firewall API
```

### WebSocket Routes
```
/ws                  - Main WebSocket endpoint
/ws/network          - Network updates
/ws/status           - System status updates
```

## Error Pages

- `/404` - Not Found (templates/404.html)
- `/501` - Not Implemented (templates/501.html)

# Templates Module
> HTML templates based on Tabler.io framework for consistent UI design.

## Structure

```
templates/
├── base.html           # Base template with Tabler.io
├── 404.html           # Not Found error page
├── 501.html           # Not Implemented error page
├── dashboard.html     # Main dashboard
├── login.html         # Authentication page
└── network/           # Network management templates
    ├── index.html     # Network overview
    ├── interfaces.html # Interface management
    ├── wan.html       # WAN configuration
    ├── lan.html       # LAN services
    ├── wireless.html  # Wireless management
    ├── routing.html   # Routing configuration
    ├── firewall.html  # Firewall management
    ├── firewall_new.html    # NFTables firewall
    ├── firewall_classic.html # IPTables firewall
    └── modal.html     # Modal dialogs
```

## Template Features

### Base Template (Tabler.io)
- Responsive design
- Modern UI components
- Consistent styling
- No static data

### Network Templates
- Tab-based navigation
- Dynamic content loading
- Form validation
- Real-time updates via WebSocket

### Modal Dialogs
- UPnP STUN Server configuration
- Static Route management
- Firewall Rule editor
- Network Interface creator
- DNS Zone editor
- WiFi connection setup

## Template Variables

Templates receive dynamic data from handlers:
- Network interface status
- Configuration settings
- System statistics
- Error messages
- Form validation results

# Static Assets
> Static files for the web interface including CSS, JavaScript, and images.

## Structure

```
static/
├── css/
│   ├── custom.css     # Custom styles
│   └── tabler.min.css # Tabler.io framework
├── js/
│   ├── app.js         # Main application JavaScript
│   ├── websocket.js   # WebSocket client
│   ├── network.js     # Network management functions
│   └── tabler.min.js  # Tabler.io JavaScript
├── img/
│   ├── logo.png       # Application logo
│   └── icons/         # UI icons
└── fonts/             # Web fonts
```

## JavaScript Modules

### WebSocket Client
- Real-time updates
- Connection management
- Event handling

### Network Management
- Form validation
- AJAX requests
- Modal dialogs
- Tab navigation

### UI Components
- Dynamic tables
- Progress indicators
- Notifications
- Form helpers

## CSS Customization

- Custom color scheme
- Network-specific styling
- Responsive adjustments
- Dark/light theme support

# Utils Module
> Utility tools and system configurators for network management.

## Structure

```
utils/
├── configurators/     # OS-specific configurators
│   ├── net/          # Network configuration
│   └── sys/          # System configuration
├── helpers/          # Helper functions
├── validators/       # Input validation
└── converters/       # Data conversion utilities
```

## Network Configurators

### Supported OS
- Armbian
- Debian  
- Ubuntu

### Network Modules
- `netplan.go` - Netplan configuration
- `interfaces.go` - Network interfaces
- `wan.go` - WAN configuration
- `dns.go` - DNS settings
- `dhcp.go` - DHCP server
- `routing.go` - Routing tables
- `firewall.go` - Firewall rules

### System Modules
- `kernel.go` - Kernel parameters
- `services.go` - System service management

## Service Operations

```go
// Service management
Start(service string) error
Stop(service string) error
Restart(service string) error
Reload(service string) error
Enable(service string) error
Disable(service string) error
```

## Configuration Management

- Backup existing configurations
- Validate new configurations
- Apply changes atomically
- Rollback on failure

# Scripts Module
> Build, deployment, and maintenance scripts for the x-routersbc project.

## Build Scripts

- `build.sh` - Build application for multiple architectures
- `build-arm.sh` - ARM-specific build (for SBC devices)
- `build-docker.sh` - Docker container build

## Deployment Scripts

- `deploy.sh` - Deploy to target system
- `install.sh` - System installation script
- `update.sh` - Application update script

## Development Scripts

- `dev.sh` - Development server with hot reload
- `test.sh` - Run test suite
- `lint.sh` - Code linting and formatting

## Maintenance Scripts

- `backup.sh` - Backup configuration and database
- `restore.sh` - Restore from backup
- `cleanup.sh` - Clean temporary files and logs

## System Scripts

- `setup-deps.sh` - Install system dependencies
- `setup-network.sh` - Configure network prerequisites
- `setup-firewall.sh` - Setup firewall prerequisites

## Usage Examples

```bash
# Build for ARM devices
./scripts/build-arm.sh

# Deploy to remote SBC
./scripts/deploy.sh user@192.168.1.1

# Start development server
./scripts/dev.sh

# Run tests
./scripts/test.sh
```
