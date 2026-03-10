package session

import "github.com/google/uuid"

// Store defines the interface for session persistence.
// The default implementation is in-memory; swap in Redis for horizontal scaling.
type Store interface {
	// Create starts a new session with the given creator username.
	Create(creator string) (*Session, error)

	// Join adds a second user to an existing session.
	Join(sessionID uuid.UUID, joiner string) (*Session, error)

	// Get retrieves a session by ID.
	Get(sessionID uuid.UUID) (*Session, error)

	// Delete removes a session.
	Delete(sessionID uuid.UUID) error
}
