package loader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/akhil-datla/athena/internal/engine"
)

// Load reads Q&A entries from a file, auto-detecting format by extension.
// Supported formats: .csv, .json
func Load(path string) ([]engine.Entry, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return LoadCSV(path)
	case ".json":
		return LoadJSON(path)
	default:
		return nil, fmt.Errorf("unsupported file format %q (use .csv or .json)", ext)
	}
}

// LoadCSV reads Q&A entries from a CSV file.
// Expected format: question,answer (no header row required).
func LoadCSV(path string) ([]engine.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening csv: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // allow variable fields
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading csv: %w", err)
	}

	entries := make([]engine.Entry, 0, len(records))
	for i, record := range records {
		if len(record) < 2 {
			continue
		}
		q := strings.TrimSpace(record[0])
		a := strings.TrimSpace(record[1])
		if q == "" {
			continue
		}
		// Skip header row if detected
		if i == 0 && strings.EqualFold(q, "question") && strings.EqualFold(a, "answer") {
			continue
		}
		entries = append(entries, engine.Entry{Question: q, Answer: a})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no valid entries found in %s", path)
	}
	return entries, nil
}

// LoadJSON reads Q&A entries from a JSON file.
// Expected format: [{"question": "...", "answer": "..."}]
func LoadJSON(path string) ([]engine.Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading json: %w", err)
	}

	var entries []engine.Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}

	// Filter out empty entries
	valid := make([]engine.Entry, 0, len(entries))
	for _, e := range entries {
		if strings.TrimSpace(e.Question) != "" {
			valid = append(valid, e)
		}
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid entries found in %s", path)
	}
	return valid, nil
}
