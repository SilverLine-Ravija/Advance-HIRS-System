package httpapi

import (
    "net/http"
    "strings"

    "github.com/savisaluwadana/Advance-HIRS-System/apps/api/internal/store"
)

// logout runs after authorize has verified the token signature, current
// membership and revocation state. A successful response means the token can
// no longer be used by the browser, desktop or AI agent.
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
    header := strings.TrimSpace(r.Header.Get("Authorization"))
    if len(header) < len("Bearer ") || !strings.EqualFold(header[:len("Bearer")], "Bearer") {
        writeError(w, http.StatusUnauthorized, "unauthorized", "bearer token required")
        return
    }
    raw := strings.TrimSpace(header[len("Bearer "):])
    claims, err := s.auth.Parse(raw)
    if err != nil || claims.ExpiresAt == nil {
        writeError(w, http.StatusUnauthorized, "unauthorized", "invalid session")
        return
    }
    if err := s.store.RevokeSession(r.Context(), store.SessionDigest(raw), claims.OrganizationID, claims.UserID, claims.ExpiresAt.Time); err != nil {
        writeError(w, http.StatusServiceUnavailable, "session_store_unavailable", "could not revoke the session")
        return
    }
    s.store.RecordAudit(r.Context(), claims.OrganizationID, claims.UserID, "auth.session_revoked", "user", claims.UserID, `{}`)
    w.Header().Set("Cache-Control", "no-store")
    writeJSON(w, http.StatusOK, map[string]bool{"revoked": true})
}
