package handlers

import (
	"bytes"
	"fmt"
	"img-resizer/internal/models"
	"img-resizer/internal/processor"
	"img-resizer/internal/queue"
	"img-resizer/internal/storage"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageHandler struct {
	storage   storage.Storage
	queue     *queue.RabbitMQ
	processor *processor.Processor
}

func NewImageHandler(storage storage.Storage, queue *queue.RabbitMQ) *ImageHandler {
	return &ImageHandler{
		storage:   storage,
		queue:     queue,
		processor: processor.NewProcessor(),
	}
}

// UploadImage handles image upload requests
func (h *ImageHandler) UploadImage(c *gin.Context) {
	// Get the file from the request
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image provided"})
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close file"})
		}
	}()

	if !isImage(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is not an image"})
		return
	}

	id := uuid.New().String()

	// Read the image data
	imageData, err := h.processor.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}

	// Save the original image
	_, err = h.storage.Save(id, models.QualityOriginal, bytes.NewReader(imageData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Create a task for processing the image
	task := &models.ImageProcessingTask{
		ID:       id,
		FilePath: filepath.Join("storage", id[:2], fmt.Sprintf("%s_%s.jpg", id, models.QualityOriginal)),
	}

	// Publish the task to the queue
	err = h.queue.PublishTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue image for processing"})
		return
	}

	// Return the image ID
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "Image uploaded successfully and queued for processing",
	})
}

// GetImage handles image retrieval requests
func (h *ImageHandler) GetImage(c *gin.Context) {
	// Get the image ID from the URL
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image ID provided"})
		return
	}

	// Get the quality from the query parameters
	qualityStr := c.DefaultQuery("quality", string(models.QualityOriginal))
	quality := models.ImageQuality(qualityStr)

	// Validate the quality
	switch quality {
	case models.QualityOriginal, models.QualityHigh, models.QualityMedium, models.QualityLow:
		// Valid quality
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quality parameter"})
		return
	}

	// Get the image from storage
	image, err := h.storage.Get(id, quality)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}
	defer func() {
		if err := image.Close(); err != nil {
			log.Printf("failed to close image: %v", err)
		}
	}()

	// Set the content type
	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=31536000")

	// Stream the image to the response
	_, err = io.Copy(c.Writer, image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream image"})
		return
	}
}

// check files for ex
func isImage(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
