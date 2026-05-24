package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu    sync.RWMutex
	conns map[uint64]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{conns: make(map[uint64]*websocket.Conn)}
}

func (h *Hub) Register(userID uint64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = conn
}

func (h *Hub) Unregister(userID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, userID)
}

func (h *Hub) Send(userID uint64, messageType string, payload any) error {
	h.mu.RLock()
	conn, ok := h.conns[userID]
	h.mu.RUnlock()
	if !ok {
		return nil
	}
	envelope := map[string]any{"type": messageType, "data": payload}
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, body)
}
