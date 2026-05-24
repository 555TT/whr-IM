package model

import "time"

type Message struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	SenderID   uint64    `gorm:"not null;index" json:"senderId"`
	ReceiverID uint64    `gorm:"not null;index" json:"receiverId"`
	Ciphertext string    `gorm:"type:text;not null" json:"ciphertext"`
	Algorithm  string    `gorm:"size:50;not null" json:"algorithm"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (Message) TableName() string {
	return "messages"
}
