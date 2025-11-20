package routes

import (
	"core-service/controllers"
	"core-service/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.Engine) {

	user := router.Group("/user")
	user.Use(middlewares.JWTAuthMiddleware())
	{
		user.GET("/me", controllers.GetCurrentUser)
		user.GET("/:userId/tasks/urgent", controllers.GetUrgentTasks)
		user.PUT("/me", controllers.UpdateProfile)
		user.PUT("/me/password", controllers.ChangePassword)
		user.DELETE("/me", controllers.DeleteUser)

		user.GET("/:userId/profile", controllers.GetUserProfile)
		user.GET("/users/:userId/mutual-groups", controllers.GetMutualGroups)
		user.GET("/groups", controllers.GetUserGroups)
		user.GET("/tasks", controllers.GetUserTasks)

	}

}
