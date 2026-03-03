package memory

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Store interface {
	Append(ctx context.Context, sessionID string, msg Message) error
	History(ctx context.Context, sessionID string, limit int) ([]Message, error)
	Close() error
}
