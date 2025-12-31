package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatMessage struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key"`

	RoomID string `gorm:"type:varchar(255);index:idx_room_timestamp,priority:1"`

	UserID uuid.UUID `gorm:"type:uuid;index"`
	User   User      `gorm:"foreignKey:UserID"`

	Text string `gorm:"type:text"`

	Timestamp time.Time `gorm:"index:idx_room_timestamp,priority:2"`

	Attachments []Attachment `gorm:"foreignKey:ChatMessageID;constraint:OnDelete:CASCADE;"`
}

type Attachment struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	ChatMessageID uuid.UUID `gorm:"type:uuid;index"`

	FileID string `gorm:"type:text;not null"`

	FileName string `gorm:"type:text"`
	FileType string `gorm:"type:varchar(50)"`
	FileSize int64  `gorm:"type:bigint"`

	CreatedAt time.Time
}

func (msg *ChatMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	return
}

func (att *Attachment) BeforeCreate(tx *gorm.DB) (err error) {
	if att.ID == uuid.Nil {
		att.ID = uuid.New()
	}
	if att.CreatedAt.IsZero() {
		att.CreatedAt = time.Now()
	}
	return
}
