package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
)

type FavoriteService struct {
	favoriteRepo     repository.FavoriteRepository
	messageRepo      repository.MessageRepository
	groupMessageRepo repository.GroupMessageRepository
	groupRepo        repository.GroupRepository
	aiChatRepo       repository.AIChatMessageRepository
	userRepo         repository.UserRepository
}

type CreateFavoriteItemInput struct {
	SourceType      string
	SourceMessageID uint64
	Content         string
}

type FavoriteView struct {
	ID               uint64 `json:"id"`
	SourceType       string `json:"sourceType"`
	SourceMessageID  uint64 `json:"sourceMessageId"`
	ConversationID   uint64 `json:"conversationId"`
	SenderID         uint64 `json:"senderId"`
	SenderName       string `json:"senderName"`
	Content          string `json:"content"`
	MessageCreatedAt string `json:"messageCreatedAt"`
	FavoritedAt      string `json:"favoritedAt"`
}

var ErrFavoriteItemsRequired = errors.New("favorite items are required")
var ErrFavoriteSourceTypeInvalid = errors.New("favorite source type is invalid")
var ErrFavoriteMessageNotFound = errors.New("favorite message not found")

func NewFavoriteService(
	favoriteRepo repository.FavoriteRepository,
	messageRepo repository.MessageRepository,
	groupMessageRepo repository.GroupMessageRepository,
	groupRepo repository.GroupRepository,
	aiChatRepo repository.AIChatMessageRepository,
	userRepo repository.UserRepository,
) *FavoriteService {
	return &FavoriteService{
		favoriteRepo:     favoriteRepo,
		messageRepo:      messageRepo,
		groupMessageRepo: groupMessageRepo,
		groupRepo:        groupRepo,
		aiChatRepo:       aiChatRepo,
		userRepo:         userRepo,
	}
}

func (s *FavoriteService) Create(userID uint64, items []CreateFavoriteItemInput) ([]FavoriteView, error) {
	if len(items) == 0 {
		return nil, ErrFavoriteItemsRequired
	}

	records := make([]model.Favorite, 0, len(items))
	for _, item := range items {
		record, err := s.buildFavoriteRecord(userID, item)
		if err != nil {
			return nil, err
		}
		records = append(records, *record)
	}

	if err := s.favoriteRepo.CreateMany(records); err != nil {
		if repository.IsFavoriteDuplicateError(err) {
			return nil, fmt.Errorf("message already favorited")
		}
		return nil, err
	}

	views := make([]FavoriteView, 0, len(records))
	for _, record := range records {
		views = append(views, toFavoriteView(record))
	}
	return views, nil
}

func (s *FavoriteService) List(userID uint64) ([]FavoriteView, error) {
	items, err := s.favoriteRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	views := make([]FavoriteView, 0, len(items))
	for _, item := range items {
		views = append(views, toFavoriteView(item))
	}
	return views, nil
}

func (s *FavoriteService) Delete(userID, favoriteID uint64) error {
	return s.favoriteRepo.DeleteByID(userID, favoriteID)
}

func (s *FavoriteService) buildFavoriteRecord(userID uint64, item CreateFavoriteItemInput) (*model.Favorite, error) {
	switch strings.TrimSpace(item.SourceType) {
	case "friend":
		return s.buildFriendFavorite(userID, item)
	case "group":
		return s.buildGroupFavorite(userID, item)
	case "ai":
		return s.buildAIFavorite(userID, item.SourceMessageID)
	default:
		return nil, ErrFavoriteSourceTypeInvalid
	}
}

func (s *FavoriteService) buildFriendFavorite(userID uint64, item CreateFavoriteItemInput) (*model.Favorite, error) {
	messageID := item.SourceMessageID
	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, ErrFavoriteMessageNotFound
	}
	if message.SenderID != userID && message.ReceiverID != userID {
		return nil, ErrFavoriteMessageNotFound
	}
	senderName, err := s.resolveUserName(message.SenderID)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(item.Content)
	if content == "" {
		content = message.SenderCiphertext
	}
	conversationID := message.SenderID
	if conversationID == userID {
		conversationID = message.ReceiverID
	}
	return &model.Favorite{
		UserID:           userID,
		SourceType:       "friend",
		SourceMessageID:  message.ID,
		ConversationID:   conversationID,
		SenderID:         message.SenderID,
		SenderName:       senderName,
		Content:          content,
		MessageCreatedAt: message.CreatedAt,
		FavoritedAt:      time.Now(),
	}, nil
}

func (s *FavoriteService) buildGroupFavorite(userID uint64, item CreateFavoriteItemInput) (*model.Favorite, error) {
	messageID := item.SourceMessageID
	message, err := s.groupMessageRepo.FindByID(messageID)
	if err != nil {
		return nil, ErrFavoriteMessageNotFound
	}
	isMember, err := s.groupRepo.IsMember(message.GroupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrFavoriteMessageNotFound
	}
	senderName, err := s.resolveUserName(message.SenderID)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(item.Content)
	if content == "" {
		content = message.ContentCiphertext
	}
	return &model.Favorite{
		UserID:           userID,
		SourceType:       "group",
		SourceMessageID:  message.ID,
		ConversationID:   message.GroupID,
		SenderID:         message.SenderID,
		SenderName:       senderName,
		Content:          content,
		MessageCreatedAt: message.CreatedAt,
		FavoritedAt:      time.Now(),
	}, nil
}

func (s *FavoriteService) buildAIFavorite(userID, messageID uint64) (*model.Favorite, error) {
	messages, err := s.aiChatRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	for _, message := range messages {
		if message.ID != messageID {
			continue
		}
		senderName := "AI 助手"
		senderID := uint64(0)
		if message.Role == "user" {
			senderID = userID
			senderName, err = s.resolveUserName(userID)
			if err != nil {
				return nil, err
			}
		}
		return &model.Favorite{
			UserID:           userID,
			SourceType:       "ai",
			SourceMessageID:  message.ID,
			ConversationID:   0,
			SenderID:         senderID,
			SenderName:       senderName,
			Content:          message.Content,
			MessageCreatedAt: message.CreatedAt,
			FavoritedAt:      time.Now(),
		}, nil
	}
	return nil, ErrFavoriteMessageNotFound
}

func (s *FavoriteService) resolveUserName(userID uint64) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(user.Nickname) != "" {
		return user.Nickname, nil
	}
	return user.Username, nil
}

func toFavoriteView(item model.Favorite) FavoriteView {
	return FavoriteView{
		ID:               item.ID,
		SourceType:       item.SourceType,
		SourceMessageID:  item.SourceMessageID,
		ConversationID:   item.ConversationID,
		SenderID:         item.SenderID,
		SenderName:       item.SenderName,
		Content:          item.Content,
		MessageCreatedAt: item.MessageCreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		FavoritedAt:      item.FavoritedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
