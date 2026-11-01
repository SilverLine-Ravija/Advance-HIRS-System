---
title: AI agent implementation
description: Agent orchestration, scoped HR tools, confirmation, security model and limitations.
section: docs
permalink: /agentic-hris/
---

# Agentic HRIS: Implementation guide

This increment adds an authenticated AI assistant without replacing existing business logic. **The HRIS API continues to enforce tenant isolation, leave accounting, manager scope and other domain invariants.** Agent tools cannot access the PostgreSQL HR data tables directly.

## Features shipped in this increment

- Conversational HR copilot backed by OpenAI-compatible Chat Completions with structured tool calls.
- Employee/manager/HR/admin role-specific tool catalogs, verified against the current JWT identity at every request.
- Leave balances, policies, holidays and scoped leave requests; leave application preparation with human confirmation.
- Attendance, correction queue, work schedules and period summaries; correction preparation with confirmation.
- Manager review proposals and limited HR record creation/updates with confirmation.
- Workforce insights based on existing dashboard and directory APIs (no invented KPIs).
- Single-use PostgreSQL-persisted action proposals, 10-minute expiry, audit outcome, retry prevention and bound actor/tenant.
- Next.js `/assistant` interface and same-origin server-side BFF API proxy.
- Separate agent container deployable as another OpenChoreo application component.

Not shipped: payroll computations, candidate recruiting pipeline, employee document ingestion, RAG over company handbooks, Slack/Teams/WhatsApp notifications, unattended scheduled jobs, automated hiring/performance decisions, or native offline Wails agent UI. Those require new business modules and explicit product requirements. AI is not silently presented as those modules.

## Running in local development

1. Clone the current repository and check out the default branch. The existing migrations script automatically runs `005_agent_actions.sql`.
2. Set an OpenAI-compatible provider key in your local environment. Never commit it or expose it through `NEXT_PUBLIC_` variables.
3. Start the stack:

```sh
export OPENAI_API_KEY='your-key-from-secret-manager'
docker compose -f docker-compose.yml -f docker-compose.agentic.yml up --build
```

4. Sign in at `http://localhost:3000` using demo-only accounts, then open `http://localhost:3000/assistant`.
5. Test read-only questions first, then test a leave request in a disposable organization. Confirm that an unconfirmed action does not write to the HRIS database.

On OpenChoreo, deploy `apps/agent` as **a private service on port 8090**, wire `AGENT_INTERNAL_URL` in the web component, `HRIS_API_URL` in agent, and inject `OPENAI_API_KEY`, `AGENT_DATABASE_URL`, `OPENAI_MODEL` from environment secrets. Do not publish the agent service directly to the internet. Connect to a managed PostgreSQL instance whose migrations have run. The desktop application remains a normal HRIS API client.

## Security model

- Browser → Next.js BFF: session token stays HttpOnly; server forwards it to internal agent service, not JavaScript.
- Agent → HRIS: the agent resolves the current principal through `/auth/me` and calls the existing role- and tenant-checked API for **every** tool operation.
- Tool outputs and user-entered documents are untrusted text. Model tool names/arguments are checked in Go against a server-side allowlist, exact allowed fields, input types and current role.
- Every mutation is staged in `agent_action_proposals` and **not executed until the exact authenticated user confirms**. The `pending → running` transition is atomic, so parallel confirmations can't execute twice. The HRIS itself continues to approve or reject leave according to its existing policies.
- The table should be accessible to a dedicated DB role with privileges restricted to agent action state. Development compose uses the demo database login for convenience only; replace it in production.
- After a request times out, status is considered uncertain; the client must not blindly retry a mutation. Operators should inspect the HRIS state and reconcile any `running` actions.
- Agent tool responses are limited in size; large responses can be truncated and may require future API pagination. Conversational history is client-held and bounded, not persisted by the agent. Models may receive HR data permitted by the user; verify your AI vendor's enterprise privacy and retention agreements before using production employee records.
- Rate limiting here is per-process and only supplemental. Set centralized ingress rate limits, model spend caps, API logging, alerting, backups, secrets rotation, session revocation and SSO/OIDC before deployment to customers.

## Current limitations / known follow-ups

1. **Test infrastructure.** CI runs Go tool tests and builds the Docker image. PostgreSQL CI tests verify one-time proposal execution under concurrent confirmations. Extend the suite for actual HRIS RBAC, revalidation after role changes and provider failures.
2. **Approval ownership.** Confirming an agent-proposed manager decision is a human action, but the existing Go HRIS manager permission checks are still the source of truth.
3. **Freshness.** Model may cite stale tool outputs; use per-request read validation before a mutation and never assume a previous balance remains unchanged. The existing HRIS transactional approval path handles balance conflicts.
4. **UX.** Dismissed proposals remain in the database until expiry. Add a UI for reviewing past action status, accessibility evaluation, and localization.
5. **Company policies.** This increment only reads structured policies already in PostgreSQL. Document and handbook RAG require authorization-aware document storage, chunking, citations and data-retention policies.
6. **Multi-agent architecture.** This intentionally starts with **one orchestrator** and role-scoped tools. Separate autonomous agents would introduce unnecessary orchestration, privacy and consistency risks before validating customer demand.
7. **AI quality.** Add evaluation datasets, prompt-injection tests, abuse tests, costs per successful HR action, human override metrics and user feedback before customer rollout.
