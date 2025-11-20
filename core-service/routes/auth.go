package routes

import (
	"core-service/controllers"

	"core-service/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine) *gin.Engine {

	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)

		auth.Use(middlewares.JWTAuthMiddleware())

		auth.POST("/logout", controllers.Logout)
	}

	return r
}
