---
title: Docker deployment and operations
description: Local Docker development, agent integration tests, and an isolated production-style Compose deployment.
section: docs
permalink: /docker-deployment/
---

# Docker deployment and operations

Advance HRIS has **separate** Docker stacks for local demo use, full agent testing, and an external-PostgreSQL deployment template. The three are intentionally not interchangeable.

## 1. Local demo

Prerequisites: Docker Engine / Desktop with Compose v2.

```sh
git clone https://github.com/SilverLine-Ravija/Advance-HIRS-System.git
cd Advance-HIRS-System
docker compose up --build
```

The normal stack runs local PostgreSQL, the migration job, the Go HRIS API and Next.js. It deliberately seeds sample accounts and uses publicly documented **demo-only passwords**. It must never be exposed to the internet or used with real employee data.

Local ports: `localhost:3000` (web), `localhost:8080` (API), `localhost:5432` (database).

To add the AI agent in **development**:

```sh
export OPENAI_API_KEY='your-provider-key'
docker compose -f docker-compose.yml -f docker-compose.agentic.yml up --build
```

The Go agent listens internally on port `8090`. The browser reaches it through the server-side Next.js session proxy, not a public direct connection.

## 2. End-to-end Docker integration without an external model

```sh
bash scripts/agent-docker-smoke.sh
```

The test starts the full local stack with `docker-compose.ci-agentic.yml` and a deterministic local `mock-model` container. It verifies user login, HRIS read tools, preparation and dismissal of a leave request, rejection of a dismissed proposal, successful confirmation of another proposal and action history. It uses synthetic employees only.

The script cleans up its containers and database volume. Run it in a disposable development environment rather than against an existing Docker Compose project with persistent data.

## 3. Production-style reference configuration

**This is an operations template, not a certified or completely production-ready deployment.** It needs infrastructure controls, a legal/privacy review and a release process before storing real personnel information.

The production template is a **standalone Compose project**. Do not add `-f docker-compose.yml` to these commands; that file contains demo credentials and demo seeding.

### Required infrastructure

- Managed PostgreSQL with restricted network access, encrypted connections and tested backups.
- A controlled `MIGRATION_DATABASE_URL` for schema migrations.
- A separate `DATABASE_URL` for the application API.
- Optionally, `AGENT_DATABASE_URL` limited to the `agent_action_proposals` table.
- An HTTPS reverse proxy on the host that forwards to `127.0.0.1:3000`.
- Secret management, host firewall, container monitoring and rate limiting.

### Configure without checking secrets into Git

```sh
cp .env.production.example .env.production
# Edit the placeholders locally. This file is gitignored.
docker compose --env-file .env.production -f docker-compose.production.yml config --quiet
```

To start **without AI**:

```sh
docker compose --env-file .env.production -f docker-compose.production.yml up -d --build
```

To start **with AI**:

```sh
docker compose --env-file .env.production \
  -f docker-compose.production.yml -f docker-compose.production.agentic.yml \
  up -d --build
```

The API and agent have **no host-published ports**. The Next.js portal binds only to the host loopback interface. Configure your HTTPS reverse proxy to serve the exact `WEB_ORIGIN` and route to `127.0.0.1:3000`. A remote desktop client needs a separately authenticated, secured API ingress; it cannot use this stack's private API URL directly.

### Initial organization provisioning

The production template sets `SEED_DEMO_DATA=false`. It does not create a usable organization unless you explicitly set `BOOTSTRAP_ORG_SLUG`, `BOOTSTRAP_ORG_NAME`, `BOOTSTRAP_ADMIN_EMAIL` and a strong `BOOTSTRAP_ADMIN_PASSWORD` in a **controlled initial deployment**. Set `BOOTSTRAP_ADMIN_NAME` as needed.

Remove the bootstrap variables from the live runtime once the administrator has been provisioned and verified. Do not use sample credentials, commit a bootstrap password or expose the service before TLS and access controls are ready.

### Important warning about environment secrets

Compose environment interpolation provides a convenient reference deployment, but variables may be visible to operators who can inspect Docker container configuration. Production should use a real secret manager/orchestrator and enforce access controls. Separate the agent DB role from the HRIS business data role.

### Check health and diagnostics

```sh
docker compose --env-file .env.production -f docker-compose.production.yml ps
docker compose --env-file .env.production -f docker-compose.production.yml logs --tail=100 api web
```

For deployments with agent overlay, include both `-f` files in all `ps`, `logs` and `up` commands.

The migration job runs `db/migrate.sh` at startup. Before real production upgrades, replace this convenience pattern with controlled, version-tracked migrations and a rollback/restore procedure.

## 4. Build and security basics

- Go API and agent are built in multi-stage Dockerfiles, running as non-root users.
- Next.js uses a multi-stage standalone build and runs as the unprivileged `node` user.
- Each application build context includes `.dockerignore` to avoid copying local secrets and dependency caches.
- Production template sets read-only filesystems for Go API/agent and mounts a private temporary directory.
- The agent never receives the HRIS application's database credentials; only its separate action-state DSN.
- API and agent are accessible only by other Compose services, not by anonymous host traffic.

## 5. Next steps for real production operations

Implement SSO/OIDC and session revocation; centralized abuse limits; durable job/action reconciliation; model privacy assessments; structured logs, metrics and traces; automatic backup and restore drills; database least-privilege; disaster recovery, and signed/verified container images. Do not represent the sample Compose files as these protections already being present.

Related: [Configuration]({{ '/configuration/' | relative_url }}) · [Security]({{ '/security-and-privacy/' | relative_url }}) · [Testing]({{ '/testing/' | relative_url }}) · [Production readiness]({{ '/production-readiness/' | relative_url }}).
