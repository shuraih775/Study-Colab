package controllers

import (
	"core-service/config"
	"core-service/models"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email" binding:"omitempty,email"`
		Avatar   string `json:"avatar"` // base64 image string
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Avatar != "" {
		if _, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(input.Avatar, "data:image/png;base64,")); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 image"})
			return
		}
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Username = input.Username
	user.Email = input.Email
	user.Avatar = input.Avatar

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetUserProfile(c *gin.Context) {
	userID := c.Param("userId")

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
	})
}

func GetUserGroups(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var groups []models.Group

	err := config.DB.Table("groups").
		Select("groups.*").
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	if groups == nil {
		groups = []models.Group{}
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
	})
}

func GetUserTasks(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var tasks []models.Task

	err := config.DB.Table("tasks").
		Select("tasks.*").
		Joins("JOIN group_members ON group_members.group_id = tasks.group_id").
		Where("group_members.user_id = ?", userID).
		Find(&tasks).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
	})
}

func GetMutualGroups(c *gin.Context) {
	userID1 := c.MustGet("user_id").(uuid.UUID)
	userID2 := c.Param("otherUserId")

	var user1GroupIDs []string
	var user2GroupIDs []string

	config.DB.Model(&models.GroupMember{}).
		Where("user_id = ? AND status = ?", userID1, "active").
		Pluck("group_id", &user1GroupIDs)

	config.DB.Model(&models.GroupMember{}).
		Where("user_id = ? AND status = ?", userID2, "active").
		Pluck("group_id", &user2GroupIDs)

	groupMap := make(map[string]bool)
	for _, id := range user1GroupIDs {
		groupMap[id] = true
	}

	var mutualGroupIDs []string
	for _, id := range user2GroupIDs {
		if groupMap[id] {
			mutualGroupIDs = append(mutualGroupIDs, id)
		}
	}

	var mutualGroups []models.Group
	if len(mutualGroupIDs) > 0 {
		config.DB.Where("id IN ?", mutualGroupIDs).Find(&mutualGroups)
	}

	c.JSON(http.StatusOK, gin.H{"mutual_groups": mutualGroups})
}
