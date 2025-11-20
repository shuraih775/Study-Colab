package main

import (
	"core-service/config"
	"core-service/models"
	"core-service/routes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()
	err := config.DB.AutoMigrate(&models.User{}, &models.Group{}, &models.GroupMember{}, &models.Task{}, &models.ChatMessage{}, &models.Attachment{})
	if err != nil {
		fmt.Println("automigrate failed")

	}
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000"
		},
		MaxAge: 12 * time.Hour,
	}))

	routes.RegisterAuthRoutes(r)
	routes.RegisterGroupRoutes(r)
	routes.RegisterUserRoutes(r)
	routes.RegisterChatRoutes(r)
	routes.RegisterMaterialRoutes(r)
	r.Static("/uploads", "./uploads")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
