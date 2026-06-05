package model

import "time"

type AIChatMessage struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"not null;index" json:"userId"`
	Role      string    `gorm:"size:20;not null;index" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func (AIChatMessage) TableName() string {
	return "ai_chat_messages"
}
