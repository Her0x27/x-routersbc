package main

import (
	"log"
	"os"

	"github.com/Her0x27/x-routersbc/core"
)

func main() {
	// Initialize the server
	server := core.NewServer()
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Start the server
	log.Printf("Starting RouterSBC server on port %s", port)
	if err := server.Start("0.0.0.0:" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
