package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type NormalGroupMessageHandler struct {
	normalGroupMessageService *service.NormalGroupMessageService
}

func NewNormalGroupMessageHandler(normalGroupMessageService *service.NormalGroupMessageService) *NormalGroupMessageHandler {
	return &NormalGroupMessageHandler{normalGroupMessageService: normalGroupMessageService}
}

type createNormalGroupMessageRequest struct {
	Content string `json:"content"`
}

func (h *NormalGroupMessageHandler) Create(c *gin.Context) {
	groupID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	var req createNormalGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	message, err := h.normalGroupMessageService.Create(c.MustGet("userID").(uint64), groupID, service.CreateNormalGroupMessageInput{
		Content: req.Content,
	})
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *NormalGroupMessageHandler) List(c *gin.Context) {
	groupID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	messages, err := h.normalGroupMessageService.List(c.MustGet("userID").(uint64), groupID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *NormalGroupMessageHandler) writeServiceError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrNormalGroupMessageNotMember):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrNormalGroupMessageContentRequired):
		status = http.StatusBadRequest
	}
	c.JSON(status, gin.H{"message": err.Error()})
}
