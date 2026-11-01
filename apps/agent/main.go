package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	port, apiURL, databaseURL, modelURL, model, apiKey string
}

type server struct {
	config   config
	client   *http.Client
	pool     *pgxpool.Pool
	limitsMu sync.Mutex
	limits   map[string]requestCount
}

type requestCount struct {
	start time.Time
	count int
}

type principal struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
	DisplayName    string `json:"display_name"`
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://127.0.0.1:" + env("AGENT_PORT", "8090") + "/health")
		if err != nil {
			os.Exit(1)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}

	cfg := config{
		port:        env("AGENT_PORT", "8090"),
		apiURL:      strings.TrimRight(env("HRIS_API_URL", "http://localhost:8080"), "/"),
		databaseURL: os.Getenv("AGENT_DATABASE_URL"),
		modelURL:    strings.TrimRight(env("OPENAI_BASE_URL", "https://api.openai.com/v1"), "/"),
		model:       env("OPENAI_MODEL", "gpt-4.1-mini"),
		apiKey:      os.Getenv("OPENAI_API_KEY"),
	}
	if cfg.apiKey == "" || cfg.databaseURL == "" {
		slog.Error("OPENAI_API_KEY and AGENT_DATABASE_URL are required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.databaseURL)
	if err != nil {
		slog.Error("database configuration failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		slog.Error("agent state database unavailable", "error", err)
		os.Exit(1)
	}

	s := &server{config: cfg, client: &http.Client{Timeout: 30 * time.Second}, pool: pool, limits: make(map[string]requestCount)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /api/agent/chat", s.chat)
	mux.HandleFunc("POST /api/agent/confirm", s.confirm)
	mux.HandleFunc("GET /api/agent/actions", s.actionsHistory)
	mux.HandleFunc("POST /api/agent/actions/{id}/dismiss", s.dismissAction)
	app := &http.Server{Addr: ":" + cfg.port, Handler: s.secure(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 35 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 45 * time.Second}
	slog.Info("agent service listening", "port", cfg.port)
	if err := app.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("agent service stopped", "error", err)
		os.Exit(1)
	}
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		jsonError(w, 503, "unavailable", "agent state database unavailable")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *server) secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		defer func() {
			if v := recover(); v != nil {
				slog.Error("unhandled agent error", "panic", fmt.Sprint(v))
				jsonError(w, 500, "internal_error", "agent request failed")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *server) authenticated(r *http.Request) (principal, string, error) {
	var p principal
	v := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(v), "bearer ") {
		return p, "", errors.New("sign in required")
	}
	token := strings.TrimSpace(v[7:])
	if len(token) < 16 || len(token) > 8192 {
		return p, "", errors.New("invalid session")
	}
	body, status, err := s.apiCall(r.Context(), token, "GET", "/api/v1/auth/me", nil)
	if err != nil || status != 200 {
		return p, "", errors.New("session expired or invalid")
	}
	var response struct {
		User principal `json:"user"`
	}
	if err := json.Unmarshal(body, &response); err != nil || response.User.UserID == "" || response.User.OrganizationID == "" {
		return p, "", errors.New("invalid session identity")
	}
	return response.User, token, nil
}

func (s *server) permitted(p principal) bool {
	return p.Role == "employee" || p.Role == "manager" || p.Role == "hr" || p.Role == "admin"
}

func (s *server) allow(userID string) bool {
	now := time.Now()
	s.limitsMu.Lock()
	defer s.limitsMu.Unlock()
	if len(s.limits) > 2000 {
		for k, v := range s.limits {
			if now.Sub(v.start) > time.Minute {
				delete(s.limits, k)
			}
		}
	}
	entry, ok := s.limits[userID]
	if !ok || now.Sub(entry.start) >= time.Minute {
		s.limits[userID] = requestCount{start: now, count: 1}
		return true
	}
	if entry.count >= 12 {
		return false
	}
	entry.count++
	s.limits[userID] = entry
	return true
}

func (s *server) apiCall(ctx context.Context, bearer, method, path string, payload []byte) ([]byte, int, error) {
	if !strings.HasPrefix(path, "/api/v1/") || strings.ContainsAny(path, "\r\n") {
		return nil, 0, errors.New("invalid API path")
	}
	var reader io.Reader
	if payload != nil {
		reader = strings.NewReader(string(payload))
	}
	req, err := http.NewRequestWithContext(ctx, method, s.config.apiURL+path, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 200*1024))
	return body, resp.StatusCode, err
}

func readJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		jsonError(w, 400, "invalid_request", "invalid JSON request")
		return false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		jsonError(w, 400, "invalid_request", "multiple JSON values are not allowed")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func jsonError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
