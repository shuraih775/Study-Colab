package models

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex" json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // public / private
	CreatedBy   uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Members     uint16    `json:"members"`
}

type GroupMember struct {
	GroupID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Status      string    `json:"status"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	RequestedAt time.Time `json:"requested_at"`
}
