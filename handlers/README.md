
# Handlers Module - HTTP обработчики

## Структура обработчиков

### 1. Аутентификация (auth.go)
```go
// ОБЯЗАТЕЛЬНАЯ валидация и обработка ошибок
func (h *AuthHandler) Login(c echo.Context) error {
    var req LoginRequest
    
    // Валидация входных данных
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, ErrorResponse{
            Error: "Invalid request format",
            Code:  "INVALID_REQUEST",
        })
    }
    
    if err := c.Validate(&req); err != nil {
        return c.JSON(400, ErrorResponse{
            Error: "Validation failed",
            Details: err.Error(),
        })
    }
    
    // Sanitize входные данные
    req.Username = strings.TrimSpace(req.Username)
    
    session, user, err := h.authManager.Login(req.Username, req.Password)
    if err != nil {
        // Логирование попыток входа
        h.logger.Warn("Login attempt failed", 
            "username", req.Username,
            "ip", c.RealIP(),
            "error", err)
            
        return c.JSON(401, ErrorResponse{
            Error: "Invalid credentials",
            Code:  "AUTH_FAILED",
        })
    }
    
    // Установка cookie с правильными флагами безопасности
    cookie := &http.Cookie{
        Name:     "session",
        Value:    session.Token,
        HttpOnly: true,
        Secure:   true, // HTTPS only
        SameSite: http.SameSiteStrictMode,
        MaxAge:   86400, // 24 hours
    }
    c.SetCookie(cookie)
    
    return c.JSON(200, LoginResponse{
        Success: true,
        User:    user,
    })
}
```

### 2. Сетевые интерфейсы (network.go)
```go
// ОБЯЗАТЕЛЬНАЯ валидация системных команд
func (h *NetworkHandler) UpdateInterface(c echo.Context) error {
    var req InterfaceUpdateRequest
    
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, ErrorResponse{
            Error: "Invalid request",
        })
    }
    
    // Валидация имени интерфейса (защита от injection)
    if !isValidInterfaceName(req.Name) {
        return c.JSON(400, ErrorResponse{
            Error: "Invalid interface name",
            Code:  "INVALID_INTERFACE",
        })
    }
    
    // Валидация IP адреса
    if req.IPAddress != "" && net.ParseIP(req.IPAddress) == nil {
        return c.JSON(400, ErrorResponse{
            Error: "Invalid IP address",
            Code:  "INVALID_IP",
        })
    }
    
    // Вызов сервиса с обработкой ошибок
    if err := h.networkService.UpdateInterface(req.Name, req); err != nil {
        h.logger.Error("Failed to update interface", 
            "interface", req.Name,
            "error", err)
            
        return c.JSON(500, ErrorResponse{
            Error: "Failed to update interface",
            Code:  "UPDATE_FAILED",
        })
    }
    
    return c.JSON(200, SuccessResponse{
        Success: true,
        Message: "Interface updated successfully",
    })
}

// Валидация имени интерфейса
func isValidInterfaceName(name string) bool {
    if len(name) == 0 || len(name) > 15 {
        return false
    }
    // Только буквы, цифры и дефис
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, name)
    return matched
}
```

### 3. Системные обработчики (system.go)
```go
// ОБЯЗАТЕЛЬНАЯ обработка системных операций
func (h *SystemHandler) GetSystemInfo(c echo.Context) error {
    ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
    defer cancel()
    
    info, err := h.systemService.GetSystemInfoWithContext(ctx)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return c.JSON(504, ErrorResponse{
                Error: "System info request timed out",
                Code:  "TIMEOUT",
            })
        }
        
        h.logger.Error("Failed to get system info", "error", err)
        return c.JSON(500, ErrorResponse{
            Error: "Failed to retrieve system information",
            Code:  "SYSTEM_ERROR",
        })
    }
    
    return c.JSON(200, SystemInfoResponse{
        Success: true,
        Data:    info,
    })
}
```

## Middleware для всех обработчиков

### 1. Аутентификация
```go
func AuthMiddleware(authManager *AuthManager) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Пропустить страницы входа
            if c.Request().URL.Path == "/login" {
                return next(c)
            }
            
            cookie, err := c.Cookie("session")
            if err != nil {
                return c.Redirect(302, "/login")
            }
            
            session, err := authManager.ValidateSession(cookie.Value)
            if err != nil {
                return c.Redirect(302, "/login")
            }
            
            c.Set("session", session)
            return next(c)
        }
    }
}
```

### 2. CORS и безопасность
```go
func SecurityMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Security headers
            c.Response().Header().Set("X-Content-Type-Options", "nosniff")
            c.Response().Header().Set("X-Frame-Options", "DENY")
            c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
            
            return next(c)
        }
    }
}
```

## ЗАПРЕЩЕННЫЕ практики:
- ❌ Прямое выполнение системных команд без валидации
- ❌ SQL injection через неэкранированные параметры
- ❌ Отсутствие валидации входных данных
- ❌ Игнорирование ошибок сервисов
- ❌ Неустановленные HTTP статус коды

## ОБЯЗАТЕЛЬНЫЕ практики:
- ✅ Валидация ВСЕХ входных данных
- ✅ Экранирование для SQL и команд
- ✅ Логирование всех ошибок
- ✅ Корректные HTTP статус коды
- ✅ Таймауты для длительных операций
