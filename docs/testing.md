---
title: Testing and quality assurance
description: How to verify the API, web, desktop, agent and static documentation.
section: docs
permalink: /testing/
---

# Testing and quality assurance

The repository contains automated GitHub Actions checks and integration tests, but successful CI does not replace user acceptance and security validation.

## API tests

```sh
cd apps/api
go test ./...
go build ./cmd/server
```

PostgreSQL-backed integration tests validate tenant scope, RBAC, invitations, leave accounting, attendance accounting and audit behavior. Follow the [development guide]({{ '/development/' | relative_url }}) to start database dependencies and set the environment expected by integration tests.

## Agent tests

```sh
cd apps/agent
go test -race ./...
go vet ./...
```

The agent CI also starts PostgreSQL and applies migrations before running agent action integration tests. Verify proposal consumption once-only under simultaneous confirmations, denied tools for insufficient roles, validation of dates/IDs, provider downtime and uncertain HRIS API outcomes.

## Web application

```sh
cd apps/web
npm install
npm run build
```

Use the signed-in browser session to test each role. Check that the assistant interface does not expose the bearer token to client-side code.

## Desktop

```sh
cd apps/desktop/frontend
npm install
npm run build

cd ../
go test ./internal/...
```

Native Wails packaging is separate from frontend unit/build checks and requires target OS testing.

## Full local smoke test

```sh
docker compose up --build
docker compose ps
curl --fail http://localhost:8080/health
```

Test web login, employee view, leave request, attendance entry, manager decision and sync. When the agent overlay is enabled, verify read-only tools and prepare/confirm functionality on disposable demo data.

## Static documentation

The public site is built by `.github/workflows/pages.yml` using Jekyll with `docs/` as the source. Pull requests run the build without deploying; merges to `main` deploy the finished site.

The documentation validator in `scripts/check_docs_site.py` checks all document slugs, navigation destinations, internal Liquid links and missing assets. Run it with:

```sh
python3 scripts/check_docs_site.py
```

To manually preview if Jekyll is installed:

```sh
bundle exec jekyll serve --source docs --baseurl /Advance-HIRS-System
```

The GitHub workflow itself is the authoritative Jekyll build test; its action container supplies the required environment.

## Minimum role acceptance checklist

- Employee cannot view company-wide directory or administer another person's data.
- Manager sees own/direct-report requests only.
- HR can manage normal staff workflows but cannot issue admin/HR invitations.
- Admin can manage organization-level access and audits.
- Leave approvals debit only once; cancellation reverses according to policy.
- Attendance corrections require a real reason and authorized review.
- Agent proposals show an exact preview, expire and never run without explicit confirmation.
- Desktop cloud pulls preserve pending local edits; version conflicts are visible.

Keep synthetic data in public screenshots and CI logs. Related: [security]({{ '/security-and-privacy/' | relative_url }}) · [Troubleshooting]({{ '/troubleshooting/' | relative_url }}).
