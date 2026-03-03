package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
)

type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type SearchResult struct {
	Doc   Document
	Score int
}

type Base interface {
	Search(query string, topK int) []SearchResult
}

func LoadDocuments(path string) ([]Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read knowledge base: %w", err)
	}
	var docs []Document
	if err := json.Unmarshal(b, &docs); err != nil {
		return nil, fmt.Errorf("parse knowledge base: %w", err)
	}
	return docs, nil
}
