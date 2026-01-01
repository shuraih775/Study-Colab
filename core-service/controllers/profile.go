package controllers

import (
	"core-service/config"
	"core-service/models"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"core-service/internal/observability/logging"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var profileTracer = otel.Tracer("controllers.user")

func UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := tracer.Start(ctx, "user.update_profile")
	defer span.End()

	userID := c.MustGet("user_id").(uuid.UUID)
	span.SetAttributes(attribute.String("user.id", userID.String()))

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email" binding:"omitempty,email"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		span.AddEvent("invalid_payload")
		log.Warn("invalid update profile payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if input.Avatar != "" {
		if _, err := base64.StdEncoding.DecodeString(
			strings.TrimPrefix(input.Avatar, "data:image/png;base64,"),
		); err != nil {
			span.AddEvent("invalid_avatar_encoding")
			log.Warn("invalid avatar encoding", zap.String("user_id", userID.String()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid avatar image"})
			return
		}
	}

	var user models.User
	if err := config.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		span.AddEvent("user_not_found")
		log.Warn("user not found during profile update", zap.String("user_id", userID.String()))
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.Username = input.Username
	user.Email = input.Email
	user.Avatar = input.Avatar

	if err := config.DB.WithContext(ctx).Save(&user).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db update failed")

		log.Error(
			"failed to update user profile",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	span.SetStatus(codes.Ok, "profile updated")

	log.Info("user profile updated", zap.String("user_id", userID.String()))
	c.JSON(http.StatusOK, user)
}

func GetUserProfile(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := materialTracer.Start(ctx, "user.get_profile")
	defer span.End()

	userID := c.Param("userId")
	span.SetAttributes(attribute.String("user.id", userID))

	var user models.User
	if err := config.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		span.AddEvent("user_not_found")
		log.Warn("user profile not found", zap.String("user_id", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	span.SetStatus(codes.Ok, "profile fetched")

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
	})
}

func GetUserGroups(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := materialTracer.Start(ctx, "user.groups")
	defer span.End()

	userID := c.MustGet("user_id").(uuid.UUID)
	span.SetAttributes(attribute.String("user.id", userID.String()))

	var groups []models.Group
	err := config.DB.WithContext(ctx).Table("groups").
		Select("groups.*").
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Find(&groups).Error

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch user groups")

		log.Error("failed to fetch user groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	if groups == nil {
		groups = []models.Group{}
	}

	span.SetAttributes(attribute.Int("groups.count", len(groups)))
	span.SetStatus(codes.Ok, "groups fetched")

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func GetUserTasks(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := materialTracer.Start(ctx, "user.tasks")
	defer span.End()

	userID := c.MustGet("user_id").(uuid.UUID)
	span.SetAttributes(attribute.String("user.id", userID.String()))

	var tasks []models.Task
	err := config.DB.WithContext(ctx).Table("tasks").
		Select("tasks.*").
		Joins("JOIN group_members ON group_members.group_id = tasks.group_id").
		Where("group_members.user_id = ?", userID).
		Find(&tasks).Error

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch user tasks")

		log.Error("failed to fetch user tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "database error"})
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))
	span.SetStatus(codes.Ok, "tasks fetched")

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func GetMutualGroups(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := materialTracer.Start(ctx, "user.mutual_groups")
	defer span.End()

	userID1 := c.MustGet("user_id").(uuid.UUID)
	userID2 := c.Param("otherUserId")

	span.SetAttributes(
		attribute.String("user.id", userID1.String()),
		attribute.String("other_user.id", userID2),
	)

	var user1GroupIDs []string
	var user2GroupIDs []string

	if err := config.DB.WithContext(ctx).Model(&models.GroupMember{}).
		Where("user_id = ? AND status = ?", userID1, "active").
		Pluck("group_id", &user1GroupIDs).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch user groups")
		log.Error("failed to fetch groups for user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if err := config.DB.WithContext(ctx).Model(&models.GroupMember{}).
		Where("user_id = ? AND status = ?", userID2, "active").
		Pluck("group_id", &user2GroupIDs).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch other user groups")
		log.Error("failed to fetch groups for other user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

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
		if err := config.DB.WithContext(ctx).Where("id IN ?", mutualGroupIDs).Find(&mutualGroups).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to fetch mutual groups")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
	}

	span.SetAttributes(attribute.Int("mutual_groups.count", len(mutualGroups)))
	span.SetStatus(codes.Ok, "mutual groups fetched")

	c.JSON(http.StatusOK, gin.H{"mutual_groups": mutualGroups})
}
