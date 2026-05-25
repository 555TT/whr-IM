package repository

import (
	"errors"
	"strings"

	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) (*GormUserRepository, error) {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return nil, err
	}
	return &GormUserRepository{db: db}, nil
}

func (r *GormUserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrUsernameExists
		}
		return err
	}
	return nil
}

func (r *GormUserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByID(id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) UpdateProfile(userID uint64, nickname string, gender int, signature string) (*model.User, error) {
	updates := map[string]interface{}{
		"nickname":  nickname,
		"gender":    gender,
		"signature": signature,
	}
	result := r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrUserNotFound
	}
	return r.FindByID(userID)
}

func (r *GormUserRepository) UpdatePublicKey(userID uint64, publicKey string, algorithm string) (*model.User, error) {
	updates := map[string]interface{}{
		"public_key":           publicKey,
		"public_key_algorithm": algorithm,
	}
	result := r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrUserNotFound
	}
	return r.FindByID(userID)
}

func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed") || strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}
