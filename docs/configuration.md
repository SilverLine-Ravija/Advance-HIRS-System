---
title: Configuration reference
description: Environment variables, Docker Compose, model provider and deployment boundaries.
section: docs
permalink: /configuration/
---

# Configuration reference

Advance HRIS has three deployable server-side components in the full agentic setup: Next.js web (`3000`), Go HRIS API (`8080`) and the optional private Go agent (`8090`). PostgreSQL is the shared backend datastore, with agent privileges separated for production.

## Core environment variables

| Variable | Component | Purpose |
| --- | --- | --- |
| `DATABASE_URL` | HRIS API/migration job | PostgreSQL DSN for HR data |
| `JWT_SECRET` | HRIS API | High-entropy signing key; at least 32 characters |
| `WEB_ORIGIN` | HRIS API | Exact allowed web origin |
| `API_PORT` | HRIS API | Service port (defaults to 8080) |
| `API_INTERNAL_URL` | Web | Server-side upstream Go API |
| `WEB_COOKIE_SECURE` | Web | Set false only for local HTTP; true in HTTPS production |
| `AGENT_INTERNAL_URL` | Web | Private agent upstream; no browser-exposed bearer token |
| `AGENT_PORT` | Agent | Service port (defaults to 8090) |
| `HRIS_API_URL` | Agent | Internal authenticated HRIS API |
| `AGENT_DATABASE_URL` | Agent | PostgreSQL DSN for agent action-state table |
| `OPENAI_API_KEY` | Agent | Model-provider secret |
| `OPENAI_MODEL` | Agent | Model identifier (default `gpt-4.1-mini` in source) |
| `OPENAI_BASE_URL` | Agent | OpenAI-compatible API base URL |
| `ADVANCE_HRIS_API_URL` | Desktop | Optional Go API URL override |

`BOOTSTRAP_ORG_*`, `BOOTSTRAP_ADMIN_*`, `SEED_DEMO_DATA` and `SEED_DEMO_PASSWORD` are intended for local bootstrap only. Disable demo seeding and remove sample accounts in all shared deployments.

## Run the core app

```sh
docker compose up --build
```

The core Compose starts a health-gated PostgreSQL → migrations → Go API → web chain. Host port overrides: `HRIS_WEB_PORT`, `HRIS_API_PORT`, and `HRIS_POSTGRES_PORT`.

## Run with the agent

```sh
export OPENAI_API_KEY='your-key-from-secret-store'
docker compose -f docker-compose.yml -f docker-compose.agentic.yml up --build
```

The overlay configures `AGENT_INTERNAL_URL=http://agent:8090` for the server-side Next.js proxy and `HRIS_API_URL=http://api:8080` for the agent.

The model provider must support OpenAI-compatible **Chat Completions tool calls**. Compatibility with a base URL does not guarantee tool-call or privacy behavior. The default model setting is a source-code default, **not** an endorsement or an assurance of current availability.

## On OpenChoreo

Deploy:

- `hris-web` as a public Next.js component;
- `hris-api` as a protected API component;
- `hris-agent` as a private/internal service;
- PostgreSQL as a managed and backed-up data service.

Inject secrets using environment-scoped secret management. Do not commit `.env` files or API keys. Set production URLs and TLS policy per environment and run migrations as a separately controlled job.

## Important boundary

The current local Compose uses one demonstration database login; this is **not** appropriate production least privilege. Provision an agent-specific role that can access only its action-state table. Never share HR data credentials with the model provider.

See [OpenChoreo deployment]({{ '/openchoreo/' | relative_url }}) · [Security guide]({{ '/security-and-privacy/' | relative_url }}) · [Development guide]({{ '/development/' | relative_url }}).
