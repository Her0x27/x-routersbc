package main

import (
	"log"
	"os"

	"github.com/Her0x27/x-routersbc/core"
)

func main() {
	// Initialize database
	if err := core.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create and configure server
	server := core.NewServer()
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Start server
	log.Printf("Starting RouterSBC server on port %s", port)
	if err := server.Start("0.0.0.0:" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
