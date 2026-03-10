package signaling

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

// HandleWebSocket is the HTTP handler for WebSocket signaling connections.
// Endpoint: GET /ws?username=<name>
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username query parameter", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // accept all origins in dev; tighten for production
	})
	if err != nil {
		log.Printf("[handler] websocket accept error: %v", err)
		return
	}

	h.Register(username, conn)
	defer func() {
		h.Unregister(username)
		conn.CloseNow()
	}()

	ctx := r.Context()
	h.readLoop(ctx, username, conn)
}

// readLoop reads messages from a WebSocket connection and dispatches them.
func (h *Hub) readLoop(ctx context.Context, username string, conn *websocket.Conn) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				log.Printf("[handler] %s disconnected: %v", username, err)
			} else {
				log.Printf("[handler] %s read error: %v", username, err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			h.SendError(ctx, username, "INVALID_MESSAGE", "malformed JSON")
			continue
		}

		msg.Sender = username // server stamps the sender
		h.dispatch(ctx, username, msg)
	}
}

// dispatch routes a message based on its type.
func (h *Hub) dispatch(ctx context.Context, username string, msg Message) {
	switch msg.Type {
	case TypeSessionCreate:
		h.handleSessionCreate(ctx, username)

	case TypeSessionJoin:
		h.handleSessionJoin(ctx, username, msg)

	case TypeSignalOffer, TypeSignalAnswer, TypeSignalICE:
		h.handleSignalRelay(ctx, username, msg)

	default:
		h.SendError(ctx, username, "UNKNOWN_TYPE", "unrecognised message type: "+msg.Type)
	}
}

// handleSessionCreate creates a new session and returns the UUID + encryption key to the creator.
func (h *Hub) handleSessionCreate(ctx context.Context, creator string) {
	sess, err := h.store.Create(creator)
	if err != nil {
		log.Printf("[handler] session create error: %v", err)
		h.SendError(ctx, creator, "SESSION_CREATE_FAILED", err.Error())
		return
	}

	payload, _ := json.Marshal(SessionCreatedPayload{
		SessionID: sess.ID.String(),
		Key:       base64.StdEncoding.EncodeToString(sess.SharedKey),
	})

	_ = h.Send(ctx, creator, Message{
		Type:    TypeSessionCreated,
		Payload: payload,
	})

	log.Printf("[handler] session created: %s by %s", sess.ID, creator)
}

// handleSessionJoin adds a user to an existing session and notifies both peers with the encryption key.
func (h *Hub) handleSessionJoin(ctx context.Context, joiner string, msg Message) {
	if msg.SessionID == "" {
		h.SendError(ctx, joiner, "MISSING_SESSION_ID", "sessionId is required")
		return
	}

	sessionID, err := uuid.Parse(msg.SessionID)
	if err != nil {
		h.SendError(ctx, joiner, "INVALID_SESSION_ID", "invalid sessionId format")
		return
	}

	sess, err := h.store.Join(sessionID, joiner)
	if err != nil {
		h.SendError(ctx, joiner, "SESSION_JOIN_FAILED", err.Error())
		return
	}

	encodedKey := base64.StdEncoding.EncodeToString(sess.SharedKey)

	// Notify the joiner
	joinerPayload, _ := json.Marshal(SessionJoinedPayload{
		SessionID: sess.ID.String(),
		Peer:      sess.Users[0], // the creator
		Key:       encodedKey,
	})
	_ = h.Send(ctx, joiner, Message{
		Type:    TypeSessionJoined,
		Payload: joinerPayload,
	})

	// Notify the creator
	creatorPayload, _ := json.Marshal(SessionJoinedPayload{
		SessionID: sess.ID.String(),
		Peer:      joiner,
		Key:       encodedKey,
	})
	_ = h.Send(ctx, sess.Users[0], Message{
		Type:    TypeSessionJoined,
		Payload: creatorPayload,
	})

	log.Printf("[handler] %s joined session %s (creator: %s)", joiner, sess.ID, sess.Users[0])
}

// handleSignalRelay forwards signaling messages (offer, answer, ICE) to the target peer.
func (h *Hub) handleSignalRelay(ctx context.Context, sender string, msg Message) {
	if msg.Target == "" {
		h.SendError(ctx, sender, "MISSING_TARGET", "target is required for signaling")
		return
	}

	// Stamp the sender and relay
	msg.Sender = sender
	if err := h.Send(ctx, msg.Target, msg); err != nil {
		h.SendError(ctx, sender, "RELAY_FAILED", "could not reach target: "+msg.Target)
	}
}
