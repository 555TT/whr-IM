package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"whr-im/server/internal/config"
)

type DeepSeekMomentAIAssistProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

type deepSeekChatRequest struct {
	Model    string                `json:"model"`
	Messages []deepSeekChatMessage `json:"messages"`
}

type deepSeekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekChatResponse struct {
	Choices []struct {
		Message deepSeekChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewDeepSeekMomentAIAssistProvider(cfg config.AIConfig) MomentAIAssistProvider {
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
		return unavailableMomentAIAssistProvider{}
	}
	return &DeepSeekMomentAIAssistProvider{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *DeepSeekMomentAIAssistProvider) Generate(input MomentAIAssistInput) (string, error) {
	userPrompt := fmt.Sprintf("请根据以下想法，生成一条适合中文社交平台朋友圈发布的文案。语气：%s。是否有配图：%t。想法：%s", normalizeTone(input.Tone), input.HasImage, input.Prompt)
	return p.complete(systemPrompt(), userPrompt)
}

func (p *DeepSeekMomentAIAssistProvider) Polish(input MomentAIAssistInput) (string, error) {
	userPrompt := fmt.Sprintf("请润色下面这条适合中文社交平台朋友圈发布的文案。语气：%s。是否有配图：%t。原文：%s", normalizeTone(input.Tone), input.HasImage, input.Content)
	return p.complete(systemPrompt(), userPrompt)
}

func (p *DeepSeekMomentAIAssistProvider) complete(system string, user string) (string, error) {
	body, err := json.Marshal(deepSeekChatRequest{
		Model: p.model,
		Messages: []deepSeekChatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
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
		return "", fmt.Errorf("%w: request failed", ErrMomentAIAssistUpstream)
	}
	defer resp.Body.Close()

	var result deepSeekChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: invalid response", ErrMomentAIAssistUpstream)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("%w: status %d", ErrMomentAIAssistUpstream, resp.StatusCode)
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("%w: empty content", ErrMomentAIAssistUpstream)
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

type unavailableMomentAIAssistProvider struct{}

func (unavailableMomentAIAssistProvider) Generate(MomentAIAssistInput) (string, error) {
	return "", ErrMomentAIAssistUnavailable
}

func (unavailableMomentAIAssistProvider) Polish(MomentAIAssistInput) (string, error) {
	return "", ErrMomentAIAssistUnavailable
}

func systemPrompt() string {
	return "你是一个中文社交平台的朋友圈文案助手。请直接输出适合用户发布的最终文案，不要解释，不要加标题，不要使用 markdown。长度控制在 1 到 3 句话，语言自然、口语化。"
}

func normalizeTone(tone string) string {
	if strings.TrimSpace(tone) == "" {
		return "自然"
	}
	return tone
}
