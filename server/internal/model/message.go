package model

import "time"

type Message struct {
	ID                 uint64    `gorm:"primaryKey" json:"id"`
	SenderID           uint64    `gorm:"not null;index" json:"senderId"`
	ReceiverID         uint64    `gorm:"not null;index" json:"receiverId"`
	SenderCiphertext   string    `gorm:"type:text;not null" json:"senderCiphertext"`
	SenderAlgorithm    string    `gorm:"size:50;not null" json:"senderAlgorithm"`
	ReceiverCiphertext string    `gorm:"type:text;not null" json:"receiverCiphertext"`
	ReceiverAlgorithm  string    `gorm:"size:50;not null" json:"receiverAlgorithm"`
	CreatedAt          time.Time `json:"createdAt"`
}

func (Message) TableName() string {
	return "messages"
}
