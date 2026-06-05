package model

import "time"

type NormalGroupMessage struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	GroupID   uint64    `gorm:"not null;index:idx_normal_group_messages_group_created_at,priority:1" json:"groupId"`
	SenderID  uint64    `gorm:"not null;index" json:"senderId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"index:idx_normal_group_messages_group_created_at,priority:2" json:"createdAt"`
}

func (NormalGroupMessage) TableName() string {
	return "normal_group_messages"
}
