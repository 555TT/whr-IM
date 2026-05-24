package handler

import (
	"net/http"
	"strconv"
	"strings"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendService *service.FriendService
}

func NewFriendHandler(friendService *service.FriendService) *FriendHandler {
	return &FriendHandler{friendService: friendService}
}

type createFriendRequest struct {
	ToUsername string `json:"toUsername"`
	Message    string `json:"message"`
}

func (h *FriendHandler) CreateRequest(c *gin.Context) {
	var req createFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	request, err := h.friendService.CreateRequest(c.MustGet("userID").(uint64), service.CreateFriendRequestInput{ToUsername: req.ToUsername, Message: req.Message})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, request)
}

func (h *FriendHandler) ListIncomingRequests(c *gin.Context) {
	requests, err := h.friendService.ListIncomingRequests(c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func (h *FriendHandler) AcceptRequest(c *gin.Context) {
	requestID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request id"})
		return
	}
	request, err := h.friendService.AcceptRequest(requestID, c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

func (h *FriendHandler) RejectRequest(c *gin.Context) {
	requestID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request id"})
		return
	}
	request, err := h.friendService.RejectRequest(requestID, c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, request)
}

func (h *FriendHandler) ListFriends(c *gin.Context) {
	friends, err := h.friendService.ListFriends(c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friends)
}
