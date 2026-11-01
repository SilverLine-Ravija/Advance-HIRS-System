---
title: Troubleshooting
description: Fix setup, login, agent, desktop sync, attendance and documentation publishing problems.
section: docs
permalink: /troubleshooting/
---

# Troubleshooting

Use this guide to narrow down common failures without modifying production data directly.

## Web login works but subsequent calls return 401

Check `WEB_COOKIE_SECURE`: local HTTP requires `false`; production HTTPS should use secure cookies. Confirm `WEB_ORIGIN` matches the web address, the login used the correct organization slug and the JWT has not expired. Current JWT access tokens expire after eight hours.

## Database or migrations fail

```sh
docker compose ps
docker compose logs migrate
docker compose logs api
```

Check PostgreSQL readiness, `DATABASE_URL` hostname (`postgres` inside Compose), credentials, port conflicts and migration ordering. Use a backed-up controlled procedure to repair an existing database. **Do not run `docker compose down -v` on customer data.**

## The AI assistant shows "AI provider unavailable"

1. Check the optional overlay is running: `docker compose -f docker-compose.yml -f docker-compose.agentic.yml ps`.
2. Verify `OPENAI_API_KEY` is present in the agent environment without printing its value.
3. Confirm the model supports the required Chat Completions tool-call protocol.
4. Inspect `docker compose ... logs agent` for sanitized error output.
5. Verify internal `AGENT_INTERNAL_URL` and `HRIS_API_URL`, and agent database connectivity.

The assistant is optional. Core leave and attendance forms should still operate without it.

## "Action unavailable" or "already processing"

A proposed action may have expired, already been confirmed or entered a failed/uncertain state. Check the underlying HRIS record. Do not reissue a mutation blindly after a timeout.

## No leave balance or an employee-link error

An HR/admin user must usually specify an employee ID for an individual's balances. Employee/manager self-service needs a linked employee profile. Check the membership and role; avoid altering tenant identifiers in requests.

## Leave request rejected

Common causes: overlapping pending/approved leave, a range with no working days, policy restrictions, invalid dates or insufficient balance when approving tracked leave. Read [leave management]({{ '/leave-management/' | relative_url }}).

## Attendance summary shows unexpected hours

Check effective dates, assigned timezone, schedule start/end, breaks, grace period and overtime threshold. Old attendance rows retain snapshotted schedule settings. Review correction history rather than rewriting payroll-ready period totals manually.

## Desktop cannot synchronize

- Confirm the cloud API is reachable and the token remains authorized.
- Verify OS keychain permissions and the locally encrypted SQLite cache.
- Review pending queue counts and the returned conflict list.
- A conflict means the server version advanced; it is not safe to discard one side automatically.
- Cloud pulls intentionally do not replace records with pending local edits.

## GitHub Pages site is missing or deploy fails

- Repository **Settings → Pages → Build and deployment → Source** must be **GitHub Actions**.
- The canonical workflow is `.github/workflows/pages.yml`; it builds the **Jekyll docs folder**, not the authenticated Next.js HR portal.
- Confirm the build job succeeded and the deploy job has a `github-pages` environment URL.
- Project-page assets use the `/Advance-HIRS-System` base path. Test a nested route such as `/docs/`.
- A GitHub Pages deployment is static: `/assistant` and HRIS APIs are **not** served at the Pages URL.

## Getting help

Open a [GitHub issue](https://github.com/SilverLine-Ravija/Advance-HIRS-System/issues) with non-sensitive reproduction steps, the affected module, expected vs actual behavior, and sanitized logs. Never attach credentials, bearer tokens or employee records.

Read [configuration]({{ '/configuration/' | relative_url }}) · [testing]({{ '/testing/' | relative_url }}) · [production readiness]({{ '/production-readiness/' | relative_url }}).


## Agent proposal remains pending after a page reload

Open `/assistant` and use the **Recent actions** panel. The authenticated action history is persisted in PostgreSQL; a pending action is still unsubmitted. You can dismiss it safely or review and confirm it while it remains valid. Expired or dismissed actions cannot be confirmed.

## Production Docker service cannot start

Use `docker compose -f docker-compose.production.yml config --quiet` with the required environment variables supplied. Missing `DATABASE_URL`, `MIGRATION_DATABASE_URL`, `JWT_SECRET` or `WEB_ORIGIN` fail intentionally. The production stack **does not provision a PostgreSQL server or any demo organization**. Verify your database endpoint, network rules, controlled migration role and first-admin provisioning separately.

See [Docker deployment]({{ '/docker-deployment/' | relative_url }}).
