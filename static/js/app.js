/**
 * RouterSBC Main Application JavaScript
 * Handles WebSocket connections, real-time updates, and common UI interactions
 */

class RouterSBCApp {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectInterval = 5000;
        this.isConnected = false;
        
        this.init();
    }
    
    init() {
        this.setupWebSocket();
        this.setupEventListeners();
        this.setupTooltips();
        this.initializeModals();
    }
    
    /**
     * Setup WebSocket connection for real-time updates
     */
    setupWebSocket() {
        if (typeof WebSocket === 'undefined') {
            console.warn('WebSocket not supported by this browser');
            return;
        }
        
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        try {
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = (event) => {
                console.log('WebSocket connected');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.updateConnectionStatus(true);
                
                // Send initial ping
                this.sendMessage('ping', {});
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleWebSocketMessage(message);
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            };
            
            this.ws.onclose = (event) => {
                console.log('WebSocket disconnected:', event.code, event.reason);
                this.isConnected = false;
                this.updateConnectionStatus(false);
                this.attemptReconnect();
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionStatus(false);
            };
            
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
        }
    }
    
    /**
     * Handle incoming WebSocket messages
     */
    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'pong':
                // Handle ping response
                break;
                
            case 'interfaces_update':
                this.handleInterfacesUpdate(message.data);
                break;
                
            case 'system_status':
                this.handleSystemStatusUpdate(message.data);
                break;
                
            case 'network_status':
                this.handleNetworkStatusUpdate(message.data);
                break;
                
            default:
                console.log('Unknown WebSocket message type:', message.type);
        }
    }
    
    /**
     * Send WebSocket message
     */
    sendMessage(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, data }));
        }
    }
    
    /**
     * Attempt to reconnect WebSocket
     */
    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect WebSocket (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(() => {
                this.setupWebSocket();
            }, this.reconnectInterval);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }
    
    /**
     * Update connection status indicator
     */
    updateConnectionStatus(connected) {
        const indicator = document.getElementById('connection-status');
        if (indicator) {
            if (connected) {
                indicator.className = 'status status-green';
                indicator.innerHTML = '<span class="status-dot status-dot-animated"></span>Connected';
            } else {
                indicator.className = 'status status-red';
                indicator.innerHTML = '<span class="status-dot"></span>Disconnected';
            }
        }
    }
    
    /**
     * Handle interfaces update
     */
    handleInterfacesUpdate(data) {
        // Trigger interface table refresh if on interfaces page
        if (window.location.pathname.includes('/network/interfaces')) {
            const event = new CustomEvent('interfacesUpdated', { detail: data });
            document.dispatchEvent(event);
        }
    }
    
    /**
     * Handle system status update
     */
    handleSystemStatusUpdate(data) {
        // Update system status widgets
        this.updateSystemStatusWidgets(data);
    }
    
    /**
     * Handle network status update
     */
    handleNetworkStatusUpdate(data) {
        // Update network status indicators
        this.updateNetworkStatusIndicators(data);
    }
    
    /**
     * Update system status widgets
     */
    updateSystemStatusWidgets(data) {
        // Memory usage
        const memoryBar = document.getElementById('memory-bar');
        const memoryPercent = document.getElementById('memory-percent');
        if (memoryBar && memoryPercent && data.memory_usage !== undefined) {
            const percent = Math.round(data.memory_usage);
            memoryBar.style.width = percent + '%';
            memoryPercent.textContent = percent + '%';
        }
        
        // Disk usage
        const diskBar = document.getElementById('disk-bar');
        const diskPercent = document.getElementById('disk-percent');
        if (diskBar && diskPercent && data.disk_usage !== undefined) {
            const percent = Math.round(data.disk_usage);
            diskBar.style.width = percent + '%';
            diskPercent.textContent = percent + '%';
        }
        
        // Temperature
        const temperature = document.getElementById('temperature');
        if (temperature && data.temperature !== undefined) {
            temperature.textContent = Math.round(data.temperature) + '°C';
        }
        
        // Load average
        const loadAverage = document.getElementById('load-average');
        if (loadAverage && data.load_average) {
            loadAverage.textContent = data.load_average;
        }
        
        // Process count
        const processes = document.getElementById('processes');
        if (processes && data.process_count !== undefined) {
            processes.textContent = data.process_count;
        }
    }
    
    /**
     * Update network status indicators
     */
    updateNetworkStatusIndicators(data) {
        // Update status dots and text for network components
        const indicators = [
            { id: 'wan-status-dot', text: 'wan-status-text', key: 'wan_connected' },
            { id: 'lan-status-dot', text: 'lan-status-text', key: 'lan_active' },
            { id: 'wifi-status-dot', text: 'wifi-status-text', key: 'wifi_active' },
            { id: 'firewall-status-dot', text: 'firewall-status-text-detail', key: 'firewall_active' }
        ];
        
        indicators.forEach(indicator => {
            const dot = document.getElementById(indicator.id);
            const text = document.getElementById(indicator.text);
            
            if (dot && data[indicator.key] !== undefined) {
                const isActive = data[indicator.key];
                dot.className = `status-dot d-inline-block ${isActive ? 'bg-green status-dot-animated' : 'bg-red'}`;
                
                if (text) {
                    switch (indicator.key) {
                        case 'wan_connected':
                            text.textContent = isActive ? 'Connected' : 'Disconnected';
                            break;
                        case 'lan_active':
                            text.textContent = isActive ? 'Active' : 'Inactive';
                            break;
                        case 'wifi_active':
                            text.textContent = isActive ? 'Active' : 'Inactive';
                            break;
                        case 'firewall_active':
                            text.textContent = isActive ? 'Protected' : 'Disabled';
                            break;
                    }
                }
            }
        });
    }
    
    /**
     * Setup global event listeners
     */
    setupEventListeners() {
        // Handle form submissions with loading states
        document.addEventListener('submit', (event) => {
            const form = event.target;
            if (form.tagName === 'FORM') {
                this.handleFormSubmit(form);
            }
        });
        
        // Handle AJAX form submissions
        document.addEventListener('click', (event) => {
            const button = event.target;
            if (button.hasAttribute('data-action')) {
                event.preventDefault();
                this.handleActionButton(button);
            }
        });
        
        // Handle file uploads with progress
        document.addEventListener('change', (event) => {
            const input = event.target;
            if (input.type === 'file') {
                this.handleFileUpload(input);
            }
        });
    }
    
    /**
     * Setup tooltips
     */
    setupTooltips() {
        // Initialize Bootstrap tooltips
        const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
        tooltipTriggerList.map(function (tooltipTriggerEl) {
            return new bootstrap.Tooltip(tooltipTriggerEl);
        });
    }
    
    /**
     * Initialize modals
     */
    initializeModals() {
        // Auto-focus first input in modals
        document.addEventListener('shown.bs.modal', (event) => {
            const modal = event.target;
            const firstInput = modal.querySelector('input, select, textarea');
            if (firstInput) {
                firstInput.focus();
            }
        });
        
        // Clear forms when modals are hidden
        document.addEventListener('hidden.bs.modal', (event) => {
            const modal = event.target;
            const forms = modal.querySelectorAll('form');
            forms.forEach(form => {
                if (!form.hasAttribute('data-keep-values')) {
                    form.reset();
                }
            });
        });
    }
    
    /**
     * Handle form submission with loading states
     */
    handleFormSubmit(form) {
        const submitButton = form.querySelector('button[type="submit"]');
        if (submitButton) {
            const originalText = submitButton.innerHTML;
            submitButton.disabled = true;
            submitButton.innerHTML = '<span class="spinner-border spinner-border-sm me-2" role="status"></span>Saving...';
            
            // Restore button after a delay if form doesn't handle it
            setTimeout(() => {
                if (submitButton.disabled) {
                    submitButton.disabled = false;
                    submitButton.innerHTML = originalText;
                }
            }, 10000);
        }
    }
    
    /**
     * Handle action buttons
     */
    handleActionButton(button) {
        const action = button.getAttribute('data-action');
        const target = button.getAttribute('data-target');
        
        switch (action) {
            case 'refresh':
                this.handleRefreshAction(target);
                break;
            case 'toggle':
                this.handleToggleAction(target);
                break;
            case 'delete':
                this.handleDeleteAction(target);
                break;
            default:
                console.warn('Unknown action:', action);
        }
    }
    
    /**
     * Handle refresh actions
     */
    handleRefreshAction(target) {
        const originalButton = document.querySelector(`[data-action="refresh"][data-target="${target}"]`);
        if (originalButton) {
            const originalText = originalButton.innerHTML;
            originalButton.disabled = true;
            originalButton.innerHTML = '<span class="spinner-border spinner-border-sm me-2" role="status"></span>Refreshing...';
            
            // Simulate refresh - replace with actual refresh logic
            setTimeout(() => {
                originalButton.disabled = false;
                originalButton.innerHTML = originalText;
            }, 2000);
        }
    }
    
    /**
     * Handle toggle actions
     */
    handleToggleAction(target) {
        // Implementation for toggle actions
        console.log('Toggle action for:', target);
    }
    
    /**
     * Handle delete actions
     */
    handleDeleteAction(target) {
        if (confirm('Are you sure you want to delete this item?')) {
            // Implementation for delete actions
            console.log('Delete action for:', target);
        }
    }
    
    /**
     * Handle file uploads
     */
    handleFileUpload(input) {
        const file = input.files[0];
        if (file) {
            // Show file size and type
            const fileInfo = input.parentElement.querySelector('.file-info');
            if (fileInfo) {
                fileInfo.textContent = `${file.name} (${this.formatFileSize(file.size)})`;
            }
        }
    }
    
    /**
     * Format file size for display
     */
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    }
    
    /**
     * Show notification
     */
    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `alert alert-${type} alert-dismissible`;
        notification.innerHTML = `
            <div class="d-flex">
                <div>
                    <svg xmlns="http://www.w3.org/2000/svg" class="icon alert-icon" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
                        <circle cx="12" cy="12" r="9"/>
                        <path d="M12 8v4"/>
                        <path d="M12 16h.01"/>
                    </svg>
                </div>
                <div>${message}</div>
            </div>
            <a class="btn-close" data-bs-dismiss="alert" aria-label="close"></a>
        `;
        
        // Add to page
        const container = document.querySelector('.page-body .container-xl');
        if (container) {
            container.insertBefore(notification, container.firstChild);
            
            // Auto-remove after 5 seconds
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.remove();
                }
            }, 5000);
        }
    }
    
    /**
     * Utility method to make API requests
     */
    async apiRequest(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };
        
        const mergedOptions = { ...defaultOptions, ...options };
        
        try {
            const response = await fetch(url, mergedOptions);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('API request failed:', error);
            this.showNotification('An error occurred while communicating with the server.', 'danger');
            throw error;
        }
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.RouterSBC = new RouterSBCApp();
});

// Global utility functions
window.RouterSBCUtils = {
    formatBytes: function(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    },
    
    formatUptime: function(seconds) {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        if (days > 0) {
            return `${days}d ${hours}h ${minutes}m`;
        } else if (hours > 0) {
            return `${hours}h ${minutes}m`;
        } else {
            return `${minutes}m`;
        }
    },
    
    debounce: function(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};
