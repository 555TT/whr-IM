package handler

import (
	"errors"
	"net/http"
	"strconv"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type NormalMessageHandler struct {
	normalMessageService *service.NormalMessageService
}

func NewNormalMessageHandler(normalMessageService *service.NormalMessageService) *NormalMessageHandler {
	return &NormalMessageHandler{normalMessageService: normalMessageService}
}

type createNormalMessageRequest struct {
	ReceiverID uint64 `json:"receiverId"`
	Content    string `json:"content"`
}

func (h *NormalMessageHandler) Create(c *gin.Context) {
	var req createNormalMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	message, err := h.normalMessageService.Create(c.MustGet("userID").(uint64), service.CreateNormalMessageInput{
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *NormalMessageHandler) List(c *gin.Context) {
	friendID, err := strconv.ParseUint(c.Query("friendId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid friendId"})
		return
	}
	messages, err := h.normalMessageService.ListConversation(c.MustGet("userID").(uint64), friendID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *NormalMessageHandler) writeServiceError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, service.ErrNormalMessageNonFriend) || errors.Is(err, service.ErrNormalMessageContentRequired) {
		status = http.StatusBadRequest
	}
	c.JSON(status, gin.H{"message": err.Error()})
}
