package service

import (
	"errors"
	"strings"
	"time"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
	"whr-im/server/internal/ws"
)

var ErrNormalGroupMessageNotMember = errors.New("you are not a member of this group")
var ErrNormalGroupMessageContentRequired = errors.New("content is required")

type NormalGroupMessageService struct {
	normalGroupMessageRepo repository.NormalGroupMessageRepository
	groupRepo              repository.GroupRepository
	hub                    *ws.Hub
}

func NewNormalGroupMessageService(normalGroupMessageRepo repository.NormalGroupMessageRepository, groupRepo repository.GroupRepository, hub *ws.Hub) *NormalGroupMessageService {
	return &NormalGroupMessageService{normalGroupMessageRepo: normalGroupMessageRepo, groupRepo: groupRepo, hub: hub}
}

type CreateNormalGroupMessageInput struct {
	Content string
}

type NormalGroupMessageView struct {
	ID        uint64 `json:"id"`
	GroupID   uint64 `json:"groupId"`
	SenderID  uint64 `json:"senderId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

func (s *NormalGroupMessageService) Create(userID, groupID uint64, input CreateNormalGroupMessageInput) (*NormalGroupMessageView, error) {
	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNormalGroupMessageNotMember
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrNormalGroupMessageContentRequired
	}

	message := &model.NormalGroupMessage{
		GroupID:  groupID,
		SenderID: userID,
		Content:  content,
	}
	if err := s.normalGroupMessageRepo.Create(message); err != nil {
		return nil, err
	}

	view := toNormalGroupMessageView(*message)
	if s.hub != nil {
		members, err := s.groupRepo.ListMembers(groupID)
		if err == nil {
			for _, member := range members {
				_ = s.hub.Send(member.UserID, "normal_group_message", view)
			}
		}
	}
	return &view, nil
}

func (s *NormalGroupMessageService) List(userID, groupID uint64) ([]NormalGroupMessageView, error) {
	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNormalGroupMessageNotMember
	}

	messages, err := s.normalGroupMessageRepo.ListByGroup(groupID)
	if err != nil {
		return nil, err
	}
	views := make([]NormalGroupMessageView, 0, len(messages))
	for _, message := range messages {
		views = append(views, toNormalGroupMessageView(message))
	}
	return views, nil
}

func toNormalGroupMessageView(message model.NormalGroupMessage) NormalGroupMessageView {
	return NormalGroupMessageView{
		ID:        message.ID,
		GroupID:   message.GroupID,
		SenderID:  message.SenderID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt.Format(time.RFC3339),
	}
}
