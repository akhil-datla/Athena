package engine

import (
	"sort"
	"strings"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Entry represents a single question-answer pair.
type Entry struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// SearchResult is a ranked search result with a relevance score.
type SearchResult struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Score    int    `json:"score"`
}

// Engine performs fuzzy search over a collection of Q&A entries.
type Engine struct {
	mu        sync.RWMutex
	entries   []Entry
	questions []string
	index     map[string]string
}

// New creates a new Engine from the given entries.
func New(entries []Entry) *Engine {
	e := &Engine{
		entries:   make([]Entry, len(entries)),
		questions: make([]string, len(entries)),
		index:     make(map[string]string, len(entries)),
	}
	copy(e.entries, entries)
	for i, entry := range entries {
		e.questions[i] = entry.Question
		e.index[strings.ToLower(entry.Question)] = entry.Answer
	}
	return e
}

// Search performs a fuzzy search and returns results ranked by relevance.
// A limit of 0 or less returns all matches.
func (e *Engine) Search(query string, limit int) []SearchResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	matches := fuzzy.RankFindNormalizedFold(query, e.questions)
	sort.Sort(matches)

	results := make([]SearchResult, 0, len(matches))
	for _, m := range matches {
		answer, ok := e.index[strings.ToLower(m.Target)]
		if !ok {
			continue
		}
		results = append(results, SearchResult{
			Question: m.Target,
			Answer:   answer,
			Score:    m.Distance,
		})
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

// Count returns the total number of entries in the engine.
func (e *Engine) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.entries)
}

// Entries returns a copy of all entries.
func (e *Engine) Entries() []Entry {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Entry, len(e.entries))
	copy(out, e.entries)
	return out
}

// Reload replaces the engine's data with new entries.
func (e *Engine) Reload(entries []Entry) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.entries = make([]Entry, len(entries))
	e.questions = make([]string, len(entries))
	e.index = make(map[string]string, len(entries))

	copy(e.entries, entries)
	for i, entry := range entries {
		e.questions[i] = entry.Question
		e.index[strings.ToLower(entry.Question)] = entry.Answer
	}
}
