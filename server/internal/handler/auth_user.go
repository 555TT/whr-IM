package handler

import (
	"errors"
	"net/http"
	"strconv"

	"whr-im/server/internal/repository"
	"whr-im/server/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthUserHandler struct {
	authService *service.AuthService
	friendRepo  repository.FriendRepository
}

func NewAuthUserHandler(authService *service.AuthService, friendRepo repository.FriendRepository) *AuthUserHandler {
	return &AuthUserHandler{authService: authService, friendRepo: friendRepo}
}

type registerRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type updateProfileRequest struct {
	Nickname  string `json:"nickname"`
	Gender    int    `json:"gender"`
	Signature string `json:"signature"`
}

type updatePublicKeyRequest struct {
	PublicKey string `json:"publicKey"`
	Algorithm string `json:"algorithm"`
}

func (h *AuthUserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	user, err := h.authService.Register(service.RegisterInput{
		Username:        req.Username,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	})
	if err != nil {
		status := http.StatusBadRequest
		if err == repository.ErrUsernameExists {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h *AuthUserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	token, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if err != service.ErrInvalidCredentials {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (h *AuthUserHandler) Me(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)
	user, err := h.authService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthUserHandler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	userID := c.MustGet("userID").(uint64)
	user, err := h.authService.UpdateProfile(userID, service.UpdateProfileInput{
		Nickname:  req.Nickname,
		Gender:    req.Gender,
		Signature: req.Signature,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthUserHandler) PublicProfile(c *gin.Context) {
	targetUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	profile, err := h.authService.GetVisibleProfile(c.MustGet("userID").(uint64), targetUserID, h.friendRepo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *AuthUserHandler) UpdateMyPublicKey(c *gin.Context) {
	var req updatePublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	userID := c.MustGet("userID").(uint64)
	user, err := h.authService.UpdatePublicKey(userID, service.UpdatePublicKeyInput{
		PublicKey: req.PublicKey,
		Algorithm: req.Algorithm,
	})
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, service.ErrInvalidPublicKey) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
