package main

import (
	"context"
	"fmt"
	"img-resizer/internal/api"
	"img-resizer/internal/config"
	"img-resizer/internal/queue"
	"img-resizer/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.NewConfig()

	// Init storage
	storageProvider, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Init RabbitMQ
	rabbitMQ, err := queue.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer func() {
		if err := rabbitMQ.Close(); err != nil {
			log.Printf("Failed to close RabbitMQ: %v", err)
		}
	}()

	router := api.SetupRouter(storageProvider, rabbitMQ)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
