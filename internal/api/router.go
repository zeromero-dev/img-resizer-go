package api

import (
	"img-resizer/internal/api/handlers"
	"img-resizer/internal/queue"
	"img-resizer/internal/storage"

	"github.com/gin-gonic/gin"
)

func SetupRouter(storage storage.Storage, queue *queue.RabbitMQ) *gin.Engine {
	router := gin.Default()

	imageHandler := handlers.NewImageHandler(storage, queue)

	api := router.Group("/api")
	{
		api.POST("/images", imageHandler.UploadImage)
		api.GET("/images/:id", imageHandler.GetImage)
	}

	return router
}
