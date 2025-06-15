package main

import (
	"log"
	"os"

	"github.com/Her0x27/x-routersbc/core"
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
