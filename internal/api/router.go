package api

import (
	"img-resizer/internal/api/handlers"
	"img-resizer/internal/queue"
	"img-resizer/internal/storage"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the Gin router
func SetupRouter(storage storage.Storage, queue *queue.RabbitMQ) *gin.Engine {
	// Create a new Gin router
	router := gin.Default()

	// Create handlers
	imageHandler := handlers.NewImageHandler(storage, queue)

	// Set up routes
	api := router.Group("/api")
	{
		// Image routes
		api.POST("/images", imageHandler.UploadImage)
		api.GET("/images/:id", imageHandler.GetImage)
	}

	return router
}
