# X-RouterSBC
> Advanced Software-Based Controller (SBC) for network routing and management.

## Project Structure

```
github.com/Her0x27/x-routersbc/
├── core/                 # Web server core (HTTP/2, WebSocket, Auth)
├── handlers/             # HTTP request handlers
├── services/             # Business logic services
├── routes/               # URL routing definitions
├── templates/            # HTML templates (Tabler.io based)
├── static/               # Static assets (CSS, JS, images)
├── utils/                # Utility tools and helpers
└── scripts/              # Build and deployment scripts
```

## Features

- **Web Server Core**: HTTP/2, WebSocket support, authentication
- **Auto-initialization**: Automatic loading of handlers, routes, and modules
- **REST API v1**: RESTful API endpoints
- **Database**: SQLite database (routersbc.sqlitedb)
- **Admin Access**: Default credentials (sbc:sbc)
- **Network Management**: Complete network configuration interface
- **Real-time Updates**: WebSocket-based interface updates

## Quick Start

1. Clone the repository
2. Build the application
3. Run with default configuration
4. Access web interface at http://localhost:8080
5. Login with admin credentials (sbc:sbc)

## Network Management Features

- Interface management (Physical, VPN, VLAN, ...)
- WAN configuration (Single/Multi WAN, Load balancing)
- LAN settings (DHCP, DNS, VLAN Bridge)
- Wireless management (AP, STA, ADHOC, Monitor)
- Static routing and UPnP
- Firewall (NFTables/IPTables support)

## Supported Operating Systems

- Armbian
- Debian
- Ubuntu

# Core Module
> Web server core module providing HTTP/2, WebSocket, authentication, and auto-initialization functionality.

## Structure

```
core/
├── server.go           # Main HTTP/2 server implementation
├── websocket.go        # WebSocket handler and management
├── auth.go             # Authentication middleware
├── init.go             # Auto-initialization of handlers/routes/modules
├── database.go         # SQLite database connection and management
└── middleware.go       # Common middleware functions
```

## Features

- **HTTP/2 Server**: High-performance web server
- **WebSocket Support**: Real-time communication
- **Authentication**: Session-based authentication system
- **Auto-loading**: Automatic discovery and loading of:
  - HTTP handlers
  - URL routes
  - Service modules
- **Database**: SQLite integration (routersbc.sqlitedb)
- **Admin Interface**: Default admin user (sbc:sbc)

## Database Schema

- Users table (authentication)
- Configuration tables (network settings)
- Logs table (system events)
- Sessions table (active sessions)

## API Endpoints

- `/api/v1/auth` - Authentication endpoints
- `/api/v1/network` - Network management
- `/api/v1/system` - System information
- `/ws` - WebSocket endpoint for real-time updates

# Handlers Module
> HTTP request handlers for all web interface endpoints and API routes.

## Structure

```
handlers/
├── auth_handler.go         # Authentication handlers
├── network_handler.go      # Network management handlers
├── interface_handler.go    # Network interface handlers
├── wan_handler.go          # WAN configuration handlers
├── lan_handler.go          # LAN configuration handlers
├── wireless_handler.go     # Wireless management handlers
├── routing_handler.go      # Routing configuration handlers
├── firewall_handler.go     # Firewall management handlers
├── dns_handler.go          # DNS configuration handlers
├── dhcp_handler.go         # DHCP server handlers
└── websocket_handler.go    # WebSocket message handlers
```

## Handler Functions

### Authentication Handlers
- Login/logout functionality
- Session management
- User authentication

### Network Handlers
- Interface management (Physical, VPN, VLAN, ...)
- WAN configuration (Wire/Wireless, Multi-WAN)
- LAN settings (DHCP, DNS, Bridge)
- Wireless management (AP, STA, ADHOC, Monitor)
- Static routing and UPnP IGD
- Firewall rules (NFTables/IPTables)

### WebSocket Handlers
- Real-time interface updates
- System status notifications
- Configuration change broadcasts

## API Response Format

All handlers return JSON responses with consistent structure:
```json
{
  "success": true,
  "data": {},
  "message": "Operation completed successfully"
}
```

# Services Module
> Business logic services that interact with system configurators and provide data processing.

## Structure

```
services/
├── auth_service.go         # Authentication business logic
├── network_service.go      # Network management service
├── interface_service.go    # Network interface service
├── wan_service.go          # WAN configuration service
├── lan_service.go          # LAN configuration service
├── wireless_service.go     # Wireless management service
├── routing_service.go      # Routing configuration service
├── firewall_service.go     # Firewall management service
├── dns_service.go          # DNS configuration service
├── dhcp_service.go         # DHCP server service
└── system_service.go       # System information service
```

## Service Functions

### Network Services
- Interface detection and management
- Configuration validation
- System integration with `/utils/configurators/net/*`

### Authentication Service
- User management
- Session handling
- Permission validation

### System Service
- OS detection (Armbian, Debian, Ubuntu)
- Service management (start/stop/restart/reload)
- Kernel parameter management

## Integration

Services integrate with:
- `/utils/configurators/` for system configuration
- Database for persistent storage
- WebSocket for real-time updates
- External system tools and APIs

## Configuration Management

Services handle:
- Network interface configuration
- Firewall rules (NFTables/IPTables)
- DNS and DHCP settings
- Wireless network management
- Static routing configuration

# Routes Module
> URL routing definitions for web interface and API endpoints.

## Structure

```
routes/
├── auth_routes.go          # Authentication routes
├── network_routes.go       # Network management routes
├── interface_routes.go     # Network interface routes
├── wan_routes.go           # WAN configuration routes
├── lan_routes.go           # LAN configuration routes
├── wireless_routes.go      # Wireless management routes
├── routing_routes.go       # Routing configuration routes
├── firewall_routes.go      # Firewall management routes
├── api_routes.go           # REST API v1 routes
└── static_routes.go        # Static file serving routes
```

## Route Definitions

### Web Interface Routes
- `/` - Dashboard
- `/network` - Network management interface
- `/network/interfaces` - Interface management
- `/network/wan` - WAN configuration
- `/network/lan` - LAN configuration
- `/network/wireless` - Wireless management
- `/network/routing` - Routing configuration
- `/network/firewall` - Firewall management

### API Routes (v1)
- `/api/v1/auth/*` - Authentication endpoints
- `/api/v1/network/*` - Network management API
- `/api/v1/interfaces/*` - Interface management API
- `/api/v1/wan/*` - WAN configuration API
- `/api/v1/lan/*` - LAN configuration API
- `/api/v1/wireless/*` - Wireless management API
- `/api/v1/routing/*` - Routing configuration API
- `/api/v1/firewall/*` - Firewall management API

### WebSocket Routes
- `/ws` - WebSocket endpoint for real-time updates

### Static Routes
- `/static/*` - Static assets (CSS, JS, images)

# Templates Module
> HTML templates based on Tabler.io framework for web interface.

## Structure

```
templates/
├── base/
│   ├── layout.html         # Base layout template
│   ├── header.html         # Common header
│   ├── sidebar.html        # Navigation sidebar
│   └── footer.html         # Common footer
├── auth/
│   ├── login.html          # Login page
│   └── logout.html         # Logout confirmation
├── network/
│   ├── index.html          # Network dashboard
│   ├── interfaces.html     # Interface management
│   ├── wan.html            # WAN configuration
│   ├── lan.html            # LAN configuration
│   ├── wireless.html       # Wireless management
│   ├── routing.html        # Routing configuration
│   ├── firewall.html       # Firewall management
│   ├── firewall_new.html   # NFTables firewall
│   ├── firewall_classic.html # IPTables firewall
│   └── modal.html          # Modal dialogs
└── dashboard/
    └── index.html          # Main dashboard
```

## Template Features

### Base Templates
- **Tabler.io Framework**: Modern, responsive design
- **No Static Data**: All data loaded dynamically
- **WebSocket Integration**: Real-time updates
- **Modal Support**: Dynamic modal dialogs

### Network Templates
- **Tabbed Interface**: Organized configuration sections
- **Dynamic Forms**: Configuration forms with validation
- **Real-time Status**: Live interface status updates

### Modal Dialogs
- Add/Edit UPnP STUN Server
- Add/Edit Static Route
- Add/Edit Firewall Rule/Chain
- Add/Edit Network Interface
- Add/Edit DNS Local Zones/Resolvers/Routing
- Add/Edit WiFi Station/Access Point connections

## Template Data Structure

Templates expect data in JSON format:
- Interface lists and configurations
- Network settings and status
- Firewall rules and chains
- DNS and DHCP configurations

# Static Assets
> Static files for web interface including CSS, JavaScript, and images.

## Structure

```
static/
├── css/
│   ├── tabler.min.css      # Tabler.io framework CSS
│   ├── custom.css          # Custom styles
│   └── network.css         # Network-specific styles
├── js/
│   ├── tabler.min.js       # Tabler.io framework JS
│   ├── websocket.js        # WebSocket client
│   ├── network.js          # Network management JS
│   ├── modal.js            # Modal dialog handling
│   └── utils.js            # Utility functions
├── images/
│   ├── logo.png            # Application logo
│   ├── icons/              # Interface icons
│   └── network/            # Network-related images
└── fonts/
    └── tabler-icons/       # Tabler icon fonts
```

## JavaScript Modules

### WebSocket Client (`websocket.js`)
- Real-time interface updates
- System status notifications
- Configuration change handling

### Network Management (`network.js`)
- Interface management functions
- Configuration form handling
- Status monitoring

### Modal Handling (`modal.js`)
- Dynamic modal creation
- Form validation
- AJAX form submission

### Utilities (`utils.js`)
- Common helper functions
- Data formatting
- Error handling

## CSS Customization

### Custom Styles (`custom.css`)
- Application-specific styling
- Theme customizations
- Responsive design adjustments

### Network Styles (`network.css`)
- Network interface styling
- Status indicators
- Configuration form styling

# Utils Module
> Utility tools and system configurators for different operating systems.

## Structure

```
utils/
├── configurators/
│   ├── net/
│   │   ├── netplan.go      # Netplan configuration (Ubuntu)
│   │   ├── interfaces.go   # Network interfaces configuration
│   │   ├── wan.go          # WAN configuration
│   │   ├── dns.go          # DNS configuration
│   │   ├── dhcp.go         # DHCP server configuration
│   │   ├── routing.go      # Static routing configuration
│   │   └── firewall.go     # Firewall configuration
│   └── sys/
│       ├── kernel.go       # Kernel parameter management
│       └── services.go     # System service management
├── helpers/
│   ├── validation.go       # Input validation helpers
│   ├── network.go          # Network utility functions
│   └── system.go           # System information helpers
└── constants/
    ├── network.go          # Network-related constants
    └── system.go           # System constants
```

## Configurators

### Network Configurators (`net/`)
- **netplan.go**: Ubuntu/Netplan configuration management
- **interfaces.go**: Debian/interfaces configuration
- **wan.go**: WAN interface configuration
- **dns.go**: DNS resolver configuration (DoT, DoH support)
- **dhcp.go**: DHCP server configuration
- **routing.go**: Static routes and UPnP configuration
- **firewall.go**: NFTables/IPTables management

### System Configurators (`sys/`)
- **kernel.go**: Kernel parameter tuning
- **services.go**: Systemd service management (start/stop/restart/reload/enable/disable)

## Supported Operating Systems

- **Armbian**: ARM-based Linux distribution
- **Debian**: Debian-based systems
- **Ubuntu**: Ubuntu and derivatives

## Configuration Features

### Network Configuration
- Interface detection and management
- VLAN and bridge configuration
- Wireless network management
- Firewall rule management
- DNS and DHCP configuration

### System Management
- Service control and monitoring
- Kernel parameter optimization
- System information gathering

# Scripts Module

Build, deployment, and maintenance scripts for the x-routersbc project.

## Structure

```
scripts/
├── build/
│   ├── build.sh            # Main build script
│   ├── cross-compile.sh    # Cross-compilation for ARM
│   └── package.sh          # Package creation script
├── deploy/
│   ├── install.sh          # Installation script
│   ├── update.sh           # Update script
│   └── uninstall.sh        # Uninstallation script
├── dev/
│   ├── setup-dev.sh        # Development environment setup
│   ├── test.sh             # Run tests
│   └── lint.sh             # Code linting
└── maintenance/
    ├── backup.sh           # Configuration backup
    ├── restore.sh          # Configuration restore
    └── cleanup.sh          # System cleanup
```

## Build Scripts

### Main Build (`build/build.sh`)
- Compile Go application
- Generate static assets
- Create distribution package

### Cross-compilation (`build/cross-compile.sh`)
- Build for ARM architectures
- Support for different ARM variants
- Optimization for embedded systems

### Packaging (`build/package.sh`)
- Create DEB/RPM packages
- Generate installation archives
- Include dependencies and configurations

## Deployment Scripts

### Installation (`deploy/install.sh`)
- System requirements check
- Service installation and configuration
- Database initialization
- Default configuration setup

### Updates (`deploy/update.sh`)
- Application updates
- Configuration migration
- Service restart management

### Uninstallation (`deploy/uninstall.sh`)
- Clean removal of application
- Configuration backup option
- Service cleanup

## Development Scripts

### Development Setup (`dev/setup-dev.sh`)
- Development environment preparation
- Dependency installation
- Test database setup

### Testing (`dev/test.sh`)
- Unit test execution
- Integration test running
- Coverage reporting

### Code Quality (`dev/lint.sh`)
- Go code linting
- Static analysis
- Code formatting validation

## Maintenance Scripts

### Backup (`maintenance/backup.sh`)
- Configuration backup
- Database backup
- System state preservation

### Restore (`maintenance/restore.sh`)
- Configuration restoration
- Database recovery
- Service reconfiguration

### Cleanup (`maintenance/cleanup.sh`)
- Log file cleanup
- Temporary file removal
