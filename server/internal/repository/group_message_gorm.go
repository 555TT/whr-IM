package repository

import (
	"errors"
	"time"

	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

// GroupMessageForUser 是按某个用户视角拉取群消息时的组合结果。
// 把 group_messages 与 group_message_keys 通过 JOIN 一次性取出,避免在 service 层 N+1。
type GroupMessageForUser struct {
	ID                uint64
	GroupID           uint64
	SenderID          uint64
	ContentCiphertext string
	ContentIV         string
	ContentAlgorithm  string
	CreatedAt         time.Time
	KeyCiphertext     string
	KeyAlgorithm      string
}

type GroupMessageRepository interface {
	Create(message *model.GroupMessage, keys []model.GroupMessageKey) error
	ListForUser(groupID, userID uint64) ([]GroupMessageForUser, error)
	GetKeyForUser(messageID, userID uint64) (*model.GroupMessageKey, error)
}

type GormGroupMessageRepository struct {
	db *gorm.DB
}

func NewGormGroupMessageRepository(db *gorm.DB) (*GormGroupMessageRepository, error) {
	if err := db.AutoMigrate(&model.GroupMessage{}, &model.GroupMessageKey{}); err != nil {
		return nil, err
	}
	return &GormGroupMessageRepository{db: db}, nil
}

func (r *GormGroupMessageRepository) Create(message *model.GroupMessage, keys []model.GroupMessageKey) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		for i := range keys {
			keys[i].MessageID = message.ID
			if err := tx.Create(&keys[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormGroupMessageRepository) ListForUser(groupID, userID uint64) ([]GroupMessageForUser, error) {
	var rows []GroupMessageForUser
	err := r.db.
		Table("group_messages").
		Select(`group_messages.id,
				group_messages.group_id,
				group_messages.sender_id,
				group_messages.content_ciphertext,
				group_messages.content_iv,
				group_messages.content_algorithm,
				group_messages.created_at,
				group_message_keys.key_ciphertext,
				group_message_keys.key_algorithm`).
		Joins("JOIN group_message_keys ON group_message_keys.message_id = group_messages.id AND group_message_keys.user_id = ?", userID).
		Where("group_messages.group_id = ?", groupID).
		Order("group_messages.created_at asc, group_messages.id asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GormGroupMessageRepository) GetKeyForUser(messageID, userID uint64) (*model.GroupMessageKey, error) {
	var key model.GroupMessageKey
	if err := r.db.Where("message_id = ? AND user_id = ?", messageID, userID).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &key, nil
}
