package controllers

import (
	"net/http"
	"time"

	"core-service/config"
	"core-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"core-service/internal/observability/logging"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var taskTracer = otel.Tracer("controllers.task")

func CreateTask(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := taskTracer.Start(ctx, "task.create")
	defer span.End()

	groupId := c.Param("groupId")
	parsedGroupID, err := uuid.Parse(groupId)
	if err != nil {
		span.AddEvent("invalid_group_id")
		log.Warn("invalid group_id", zap.String("group_id", groupId))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	userId := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.id", parsedGroupID.String()),
		attribute.String("user.id", userId.String()),
	)

	var body struct {
		Title       string `json:"title" binding:"required"`
		Deadline    string `json:"deadline" binding:"required"`
		Status      string `json:"status" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		span.AddEvent("invalid_payload")
		log.Warn("invalid create task payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	const layout = "2006-01-02T15:04"
	deadline, err := time.Parse(layout, body.Deadline)
	if err != nil {
		span.AddEvent("invalid_deadline_format")
		log.Warn("invalid deadline format", zap.String("deadline", body.Deadline))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deadline format"})
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

	span.SetAttributes(attribute.String("task.id", task.ID.String()))

	if err := config.DB.WithContext(ctx).Create(&task).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "task creation failed")

		log.Error(
			"failed to create task",
			zap.String("task_id", task.ID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	span.SetStatus(codes.Ok, "task created")

	log.Info(
		"task created",
		zap.String("task_id", task.ID.String()),
		zap.String("group_id", parsedGroupID.String()),
		zap.String("user_id", userId.String()),
	)

	c.JSON(http.StatusCreated, task)
}

func ListTasks(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := taskTracer.Start(ctx, "task.list")
	defer span.End()

	groupId := c.Param("groupId")
	sortOrder := c.DefaultQuery("sort", "asc")

	order := "deadline asc"
	if sortOrder == "desc" {
		order = "deadline desc"
	}

	span.SetAttributes(
		attribute.String("group.id", groupId),
		attribute.String("sort", sortOrder),
	)

	var tasks []models.Task
	if err := config.DB.WithContext(ctx).
		Where("group_id = ?", groupId).
		Order(order).
		Find(&tasks).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list tasks")

		log.Error("failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tasks"})
		return
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	span.SetStatus(codes.Ok, "tasks listed")

	c.JSON(http.StatusOK, tasks)
}

func GetUrgentTasks(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := taskTracer.Start(ctx, "task.urgent")
	defer span.End()

	userId := c.Param("userId")
	now := time.Now()
	threshold := now.Add(48 * time.Hour)

	span.SetAttributes(
		attribute.String("user.id", userId),
		attribute.String("window", "48h"),
	)

	var tasks []models.Task

	err := config.DB.WithContext(ctx).
		Table("tasks").
		Joins("JOIN group_members gm ON gm.group_id = tasks.group_id").
		Where("gm.user_id = ?", userId).
		Where("tasks.deadline BETWEEN ? AND ?", now, threshold).
		Order("tasks.deadline asc").
		Find(&tasks).Error

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "urgent task query failed")

		log.Error(
			"failed to fetch urgent tasks",
			zap.String("user_id", userId),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch urgent tasks"})
		return
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	span.SetStatus(codes.Ok, "urgent tasks fetched")

	c.JSON(http.StatusOK, tasks)
}
