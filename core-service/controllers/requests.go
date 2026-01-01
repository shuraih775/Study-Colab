package controllers

import (
	"core-service/config"
	"core-service/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"core-service/internal/observability/logging"
	"core-service/internal/observability/metrics"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var requestTracer = otel.Tracer("controllers.group")

func AcceptMemberRequest(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := tracer.Start(ctx, "group.accept_member")
	defer span.End()

	groupID := c.Param("groupId")
	targetUserID := c.Param("userId")
	currentUserID := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.id", groupID),
		attribute.String("admin.id", currentUserID.String()),
		attribute.String("member.id", targetUserID),
	)

	auditFields := []zap.Field{
		zap.String("group_id", groupID),
		zap.String("admin_id", currentUserID.String()),
		zap.String("target_user_id", targetUserID),
	}

	var admin models.GroupMember
	if err := config.DB.WithContext(ctx).First(
		&admin,
		"group_id = ? AND user_id = ? AND role = ?",
		groupID, currentUserID, "admin",
	).Error; err != nil {

		span.AddEvent("admin_check_failed")
		log.Warn("unauthorized attempt to accept member", auditFields...)
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can accept requests"})
		return
	}

	if err := config.DB.WithContext(ctx).Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, targetUserID, "pending").
		Updates(map[string]interface{}{
			"status":    "joined",
			"joined_at": time.Now(),
		}).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to accept member")

		log.Error("failed to update member status to joined", append(auditFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept member"})
		return
	}

	span.SetStatus(codes.Ok, "member accepted")
	metrics.GroupsJoined.Add(ctx, 1)
	log.Info("member request accepted", auditFields...)
	c.JSON(http.StatusOK, gin.H{"message": "Member accepted"})
}

func RejectMemberRequest(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := requestTracer.Start(ctx, "group.reject_member")
	defer span.End()

	groupID := c.Param("groupId")
	targetUserID := c.Param("userId")
	currentUserID := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.id", groupID),
		attribute.String("admin.id", currentUserID.String()),
		attribute.String("member.id", targetUserID),
	)

	auditFields := []zap.Field{
		zap.String("group_id", groupID),
		zap.String("admin_id", currentUserID.String()),
		zap.String("target_user_id", targetUserID),
	}

	var admin models.GroupMember
	if err := config.DB.WithContext(ctx).First(
		&admin,
		"group_id = ? AND user_id = ? AND role = ?",
		groupID, currentUserID, "admin",
	).Error; err != nil {

		span.AddEvent("admin_check_failed")
		log.Warn("unauthorized attempt to reject member", auditFields...)
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can reject requests"})
		return
	}

	if err := config.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, targetUserID, "pending").
		Delete(&models.GroupMember{}).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to reject member")

		log.Error("failed to delete pending member request", append(auditFields, zap.Error(err))...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject member"})
		return
	}

	span.SetStatus(codes.Ok, "member rejected")
	log.Info("member request rejected", auditFields...)
	c.JSON(http.StatusOK, gin.H{"message": "Member request rejected"})
}

func GetMemberRequests(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := requestTracer.Start(ctx, "group.pending_requests")
	defer span.End()

	groupID := c.Param("groupId")
	currentUserID := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.id", groupID),
		attribute.String("admin.id", currentUserID.String()),
	)

	var admin models.GroupMember
	if err := config.DB.WithContext(ctx).First(
		&admin,
		"group_id = ? AND user_id = ? AND role = ?",
		groupID, currentUserID, "admin",
	).Error; err != nil {

		span.AddEvent("admin_check_failed")
		log.Warn("unauthorized access to member requests",
			zap.String("group_id", groupID),
			zap.String("user_id", currentUserID.String()),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view requests"})
		return
	}

	parsedGroupID, err := uuid.Parse(groupID)
	if err != nil {
		span.AddEvent("invalid_group_id")
		log.Warn("invalid group id format", zap.String("group_id", groupID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var results []struct {
		ID          uuid.UUID `json:"id"`
		User        string    `json:"user"`
		Email       string    `json:"email"`
		RequestedAt time.Time `json:"requestedAt"`
	}

	if err := config.DB.WithContext(ctx).Table("group_members").
		Select("users.id as id, users.username as user, users.email as email, group_members.requested_at").
		Joins("join users on users.id = group_members.user_id").
		Where("group_members.group_id = ? AND group_members.status = ?", parsedGroupID, "pending").
		Find(&results).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch pending requests")

		log.Error("failed to fetch pending member requests",
			zap.String("group_id", groupID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	span.SetAttributes(attribute.Int("requests.count", len(results)))
	span.SetStatus(codes.Ok, "pending requests fetched")

	c.JSON(http.StatusOK, results)
}
