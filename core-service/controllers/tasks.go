package controllers

import (
	"net/http"
	"time"

	"core-service/config"
	"core-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateTask(c *gin.Context) {
	groupId := c.Param("groupId")
	parsedGroupID := uuid.MustParse(groupId)

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userId, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var body struct {
		Title       string `json:"title" binding:"required"`
		Deadline    string `json:"deadline" binding:"required"`
		Status      string `json:"status" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	const layout = "2006-01-02T15:04"
	deadline, err := time.Parse(layout, body.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format"})
		return
	}

	task := models.Task{
		ID:          uuid.New(),
		GroupID:     parsedGroupID,
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
		Deadline:    deadline,
		AssignedBy:  userId.String(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := config.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func ListTasks(c *gin.Context) {
	groupId := c.Param("groupId")
	sortOrder := c.DefaultQuery("sort", "asc")

	var tasks []models.Task
	order := "deadline asc"
	if sortOrder == "desc" {
		order = "deadline desc"
	}

	if err := config.DB.Where("group_id = ?", groupId).Order(order).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func GetUrgentTasks(c *gin.Context) {
	userId := c.Param("userId")
	now := time.Now()
	threshold := now.Add(48 * time.Hour)

	var tasks []models.Task

	err := config.DB.
		Table("tasks").
		Joins("JOIN group_members gm ON gm.group_id = tasks.group_id").
		Where("gm.user_id = ?", userId).
		Where("tasks.deadline BETWEEN ? AND ?", now, threshold).
		Order("tasks.deadline asc").
		Find(&tasks).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch urgent tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
