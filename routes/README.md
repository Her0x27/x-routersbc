
# Routes Module - Маршрутизация

## Структура маршрутов

### ОБЯЗАТЕЛЬНЫЕ Middleware
```go
func SetupRoutes(e *echo.Echo, handlers *Handlers) {
    // ОБЯЗАТЕЛЬНАЯ обработка паники
    e.Use(middleware.Recover())
    
    // ОБЯЗАТЕЛЬНОЕ логирование запросов
    e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
        Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
    }))
    
    // ОБЯЗАТЕЛЬНАЯ безопасность
    e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
        XSSProtection:         "1; mode=block",
        ContentTypeNosniff:    "nosniff",
        XFrameOptions:         "DENY",
        HSTSMaxAge:           3600,
    }))
    
    // ОБЯЗАТЕЛЬНЫЕ ограничения
    e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
        Skipper: middleware.DefaultSkipper,
        Store: middleware.NewRateLimiterMemoryStoreWithConfig(
            middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(20), Burst: 30}),
        IdentifierExtractor: func(ctx echo.Context) (string, error) {
            id := ctx.RealIP()
            return id, nil
        },
        ErrorHandler: func(context echo.Context, err error) error {
            return context.JSON(429, map[string]string{
                "error": "Too many requests",
            })
        },
    }))
    
    // ОБЯЗАТЕЛЬНЫЕ таймауты
    e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
        Timeout: 30 * time.Second,
        ErrorMessage: "Request timeout",
    }))
}
```

### Маршруты аутентификации
```go
func setupAuthRoutes(e *echo.Echo, auth *AuthHandler) {
    authGroup := e.Group("/auth")
    
    // ОБЯЗАТЕЛЬНАЯ валидация для логина
    authGroup.POST("/login", auth.Login, validateLoginRequest())
    authGroup.POST("/logout", auth.Logout, AuthRequiredMiddleware())
    authGroup.GET("/status", auth.Status, AuthRequiredMiddleware())
}

// Middleware валидации входа
func validateLoginRequest() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            var req LoginRequest
            if err := c.Bind(&req); err != nil {
                return c.JSON(400, map[string]string{
                    "error": "Invalid request format",
                })
            }
            
            // Валидация полей
            if req.Username == "" || req.Password == "" {
                return c.JSON(400, map[string]string{
                    "error": "Username and password are required",
                })
            }
            
            if len(req.Username) > 50 || len(req.Password) > 128 {
                return c.JSON(400, map[string]string{
                    "error": "Username or password too long",
                })
            }
            
            return next(c)
        }
    }
}
```

### API маршруты с обработкой ошибок
```go
func setupAPIRoutes(e *echo.Echo, handlers *Handlers) {
    api := e.Group("/api", AuthRequiredMiddleware())
    
    // ОБЯЗАТЕЛЬНАЯ валидация для всех API
    api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Установка заголовков API
            c.Response().Header().Set("Content-Type", "application/json")
            
            // Обработка паники в API
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("API panic: %v", r)
                    c.JSON(500, map[string]string{
                        "error": "Internal server error",
                        "code":  "INTERNAL_ERROR",
                    })
                }
            }()
            
            return next(c)
        }
    })
    
    // Network API
    network := api.Group("/network")
    network.GET("/interfaces", handlers.Network.GetInterfaces)
    network.POST("/interfaces", handlers.Network.CreateInterface, validateInterfaceRequest())
    network.PUT("/interfaces/:name", handlers.Network.UpdateInterface, validateInterfaceRequest())
    network.DELETE("/interfaces/:name", handlers.Network.DeleteInterface)
    
    // System API  
    system := api.Group("/system")
    system.GET("/info", handlers.System.GetSystemInfo)
    system.GET("/usb-devices", handlers.System.GetUSBDevices)
}

// Валидация интерфейсов
func validateInterfaceRequest() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            var req InterfaceRequest
            if err := c.Bind(&req); err != nil {
                return c.JSON(400, map[string]string{
                    "error": "Invalid request format",
                })
            }
            
            // Валидация имени интерфейса
            if req.Name != "" {
                if !regexp.MustMatch(`^[a-zA-Z0-9-]{1,15}$`, req.Name) {
                    return c.JSON(400, map[string]string{
                        "error": "Invalid interface name",
                    })
                }
            }
            
            // Валидация IP адреса
            if req.IPAddress != "" {
                if net.ParseIP(req.IPAddress) == nil {
                    return c.JSON(400, map[string]string{
                        "error": "Invalid IP address",
                    })
                }
            }
            
            return next(c)
        }
    }
}
```

### WebSocket маршруты
```go
func setupWebSocketRoutes(e *echo.Echo, ws *WSHandler) {
    wsGroup := e.Group("/ws", AuthRequiredMiddleware())
    
    // ОБЯЗАТЕЛЬНАЯ обработка ошибок соединения
    wsGroup.GET("/system", func(c echo.Context) error {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("WebSocket panic: %v", r)
            }
        }()
        
        return ws.HandleSystemUpdates(c)
    })
    
    wsGroup.GET("/network", func(c echo.Context) error {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("WebSocket panic: %v", r)
            }
        }()
        
        return ws.HandleNetworkUpdates(c)
    })
}
```

## ЗАПРЕЩЕННЫЕ практики:
- ❌ Маршруты без middleware аутентификации
- ❌ API без валидации входных данных
- ❌ Отсутствие обработки паники
- ❌ Маршруты без таймаутов
- ❌ Незащищенные административные функции

## ОБЯЗАТЕЛЬНЫЕ практики:
- ✅ Все маршруты с middleware безопасности
- ✅ Валидация для всех POST/PUT запросов
- ✅ Rate limiting для API
- ✅ Логирование всех запросов
- ✅ Централизованная обработка ошибок
