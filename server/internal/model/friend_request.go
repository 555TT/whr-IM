package model

type FriendRequest struct {
	ID         uint64  `gorm:"primaryKey" json:"id"`
	FromUserID uint64  `gorm:"not null;index" json:"fromUserId"`
	ToUserID   uint64  `gorm:"not null;index" json:"toUserId"`
	Message    string `gorm:"size:255;not null;default:''" json:"message"`
	Status     string `gorm:"size:20;not null;default:pending;index" json:"status"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}
