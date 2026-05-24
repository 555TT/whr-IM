package handler

import (
	"net/http"
	"strconv"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

type createMessageRequest struct {
	ReceiverID uint64 `json:"receiverId"`
	Ciphertext string `json:"ciphertext"`
	Algorithm  string `json:"algorithm"`
}

func (h *MessageHandler) Create(c *gin.Context) {
	var req createMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	message, err := h.messageService.Create(c.MustGet("userID").(uint64), service.CreateMessageInput{
		ReceiverID: req.ReceiverID,
		Ciphertext: req.Ciphertext,
		Algorithm:  req.Algorithm,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *MessageHandler) List(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Query("friendId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid friendId"})
		return
	}
	messages, err := h.messageService.ListConversation(c.MustGet("userID").(uint64), friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}
