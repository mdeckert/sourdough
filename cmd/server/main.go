package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdeckert/sourdough/internal/ecobee"
	"github.com/mdeckert/sourdough/internal/server"
	"github.com/mdeckert/sourdough/internal/storage"
)

func main() {
	// Get configuration from environment variables
	port := os.Getenv("SOURDOUGH_PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := os.Getenv("SOURDOUGH_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	// Ecobee configuration via Home Assistant (optional)
	haURL := os.Getenv("HA_URL")
	haToken := os.Getenv("HA_TOKEN")
	ecobeeEntity := os.Getenv("ECOBEE_ENTITY")

	// Initialize storage
	store, err := storage.New(dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize Ecobee client (can be disabled)
	ecobeeClient := ecobee.New(haURL, haToken, ecobeeEntity)
	if ecobeeClient.IsEnabled() {
		log.Printf("Ecobee integration enabled via Home Assistant: %s", ecobeeEntity)
	} else {
		log.Printf("Ecobee integration disabled (set HA_URL, HA_TOKEN, and ECOBEE_ENTITY to enable)")
	}

	// Create server
	srv := server.New(store, ecobeeClient, port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		os.Exit(0)
	}()

	// Start server
	log.Printf("Starting Sourdough Server on port %s, data directory: %s", port, dataDir)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
