package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"whr-im/server/internal/config"
	"whr-im/server/internal/model"
	"whr-im/server/internal/repository"
)

type AIChatProvider interface {
	Reply(history []AIChatHistoryMessage) (string, error)
}

type AIChatHistoryMessage struct {
	Role    string
	Content string
}

type AIChatService struct {
	messageRepo repository.AIChatMessageRepository
	aiProvider  AIChatProvider
}

type CreateAIChatMessageInput struct {
	Content string
}

var ErrAIChatContentRequired = errors.New("content is required")
var ErrAIChatUnavailable = errors.New("ai chat is unavailable")
var ErrAIChatUpstream = errors.New("ai chat upstream failed")

func NewAIChatService(messageRepo repository.AIChatMessageRepository, aiProvider AIChatProvider) *AIChatService {
	if aiProvider == nil {
		aiProvider = stubAIChatProvider{}
	}
	return &AIChatService{messageRepo: messageRepo, aiProvider: aiProvider}
}

func (s *AIChatService) ListMessages(userID uint64) ([]model.AIChatMessage, error) {
	return s.messageRepo.ListByUser(userID)
}

func (s *AIChatService) CreateAndReply(userID uint64, input CreateAIChatMessageInput) ([]model.AIChatMessage, error) {
	input.Content = strings.TrimSpace(input.Content)
	if input.Content == "" {
		return nil, ErrAIChatContentRequired
	}

	userMessage := &model.AIChatMessage{UserID: userID, Role: "user", Content: input.Content}
	if err := s.messageRepo.Create(userMessage); err != nil {
		return nil, err
	}

	history, err := s.messageRepo.ListRecentByUser(userID, 20)
	if err != nil {
		return nil, err
	}
	providerHistory := make([]AIChatHistoryMessage, 0, len(history))
	for _, item := range history {
		providerHistory = append(providerHistory, AIChatHistoryMessage{Role: item.Role, Content: item.Content})
	}

	reply, err := s.aiProvider.Reply(providerHistory)
	if err != nil {
		return nil, err
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return nil, fmt.Errorf("%w: empty content", ErrAIChatUpstream)
	}

	assistantMessage := &model.AIChatMessage{UserID: userID, Role: "assistant", Content: reply}
	if err := s.messageRepo.Create(assistantMessage); err != nil {
		return nil, err
	}

	return []model.AIChatMessage{*userMessage, *assistantMessage}, nil
}

type stubAIChatProvider struct{}

func (stubAIChatProvider) Reply(history []AIChatHistoryMessage) (string, error) {
	if len(history) == 0 {
		return "你好，我是 AI 助手。", nil
	}
	return "AI 回复：" + history[len(history)-1].Content, nil
}

type DeepSeekAIChatProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

type deepSeekAIChatRequest struct {
	Model    string                `json:"model"`
	Messages []deepSeekAIChatEntry `json:"messages"`
}

type deepSeekAIChatEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekAIChatResponse struct {
	Choices []struct {
		Message deepSeekAIChatEntry `json:"message"`
	} `json:"choices"`
}

func NewDeepSeekAIChatProvider(cfg config.AIConfig) AIChatProvider {
	baseURL := strings.TrimRight(cfg.DeepSeekBaseURL, "/")
	apiKey := cfg.DeepSeekAPIKey
	model := cfg.DeepSeekModel
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-v4-flash"
	}
	if apiKey == "" {
		return unavailableAIChatProvider{}
	}
	return &DeepSeekAIChatProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *DeepSeekAIChatProvider) Reply(history []AIChatHistoryMessage) (string, error) {
	messages := make([]deepSeekAIChatEntry, 0, len(history)+1)
	messages = append(messages, deepSeekAIChatEntry{Role: "system", Content: "你是一个中文即时通讯应用内的 AI 聊天助手。请直接回答用户问题，语气自然、简洁。不要使用 markdown 表格，避免过长输出。"})
	for _, item := range history {
		messages = append(messages, deepSeekAIChatEntry{Role: item.Role, Content: item.Content})
	}
	body, err := json.Marshal(deepSeekAIChatRequest{Model: p.model, Messages: messages})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: request failed", ErrAIChatUpstream)
	}
	defer resp.Body.Close()
	var result deepSeekAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: invalid response", ErrAIChatUpstream)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("%w: status %d", ErrAIChatUpstream, resp.StatusCode)
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("%w: empty content", ErrAIChatUpstream)
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

type unavailableAIChatProvider struct{}

func (unavailableAIChatProvider) Reply([]AIChatHistoryMessage) (string, error) {
	return "", ErrAIChatUnavailable
}
