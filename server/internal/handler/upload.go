package handler

import (
	"errors"
	"io"
	"net/http"

	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to open file"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read file"})
		return
	}

	userID := c.MustGet("userID").(uint64)
	uploaded, err := h.uploadService.UploadImage(c.Request.Context(), userID, fileHeader.Filename, fileHeader.Header.Get("Content-Type"), data)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidUpload) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, uploaded)
}
