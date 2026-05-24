package repository

import (
	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *model.Message) error
	ListConversation(userID uint64, friendID uint64) ([]model.Message, error)
}

type GormMessageRepository struct {
	db *gorm.DB
}

func NewGormMessageRepository(db *gorm.DB) (*GormMessageRepository, error) {
	if err := db.AutoMigrate(&model.Message{}); err != nil {
		return nil, err
	}
	return &GormMessageRepository{db: db}, nil
}

func (r *GormMessageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}

func (r *GormMessageRepository) ListConversation(userID uint64, friendID uint64) ([]model.Message, error) {
	var messages []model.Message
	if err := r.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, friendID, friendID, userID).Order("id asc").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
