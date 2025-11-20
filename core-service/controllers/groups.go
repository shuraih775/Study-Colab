package controllers

import (
	"core-service/config"
	"core-service/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateGroup(c *gin.Context) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

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

	groupID := uuid.New()

	var existing models.Group
	if err := config.DB.First(&existing, "name = ?", body.Name).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already exists"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`INSERT INTO groups (id, name, description, type, created_by, created_at,members)
			VALUES (?, ?, ?, ?, ?, ?,?)`,
			groupID, body.Name, body.Description, body.Type, userId, time.Now(), 1).Error; err != nil {
			return err
		}

		if err := tx.Exec(`INSERT INTO group_members (user_id, group_id, status, role, joined_at)
			VALUES (?, ?,'joined', 'admin', ?)`, userId, groupID, time.Now()).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group and membership"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Group created",
		"group_id": groupID,
	})
}

func ListGroups(c *gin.Context) {
	var groups []models.Group

	if err := config.DB.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}
func JoinGroup(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	groupIDStr := c.Param("groupId")

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var existing models.GroupMember
	if err := config.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already a member or pending request exists"})
		return
	}

	member := models.GroupMember{
		UserID:      userID,
		GroupID:     groupID,
		RequestedAt: time.Now(),
	}

	if group.Type == "public" {
		member.JoinedAt = time.Now()
		member.Status = "joined"
		member.Role = "member"

		if err := config.DB.Create(&member).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join group"})
			return
		}

		if err := config.DB.Model(&group).UpdateColumn("members", gorm.Expr("members + ?", 1)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member count"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Joined group"})

	} else if group.Type == "private" {
		member.Status = "pending"
		member.Role = "member"

		if err := config.DB.Create(&member).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Request failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Join request sent"})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown group type"})
	}
}

func GetGroupDetails(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}
	var group models.Group

	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	type MemberInfo struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	var members []MemberInfo
	config.DB.Table("users").
		Select("users.id, users.username, users.email, group_members.role").
		Joins("JOIN group_members ON group_members.user_id = users.id").
		Where("group_members.group_id = ? AND group_members.status = ?", groupID, "joined").
		Scan(&members)

	memberCount := len(members)

	c.JSON(http.StatusOK, gin.H{
		"group_id":     group.ID.String(),
		"name":         group.Name,
		"description":  group.Description,
		"type":         group.Type,
		"created_by":   group.CreatedBy.String(),
		"created_at":   group.CreatedAt,
		"members":      members,
		"member_count": memberCount,
	})
}

func GetMembers(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	currentUserID := c.MustGet("user_id").(uuid.UUID)

	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var membership models.GroupMember
	if err := config.DB.First(&membership, "group_id = ? AND user_id = ? AND status = ?", groupID, currentUserID, "joined").Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be a member of this group to view its members"})
		return
	}

	type GroupMember struct {
		UserID   uuid.UUID `json:"user_id"`
		Status   string    `json:"status"`
		Role     string    `json:"role"`
		JoinedAt time.Time `json:"joined_at"`
	}

	var groupMembers []GroupMember

	if err := config.DB.Table("group_members").
		Select("user_id, status, role, joined_at").
		Where("group_id = ? AND status = ?", groupID, "joined").
		Order("joined_at ASC").
		Scan(&groupMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group members"})
		return
	}

	userIDs := make([]uuid.UUID, len(groupMembers))
	for i, m := range groupMembers {
		userIDs[i] = m.UserID
	}

	type User struct {
		ID       uuid.UUID
		Username string
		Email    string
	}

	var users []User
	if len(userIDs) > 0 {
		if err := config.DB.Table("users").
			Select("id, username, email").
			Where("id IN ?", userIDs).
			Scan(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user details"})
			return
		}
	}

	userMap := make(map[uuid.UUID]User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	type MemberInfo struct {
		UserID   uuid.UUID `json:"id"`
		Username string    `json:"username"`
		Email    string    `json:"email"`
		Role     string    `json:"role"`
		Status   string    `json:"status"`
		JoinedAt time.Time `json:"joined_at"`
	}

	var members []MemberInfo
	for _, gm := range groupMembers {
		if user, ok := userMap[gm.UserID]; ok {
			members = append(members, MemberInfo{
				UserID:   gm.UserID,
				Username: user.Username,
				Email:    user.Email,
				Role:     gm.Role,
				Status:   gm.Status,
				JoinedAt: gm.JoinedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"group_id": group.ID,
		"members":  members,
		"count":    len(members),
	})
}

func FetchMembershipStatus(c *gin.Context) {
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

	groupId := c.Param("groupId")
	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var member models.GroupMember
	err := config.DB.
		Where("group_id = ? AND user_id = ?", group.ID, userID).
		First(&member).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"status": "not_a_member"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": member.Status})
}
func UpdateGroupMember(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	memberIDStr := c.Param("memberid")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	currentUserID := c.MustGet("user_id").(uuid.UUID)

	var admin models.GroupMember
	if err := config.DB.First(&admin, "group_id = ? AND user_id = ? AND role = ?", groupID, currentUserID, "admin").Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update members"})
		return
	}

	if currentUserID == memberID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admins cannot modify their own role or status"})
		return
	}

	var body struct {
		Role   *string `json:"role,omitempty"`
		Status *string `json:"status,omitempty"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if body.Role == nil && body.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	var member models.GroupMember
	if err := config.DB.First(&member, "group_id = ? AND user_id = ?", groupID, memberID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find member"})
		return
	}

	prevStatus := member.Status

	tx := config.DB.Begin()
	if tx.Error != nil {
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
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member"})
		return
	}

	if prevStatus == "pending" && body.Status != nil && *body.Status == "joined" {
		if err := tx.Model(&models.Group{}).Where("id = ?", groupID).
			UpdateColumn("members", gorm.Expr("members + ?", 1)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group member count (increment)"})
			return
		}
	}

	if prevStatus == "joined" && body.Status != nil && *body.Status != "joined" {
		if err := tx.Model(&models.Group{}).Where("id = ?", groupID).
			UpdateColumn("members", gorm.Expr("members - ?", 1)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group member count (decrement)"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member updated"})
}

func UpdateGroupDetails(c *gin.Context) {
	groupID := c.Param("groupId")

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"` // public/private
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

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

	var group models.Group
	if err := config.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var member models.GroupMember
	if err := config.DB.First(&member, "group_id = ? AND user_id = ? AND role = 'admin'", groupID, userId).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this group"})
		return
	}

	if body.Name != "" && body.Name != group.Name {
		var existing models.Group
		if err := config.DB.First(&existing, "name = ?", body.Name).Error; err == nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	if err := config.DB.Model(&group).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Group details updated successfully",
		"group":   group,
	})
}
