package routes

import (
	"core-service/controllers"
	"core-service/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterChatRoutes(router *gin.Engine) {
	chat := router.Group("/chat")
	chat.Use(middlewares.JWTAuthMiddleware())

	chat.GET("", controllers.HandleConnection)
}
