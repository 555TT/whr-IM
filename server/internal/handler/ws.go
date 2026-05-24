package handler

import (
	"net/http"

	"whr-im/server/internal/service"
	"whr-im/server/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	authService *service.AuthService
	hub         *ws.Hub
	upgrader    websocket.Upgrader
}

func NewWebSocketHandler(authService *service.AuthService, hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		authService: authService,
		hub:         hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *WebSocketHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	userID, err := h.authService.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.Register(userID, conn)
	defer func() {
		h.hub.Unregister(userID)
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
