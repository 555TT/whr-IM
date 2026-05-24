package model

type Friend struct {
	ID       uint64 `gorm:"primaryKey" json:"id"`
	UserID   uint64 `gorm:"not null;uniqueIndex:uk_friends_user_friend" json:"userId"`
	FriendID uint64 `gorm:"not null;uniqueIndex:uk_friends_user_friend" json:"friendId"`
}

func (Friend) TableName() string {
	return "friends"
}
