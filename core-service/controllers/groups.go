package controllers

import (
	"core-service/config"
	"core-service/internal/observability/logging"
	"core-service/internal/observability/metrics"
	"core-service/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var grouptracer = otel.Tracer("controllers.group")

func CreateGroup(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.create")
	defer span.End()

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		span.AddEvent("invalid_payload")
		log.Warn("invalid create group payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.name", body.Name),
		attribute.String("group.type", body.Type),
		attribute.String("user.id", userID.String()),
	)

	var existing models.Group
	if err := config.DB.WithContext(ctx).First(&existing, "name = ?", body.Name).Error; err == nil {
		span.AddEvent("group_name_conflict")
		log.Warn("group name already exists", zap.String("group_name", body.Name))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already exists"})
		return
	}

	groupID := uuid.New()

	if err := config.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`INSERT INTO groups (id, name, description, type, created_by, created_at, members)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			groupID, body.Name, body.Description, body.Type, userID, time.Now(), 1,
		).Error; err != nil {
			return err
		}

		if err := tx.Exec(
			`INSERT INTO group_members (user_id, group_id, status, role, joined_at)
			 VALUES (?, ?, 'joined', 'admin', ?)`,
			userID, groupID, time.Now(),
		).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "group creation failed")

		log.Error("failed to create group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	span.SetStatus(codes.Ok, "group created")
	metrics.GroupsCreated.Add(ctx, 1)

	log.Info(
		"group created",
		zap.String("group_id", groupID.String()),
		zap.String("user_id", userID.String()),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Group created",
		"group_id": groupID,
	})
}

func ListGroups(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.list")
	defer span.End()

	var groups []models.Group
	if err := config.DB.WithContext(ctx).Find(&groups).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list groups")
		log.Error("failed to list groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	span.SetAttributes(attribute.Int("groups.count", len(groups)))
	span.SetStatus(codes.Ok, "groups listed")

	c.JSON(http.StatusOK, groups)
}

func JoinGroup(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.join")
	defer span.End()

	userID := c.MustGet("user_id").(uuid.UUID)
	groupIDStr := c.Param("groupId")

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		span.AddEvent("invalid_group_id")
		log.Warn("invalid group_id", zap.String("group_id", groupIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
	)

	var group models.Group
	if err := config.DB.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		span.AddEvent("group_not_found")
		log.Warn("group not found", zap.String("group_id", groupID.String()))
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	var existing models.GroupMember
	if err := config.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&existing).Error; err == nil {

		span.AddEvent("already_member_or_pending")
		log.Warn("user already member or pending", zap.String("group_id", groupID.String()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "already a member or request exists"})
		return
	}

	member := models.GroupMember{
		UserID:      userID,
		GroupID:     groupID,
		RequestedAt: time.Now(),
		Role:        "member",
	}

	switch group.Type {
	case "public":
		member.Status = "joined"
		member.JoinedAt = time.Now()

		if err := config.DB.WithContext(ctx).Create(&member).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to join public group")
			log.Error("failed to join public group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join group"})
			return
		}

		if err := config.DB.WithContext(ctx).Model(&group).
			UpdateColumn("members", gorm.Expr("members + 1")).Error; err != nil {

			span.RecordError(err)
			span.SetStatus(codes.Error, "member count update failed")
			log.Error("failed to update member count", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update member count"})
			return
		}

		span.SetStatus(codes.Ok, "joined public group")
		metrics.GroupsJoined.Add(ctx, 1)
		c.JSON(http.StatusOK, gin.H{"message": "joined group"})

	case "private":
		member.Status = "pending"

		if err := config.DB.WithContext(ctx).Create(&member).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to request join")
			log.Error("failed to create join request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "request failed"})
			return
		}

		span.SetStatus(codes.Ok, "join request sent")
		metrics.GroupsJoinRequest.Add(ctx, 1)
		c.JSON(http.StatusOK, gin.H{"message": "join request sent"})

	default:
		span.AddEvent("unknown_group_type")
		log.Error("unknown group type", zap.String("group_type", group.Type))
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown group type"})
	}
}

func GetGroupDetails(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.details")
	defer span.End()

	groupIDStr := c.Param("groupId")
	span.SetAttributes(attribute.String("group.id", groupIDStr))

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		span.AddEvent("invalid_group_id")
		log.Warn("invalid group_id", zap.String("group_id", groupIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	var group models.Group
	if err := config.DB.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		span.AddEvent("group_not_found")
		log.Warn("group not found", zap.String("group_id", groupID.String()))
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	var members []struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Email    string    `json:"email"`
		Role     string    `json:"role"`
	}

	if err := config.DB.WithContext(ctx).Table("users").
		Select("users.id, users.username, users.email, group_members.role").
		Joins("JOIN group_members ON group_members.user_id = users.id").
		Where("group_members.group_id = ? AND group_members.status = ?", groupID, "joined").
		Scan(&members).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch members")
		log.Error("failed to fetch group members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch members"})
		return
	}

	span.SetAttributes(attribute.Int("members.count", len(members)))
	span.SetStatus(codes.Ok, "group details fetched")

	c.JSON(http.StatusOK, gin.H{
		"group_id":     group.ID.String(),
		"name":         group.Name,
		"description":  group.Description,
		"type":         group.Type,
		"created_by":   group.CreatedBy.String(),
		"created_at":   group.CreatedAt,
		"members":      members,
		"member_count": len(members),
	})
}

func GetMembers(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.get_members")
	defer span.End()

	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		span.AddEvent("invalid_group_id")
		log.Warn("invalid group_id", zap.String("group_id", groupIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	currentUserID := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", currentUserID.String()),
	)

	var group models.Group
	if err := config.DB.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		span.AddEvent("group_not_found")
		log.Warn("group not found", zap.String("group_id", groupID.String()))
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var membership models.GroupMember
	if err := config.DB.WithContext(ctx).First(
		&membership,
		"group_id = ? AND user_id = ? AND status = ?",
		groupID, currentUserID, "joined",
	).Error; err != nil {

		span.AddEvent("non_member_access")
		log.Warn("non-member attempted to view members",
			zap.String("group_id", groupID.String()),
			zap.String("user_id", currentUserID.String()),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be a member of this group"})
		return
	}

	var groupMembers []struct {
		UserID   uuid.UUID
		Status   string
		Role     string
		JoinedAt time.Time
	}

	if err := config.DB.WithContext(ctx).Table("group_members").
		Select("user_id, status, role, joined_at").
		Where("group_id = ? AND status = ?", groupID, "joined").
		Order("joined_at ASC").
		Scan(&groupMembers).Error; err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch group members")
		log.Error("failed to fetch group members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group members"})
		return
	}

	userIDs := make([]uuid.UUID, 0, len(groupMembers))
	for _, m := range groupMembers {
		userIDs = append(userIDs, m.UserID)
	}

	var users []struct {
		ID       uuid.UUID
		Username string
		Email    string
	}

	if len(userIDs) > 0 {
		if err := config.DB.WithContext(ctx).Table("users").
			Select("id, username, email").
			Where("id IN ?", userIDs).
			Scan(&users).Error; err != nil {

			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to fetch users")
			log.Error("failed to fetch user details", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user details"})
			return
		}
	}

	userMap := make(map[uuid.UUID]struct {
		Username string
		Email    string
	})
	for _, u := range users {
		userMap[u.ID] = struct {
			Username string
			Email    string
		}{u.Username, u.Email}
	}

	type MemberInfo struct {
		ID       uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Email    string    `json:"email"`
		Role     string    `json:"role"`
		Status   string    `json:"status"`
		JoinedAt time.Time `json:"joined_at"`
	}

	var members []MemberInfo
	for _, gm := range groupMembers {
		if u, ok := userMap[gm.UserID]; ok {
			members = append(members, MemberInfo{
				ID:       gm.UserID,
				Username: u.Username,
				Email:    u.Email,
				Role:     gm.Role,
				Status:   gm.Status,
				JoinedAt: gm.JoinedAt,
			})
		}
	}

	span.SetAttributes(attribute.Int("members.count", len(members)))
	span.SetStatus(codes.Ok, "members fetched")

	c.JSON(http.StatusOK, gin.H{
		"group_id": group.ID,
		"members":  members,
		"count":    len(members),
	})
}

func FetchMembershipStatus(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.membership_status")
	defer span.End()

	userID := c.MustGet("user_id").(uuid.UUID)
	groupIDStr := c.Param("groupId")

	span.SetAttributes(
		attribute.String("group.id", groupIDStr),
		attribute.String("user.id", userID.String()),
	)

	var group models.Group
	if err := config.DB.WithContext(ctx).First(&group, "id = ?", groupIDStr).Error; err != nil {
		span.AddEvent("group_not_found")
		log.Warn("group not found", zap.String("group_id", groupIDStr))
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}

	var member models.GroupMember
	err := config.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", group.ID, userID).
		First(&member).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			span.AddEvent("not_a_member")
			c.JSON(http.StatusOK, gin.H{"status": "not_a_member"})
			return
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch membership status")
		log.Error("failed to fetch membership status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	span.SetStatus(codes.Ok, "membership status fetched")
	c.JSON(http.StatusOK, gin.H{"status": member.Status})
}

func UpdateGroupMember(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.update_member")
	defer span.End()

	groupIDStr := c.Param("groupId")
	memberIDStr := c.Param("memberid")

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		span.AddEvent("invalid_group_id")
		log.Warn("invalid group_id", zap.String("group_id", groupIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		span.AddEvent("invalid_member_id")
		log.Warn("invalid member_id", zap.String("member_id", memberIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	currentUserID := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("admin.id", currentUserID.String()),
		attribute.String("member.id", memberID.String()),
	)

	var admin models.GroupMember
	if err := config.DB.WithContext(ctx).First(
		&admin,
		"group_id = ? AND user_id = ? AND role = ?",
		groupID, currentUserID, "admin",
	).Error; err != nil {

		span.AddEvent("admin_check_failed")
		log.Warn("non-admin attempted member update")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update members"})
		return
	}

	if currentUserID == memberID {
		span.AddEvent("self_modification_attempt")
		c.JSON(http.StatusForbidden, gin.H{"error": "Admins cannot modify themselves"})
		return
	}

	var body struct {
		Role   *string `json:"role,omitempty"`
		Status *string `json:"status,omitempty"`
	}

	if err := c.BindJSON(&body); err != nil {
		span.AddEvent("invalid_payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if body.Role == nil && body.Status == nil {
		span.AddEvent("no_fields_to_update")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	var member models.GroupMember
	if err := config.DB.WithContext(ctx).First(
		&member,
		"group_id = ? AND user_id = ?",
		groupID, memberID,
	).Error; err != nil {

		span.AddEvent("member_not_found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	prevStatus := member.Status

	tx := config.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, "transaction begin failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	updates := map[string]interface{}{}
	if body.Role != nil {
		updates["role"] = *body.Role
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}

	if err := tx.Model(&member).Updates(updates).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "member update failed")
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member"})
		return
	}

	if prevStatus == "pending" && body.Status != nil && *body.Status == "joined" {
		if err := tx.Model(&models.Group{}).
			Where("id = ?", groupID).
			UpdateColumn("members", gorm.Expr("members + 1")).Error; err != nil {

			span.RecordError(err)
			span.SetStatus(codes.Error, "member count increment failed")
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member count"})
			return
		}
	}

	if prevStatus == "joined" && body.Status != nil && *body.Status != "joined" {
		if err := tx.Model(&models.Group{}).
			Where("id = ?", groupID).
			UpdateColumn("members", gorm.Expr("members - 1")).Error; err != nil {

			span.RecordError(err)
			span.SetStatus(codes.Error, "member count decrement failed")
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member count"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction commit failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	span.SetStatus(codes.Ok, "member updated")

	log.Info("group member updated",
		zap.String("group_id", groupID.String()),
		zap.String("member_id", memberID.String()),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Member updated"})
}

func UpdateGroupDetails(c *gin.Context) {
	ctx := c.Request.Context()
	log := logging.Logger(ctx)

	ctx, span := grouptracer.Start(ctx, "group.update_details")
	defer span.End()

	groupIDStr := c.Param("groupId")

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"` // public/private
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		span.AddEvent("invalid_payload")
		log.Warn("invalid update group payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	span.SetAttributes(
		attribute.String("group.id", groupIDStr),
		attribute.String("user.id", userID.String()),
	)

	var group models.Group
	if err := config.DB.WithContext(ctx).First(&group, "id = ?", groupIDStr).Error; err != nil {
		span.AddEvent("group_not_found")
		log.Warn("group not found", zap.String("group_id", groupIDStr))
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var member models.GroupMember
	if err := config.DB.WithContext(ctx).First(
		&member,
		"group_id = ? AND user_id = ? AND role = 'admin'",
		groupIDStr, userID,
	).Error; err != nil {

		span.AddEvent("admin_check_failed")
		log.Warn(
			"non-admin attempted group update",
			zap.String("group_id", groupIDStr),
			zap.String("user_id", userID.String()),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this group"})
		return
	}

	if body.Name != "" && body.Name != group.Name {
		var existing models.Group
		if err := config.DB.WithContext(ctx).First(&existing, "name = ?", body.Name).Error; err == nil {
			span.AddEvent("group_name_conflict")
			log.Warn("group name already exists", zap.String("group_name", body.Name))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already exists"})
			return
		}
	}

	updates := map[string]interface{}{}
	if body.Name != "" {
		updates["name"] = body.Name
	}
	if body.Description != "" {
		updates["description"] = body.Description
	}
	if body.Type != "" {
		updates["type"] = body.Type
	}

	if len(updates) == 0 {
		span.AddEvent("no_fields_to_update")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	if err := config.DB.WithContext(ctx).Model(&group).Updates(updates).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update group details")

		log.Error(
			"failed to update group details",
			zap.String("group_id", groupIDStr),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	span.SetStatus(codes.Ok, "group details updated")

	c.JSON(http.StatusOK, gin.H{
		"message": "Group details updated successfully",
		"group":   group,
	})
}
