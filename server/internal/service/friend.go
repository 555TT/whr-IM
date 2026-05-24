package service

import (
	"fmt"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
)

type FriendService struct {
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
}

func NewFriendService(friendRepo repository.FriendRepository, userRepo repository.UserRepository) *FriendService {
	return &FriendService{friendRepo: friendRepo, userRepo: userRepo}
}

type CreateFriendRequestInput struct {
	ToUserID uint64
	Message  string
}

type FriendListItem struct {
	UserID    uint64  `json:"userId"`
	FriendID  uint64  `json:"friendId"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
}

func (s *FriendService) CreateRequest(fromUserID uint64, input CreateFriendRequestInput) (*model.FriendRequest, error) {
	if fromUserID == input.ToUserID {
		return nil, fmt.Errorf("cannot add yourself")
	}
	if _, err := s.userRepo.FindByID(input.ToUserID); err != nil {
		return nil, err
	}
	request := &model.FriendRequest{FromUserID: fromUserID, ToUserID: input.ToUserID, Message: input.Message, Status: "pending"}
	if err := s.friendRepo.CreateRequest(request); err != nil {
		return nil, err
	}
	return request, nil
}

func (s *FriendService) ListIncomingRequests(userID uint64) ([]model.FriendRequest, error) {
	return s.friendRepo.ListIncomingRequests(userID)
}

func (s *FriendService) AcceptRequest(requestID uint64, userID uint64) (*model.FriendRequest, error) {
	request, err := s.friendRepo.HandleRequest(requestID, userID, "accepted")
	if err != nil {
		return nil, err
	}
	if err := s.friendRepo.CreateFriendPair(request.FromUserID, request.ToUserID); err != nil {
		return nil, err
	}
	return request, nil
}

func (s *FriendService) RejectRequest(requestID uint64, userID uint64) (*model.FriendRequest, error) {
	return s.friendRepo.HandleRequest(requestID, userID, "rejected")
}

func (s *FriendService) ListFriends(userID uint64) ([]FriendListItem, error) {
	friends, err := s.friendRepo.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	items := make([]FriendListItem, 0, len(friends))
	for _, friend := range friends {
		profile, err := s.userRepo.FindByID(friend.FriendID)
		if err != nil {
			return nil, err
		}
		items = append(items, FriendListItem{UserID: friend.UserID, FriendID: friend.FriendID, Nickname: profile.Nickname, Avatar: profile.Avatar, Signature: profile.Signature})
	}
	return items, nil
}
