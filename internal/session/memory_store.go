package session

import (
	"errors"
	"sync"
	"time"

	"github.com/ayush/lowkey/internal/crypto"
	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionFull     = errors.New("session is full")
	ErrAlreadyInSession = errors.New("user already in session")
)

// MemoryStore is a thread-safe, in-memory session store with automatic TTL cleanup.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
	ttl      time.Duration
	done     chan struct{}
}

// NewMemoryStore creates a new in-memory store and starts the background cleanup goroutine.
func NewMemoryStore(ttl time.Duration) *MemoryStore {
	s := &MemoryStore{
		sessions: make(map[uuid.UUID]*Session),
		ttl:      ttl,
		done:     make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Stop terminates the background cleanup goroutine.
func (s *MemoryStore) Stop() {
	close(s.done)
}

func (s *MemoryStore) Create(creator string) (*Session, error) {
	key, err := crypto.GenerateSessionKey()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sess := &Session{
		ID:        uuid.New(),
		Users:     [2]string{creator, ""},
		UserCount: 1,
		SharedKey: key,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}

	s.mu.Lock()
	s.sessions[sess.ID] = sess
	s.mu.Unlock()

	return sess, nil
}

func (s *MemoryStore) Join(sessionID uuid.UUID, joiner string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	if sess.UserCount >= 2 {
		return nil, ErrSessionFull
	}
	if sess.Users[0] == joiner {
		return nil, ErrAlreadyInSession
	}

	sess.Users[1] = joiner
	sess.UserCount = 2
	// Refresh expiry on join
	sess.ExpiresAt = time.Now().Add(s.ttl)

	return sess, nil
}

func (s *MemoryStore) Get(sessionID uuid.UUID) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return sess, nil
}

func (s *MemoryStore) Delete(sessionID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[sessionID]; !ok {
		return ErrSessionNotFound
	}
	delete(s.sessions, sessionID)
	return nil
}

// cleanup periodically removes expired sessions.
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for id, sess := range s.sessions {
				if now.After(sess.ExpiresAt) {
					delete(s.sessions, id)
				}
			}
			s.mu.Unlock()
		}
	}
}
