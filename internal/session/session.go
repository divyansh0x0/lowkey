package session

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a P2P chat session between exactly two users.
type Session struct {
	ID        uuid.UUID `json:"id"`
	Users     [2]string `json:"users"`     // Users[0] = creator, Users[1] = joiner
	UserCount int       `json:"userCount"` // 1 after creation, 2 after join
	SharedKey []byte    `json:"-"`         // symmetric key — never serialised to clients directly
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}
