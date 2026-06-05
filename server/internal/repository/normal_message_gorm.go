package repository

import (
	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type NormalMessageRepository interface {
	Create(message *model.NormalMessage) error
	ListConversation(userID uint64, friendID uint64) ([]model.NormalMessage, error)
}

type GormNormalMessageRepository struct {
	db *gorm.DB
}

func NewGormNormalMessageRepository(db *gorm.DB) (*GormNormalMessageRepository, error) {
	if err := db.AutoMigrate(&model.NormalMessage{}); err != nil {
		return nil, err
	}
	return &GormNormalMessageRepository{db: db}, nil
}

func (r *GormNormalMessageRepository) Create(message *model.NormalMessage) error {
	return r.db.Create(message).Error
}

func (r *GormNormalMessageRepository) ListConversation(userID uint64, friendID uint64) ([]model.NormalMessage, error) {
	var messages []model.NormalMessage
	if err := r.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, friendID, friendID, userID).Order("created_at asc, id asc").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
