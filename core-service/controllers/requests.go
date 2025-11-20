package controllers

import (
	"core-service/config"
	"core-service/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AcceptMemberRequest(c *gin.Context) {
	groupID := c.Param("groupId")
	targetUserID := c.Param("userId")
	currentUserID := c.MustGet("user_id").(uuid.UUID)

	var admin models.GroupMember
	if err := config.DB.First(&admin, "group_id = ? AND user_id = ? AND role = ?", groupID, currentUserID, "admin").Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can accept requests"})
		return
	}

	if err := config.DB.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ? AND status = ?", groupID, targetUserID, "pending").
		Update("status", "joined").Update("joined_at", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member accepted"})
}
func RejectMemberRequest(c *gin.Context) {
	groupID := c.Param("groupId")
	targetUserID := c.Param("userId")
	currentUserID := c.MustGet("user_id").(uuid.UUID)

	var admin models.GroupMember
	if err := config.DB.First(&admin, "group_id = ? AND user_id = ? AND role = ?", groupID, currentUserID, "admin").Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can reject requests"})
		return
	}

	if err := config.DB.Where("group_id = ? AND user_id = ? AND status = ?", groupID, targetUserID, "pending").
		Delete(&models.GroupMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member request rejected"})
}

func GetMemberRequests(c *gin.Context) {
	groupID := c.Param("groupId")
	currentUserID := c.MustGet("user_id").(uuid.UUID)

	var admin models.GroupMember
	if err := config.DB.First(&admin, "group_id = ? AND user_id = ? AND role = ?", groupID, currentUserID, "admin").Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view requests"})
		return
	}

	type MemberRequestResponse struct {
		ID          uuid.UUID `json:"id"`
		User        string    `json:"user"`
		Email       string    `json:"email"`
		RequestedAt time.Time `json:"requestedAt"`
	}

	var results []MemberRequestResponse
	parsedGroupID := uuid.MustParse(groupID)
	if err := config.DB.Table("group_members").
		Select("users.id as id, users.username as user, users.email as email, group_members.requested_at").
		Joins("join users on users.id = group_members.user_id").
		Where("group_members.group_id = ? AND group_members.status = ?", parsedGroupID, "pending").
		Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
		return
	}

	// Return the joined results
	c.JSON(http.StatusOK, results)
}
