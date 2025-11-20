package models

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	GroupID     uuid.UUID `json:"group_id" gorm:"type:uuid;not null;index"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	AssignedBy  string    `json:"assigned_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
