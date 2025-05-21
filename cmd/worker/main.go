package main

import (
	"fmt"
	"img-resizer/internal/config"
	"img-resizer/internal/models"
	"img-resizer/internal/processor"
	"img-resizer/internal/queue"
	"img-resizer/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
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
	defer func() {
		if err := rabbitMQ.Close(); err != nil {
			log.Printf("Failed to close RabbitMQ: %v", err)
		}
	}()

	// Initialize processor
	proc := processor.NewProcessor()

	// Set up signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming tasks in a separate goroutine
	go func() {
		log.Println("Worker started, waiting for tasks...")
		err := rabbitMQ.ConsumeTask(func(task *models.ImageProcessingTask) error {
			return processImage(task, storageProvider, proc)
		})
		if err != nil {
			log.Fatalf("Failed to consume tasks: %v", err)
		}
	}()

	// Wait for termination signal
	<-signals
	log.Println("Shutting down worker...")
}

// processImage processes an image from a task
func processImage(task *models.ImageProcessingTask, storage storage.Storage, proc *processor.Processor) error {
	log.Printf("Processing image: %s", task.ID)

	// Get the original image from storage
	originalImage, err := storage.Get(task.ID, models.QualityOriginal)
	if err != nil {
		return fmt.Errorf("failed to get original image: %w", err)
	}
	defer func() {
		if err := originalImage.Close(); err != nil {
			log.Printf("Failed to close original image: %v", err)
		}
	}()

	// Read the image data
	imageData, err := proc.ReadAll(originalImage)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	// Process the image
	variants, err := proc.ProcessImage(imageData)
	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}

	// Save the processed images
	for quality, data := range variants {
		// Skip the original image as it's already saved
		if quality == models.QualityOriginal {
			continue
		}

		// Save the processed image
		_, err := storage.Save(task.ID, quality, proc.CreateReader(data))
		if err != nil {
			return fmt.Errorf("failed to save processed image with quality %s: %w", quality, err)
		}

		log.Printf("Saved image %s with quality %s", task.ID, quality)
	}

	log.Printf("Image processing completed: %s", task.ID)
	return nil
}
