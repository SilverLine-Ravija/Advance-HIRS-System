package store

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "time"
)

// SessionDigest is a one-way fingerprint of a signed bearer access token.
// Never store the bearer token itself, even as audit metadata.
func SessionDigest(token string) string {
    digest := sha256.Sum256([]byte(token))
    return hex.EncodeToString(digest[:])
}

// SessionActive is checked for every protected request. It intentionally fails
// closed if PostgreSQL is unavailable and rejects role or membership changes
// without waiting for the JWT's expiration time.
func (s *Store) SessionActive(ctx context.Context, tokenDigest, orgID, userID, role string) (bool, error) {
    var active bool
    err := s.pool.QueryRow(ctx, `
      SELECT
        EXISTS (
          SELECT 1 FROM users u
          JOIN organization_memberships m ON m.user_id = u.id
          WHERE u.id = $1::uuid AND u.status = 'active'
            AND m.organization_id = $2::uuid
            AND m.status = 'active' AND m.role = $3
        )
        AND NOT EXISTS (
          SELECT 1 FROM revoked_sessions r
          WHERE r.token_hash = $4 AND r.expires_at > now()
        )`,
      userID, orgID, role, tokenDigest).Scan(&active)
    return active, err
}

// RevokeSession prevents this specific signed token from being used again.
// The caller's identity is taken from the verified JWT, not the request body.
func (s *Store) RevokeSession(ctx context.Context, tokenDigest, orgID, userID string, expiresAt time.Time) error {
    _, err := s.pool.Exec(ctx, `
      INSERT INTO revoked_sessions (token_hash,organization_id,user_id,expires_at)
      VALUES ($1, $2::uuid, $3::uuid, $4)
      ON CONFLICT (token_hash) DO NOTHING`,
      tokenDigest, orgID, userID, expiresAt)
    return err
}
