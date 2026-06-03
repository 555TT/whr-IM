package service

import (
	"context"
	"encoding/json"
	"fmt"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
)

type MomentService struct {
	momentRepo repository.MomentRepository
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
	storage    ObjectStorage
}

func NewMomentService(momentRepo repository.MomentRepository, friendRepo repository.FriendRepository, userRepo repository.UserRepository, storage ObjectStorage) *MomentService {
	return &MomentService{momentRepo: momentRepo, friendRepo: friendRepo, userRepo: userRepo, storage: storage}
}

type CreateMomentInput struct {
	Content   string
	ImageKeys []string
}

type CreateMomentCommentInput struct {
	Content string
}

type MomentCommentView struct {
	ID        uint64 `json:"id"`
	UserID    uint64 `json:"userId"`
	Nickname  string `json:"nickname"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

var ErrMomentNotVisible = fmt.Errorf("moment is not visible to current user")

type MomentView struct {
	ID        uint64              `json:"id"`
	UserID    uint64              `json:"userId"`
	Nickname  string              `json:"nickname"`
	Avatar    string              `json:"avatar"`
	Content   string              `json:"content"`
	Images    []string            `json:"images"`
	LikeCount int                 `json:"likeCount"`
	LikedByMe bool                `json:"likedByMe"`
	Comments  []MomentCommentView `json:"comments"`
	CreatedAt string              `json:"createdAt"`
}

func (s *MomentService) Create(userID uint64, input CreateMomentInput) (*MomentView, error) {
	if input.Content == "" {
		return nil, fmt.Errorf("content is required")
	}
	imagesJSON, err := json.Marshal(input.ImageKeys)
	if err != nil {
		return nil, err
	}
	moment := &model.Moment{
		UserID:     userID,
		Content:    input.Content,
		ImagesJSON: string(imagesJSON),
	}
	if err := s.momentRepo.Create(moment); err != nil {
		return nil, err
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return s.buildMomentViewForUser(moment, userID, user.Nickname, user.Avatar)
}

func (s *MomentService) Like(userID uint64, momentID uint64) error {
	moment, err := s.momentRepo.FindByID(momentID)
	if err != nil {
		return err
	}
	if !s.canViewMoment(userID, moment.UserID) {
		return ErrMomentNotVisible
	}
	return s.momentRepo.Like(momentID, userID)
}

func (s *MomentService) Unlike(userID uint64, momentID uint64) error {
	moment, err := s.momentRepo.FindByID(momentID)
	if err != nil {
		return err
	}
	if !s.canViewMoment(userID, moment.UserID) {
		return ErrMomentNotVisible
	}
	return s.momentRepo.Unlike(momentID, userID)
}

func (s *MomentService) CreateComment(userID uint64, momentID uint64, input CreateMomentCommentInput) error {
	if input.Content == "" {
		return fmt.Errorf("content is required")
	}
	moment, err := s.momentRepo.FindByID(momentID)
	if err != nil {
		return err
	}
	if !s.canViewMoment(userID, moment.UserID) {
		return ErrMomentNotVisible
	}
	comment := &model.MomentComment{
		MomentID: momentID,
		UserID:   userID,
		Content:  input.Content,
	}
	return s.momentRepo.CreateComment(comment)
}

func (s *MomentService) Delete(userID uint64, momentID uint64) error {
	moment, err := s.momentRepo.FindByID(momentID)
	if err != nil {
		return err
	}
	if moment.UserID != userID {
		return fmt.Errorf("only the author can delete this moment")
	}
	return s.momentRepo.Delete(momentID)
}

func (s *MomentService) ListVisible(userID uint64) ([]MomentView, error) {
	friends, err := s.friendRepo.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	friendIDs := make([]uint64, 0, len(friends))
	for _, friend := range friends {
		friendIDs = append(friendIDs, friend.FriendID)
	}
	moments, err := s.momentRepo.ListVisibleForUser(userID, friendIDs)
	if err != nil {
		return nil, err
	}
	views := make([]MomentView, 0, len(moments))
	for _, moment := range moments {
		user, err := s.userRepo.FindByID(moment.UserID)
		if err != nil {
			return nil, err
		}
		view, err := s.buildMomentViewForUser(&moment, userID, user.Nickname, user.Avatar)
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func (s *MomentService) ListVisibleByUser(viewerID uint64, ownerID uint64) ([]MomentView, error) {
	if !s.canViewMoment(viewerID, ownerID) {
		return nil, ErrMomentNotVisible
	}
	moments, err := s.momentRepo.ListByUserID(ownerID)
	if err != nil {
		return nil, err
	}
	owner, err := s.userRepo.FindByID(ownerID)
	if err != nil {
		return nil, err
	}
	views := make([]MomentView, 0, len(moments))
	for _, moment := range moments {
		view, err := s.buildMomentViewForUser(&moment, viewerID, owner.Nickname, owner.Avatar)
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func (s *MomentService) canViewMoment(userID uint64, ownerID uint64) bool {
	if userID == ownerID {
		return true
	}
	friends, err := s.friendRepo.ListFriends(userID)
	if err != nil {
		return false
	}
	for _, friend := range friends {
		if friend.FriendID == ownerID {
			return true
		}
	}
	return false
}

func (s *MomentService) buildMomentViewForUser(moment *model.Moment, viewerID uint64, nickname string, avatar string) (*MomentView, error) {
	imageKeys := make([]string, 0)
	if moment.ImagesJSON != "" {
		if err := json.Unmarshal([]byte(moment.ImagesJSON), &imageKeys); err != nil {
			return nil, err
		}
	}
	images := make([]string, 0, len(imageKeys))
	for _, objectKey := range imageKeys {
		if s.storage == nil {
			images = append(images, objectKey)
			continue
		}
		resolved, err := s.storage.ObjectURL(context.Background(), objectKey)
		if err != nil {
			return nil, err
		}
		images = append(images, resolved)
	}
	likeCount, err := s.momentRepo.CountLikes(moment.ID)
	if err != nil {
		return nil, err
	}
	likedByMe, err := s.momentRepo.HasLiked(moment.ID, viewerID)
	if err != nil {
		return nil, err
	}
	comments, err := s.momentRepo.ListComments(moment.ID)
	if err != nil {
		return nil, err
	}
	commentViews := make([]MomentCommentView, 0, len(comments))
	for _, comment := range comments {
		user, err := s.userRepo.FindByID(comment.UserID)
		if err != nil {
			return nil, err
		}
		commentViews = append(commentViews, MomentCommentView{
			ID:        comment.ID,
			UserID:    comment.UserID,
			Nickname:  user.Nickname,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return &MomentView{
		ID:        moment.ID,
		UserID:    moment.UserID,
		Nickname:  nickname,
		Avatar:    avatar,
		Content:   moment.Content,
		Images:    images,
		LikeCount: int(likeCount),
		LikedByMe: likedByMe,
		Comments:  commentViews,
		CreatedAt: moment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
