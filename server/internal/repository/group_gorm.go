package repository

import (
	"errors"

	"whr-im/server/internal/model"

	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(group *model.Group, memberIDs []uint64) error
	FindByID(id uint64) (*model.Group, error)
	ListByUser(userID uint64) ([]model.Group, error)
	IsMember(groupID, userID uint64) (bool, error)
	ListMembers(groupID uint64) ([]model.GroupMember, error)
	AddMembers(groupID uint64, memberIDs []uint64) error
	RemoveMember(groupID, userID uint64) error
}

type GormGroupRepository struct {
	db *gorm.DB
}

func NewGormGroupRepository(db *gorm.DB) (*GormGroupRepository, error) {
	if err := db.AutoMigrate(&model.Group{}, &model.GroupMember{}); err != nil {
		return nil, err
	}
	return &GormGroupRepository{db: db}, nil
}

func (r *GormGroupRepository) Create(group *model.Group, memberIDs []uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(group).Error; err != nil {
			return err
		}
		for _, uid := range memberIDs {
			member := model.GroupMember{GroupID: group.ID, UserID: uid}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormGroupRepository) FindByID(id uint64) (*model.Group, error) {
	var group model.Group
	if err := r.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &group, nil
}

func (r *GormGroupRepository) ListByUser(userID uint64) ([]model.Group, error) {
	var groups []model.Group
	err := r.db.
		Table("chat_groups").
		Joins("JOIN group_members ON group_members.group_id = chat_groups.id").
		Where("group_members.user_id = ?", userID).
		Order("chat_groups.id desc").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GormGroupRepository) IsMember(groupID, userID uint64) (bool, error) {
	var count int64
	if err := r.db.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormGroupRepository) ListMembers(groupID uint64) ([]model.GroupMember, error) {
	var members []model.GroupMember
	if err := r.db.Where("group_id = ?", groupID).Order("id asc").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *GormGroupRepository) AddMembers(groupID uint64, memberIDs []uint64) error {
	if len(memberIDs) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, uid := range memberIDs {
			member := model.GroupMember{GroupID: groupID, UserID: uid}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormGroupRepository) RemoveMember(groupID, userID uint64) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&model.GroupMember{}).Error
}
