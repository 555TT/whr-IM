package model

import "time"

type Favorite struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	UserID           uint64    `gorm:"not null;index:idx_favorites_user_favorited" json:"userId"`
	SourceType       string    `gorm:"size:20;not null;uniqueIndex:uk_favorites_user_source,priority:2" json:"sourceType"`
	SourceMessageID  uint64    `gorm:"not null;uniqueIndex:uk_favorites_user_source,priority:3" json:"sourceMessageId"`
	ConversationID   uint64    `gorm:"not null;default:0" json:"conversationId"`
	SenderID         uint64    `gorm:"not null" json:"senderId"`
	SenderName       string    `gorm:"size:100;not null" json:"senderName"`
	Content          string    `gorm:"type:text;not null" json:"content"`
	MessageCreatedAt time.Time `gorm:"not null" json:"messageCreatedAt"`
	FavoritedAt      time.Time `gorm:"not null;index:idx_favorites_user_favorited;autoCreateTime" json:"favoritedAt"`
}

func (Favorite) TableName() string {
	return "favorites"
}
