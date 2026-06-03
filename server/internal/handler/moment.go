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

type momentAIAssistRequest struct {
	Mode     string `json:"mode"`
	Prompt   string `json:"prompt"`
	Content  string `json:"content"`
	Tone     string `json:"tone"`
	HasImage bool   `json:"hasImage"`
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

func (h *MomentHandler) AIAssist(c *gin.Context) {
	var req momentAIAssistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	text, err := h.momentService.AIAssist(service.MomentAIAssistInput{
		Mode:     req.Mode,
		Prompt:   req.Prompt,
		Content:  req.Content,
		Tone:     req.Tone,
		HasImage: req.HasImage,
	})
	if err != nil {
		switch {
		case err == service.ErrMomentAIAssistPromptRequired:
			c.JSON(http.StatusBadRequest, gin.H{"message": "请输入想法后再生成文案"})
		case err == service.ErrMomentAIAssistContentRequired:
			c.JSON(http.StatusBadRequest, gin.H{"message": "请先输入正文后再进行润色"})
		case err == service.ErrMomentAIAssistInvalidMode:
			c.JSON(http.StatusBadRequest, gin.H{"message": "AI 助手模式无效"})
		case err == service.ErrMomentAIAssistUnavailable:
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "AI 助手暂未配置，请稍后再试"})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"message": "AI 生成暂时失败，请稍后重试"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"text": text})
}

func (h *MomentHandler) List(c *gin.Context) {
	moments, err := h.momentService.ListVisible(c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, moments)
}

func (h *MomentHandler) ListByUser(c *gin.Context) {
	targetUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	moments, err := h.momentService.ListVisibleByUser(c.MustGet("userID").(uint64), targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
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
