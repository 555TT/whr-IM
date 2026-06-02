package handler

import (
	"net/http"
	"strconv"
	"strings"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupMessageHandler struct {
	groupMessageService *service.GroupMessageService
}

func NewGroupMessageHandler(groupMessageService *service.GroupMessageService) *GroupMessageHandler {
	return &GroupMessageHandler{groupMessageService: groupMessageService}
}

type createGroupMessageMemberKey struct {
	UserID        uint64 `json:"userId"`
	KeyCiphertext string `json:"keyCiphertext"`
	KeyAlgorithm  string `json:"keyAlgorithm"`
}

type createGroupMessageRequest struct {
	ContentCiphertext string                        `json:"contentCiphertext"`
	ContentIV         string                        `json:"contentIv"`
	ContentAlgorithm  string                        `json:"contentAlgorithm"`
	MemberKeys        []createGroupMessageMemberKey `json:"memberKeys"`
}

func (h *GroupMessageHandler) Create(c *gin.Context) {
	groupID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	var req createGroupMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	memberKeys := make([]service.CreateGroupMessageMemberKey, 0, len(req.MemberKeys))
	for _, k := range req.MemberKeys {
		memberKeys = append(memberKeys, service.CreateGroupMessageMemberKey{
			UserID:        k.UserID,
			KeyCiphertext: k.KeyCiphertext,
			KeyAlgorithm:  k.KeyAlgorithm,
		})
	}
	view, err := h.groupMessageService.Create(c.MustGet("userID").(uint64), groupID, service.CreateGroupMessageInput{
		ContentCiphertext: req.ContentCiphertext,
		ContentIV:         req.ContentIV,
		ContentAlgorithm:  req.ContentAlgorithm,
		MemberKeys:        memberKeys,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *GroupMessageHandler) List(c *gin.Context) {
	groupID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	views, err := h.groupMessageService.List(c.MustGet("userID").(uint64), groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, views)
}
