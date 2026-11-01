package main

import (
    "errors"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/jackc/pgx/v5"
)

type actionHistoryItem struct {
    ID           string     `json:"id"`
    Tool         string     `json:"tool"`
    Preview      string     `json:"preview"`
    Status       string     `json:"status"`
    CreatedAt    time.Time  `json:"created_at"`
    ExpiresAt    time.Time  `json:"expires_at"`
    ExecutedAt   *time.Time `json:"executed_at,omitempty"`
    ErrorMessage string     `json:"error_message,omitempty"`
}

// GET /api/agent/actions returns only the current actor's action history.
// Arguments and provider/tool output are deliberately omitted from the response.
func (s *server) actionsHistory(w http.ResponseWriter, r *http.Request) {
    p, _, err := s.authenticated(r)
    if err != nil {
        jsonError(w, http.StatusUnauthorized, "unauthorized", "Sign in required")
        return
    }
    if !s.permitted(p) {
        jsonError(w, http.StatusForbidden, "forbidden", "Not authorized")
        return
    }
    if !s.allow(p.UserID) {
        jsonError(w, http.StatusTooManyRequests, "rate_limit", "Too many assistant requests")
        return
    }
    limit := 20
    if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
        n, parseErr := strconv.Atoi(raw)
        if parseErr != nil || n < 1 || n > 50 {
            jsonError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 50")
            return
        }
        limit = n
    }
    rows, err := s.pool.Query(r.Context(), `SELECT id::text, tool_name, preview,
        CASE WHEN status='pending' AND expires_at <= now() THEN 'expired' ELSE status END,
        created_at, expires_at, executed_at, COALESCE(error_message,'')
        FROM agent_action_proposals
        WHERE organization_id=$1::uuid AND actor_user_id=$2::uuid
        ORDER BY created_at DESC, id DESC LIMIT $3`,
        p.OrganizationID, p.UserID, limit)
    if err != nil {
        jsonError(w, http.StatusServiceUnavailable, "storage_unavailable", "Action history unavailable")
        return
    }
    defer rows.Close()
    actions := make([]actionHistoryItem, 0, limit)
    for rows.Next() {
        var item actionHistoryItem
        if err := rows.Scan(&item.ID, &item.Tool, &item.Preview, &item.Status,
            &item.CreatedAt, &item.ExpiresAt, &item.ExecutedAt, &item.ErrorMessage); err != nil {
            jsonError(w, http.StatusServiceUnavailable, "storage_unavailable", "Action history unavailable")
            return
        }
        actions = append(actions, item)
    }
    if rows.Err() != nil {
        jsonError(w, http.StatusServiceUnavailable, "storage_unavailable", "Action history unavailable")
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"data": actions, "total": len(actions)})
}

// POST /api/agent/actions/{id}/dismiss atomically revokes a pending proposal.
// It cannot cancel an already running or completed HRIS mutation.
func (s *server) dismissAction(w http.ResponseWriter, r *http.Request) {
    p, _, err := s.authenticated(r)
    if err != nil {
        jsonError(w, http.StatusUnauthorized, "unauthorized", "Sign in required")
        return
    }
    if !s.permitted(p) {
        jsonError(w, http.StatusForbidden, "forbidden", "Not authorized")
        return
    }
    if !s.allow(p.UserID) {
        jsonError(w, http.StatusTooManyRequests, "rate_limit", "Too many assistant requests")
        return
    }
    id := r.PathValue("id")
    if !proposalIDPattern.MatchString(id) {
        jsonError(w, http.StatusBadRequest, "invalid_action", "Invalid action ID")
        return
    }
    var dismissed string
    err = s.pool.QueryRow(r.Context(), `UPDATE agent_action_proposals
        SET status='dismissed', dismissed_at=now()
        WHERE id=$1::uuid AND organization_id=$2::uuid AND actor_user_id=$3::uuid
          AND status='pending' AND expires_at > now()
        RETURNING id::text`, id, p.OrganizationID, p.UserID).Scan(&dismissed)
    if errors.Is(err, pgx.ErrNoRows) {
        jsonError(w, http.StatusConflict, "action_unavailable", "Action is no longer pending or available")
        return
    }
    if err != nil {
        jsonError(w, http.StatusServiceUnavailable, "storage_unavailable", "Could not dismiss the action")
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"action_id": dismissed, "status": "dismissed"})
}
