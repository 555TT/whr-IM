package model

import "time"

type Message struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	SenderID   uint64    `gorm:"not null;index" json:"senderId"`
	ReceiverID uint64    `gorm:"not null;index" json:"receiverId"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	MsgType    string    `gorm:"size:20;not null;default:text" json:"msgType"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (Message) TableName() string {
	return "messages"
}
