package repository

import (
	"errors"

	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type FriendRepository interface {
	CreateRequest(request *model.FriendRequest) error
	ListIncomingRequests(userID uint64) ([]model.FriendRequest, error)
	HandleRequest(requestID uint64, userID uint64, status string) (*model.FriendRequest, error)
	CreateFriendPair(userID uint64, friendID uint64) error
	ListFriends(userID uint64) ([]model.Friend, error)
}

type GormFriendRepository struct {
	db *gorm.DB
}

func NewGormFriendRepository(db *gorm.DB) (*GormFriendRepository, error) {
	if err := db.AutoMigrate(&model.FriendRequest{}, &model.Friend{}); err != nil {
		return nil, err
	}
	return &GormFriendRepository{db: db}, nil
}

func (r *GormFriendRepository) CreateRequest(request *model.FriendRequest) error {
	return r.db.Create(request).Error
}

func (r *GormFriendRepository) ListIncomingRequests(userID uint64) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	if err := r.db.Where("to_user_id = ?", userID).Order("id asc").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *GormFriendRepository) HandleRequest(requestID uint64, userID uint64, status string) (*model.FriendRequest, error) {
	var request model.FriendRequest
	if err := r.db.Where("id = ? AND to_user_id = ?", requestID, userID).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	request.Status = status
	if err := r.db.Save(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *GormFriendRepository) CreateFriendPair(userID uint64, friendID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		pair := []model.Friend{{UserID: userID, FriendID: friendID}, {UserID: friendID, FriendID: userID}}
		for _, friend := range pair {
			if err := tx.Create(&friend).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormFriendRepository) ListFriends(userID uint64) ([]model.Friend, error) {
	var friends []model.Friend
	if err := r.db.Where("user_id = ?", userID).Order("id asc").Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}
