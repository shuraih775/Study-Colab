package routes

import (
	"core-service/controllers"
	"core-service/internal/file"
	"core-service/middlewares"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func RegisterMaterialRoutes(router *gin.Engine) {
	materials := router.Group("/materials")
	materials.Use(middlewares.JWTAuthMiddleware())

	grpcConfig := file.FileClientConfig{
		Addr:        os.Getenv("FILE_SERVICE_ADDR"),
		InternalKey: os.Getenv("FILE_SERVICE_KEY"),
	}

	fileClient, err := file.NewFileClient(grpcConfig)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	fileHandler := &controllers.FileHandler{
		GrpcClient: fileClient,
	}

	materials.POST("/presigned-url", fileHandler.GeneratePresignedURL)
}
