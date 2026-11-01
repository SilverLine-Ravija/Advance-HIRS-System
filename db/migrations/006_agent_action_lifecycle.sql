-- Allow a user to revoke a pending AI proposal without executing the HRIS mutation.
-- Idempotent because db/migrate.sh replays every migration for local development.
ALTER TABLE agent_action_proposals
    ADD COLUMN IF NOT EXISTS dismissed_at TIMESTAMPTZ;

ALTER TABLE agent_action_proposals
    DROP CONSTRAINT IF EXISTS agent_action_proposals_status_check;

ALTER TABLE agent_action_proposals
    ADD CONSTRAINT agent_action_proposals_status_check
    CHECK (status IN ('pending', 'running', 'completed', 'failed', 'dismissed'));

-- Existing actor-index covers newest-first history reads.
-- Historical "expired" is computed from expires_at; pending rows need no cleanup to be safe.
