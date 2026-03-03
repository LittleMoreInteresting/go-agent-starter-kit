package memory

import (
	"context"
	"sync"
)

type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string][]Message
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string][]Message)}
}

func (s *InMemoryStore) Append(_ context.Context, sessionID string, msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sessionID] = append(s.data[sessionID], msg)
	return nil
}

func (s *InMemoryStore) History(_ context.Context, sessionID string, limit int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := s.data[sessionID]
	if limit <= 0 || len(msgs) <= limit {
		cp := make([]Message, len(msgs))
		copy(cp, msgs)
		return cp, nil
	}
	start := len(msgs) - limit
	cp := make([]Message, limit)
	copy(cp, msgs[start:])
	return cp, nil
}

func (s *InMemoryStore) Close() error { return nil }
