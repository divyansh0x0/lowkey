package signaling

import "encoding/json"

// Message types used in the signaling protocol.
const (
	// Session lifecycle
	TypeSessionCreate  = "session:create"
	TypeSessionCreated = "session:created"
	TypeSessionJoin    = "session:join"
	TypeSessionJoined  = "session:joined"

	// WebRTC signaling
	TypeSignalOffer  = "signal:offer"
	TypeSignalAnswer = "signal:answer"
	TypeSignalICE    = "signal:ice"

	// Errors
	TypeError = "error"
)

// Message is the envelope for all WebSocket communication.
type Message struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId,omitempty"`
	Target    string          `json:"target,omitempty"`
	Sender    string          `json:"sender,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// SessionCreatedPayload is sent to the creator after session creation.
type SessionCreatedPayload struct {
	SessionID string `json:"sessionId"`
	Key       string `json:"key"` // base64-encoded symmetric key
}

// SessionJoinedPayload is sent to both users when the joiner connects.
type SessionJoinedPayload struct {
	SessionID string `json:"sessionId"`
	Peer      string `json:"peer"` // the other user's username
	Key       string `json:"key"`  // base64-encoded symmetric key
}

// ErrorPayload carries error information.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
