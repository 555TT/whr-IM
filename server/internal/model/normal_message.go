package model

import "time"

type NormalMessage struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	SenderID   uint64    `gorm:"not null;index" json:"senderId"`
	ReceiverID uint64    `gorm:"not null;index" json:"receiverId"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (NormalMessage) TableName() string {
	return "normal_messages"
}
