package model

import "time"

type MomentLike struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	MomentID  uint64    `gorm:"not null;uniqueIndex:uk_moment_likes_moment_user,priority:1;index" json:"momentId"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_moment_likes_moment_user,priority:2;index" json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (MomentLike) TableName() string {
	return "moment_likes"
}
