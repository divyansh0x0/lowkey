package signaling

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/ayush/lowkey/internal/session"
	"github.com/coder/websocket"
)

// Hub manages active WebSocket connections and routes signaling messages.
type Hub struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn // username → connection
	store session.Store
}

// NewHub creates a new signaling hub.
func NewHub(store session.Store) *Hub {
	return &Hub{
		conns: make(map[string]*websocket.Conn),
		store: store,
	}
}

// Register associates a username with a WebSocket connection.
func (h *Hub) Register(username string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// If user already has a connection, close the old one.
	if old, ok := h.conns[username]; ok {
		old.Close(websocket.StatusGoingAway, "replaced by new connection")
	}
	h.conns[username] = conn
	log.Printf("[hub] registered: %s", username)
}

// Unregister removes a user's connection.
func (h *Hub) Unregister(username string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.conns, username)
	log.Printf("[hub] unregistered: %s", username)
}

// Send delivers a JSON message to the specified user.
func (h *Hub) Send(ctx context.Context, username string, msg Message) error {
	h.mu.RLock()
	conn, ok := h.conns[username]
	h.mu.RUnlock()

	if !ok {
		return ErrUserOffline
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return conn.Write(ctx, websocket.MessageText, data)
}

// SendError sends an error message to a specific user.
func (h *Hub) SendError(ctx context.Context, username, code, message string) {
	payload, _ := json.Marshal(ErrorPayload{Code: code, Message: message})
	_ = h.Send(ctx, username, Message{
		Type:    TypeError,
		Payload: payload,
	})
}

// IsOnline checks if a user is currently connected.
func (h *Hub) IsOnline(username string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.conns[username]
	return ok
}
