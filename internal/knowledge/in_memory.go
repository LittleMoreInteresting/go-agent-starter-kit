package knowledge

import (
	"sort"
	"strings"
)

type InMemoryBase struct {
	docs []Document
}

func NewInMemoryBase(docs []Document) *InMemoryBase {
	return &InMemoryBase{docs: docs}
}

func (b *InMemoryBase) Search(query string, topK int) []SearchResult {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" || topK <= 0 {
		return nil
	}
	terms := strings.Fields(q)
	results := make([]SearchResult, 0)
	for _, d := range b.docs {
		text := strings.ToLower(d.Title + "\n" + d.Content)
		score := 0
		if strings.Contains(text, q) {
			score += 5
		}
		for _, term := range terms {
			if term != "" && strings.Contains(text, term) {
				score++
			}
		}
		if score > 0 {
			results = append(results, SearchResult{Doc: d, Score: score})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Doc.ID < results[j].Doc.ID
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > topK {
		return results[:topK]
	}
	return results
}
