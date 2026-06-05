package handler

import (
	"net/http"
	"strconv"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
	favoriteService *service.FavoriteService
}

type createFavoriteItemRequest struct {
	SourceType      string `json:"sourceType"`
	SourceMessageID uint64 `json:"sourceMessageId"`
	Content         string `json:"content"`
}

type createFavoriteRequest struct {
	Items []createFavoriteItemRequest `json:"items"`
}

func NewFavoriteHandler(favoriteService *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

func (h *FavoriteHandler) Create(c *gin.Context) {
	var req createFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	items := make([]service.CreateFavoriteItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.CreateFavoriteItemInput{SourceType: item.SourceType, SourceMessageID: item.SourceMessageID, Content: item.Content})
	}
	views, err := h.favoriteService.Create(c.MustGet("userID").(uint64), items)
	if err != nil {
		switch err {
		case service.ErrFavoriteItemsRequired:
			c.JSON(http.StatusBadRequest, gin.H{"message": "请至少选择一条消息"})
		case service.ErrFavoriteSourceTypeInvalid, service.ErrFavoriteMessageNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, views)
}

func (h *FavoriteHandler) List(c *gin.Context) {
	views, err := h.favoriteService.List(c.MustGet("userID").(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, views)
}

func (h *FavoriteHandler) Delete(c *gin.Context) {
	favoriteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid favorite id"})
		return
	}
	if err := h.favoriteService.Delete(c.MustGet("userID").(uint64), favoriteID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
