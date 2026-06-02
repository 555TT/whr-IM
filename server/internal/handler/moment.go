package handler

import (
	"net/http"
	"strconv"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type MomentHandler struct {
	momentService *service.MomentService
}

func NewMomentHandler(momentService *service.MomentService) *MomentHandler {
	return &MomentHandler{momentService: momentService}
}

type createMomentRequest struct {
	Content   string   `json:"content"`
	ImageKeys []string `json:"imageKeys"`
}

type createMomentCommentRequest struct {
	Content string `json:"content"`
}

func (h *MomentHandler) Create(c *gin.Context) {
	var req createMomentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	moment, err := h.momentService.Create(c.MustGet("userID").(uint64), service.CreateMomentInput{
		Content:   req.Content,
		ImageKeys: req.ImageKeys,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, moment)
}

func (h *MomentHandler) List(c *gin.Context) {
	moments, err := h.momentService.ListVisible(c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, moments)
}

func (h *MomentHandler) Like(c *gin.Context) {
	momentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid moment id"})
		return
	}
	if err := h.momentService.Like(c.MustGet("userID").(uint64), momentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *MomentHandler) Unlike(c *gin.Context) {
	momentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid moment id"})
		return
	}
	if err := h.momentService.Unlike(c.MustGet("userID").(uint64), momentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MomentHandler) CreateComment(c *gin.Context) {
	momentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid moment id"})
		return
	}
	var req createMomentCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	if err := h.momentService.CreateComment(c.MustGet("userID").(uint64), momentID, service.CreateMomentCommentInput{Content: req.Content}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *MomentHandler) Delete(c *gin.Context) {
	momentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid moment id"})
		return
	}
	if err := h.momentService.Delete(c.MustGet("userID").(uint64), momentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
