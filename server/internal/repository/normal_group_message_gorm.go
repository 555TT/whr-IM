package repository

import (
	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type NormalGroupMessageRepository interface {
	Create(message *model.NormalGroupMessage) error
	ListByGroup(groupID uint64) ([]model.NormalGroupMessage, error)
}

type GormNormalGroupMessageRepository struct {
	db *gorm.DB
}

func NewGormNormalGroupMessageRepository(db *gorm.DB) (*GormNormalGroupMessageRepository, error) {
	if err := db.AutoMigrate(&model.NormalGroupMessage{}); err != nil {
		return nil, err
	}
	return &GormNormalGroupMessageRepository{db: db}, nil
}

func (r *GormNormalGroupMessageRepository) Create(message *model.NormalGroupMessage) error {
	return r.db.Create(message).Error
}

func (r *GormNormalGroupMessageRepository) ListByGroup(groupID uint64) ([]model.NormalGroupMessage, error) {
	var messages []model.NormalGroupMessage
	if err := r.db.Where("group_id = ?", groupID).Order("created_at asc, id asc").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
