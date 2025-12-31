package routes

import (
	"core-service/controllers"
	"core-service/internal/file"
	"core-service/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterMaterialRoutes(router *gin.Engine, client *file.Client) {
	materials := router.Group("/materials")
	materials.Use(middlewares.JWTAuthMiddleware())

	fileHandler := &controllers.FileHandler{
		GrpcClient: client,
	}

	materials.POST("/presigned-url", fileHandler.GeneratePresignedURL)
}
