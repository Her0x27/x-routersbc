# Core Module

Основной модуль веб-сервера, обеспечивающий базовую функциональность системы управления SBC (Single Board Computer).

## Структура

```
core/
├── server.go          # Основной веб-сервер
├── config.go          # Конфигурация приложения
├── database.go        # Работа с SQLite базой данных
├── auth.go            # Система аутентификации
├── websocket.go       # WebSocket соединения
├── middleware.go      # Промежуточное ПО
├── router.go          # Маршрутизация запросов
└── init.go           # Инициализация модулей
```

## Основной функционал

### HTTP/2 Server
- Поддержка HTTP/2 протокола
- Graceful shutdown
- Middleware для логирования и восстановления после паники

### WebSocket
- Реальное время обновления интерфейса
- Уведомления о состоянии системы
- Двусторонняя связь клиент-сервер
- Автоматическое переподключение

### Аутентификация
- Сессионная аутентификация
- JWT токены для API
- Защита от CSRF атак
- Управление правами доступа

### База данных
- SQLite для хранения конфигурации
- Автоматическая миграция схемы
- Администратор по умолчанию: `sbc:sbc`
- Резервное копирование настроек

### Auto-loading система
- Автоматическая загрузка обработчиков из `/handlers`
- Автоматическая регистрация маршрутов из `/routes`
- Динамическая загрузка сервисов из `/services`
- Hot-reload в режиме разработки

### REST API v1
- RESTful архитектура
- JSON формат данных
- Версионирование API
- Swagger документация
- Rate limiting

## Конфигурация

### Переменные окружения
```bash
SBC_PORT=8080                   # Порт веб-сервера
SBC_DB_PATH=./sbc.db            # Путь к базе данных
SBC_DEBUG=false                 # Режим отладки
SBC_SESSION_SECRET=secret       # Секрет для сессий
```

### Файл конфигурации
```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  path: "./sbc.db"
  max_connections: 10
  timeout: 30s

auth:
  session_timeout: 24h
  jwt_expiry: 1h
  bcrypt_cost: 12

websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  ping_period: 54s
```

## API Endpoints

### Аутентификация
- `POST /api/v1/auth/login` - Вход в систему
- `POST /api/v1/auth/logout` - Выход из системы
- `GET /api/v1/auth/me` - Информация о пользователе

### Система
- `GET /api/v1/system/info` - Информация о системе
- `GET /api/v1/system/status` - Статус системы
- `POST /api/v1/system/reboot` - Перезагрузка системы

### WebSocket
- `WS /ws` - WebSocket соединение для реального времени

## Использование

```go
package main

import (
    "log"
    "github.com/Her0x27/x-routersbc/core"
)

func main() {
    server := core.NewServer()
    
    if err := server.Initialize(); err != nil {
        log.Fatal("Failed to initialize server:", err)
    }
    
    if err := server.Start(); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

## Зависимости

- `github.com/gin-gonic/gin` - HTTP веб-фреймворк
- `github.com/gorilla/websocket` - WebSocket поддержка
- `github.com/mattn/go-sqlite3` - SQLite драйвер
- `github.com/golang-jwt/jwt/v4` - JWT токены
- `golang.org/x/crypto/bcrypt` - Хеширование паролей
