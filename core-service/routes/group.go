package routes

import (
	"core-service/controllers"
	"core-service/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterGroupRoutes(router *gin.Engine) {
	group := router.Group("/groups")
	group.GET("/", controllers.ListGroups)

	group.Use(middlewares.JWTAuthMiddleware())

	group.POST("", controllers.CreateGroup)

	group.POST("/:groupId/join", controllers.JoinGroup)
	group.PUT(":groupId/requests/:userId", controllers.AcceptMemberRequest)
	group.DELETE(":groupId/requests/:userId", controllers.RejectMemberRequest)
	group.GET("/:groupId", controllers.GetGroupDetails)
	group.GET("/:groupId/members", controllers.GetMembers)
	group.PUT("/:groupId", controllers.UpdateGroupDetails)
	group.GET("/:groupId/user/status", controllers.FetchMembershipStatus)
	group.GET("/:groupId/requests", controllers.GetMemberRequests)
	group.PUT("/:groupId/members/:memberid", controllers.UpdateGroupMember)
	group.POST("/:groupId/tasks", controllers.CreateTask)
	group.GET("/:groupId/tasks", controllers.ListTasks)

}
