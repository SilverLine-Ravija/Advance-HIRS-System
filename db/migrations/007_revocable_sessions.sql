-- Signed access tokens are never stored in plaintext.
-- Persist only SHA-256 digests of explicitly revoked bearer JWTs.
-- Safe to replay through db/migrate.sh.
CREATE TABLE IF NOT EXISTS revoked_sessions (
    token_hash CHAR(64) PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_revoked_sessions_expires ON revoked_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_revoked_sessions_actor ON revoked_sessions(organization_id,user_id,revoked_at DESC);
