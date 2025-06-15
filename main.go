package main

import (
        "database/sql"
        "log"
        "os"

        "github.com/labstack/echo/v4"
        "github.com/Her0x27/x-routersbc/core"
        "github.com/Her0x27/x-routersbc/handlers"
        "github.com/Her0x27/x-routersbc/routes"
)

func main() {
        // Initialize database
        db, err := core.InitDatabase()
        if err != nil {
                log.Fatalf("Failed to initialize database: %v", err)
        }
        defer db.Close()

        // Initialize server
        server := core.NewServer(db)
        
        // Setup routes using a function to avoid import cycles
        server.SetupRoutes(func(e *echo.Echo, db *sql.DB) {
                setupAllRoutes(e, db)
        })
        
        // Start server on port 5000
        port := os.Getenv("PORT")
        if port == "" {
                port = "5000"
        }
        
        log.Printf("Starting router SBC management interface on port %s", port)
        if err := server.Start("0.0.0.0:" + port); err != nil {
                log.Fatalf("Failed to start server: %v", err)
        }
}

func setupAllRoutes(e *echo.Echo, db *sql.DB) {
        // Initialize handlers
        authHandler := handlers.NewAuthHandler(db)
        networkHandler := handlers.NewNetworkHandler(db)
        systemHandler := handlers.NewSystemHandler(db)
        
        // Setup routes
        routes.SetupRoutes(e, authHandler, networkHandler, systemHandler)
}
