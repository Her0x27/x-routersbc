
# Core Module - Веб-сервер и аутентификация

## Компоненты

### 1. Веб-сервер (server.go)
```go
// ОБЯЗАТЕЛЬНАЯ обработка ошибок для всех маршрутов
func (s *Server) setupRoutes() {
    // Middleware для обработки ошибок
    s.echo.Use(middleware.Recover())
    s.echo.Use(s.errorHandler)
    
    // Централизованная обработка паники
    s.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            defer func() {
                if r := recover(); r != nil {
                    err := fmt.Errorf("panic recovered: %v", r)
                    s.logger.Error("Panic in handler", "error", err)
                    c.JSON(500, map[string]string{"error": "Internal server error"})
                }
            }()
            return next(c)
        }
    })
}

// Обязательная обработка ошибок запуска сервера
func (s *Server) Start() error {
    if err := s.setupDatabase(); err != nil {
        return fmt.Errorf("failed to setup database: %w", err)
    }
    
    if err := s.setupAuth(); err != nil {
        return fmt.Errorf("failed to setup auth: %w", err)
    }
    
    // НИКОГДА не игнорировать ошибки запуска
    return s.echo.Start(":5000")
}
```

### 2. База данных (database.go)
```go
// ОБЯЗАТЕЛЬНАЯ обработка всех SQL ошибок
func (db *Database) Connect() error {
    conn, err := sql.Open("sqlite3", "routersbc.sqlitedb")
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }
    
    if err := conn.Ping(); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }
    
    db.conn = conn
    return db.createTables()
}

// Обработка транзакций с rollback
func (db *Database) ExecuteInTransaction(fn func(*sql.Tx) error) error {
    tx, err := db.conn.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()
    
    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rbErr)
        }
        return err
    }
    
    return tx.Commit()
}
```

### 3. Аутентификация (auth.go)
```go
// ОБЯЗАТЕЛЬНАЯ валидация входных данных
func (am *AuthManager) Login(username, password string) (*Session, *User, error) {
    if strings.TrimSpace(username) == "" {
        return nil, nil, errors.New("username cannot be empty")
    }
    if len(password) < 3 {
        return nil, nil, errors.New("password too short")
    }
    
    // Защита от timing attacks
    time.Sleep(100 * time.Millisecond)
    
    user, err := am.getUserByUsername(username)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil, errors.New("invalid credentials")
        }
        return nil, nil, fmt.Errorf("database error: %w", err)
    }
    
    // ОБЯЗАТЕЛЬНАЯ проверка хешированного пароля
    if !VerifyPassword(password, user.PasswordHash) {
        return nil, nil, errors.New("invalid credentials")
    }
    
    return am.createSession(user.ID)
}
```

### 4. WebSocket (websocket.go)
```go
// ОБЯЗАТЕЛЬНАЯ обработка ошибок соединения
func (h *WSHandler) HandleConnection(c echo.Context) error {
    ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
    if err != nil {
        return fmt.Errorf("failed to upgrade connection: %w", err)
    }
    defer ws.Close()
    
    // Обработка паники в горутинах
    defer func() {
        if r := recover(); r != nil {
            log.Printf("WebSocket panic: %v", r)
            ws.Close()
        }
    }()
    
    // Таймауты для чтения/записи
    ws.SetReadDeadline(time.Now().Add(60 * time.Second))
    ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
    
    return h.handleMessages(ws)
}
```

## ЗАПРЕЩЕННЫЕ практики:
- ❌ `if err != nil { /* игнорирование */ }`
- ❌ Неинициализированные переменные
- ❌ SQL без prepared statements
- ❌ Отсутствие валидации входных данных
- ❌ Незакрытые ресурсы (DB, файлы, сокеты)

## ОБЯЗАТЕЛЬНЫЕ практики:
- ✅ Обработка ВСЕХ ошибок
- ✅ Использование context.Context для таймаутов
- ✅ Валидация всех входных данных
- ✅ Логирование критических ошибок
- ✅ Graceful shutdown сервера
