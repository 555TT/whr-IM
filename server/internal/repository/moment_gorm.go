package repository

import (
	"errors"

	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type MomentRepository interface {
	Create(moment *model.Moment) error
	ListVisibleForUser(userID uint64, friendIDs []uint64) ([]model.Moment, error)
	FindByID(id uint64) (*model.Moment, error)
	Delete(momentID uint64) error
	Like(momentID uint64, userID uint64) error
	Unlike(momentID uint64, userID uint64) error
	CountLikes(momentID uint64) (int64, error)
	HasLiked(momentID uint64, userID uint64) (bool, error)
	CreateComment(comment *model.MomentComment) error
	ListComments(momentID uint64) ([]model.MomentComment, error)
}

type GormMomentRepository struct {
	db *gorm.DB
}

func NewGormMomentRepository(db *gorm.DB) (*GormMomentRepository, error) {
	if err := db.AutoMigrate(&model.Moment{}, &model.MomentLike{}, &model.MomentComment{}); err != nil {
		return nil, err
	}
	return &GormMomentRepository{db: db}, nil
}

func (r *GormMomentRepository) Create(moment *model.Moment) error {
	return r.db.Create(moment).Error
}

func (r *GormMomentRepository) ListVisibleForUser(userID uint64, friendIDs []uint64) ([]model.Moment, error) {
	visibleUserIDs := make([]uint64, 0, len(friendIDs)+1)
	visibleUserIDs = append(visibleUserIDs, userID)
	visibleUserIDs = append(visibleUserIDs, friendIDs...)

	var moments []model.Moment
	if err := r.db.Where("user_id IN ?", visibleUserIDs).Order("created_at desc, id desc").Find(&moments).Error; err != nil {
		return nil, err
	}
	return moments, nil
}

func (r *GormMomentRepository) FindByID(id uint64) (*model.Moment, error) {
	var moment model.Moment
	if err := r.db.First(&moment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &moment, nil
}

func (r *GormMomentRepository) Delete(momentID uint64) error {
	return r.db.Where("id = ?", momentID).Delete(&model.Moment{}).Error
}

func (r *GormMomentRepository) Like(momentID uint64, userID uint64) error {
	like := &model.MomentLike{MomentID: momentID, UserID: userID}
	return r.db.FirstOrCreate(like, model.MomentLike{MomentID: momentID, UserID: userID}).Error
}

func (r *GormMomentRepository) Unlike(momentID uint64, userID uint64) error {
	return r.db.Where("moment_id = ? AND user_id = ?", momentID, userID).Delete(&model.MomentLike{}).Error
}

func (r *GormMomentRepository) CountLikes(momentID uint64) (int64, error) {
	var count int64
	if err := r.db.Model(&model.MomentLike{}).Where("moment_id = ?", momentID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormMomentRepository) HasLiked(momentID uint64, userID uint64) (bool, error) {
	var count int64
	if err := r.db.Model(&model.MomentLike{}).Where("moment_id = ? AND user_id = ?", momentID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormMomentRepository) CreateComment(comment *model.MomentComment) error {
	return r.db.Create(comment).Error
}

func (r *GormMomentRepository) ListComments(momentID uint64) ([]model.MomentComment, error) {
	var comments []model.MomentComment
	if err := r.db.Where("moment_id = ?", momentID).Order("created_at asc, id asc").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}
