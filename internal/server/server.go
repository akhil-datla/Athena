package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/akhil-datla/athena/internal/engine"
	"github.com/akhil-datla/athena/internal/web"
)

// Server is the Athena HTTP server.
type Server struct {
	engine    *engine.Engine
	port      int
	logger    *slog.Logger
	startTime time.Time
	server    *http.Server
}

// New creates a new Server.
func New(eng *engine.Engine, port int, logger *slog.Logger) *Server {
	return &Server{
		engine: eng,
		port:   port,
		logger: logger,
	}
}

// SearchResponse is the JSON response for search queries.
type SearchResponse struct {
	Query   string                `json:"query"`
	Count   int                   `json:"count"`
	Results []engine.SearchResult `json:"results"`
}

// StatsResponse is the JSON response for the stats endpoint.
type StatsResponse struct {
	Entries int    `json:"entries"`
	Uptime  string `json:"uptime"`
}

// Start starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/search", s.handleSearch)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/stats", s.handleStats)

	// Web UI
	mux.Handle("/", web.Handler())

	handler := s.withCORS(s.withLogging(mux))

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.startTime = time.Now()

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("server started", "port", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.writeError(w, http.StatusBadRequest, "missing required query parameter: q")
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 0 {
			s.writeError(w, http.StatusBadRequest, "limit must be a non-negative integer")
			return
		}
		limit = parsed
	}

	results := s.engine.Search(query, limit)

	s.writeJSON(w, http.StatusOK, SearchResponse{
		Query:   query,
		Count:   len(results),
		Results: results,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStats(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, StatsResponse{
		Entries: s.engine.Count(),
		Uptime:  time.Since(s.startTime).Round(time.Second).String(),
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("failed to write response", "error", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	s.writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"duration", time.Since(start).String(),
		)
	})
}
