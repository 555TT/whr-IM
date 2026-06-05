package service

import (
	"errors"
	"strings"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
	"whr-im/server/internal/ws"
)

var ErrNormalMessageNonFriend = errors.New("non-friend users cannot chat")
var ErrNormalMessageContentRequired = errors.New("content is required")

type NormalMessageService struct {
	normalMessageRepo repository.NormalMessageRepository
	friendRepo        repository.FriendRepository
	hub               *ws.Hub
}

func NewNormalMessageService(normalMessageRepo repository.NormalMessageRepository, friendRepo repository.FriendRepository, hub *ws.Hub) *NormalMessageService {
	return &NormalMessageService{normalMessageRepo: normalMessageRepo, friendRepo: friendRepo, hub: hub}
}

type CreateNormalMessageInput struct {
	ReceiverID uint64
	Content    string
}

func (s *NormalMessageService) Create(userID uint64, input CreateNormalMessageInput) (*model.NormalMessage, error) {
	isFriend, err := s.friendRepo.AreFriends(userID, input.ReceiverID)
	if err != nil {
		return nil, err
	}
	if !isFriend {
		return nil, ErrNormalMessageNonFriend
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrNormalMessageContentRequired
	}

	message := &model.NormalMessage{
		SenderID:   userID,
		ReceiverID: input.ReceiverID,
		Content:    content,
	}
	if err := s.normalMessageRepo.Create(message); err != nil {
		return nil, err
	}
	if s.hub != nil {
		_ = s.hub.Send(input.ReceiverID, "normal_chat_message", message)
	}
	return message, nil
}

func (s *NormalMessageService) ListConversation(userID uint64, friendID uint64) ([]model.NormalMessage, error) {
	isFriend, err := s.friendRepo.AreFriends(userID, friendID)
	if err != nil {
		return nil, err
	}
	if !isFriend {
		return nil, ErrNormalMessageNonFriend
	}

	return s.normalMessageRepo.ListConversation(userID, friendID)
}
