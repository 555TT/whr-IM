package repository

import (
	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type AIChatMessageRepository interface {
	Create(message *model.AIChatMessage) error
	ListByUser(userID uint64) ([]model.AIChatMessage, error)
	ListRecentByUser(userID uint64, limit int) ([]model.AIChatMessage, error)
}

type GormAIChatMessageRepository struct {
	db *gorm.DB
}

func NewGormAIChatMessageRepository(db *gorm.DB) (*GormAIChatMessageRepository, error) {
	if err := db.AutoMigrate(&model.AIChatMessage{}); err != nil {
		return nil, err
	}
	return &GormAIChatMessageRepository{db: db}, nil
}

func (r *GormAIChatMessageRepository) Create(message *model.AIChatMessage) error {
	return r.db.Create(message).Error
}

func (r *GormAIChatMessageRepository) ListByUser(userID uint64) ([]model.AIChatMessage, error) {
	var messages []model.AIChatMessage
	if err := r.db.Where("user_id = ?", userID).Order("created_at asc, id asc").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *GormAIChatMessageRepository) ListRecentByUser(userID uint64, limit int) ([]model.AIChatMessage, error) {
	var messages []model.AIChatMessage
	query := r.db.Where("user_id = ?", userID).Order("created_at desc, id desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
	return messages, nil
}
