package main

import (
	"fmt"
	"img-resizer/internal/api"
	"img-resizer/internal/config"
	"img-resizer/internal/queue"
	"img-resizer/internal/storage"
	"log"
)

func main() {

	cfg := config.NewConfig()

	// Initialize storage
	storageProvider, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize RabbitMQ
	rabbitMQ, err := queue.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	// Set up the router
	router := api.SetupRouter(storageProvider, rabbitMQ)

	// Start the server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
