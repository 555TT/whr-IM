package service

import (
	"fmt"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
	"whr-im/server/internal/ws"
)

const supportedMessageAlgorithm = "rsa-oaep-sha256"

type MessageService struct {
	messageRepo repository.MessageRepository
	friendRepo  repository.FriendRepository
	hub         *ws.Hub
}

func NewMessageService(messageRepo repository.MessageRepository, friendRepo repository.FriendRepository, hub *ws.Hub) *MessageService {
	return &MessageService{messageRepo: messageRepo, friendRepo: friendRepo, hub: hub}
}

type CreateMessageInput struct {
	ReceiverID         uint64
	SenderCiphertext   string
	SenderAlgorithm    string
	ReceiverCiphertext string
	ReceiverAlgorithm  string
}

func (s *MessageService) Create(userID uint64, input CreateMessageInput) (*model.Message, error) {
	friends, err := s.friendRepo.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	isFriend := false
	for _, friend := range friends {
		if friend.FriendID == input.ReceiverID {
			isFriend = true
			break
		}
	}
	if !isFriend {
		return nil, fmt.Errorf("non-friend users cannot chat")
	}
	if input.SenderCiphertext == "" || input.ReceiverCiphertext == "" {
		return nil, fmt.Errorf("ciphertext is required")
	}
	if input.SenderAlgorithm == "" || input.ReceiverAlgorithm == "" {
		return nil, fmt.Errorf("algorithm is required")
	}
	if input.SenderAlgorithm != supportedMessageAlgorithm || input.ReceiverAlgorithm != supportedMessageAlgorithm {
		return nil, fmt.Errorf("unsupported algorithm")
	}
	message := &model.Message{
		SenderID:           userID,
		ReceiverID:         input.ReceiverID,
		SenderCiphertext:   input.SenderCiphertext,
		SenderAlgorithm:    input.SenderAlgorithm,
		ReceiverCiphertext: input.ReceiverCiphertext,
		ReceiverAlgorithm:  input.ReceiverAlgorithm,
	}
	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}
	if s.hub != nil {
		_ = s.hub.Send(input.ReceiverID, "chat_message", message)
	}
	return message, nil
}

func (s *MessageService) ListConversation(userID uint64, friendID uint64) ([]model.Message, error) {
	return s.messageRepo.ListConversation(userID, friendID)
}
