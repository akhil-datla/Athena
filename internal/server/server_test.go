package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/akhil-datla/athena/internal/engine"
)

func testServer() *Server {
	entries := []engine.Entry{
		{Question: "How do I reset my password?", Answer: "Go to Settings > Security."},
		{Question: "What are your hours?", Answer: "9am-5pm Monday to Friday."},
		{Question: "How do I contact support?", Answer: "Email support@example.com."},
	}
	eng := engine.New(entries)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return New(eng, 0, logger)
}

func TestSearchEndpoint(t *testing.T) {
	srv := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=password&limit=5", nil)
	w := httptest.NewRecorder()

	srv.handleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp SearchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.Query != "password" {
		t.Errorf("expected query 'password', got %q", resp.Query)
	}
	if resp.Count == 0 {
		t.Error("expected at least one result")
	}
	if len(resp.Results) != resp.Count {
		t.Errorf("count %d doesn't match results length %d", resp.Count, len(resp.Results))
	}
}

func TestSearchMissingQuery(t *testing.T) {
	srv := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	w := httptest.NewRecorder()

	srv.handleSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSearchInvalidLimit(t *testing.T) {
	srv := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=test&limit=abc", nil)
	w := httptest.NewRecorder()

	srv.handleSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	srv := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
}

func TestStatsEndpoint(t *testing.T) {
	srv := testServer()
	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	srv.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp StatsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Entries != 3 {
		t.Errorf("expected 3 entries, got %d", resp.Entries)
	}
}

func TestCORSHeaders(t *testing.T) {
	srv := testServer()
	handler := srv.withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected CORS origin '*', got %q", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	srv := testServer()
	handler := srv.withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for preflight, got %d", w.Code)
	}
}
