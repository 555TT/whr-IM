package repository

import (
	"strings"

	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type FavoriteRepository interface {
	CreateMany(items []model.Favorite) error
	ListByUser(userID uint64) ([]model.Favorite, error)
	DeleteByID(userID, favoriteID uint64) error
}

type GormFavoriteRepository struct {
	db *gorm.DB
}

func NewGormFavoriteRepository(db *gorm.DB) (*GormFavoriteRepository, error) {
	if err := db.AutoMigrate(&model.Favorite{}); err != nil {
		return nil, err
	}
	return &GormFavoriteRepository{db: db}, nil
}

func (r *GormFavoriteRepository) CreateMany(items []model.Favorite) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

func (r *GormFavoriteRepository) ListByUser(userID uint64) ([]model.Favorite, error) {
	var items []model.Favorite
	if err := r.db.Where("user_id = ?", userID).Order("favorited_at desc, id desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormFavoriteRepository) DeleteByID(userID, favoriteID uint64) error {
	result := r.db.Where("id = ? AND user_id = ?", favoriteID, userID).Delete(&model.Favorite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func IsFavoriteDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	if err == gorm.ErrDuplicatedKey {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "unique constraint failed") || strings.Contains(lower, "duplicate entry")
}
