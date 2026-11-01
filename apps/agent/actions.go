package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

var proposalIDPattern = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)

type confirmRequest struct {
	ActionID string `json:"action_id"`
}
type confirmResponse struct {
	ActionID string `json:"action_id"`
	Status   string `json:"status"`
	Result   any    `json:"result,omitempty"`
	Message  string `json:"message,omitempty"`
}

func (s *server) stageAction(ctx context.Context, p principal, plan plannedTool, args map[string]any) (actionPreview, error) {
	if !plan.Mutation {
		return actionPreview{}, errors.New("not a mutation")
	}
	argumentJSON, err := json.Marshal(args)
	if err != nil {
		return actionPreview{}, err
	}
	var id string
	expiry := time.Now().UTC().Add(10 * time.Minute)
	err = s.pool.QueryRow(ctx, `INSERT INTO agent_action_proposals
		(organization_id,actor_user_id,tool_name,arguments,preview,expires_at)
		VALUES ($1::uuid,$2::uuid,$3,$4::jsonb,$5,$6) RETURNING id::text`,
		p.OrganizationID, p.UserID, plan.Name, string(argumentJSON), plan.Preview, expiry).Scan(&id)
	if err != nil {
		return actionPreview{}, err
	}
	return actionPreview{ID: id, Tool: plan.Name, Preview: plan.Preview, ExpiresAt: expiry}, nil
}

func (s *server) confirm(w http.ResponseWriter, r *http.Request) {
	p, token, err := s.authenticated(r)
	if err != nil {
		jsonError(w, 401, "unauthorized", err.Error())
		return
	}
	if !s.permitted(p) {
		jsonError(w, 403, "forbidden", "not authorized")
		return
	}
	if !s.allow(p.UserID) {
		jsonError(w, 429, "rate_limit", "too many agent requests")
		return
	}
	var input confirmRequest
	if !readJSON(w, r, &input) {
		return
	}
	if !proposalIDPattern.MatchString(input.ActionID) {
		jsonError(w, 400, "invalid_action", "invalid action ID")
		return
	}

	// Single-use state transition is atomic. Parallel requests cannot execute a proposal twice.
	var name string
	var argsJSON []byte
	err = s.pool.QueryRow(r.Context(), `UPDATE agent_action_proposals SET status='running'
		WHERE id=$1::uuid AND organization_id=$2::uuid AND actor_user_id=$3::uuid
		AND status='pending' AND expires_at>now()
		RETURNING tool_name,arguments`, input.ActionID, p.OrganizationID, p.UserID).Scan(&name, &argsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		var status string
		var result []byte
		e := s.pool.QueryRow(r.Context(), `SELECT status,result FROM agent_action_proposals
			WHERE id=$1::uuid AND organization_id=$2::uuid AND actor_user_id=$3::uuid`, input.ActionID, p.OrganizationID, p.UserID).Scan(&status, &result)
		if e != nil {
			jsonError(w, 404, "action_unavailable", "action was not found or has expired")
			return
		}
		if status == "completed" {
			writeJSON(w, 200, confirmResponse{ActionID: input.ActionID, Status: status, Result: json.RawMessage(result), Message: "Previously executed; not repeated"})
			return
		}
		jsonError(w, 409, "action_unavailable", "action is already processing, failed or expired; it will not be repeated")
		return
	}
	if err != nil {
		jsonError(w, 503, "storage_unavailable", "could not acquire action")
		return
	}
	var args map[string]any
	decoderErr := json.Unmarshal(argsJSON, &args)
	plan, planErr := planTool(p, name, args)
	if decoderErr != nil || planErr != nil || !plan.Mutation {
		s.finalize(r.Context(), input.ActionID, "failed", nil, "Action is no longer permitted")
		jsonError(w, 403, "action_denied", "action is no longer authorized")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	response, status, callErr := s.apiCall(ctx, token, plan.Method, plan.Path, plan.Body)
	if callErr != nil {
		// A network timeout can occur after the HRIS accepted a mutation. Never retry automatically.
		s.finalize(context.WithoutCancel(r.Context()), input.ActionID, "failed", nil, "Backend status uncertain; check the HRIS before retrying")
		jsonError(w, 502, "outcome_uncertain", "backend response was interrupted; check the HRIS record before trying again")
		return
	}
	stored := map[string]any{"http_status": status, "body": limitText(string(response), 20000)}
	if status < 200 || status >= 300 {
		s.finalize(context.WithoutCancel(r.Context()), input.ActionID, "failed", stored, fmt.Sprintf("HRIS API rejected action (%d)", status))
		writeJSON(w, status, confirmResponse{ActionID: input.ActionID, Status: "failed", Result: stored, Message: "Action was rejected by the HRIS"})
		return
	}
	if err := s.finalize(context.WithoutCancel(r.Context()), input.ActionID, "completed", stored, ""); err != nil {
		jsonError(w, 503, "recording_failed", "HRIS accepted the action, but audit recording failed; check the record before retrying")
		return
	}
	writeJSON(w, http.StatusOK, confirmResponse{ActionID: input.ActionID, Status: "completed", Result: stored, Message: "HRIS confirmed the action"})
}

func (s *server) finalize(parent context.Context, id, status string, result any, errorMessage string) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	var resultJSON []byte
	if result != nil {
		resultJSON, _ = json.Marshal(result)
	}
	_, err := s.pool.Exec(ctx, `UPDATE agent_action_proposals
		SET status=$2,executed_at=now(),result=$3::jsonb,error_message=$4
		WHERE id=$1::uuid AND status='running'`, id, status, resultJSON, errorMessage)
	return err
}
