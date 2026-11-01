package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CI runs this against PostgreSQL with the normal HRIS schema/migrations applied.
// It verifies single-use DB action claims under parallel user confirmations.
func TestSingleUseConfirmationIntegration(t *testing.T) {
	dsn := os.Getenv("AGENT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set AGENT_TEST_DATABASE_URL to enable PostgreSQL integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	slug := fmt.Sprintf("agent-test-%d", time.Now().UnixNano())
	var orgID, userID string
	if err := pool.QueryRow(ctx, `INSERT INTO organizations(name,slug) VALUES ('Agent Test',$1) RETURNING id::text`, slug).Scan(&orgID); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(context.Background(), `DELETE FROM organizations WHERE id=$1::uuid`, orgID)
	if err := pool.QueryRow(ctx, `INSERT INTO users(email,password_hash,display_name) VALUES ($1,'test-password-hash','Agent Test') RETURNING id::text`, slug+"@example.invalid").Scan(&userID); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(context.Background(), `DELETE FROM users WHERE id=$1::uuid`, userID)
	p := principal{UserID: userID, OrganizationID: orgID, Role: "employee"}
	var writes atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/me":
			_ = json.NewEncoder(w).Encode(map[string]any{"user": p})
		case "/api/v1/leave/requests":
			writes.Add(1)
			time.Sleep(30 * time.Millisecond)
			w.WriteHeader(201)
			_, _ = w.Write([]byte(`{"id":"leave-test-1"}`))
		default:
			http.Error(w, "unsupported", 404)
		}
	}))
	defer upstream.Close()
	s := &server{config: config{apiURL: upstream.URL}, client: &http.Client{Timeout: 2 * time.Second}, pool: pool, limits: map[string]requestCount{}}
	args := map[string]any{"leave_type": "annual", "start_date": "2026-11-16", "end_date": "2026-11-17", "reason": "Test only"}
	plan, err := planTool(p, "submit_leave_request", args)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := s.stageAction(ctx, p, plan, args)
	if err != nil {
		t.Fatal(err)
	}
	if preview.ID == "" {
		t.Fatal("no proposal ID")
	}

	responses := make([]int, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/api/agent/confirm", bytes.NewBufferString(fmt.Sprintf(`{"action_id":%q}`, preview.ID)))
			req.Header.Set("Authorization", "Bearer integration-test-token")
			recorder := httptest.NewRecorder()
			s.confirm(recorder, req)
			responses[i] = recorder.Code
		}(i)
	}
	wg.Wait()
	if writes.Load() != 1 {
		t.Fatalf("expected one real HRIS write, got %d; HTTP statuses: %v", writes.Load(), responses)
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM agent_action_proposals WHERE id=$1::uuid`, preview.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "completed" {
		t.Fatalf("expected completed action, got %s", status)
	}

	// A later replay returns the recorded result instead of invoking the API again.
	req := httptest.NewRequest(http.MethodPost, "/api/agent/confirm", strings.NewReader(fmt.Sprintf(`{"action_id":%q}`, preview.ID)))
	req.Header.Set("Authorization", "Bearer integration-test-token")
	recorder := httptest.NewRecorder()
	s.confirm(recorder, req)
	if recorder.Code != 200 || writes.Load() != 1 {
		t.Fatalf("replay was not safely ignored: HTTP %d, writes %d", recorder.Code, writes.Load())
	}
}
