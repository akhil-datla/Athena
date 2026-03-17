package engine

import (
	"sync"
	"testing"
)

func sampleEntries() []Entry {
	return []Entry{
		{Question: "How do I reset my password?", Answer: "Go to Settings > Security > Reset Password."},
		{Question: "What are your business hours?", Answer: "We are open Monday to Friday, 9am to 5pm."},
		{Question: "How do I contact support?", Answer: "Email support@example.com or call 1-800-HELP."},
		{Question: "Where is my order?", Answer: "Check your order status at orders.example.com."},
		{Question: "How do I cancel my subscription?", Answer: "Go to Settings > Billing > Cancel."},
	}
}

func TestNew(t *testing.T) {
	eng := New(sampleEntries())
	if eng.Count() != 5 {
		t.Errorf("expected 5 entries, got %d", eng.Count())
	}
}

func TestSearch(t *testing.T) {
	eng := New(sampleEntries())

	tests := []struct {
		name     string
		query    string
		limit    int
		minCount int
	}{
		{"exact match", "reset my password", 0, 1},
		{"partial match", "password", 0, 1},
		{"limited results", "how", 2, 2},
		{"no match", "xyznonexistent", 0, 0},
		{"case insensitive", "RESET PASSWORD", 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := eng.Search(tt.query, tt.limit)
			if len(results) < tt.minCount {
				t.Errorf("expected at least %d results, got %d", tt.minCount, len(results))
			}
			if tt.limit > 0 && len(results) > tt.limit {
				t.Errorf("expected at most %d results, got %d", tt.limit, len(results))
			}
		})
	}
}

func TestSearchReturnsScores(t *testing.T) {
	eng := New(sampleEntries())
	results := eng.Search("reset password", 0)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	// First result should be the best match
	for i := 1; i < len(results); i++ {
		if results[i].Score < results[i-1].Score {
			t.Errorf("results not sorted by score: index %d (score %d) < index %d (score %d)",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestEntries(t *testing.T) {
	entries := sampleEntries()
	eng := New(entries)
	got := eng.Entries()
	if len(got) != len(entries) {
		t.Errorf("expected %d entries, got %d", len(entries), len(got))
	}
	// Ensure it's a copy
	got[0].Question = "modified"
	original := eng.Entries()
	if original[0].Question == "modified" {
		t.Error("Entries() should return a copy, not a reference")
	}
}

func TestReload(t *testing.T) {
	eng := New(sampleEntries())
	if eng.Count() != 5 {
		t.Fatalf("expected 5 entries, got %d", eng.Count())
	}

	newEntries := []Entry{
		{Question: "New question?", Answer: "New answer."},
	}
	eng.Reload(newEntries)
	if eng.Count() != 1 {
		t.Errorf("expected 1 entry after reload, got %d", eng.Count())
	}

	results := eng.Search("new question", 0)
	if len(results) == 0 {
		t.Error("expected to find new entry after reload")
	}
}

func TestConcurrentAccess(t *testing.T) {
	eng := New(sampleEntries())
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eng.Search("password", 5)
		}()
	}

	wg.Wait()
}
