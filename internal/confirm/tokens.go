// ABOUTME: Generates and validates confirmation tokens for destructive operations.
// ABOUTME: Tokens are single-use, time-limited, and tied to specific operations.

package confirm

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// PendingToken holds information about a pending confirmation.
type PendingToken struct {
	Operation string
	Resource  string
	ExpiresAt time.Time
}

// TokenStore manages confirmation tokens.
type TokenStore struct {
	ttl    time.Duration
	mu     sync.Mutex
	tokens map[string]*PendingToken
}

// NewTokenStore creates a new token store with the given TTL.
func NewTokenStore(ttl time.Duration) *TokenStore {
	return &TokenStore{
		ttl:    ttl,
		tokens: make(map[string]*PendingToken),
	}
}

// Generate creates a new confirmation token for an operation.
func (s *TokenStore) Generate(operation, resource string) string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	token := "tok_" + hex.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token] = &PendingToken{
		Operation: operation,
		Resource:  resource,
		ExpiresAt: time.Now().Add(s.ttl),
	}

	return token
}

// Validate checks if a token is valid for the given operation and resource.
// Valid tokens are consumed (single-use).
func (s *TokenStore) Validate(token, operation, resource string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending, exists := s.tokens[token]
	if !exists {
		return false
	}

	// Check expiry
	if time.Now().After(pending.ExpiresAt) {
		delete(s.tokens, token)
		return false
	}

	// Check operation and resource match
	if pending.Operation != operation || pending.Resource != resource {
		return false
	}

	// Consume token
	delete(s.tokens, token)
	return true
}

// GetPending returns information about a pending token without consuming it.
func (s *TokenStore) GetPending(token string) *PendingToken {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending, exists := s.tokens[token]
	if !exists {
		return nil
	}

	if time.Now().After(pending.ExpiresAt) {
		delete(s.tokens, token)
		return nil
	}

	return pending
}

// Cleanup removes expired tokens. Call periodically if needed.
func (s *TokenStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, pending := range s.tokens {
		if now.After(pending.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}
