-- Durable, actor-bound and single-use proposals for AI-initiated HR mutations.
-- The agent service must have access only to this table, never to HR domain tables.
CREATE TABLE IF NOT EXISTS agent_action_proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    arguments JSONB NOT NULL,
    preview TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
      CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    executed_at TIMESTAMPTZ,
    result JSONB,
    error_message TEXT
);
CREATE INDEX IF NOT EXISTS agent_action_proposals_actor_idx
    ON agent_action_proposals (organization_id, actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS agent_action_proposals_expiry_idx
    ON agent_action_proposals (expires_at);

-- For production, create an agent-specific database role and grant it only
-- SELECT, INSERT, UPDATE on agent_action_proposals, plus schema usage.
-- Never expose this database URL in the Next.js client or the LLM request.
