package handler

import (
	"net/http"
	"strconv"
	"strings"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

type createGroupRequest struct {
	Name      string   `json:"name"`
	MemberIDs []uint64 `json:"memberIds"`
}

func (h *GroupHandler) Create(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	detail, err := h.groupService.CreateGroup(c.MustGet("userID").(uint64), service.CreateGroupInput{
		Name:      req.Name,
		MemberIDs: req.MemberIDs,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, detail)
}

func (h *GroupHandler) ListMine(c *gin.Context) {
	items, err := h.groupService.ListMyGroups(c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *GroupHandler) Detail(c *gin.Context) {
	groupID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	detail, err := h.groupService.GetGroupDetail(c.MustGet("userID").(uint64), groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

type addGroupMembersRequest struct {
	MemberIDs []uint64 `json:"memberIds"`
}

func (h *GroupHandler) AddMembers(c *gin.Context) {
	groupID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	var req addGroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	detail, err := h.groupService.AddMembers(c.MustGet("userID").(uint64), groupID, req.MemberIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid group id"})
		return
	}
	if err := h.groupService.LeaveGroup(c.MustGet("userID").(uint64), groupID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
