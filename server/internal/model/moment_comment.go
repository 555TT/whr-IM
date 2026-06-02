package model

import "time"

type MomentComment struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	MomentID  uint64    `gorm:"not null;index" json:"momentId"`
	UserID    uint64    `gorm:"not null;index" json:"userId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func (MomentComment) TableName() string {
	return "moment_comments"
}
