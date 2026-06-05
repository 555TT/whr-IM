package handler

import (
	"net/http"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type AIChatHandler struct {
	aiChatService *service.AIChatService
}

type createAIChatMessageRequest struct {
	Content string `json:"content"`
}

func NewAIChatHandler(aiChatService *service.AIChatService) *AIChatHandler {
	return &AIChatHandler{aiChatService: aiChatService}
}

func (h *AIChatHandler) ListMessages(c *gin.Context) {
	messages, err := h.aiChatService.ListMessages(c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *AIChatHandler) CreateMessage(c *gin.Context) {
	var req createAIChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	messages, err := h.aiChatService.CreateAndReply(c.MustGet("userID").(uint64), service.CreateAIChatMessageInput{Content: req.Content})
	if err != nil {
		switch err {
		case service.ErrAIChatContentRequired:
			c.JSON(http.StatusBadRequest, gin.H{"message": "请输入消息内容"})
		case service.ErrAIChatUnavailable:
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "AI 聊天暂未配置，请稍后再试"})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"message": "AI 回复暂时失败，请稍后重试"})
		}
		return
	}
	c.JSON(http.StatusCreated, messages)
}
