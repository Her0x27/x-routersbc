
# Templates Module - HTML шаблоны

## ПРАВИЛЬНАЯ структура шаблонов

### Базовый шаблон (base/layout.html)
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{block "title" .}}RouterSBC{{end}}</title>
    <link href="/static/css/tabler.min.css" rel="stylesheet">
    <link href="/static/css/tabler-icons.min.css" rel="stylesheet">
</head>
<body>
    {{template "navbar" .}}
    
    <div class="page-wrapper">
        {{block "body" .}}{{end}}
    </div>
    
    <script src="/static/js/tabler.min.js"></script>
    <script src="/static/js/app.js"></script>
    {{block "scripts" .}}{{end}}
</body>
</html>
```

### ПРАВИЛЬНЫЙ формат страниц
```html
{{template "base/layout.html" .}}

{{define "title"}}Network Interfaces - RouterSBC{{end}}

{{define "body"}}
<div class="page-header d-print-none">
    <div class="container-xl">
        <div class="row g-2 align-items-center">
            <div class="col">
                <h2 class="page-title">Network Interfaces</h2>
            </div>
        </div>
    </div>
</div>

<div class="page-body">
    <div class="container-xl">
        <!-- ТОЛЬКО реальные данные -->
        {{if .interfaces}}
            {{range .interfaces}}
            <div class="card mb-3">
                <div class="card-body">
                    <h3>{{.Name}}</h3>
                    <p>Status: 
                        {{if .IsUp}}
                            <span class="status status-green">
                                <span class="status-dot status-dot-animated"></span>
                                UP
                            </span>
                        {{else}}
                            <span class="status status-red">
                                <span class="status-dot"></span>
                                DOWN
                            </span>
                        {{end}}
                    </p>
                    {{if .IPAddress}}
                    <p>IP: {{.IPAddress}}{{if .Netmask}}/{{.Netmask}}{{end}}</p>
                    {{end}}
                </div>
            </div>
            {{end}}
        {{else}}
            <div class="alert alert-info">
                No network interfaces found.
            </div>
        {{end}}
    </div>
</div>
{{end}}

{{define "scripts"}}
<script>
// ТОЛЬКО реальные данные через API
async function loadInterfaces() {
    try {
        const response = await fetch('/api/network/interfaces');
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        updateInterfaceTable(data.interfaces);
    } catch (error) {
        console.error('Failed to load interfaces:', error);
        showError('Failed to load network interfaces');
    }
}

function showError(message) {
    const alertDiv = document.createElement('div');
    alertDiv.className = 'alert alert-danger';
    alertDiv.textContent = message;
    document.querySelector('.page-body .container-xl').prepend(alertDiv);
}

// ОБЯЗАТЕЛЬНАЯ обработка ошибок загрузки
document.addEventListener('DOMContentLoaded', loadInterfaces);
</script>
{{end}}
```

## ИСПРАВЛЕНИЯ ошибок в текущих шаблонах

### ❌ ОШИБКА в system/portable-devices.html:
```html
<!-- НЕПРАВИЛЬНО: статические данные -->
<td>USB WiFi Adapter</td>
<td><span class="badge bg-blue">wifi</span></td>

<!-- ПРАВИЛЬНО: только реальные данные -->
<td>{{.Name}}</td>
<td><span class="badge bg-{{getDeviceColor .Type}}">{{.Type}}</span></td>
```

### ❌ ОШИБКА в network/interfaces.html:
```html
<!-- НЕПРАВИЛЬНО: игнорирование ошибок -->
{{range .interfaces}}
<tr>
    <td>{{.Name}}</td>
</tr>
{{end}}

<!-- ПРАВИЛЬНО: обработка пустых данных -->
{{if .interfaces}}
    {{range .interfaces}}
    <tr>
        <td>{{.Name}}</td>
        <td>
            {{if .IsUp}}
                <span class="status status-green">
                    <span class="status-dot status-dot-animated"></span>
                    UP
                </span>
            {{else}}
                <span class="status status-red">
                    <span class="status-dot"></span>
                    DOWN
                </span>
            {{end}}
        </td>
    </tr>
    {{end}}
{{else}}
    <tr>
        <td colspan="6" class="text-center text-muted">
            No interfaces available
        </td>
    </tr>
{{end}}
```

## JavaScript - ОБЯЗАТЕЛЬНАЯ обработка ошибок

### ПРАВИЛЬНАЯ загрузка данных
```javascript
// ОБЯЗАТЕЛЬНАЯ обработка всех ошибок
async function loadData(endpoint) {
    try {
        const response = await fetch(endpoint);
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const contentType = response.headers.get('content-type');
        if (!contentType || !contentType.includes('application/json')) {
            throw new Error('Response is not JSON');
        }
        
        return await response.json();
    } catch (error) {
        console.error(`Failed to load data from ${endpoint}:`, error);
        showErrorMessage(`Failed to load data: ${error.message}`);
        throw error;
    }
}

// ОБЯЗАТЕЛЬНОЕ отображение ошибок пользователю
function showErrorMessage(message) {
    const alertDiv = document.createElement('div');
    alertDiv.className = 'alert alert-danger alert-dismissible';
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    const container = document.querySelector('.page-body .container-xl');
    if (container) {
        container.insertBefore(alertDiv, container.firstChild);
    }
}
```

## WebSocket обновления - ОБЯЗАТЕЛЬНАЯ обработка соединения

```javascript
class SystemWebSocket {
    constructor(endpoint) {
        this.endpoint = endpoint;
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.reconnectAttempts = 0;
    }
    
    connect() {
        try {
            this.ws = new WebSocket(`ws://${window.location.host}${this.endpoint}`);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
                this.reconnectDelay = 1000;
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleMessage(data);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };
            
            this.ws.onclose = () => {
                console.log('WebSocket disconnected');
                this.scheduleReconnect();
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
            
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.scheduleReconnect();
        }
    }
    
    scheduleReconnect() {
        setTimeout(() => {
            this.reconnectAttempts++;
            this.reconnectDelay = Math.min(
                this.reconnectDelay * 2, 
                this.maxReconnectDelay
            );
            this.connect();
        }, this.reconnectDelay);
    }
    
    handleMessage(data) {
        // Переопределить в наследуемых классах
        console.log('Received WebSocket message:', data);
    }
}
```

## ЗАПРЕЩЕННЫЕ практики:
- ❌ `{{define "content"}}` - вызывает дублирование
- ❌ Статические/фиктивные данные в шаблонах
- ❌ JavaScript без обработки ошибок
- ❌ Игнорирование ошибок fetch/WebSocket
- ❌ Отсутствие проверки пустых данных

## ОБЯЗАТЕЛЬНЫЕ практики:
- ✅ `{{template "base/layout.html" .}}` в начале каждого шаблона
- ✅ `{{define "body"}}` вместо `{{define "content"}}`
- ✅ Проверка данных на пустоту: `{{if .data}}`
- ✅ try-catch для всех async операций
- ✅ Отображение ошибок пользователю
- ✅ Переподключение WebSocket при обрыве
