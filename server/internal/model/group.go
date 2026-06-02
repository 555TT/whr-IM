package model

import "time"

type Group struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	OwnerID   uint64    `gorm:"not null;index" json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Group) TableName() string {
	return "chat_groups"
}

type GroupMember struct {
	ID       uint64    `gorm:"primaryKey" json:"id"`
	GroupID  uint64    `gorm:"not null;uniqueIndex:uk_group_members_group_user,priority:1" json:"groupId"`
	UserID   uint64    `gorm:"not null;uniqueIndex:uk_group_members_group_user,priority:2;index" json:"userId"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joinedAt"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

type GroupMessage struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	GroupID           uint64    `gorm:"not null;index:idx_group_messages_group_created_at,priority:1" json:"groupId"`
	SenderID          uint64    `gorm:"not null;index" json:"senderId"`
	ContentCiphertext string    `gorm:"type:text;not null" json:"contentCiphertext"`
	ContentIV         string    `gorm:"size:64;not null;column:content_iv" json:"contentIv"`
	ContentAlgorithm  string    `gorm:"size:50;not null" json:"contentAlgorithm"`
	CreatedAt         time.Time `gorm:"index:idx_group_messages_group_created_at,priority:2" json:"createdAt"`
}

func (GroupMessage) TableName() string {
	return "group_messages"
}

type GroupMessageKey struct {
	ID            uint64 `gorm:"primaryKey" json:"id"`
	MessageID     uint64 `gorm:"not null;uniqueIndex:uk_group_message_keys_message_user,priority:1" json:"messageId"`
	UserID        uint64 `gorm:"not null;uniqueIndex:uk_group_message_keys_message_user,priority:2;index" json:"userId"`
	KeyCiphertext string `gorm:"type:text;not null" json:"keyCiphertext"`
	KeyAlgorithm  string `gorm:"size:50;not null" json:"keyAlgorithm"`
}

func (GroupMessageKey) TableName() string {
	return "group_message_keys"
}
