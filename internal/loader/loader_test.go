package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	content := `How do I reset my password?,Go to Settings > Security > Reset Password.
What are your hours?,Monday to Friday 9am-5pm.
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := LoadCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Question != "How do I reset my password?" {
		t.Errorf("unexpected question: %s", entries[0].Question)
	}
}

func TestLoadCSVWithHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	content := `question,answer
How do I reset my password?,Go to Settings.
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := LoadCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry (header skipped), got %d", len(entries))
	}
}

func TestLoadCSVEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadCSV(path)
	if err == nil {
		t.Error("expected error for empty CSV")
	}
}

func TestLoadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `[
  {"question": "How do I reset my password?", "answer": "Go to Settings."},
  {"question": "What are your hours?", "answer": "9am-5pm."}
]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLoadJSONInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadJSON(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadAutoDetect(t *testing.T) {
	dir := t.TempDir()

	// CSV
	csvPath := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(csvPath, []byte("Q?,A.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := Load(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	// JSON
	jsonPath := filepath.Join(dir, "data.json")
	if err := os.WriteFile(jsonPath, []byte(`[{"question":"Q?","answer":"A."}]`), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err = Load(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}

	// Unsupported
	txtPath := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(txtPath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = Load(txtPath)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/file.csv")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
