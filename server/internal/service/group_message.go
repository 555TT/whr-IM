package service

import (
	"fmt"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
	"whr-im/server/internal/ws"
)

const (
	supportedGroupContentAlgorithm = "aes-gcm-256"
	supportedGroupKeyAlgorithm     = "rsa-oaep-sha256"
)

type GroupMessageService struct {
	groupMessageRepo repository.GroupMessageRepository
	groupRepo        repository.GroupRepository
	hub              *ws.Hub
}

func NewGroupMessageService(groupMessageRepo repository.GroupMessageRepository, groupRepo repository.GroupRepository, hub *ws.Hub) *GroupMessageService {
	return &GroupMessageService{groupMessageRepo: groupMessageRepo, groupRepo: groupRepo, hub: hub}
}

type CreateGroupMessageMemberKey struct {
	UserID        uint64
	KeyCiphertext string
	KeyAlgorithm  string
}

type CreateGroupMessageInput struct {
	ContentCiphertext string
	ContentIV         string
	ContentAlgorithm  string
	MemberKeys        []CreateGroupMessageMemberKey
}

// GroupMessageView 是给单个用户视角的群消息(只携带其自己那份 keyCiphertext)
type GroupMessageView struct {
	ID                uint64 `json:"id"`
	GroupID           uint64 `json:"groupId"`
	SenderID          uint64 `json:"senderId"`
	ContentCiphertext string `json:"contentCiphertext"`
	ContentIV         string `json:"contentIv"`
	ContentAlgorithm  string `json:"contentAlgorithm"`
	CreatedAt         string `json:"createdAt"`
	KeyCiphertext     string `json:"keyCiphertext"`
	KeyAlgorithm      string `json:"keyAlgorithm"`
}

func (s *GroupMessageService) Create(userID, groupID uint64, input CreateGroupMessageInput) (*GroupMessageView, error) {
	ok, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("you are not a member of this group")
	}

	if input.ContentCiphertext == "" || input.ContentIV == "" {
		return nil, fmt.Errorf("contentCiphertext and contentIv are required")
	}
	if input.ContentAlgorithm != supportedGroupContentAlgorithm {
		return nil, fmt.Errorf("unsupported content algorithm")
	}

	members, err := s.groupRepo.ListMembers(groupID)
	if err != nil {
		return nil, err
	}
	memberIDSet := make(map[uint64]struct{}, len(members))
	for _, m := range members {
		memberIDSet[m.UserID] = struct{}{}
	}

	if len(input.MemberKeys) != len(members) {
		return nil, fmt.Errorf("memberKeys must cover exactly all current group members")
	}
	keyByUser := make(map[uint64]CreateGroupMessageMemberKey, len(input.MemberKeys))
	for _, k := range input.MemberKeys {
		if _, dup := keyByUser[k.UserID]; dup {
			return nil, fmt.Errorf("duplicate memberKey for user %d", k.UserID)
		}
		if _, ok := memberIDSet[k.UserID]; !ok {
			return nil, fmt.Errorf("user %d is not a member of this group", k.UserID)
		}
		if k.KeyCiphertext == "" {
			return nil, fmt.Errorf("keyCiphertext is required for user %d", k.UserID)
		}
		if k.KeyAlgorithm != supportedGroupKeyAlgorithm {
			return nil, fmt.Errorf("unsupported key algorithm for user %d", k.UserID)
		}
		keyByUser[k.UserID] = k
	}
	for uid := range memberIDSet {
		if _, ok := keyByUser[uid]; !ok {
			return nil, fmt.Errorf("missing memberKey for user %d", uid)
		}
	}

	message := &model.GroupMessage{
		GroupID:           groupID,
		SenderID:          userID,
		ContentCiphertext: input.ContentCiphertext,
		ContentIV:         input.ContentIV,
		ContentAlgorithm:  input.ContentAlgorithm,
	}
	keys := make([]model.GroupMessageKey, 0, len(input.MemberKeys))
	for _, k := range input.MemberKeys {
		keys = append(keys, model.GroupMessageKey{
			UserID:        k.UserID,
			KeyCiphertext: k.KeyCiphertext,
			KeyAlgorithm:  k.KeyAlgorithm,
		})
	}
	if err := s.groupMessageRepo.Create(message, keys); err != nil {
		return nil, err
	}

	if s.hub != nil {
		for _, m := range members {
			view := s.buildView(message, keyByUser[m.UserID])
			_ = s.hub.Send(m.UserID, "group_message", view)
		}
	}

	senderKey := keyByUser[userID]
	view := s.buildView(message, senderKey)
	return &view, nil
}

func (s *GroupMessageService) List(userID, groupID uint64) ([]GroupMessageView, error) {
	ok, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("you are not a member of this group")
	}
	rows, err := s.groupMessageRepo.ListForUser(groupID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]GroupMessageView, 0, len(rows))
	for _, r := range rows {
		out = append(out, GroupMessageView{
			ID:                r.ID,
			GroupID:           r.GroupID,
			SenderID:          r.SenderID,
			ContentCiphertext: r.ContentCiphertext,
			ContentIV:         r.ContentIV,
			ContentAlgorithm:  r.ContentAlgorithm,
			CreatedAt:         r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			KeyCiphertext:     r.KeyCiphertext,
			KeyAlgorithm:      r.KeyAlgorithm,
		})
	}
	return out, nil
}

func (s *GroupMessageService) buildView(message *model.GroupMessage, key CreateGroupMessageMemberKey) GroupMessageView {
	return GroupMessageView{
		ID:                message.ID,
		GroupID:           message.GroupID,
		SenderID:          message.SenderID,
		ContentCiphertext: message.ContentCiphertext,
		ContentIV:         message.ContentIV,
		ContentAlgorithm:  message.ContentAlgorithm,
		CreatedAt:         message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		KeyCiphertext:     key.KeyCiphertext,
		KeyAlgorithm:      key.KeyAlgorithm,
	}
}
