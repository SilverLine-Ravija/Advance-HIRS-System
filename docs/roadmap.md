---
title: Product status and roadmap
description: Current HRIS feature boundaries, proposed development layers and production launch criteria.
section: docs
permalink: /roadmap/
---

# Product status and roadmap

This is a source-backed status overview, **not** a statement of product availability, release timing or customer validation.

## Implemented in the repository

| Area | What exists |
| --- | --- |
| Web portal | Role-aware Next.js HR workflows and server-side session proxy |
| Employees | CRUD/archive, directory scopes and one-time invitations |
| Leave | Policies, balances, carry-over, holidays, manager review and ledger |
| Attendance | Timekeeping, schedules, corrections, period summaries |
| Desktop | Wails/React interface, encrypted local employee records and sync queue |
| Platform | Tenant-aware Go API, PostgreSQL and audit records |
| Agent | Authenticated tool calls, role allowlisting and single-use confirmed writes |
| Documentation | GitHub Pages product site, operator guides and API examples |

Implemented in code does **not** mean every edge case is audited or suitable for production workloads.

## Near-term engineering priorities

1. **Reliability:** test all role/tenant boundaries, action-result reconciliation and real user workflows.
2. **Security:** implement session revocation/refresh flow, SSO/OIDC, centralized rate limits, secret isolation and model-provider controls.
3. **Operations:** structured logs/traces/metrics, backup restore drills, migration versioning and incident response procedures.
4. **User experience:** accessible policy forms, desktop conflict resolution UI, role-specific navigation and better error recovery.
5. **Customer validation:** select one measurable HR automation use case and demonstrate a real labor/time saving with pilot customers.

## Future business modules

- Full statutory payroll, compensation and tax integrations.
- Recruitment pipelines and candidate coordination.
- Goals, reviews and performance management.
- Employee document repository with retrieval and citations.
- Enterprise notifications and external integration adapters.
- Unattended recurring agent jobs governed by per-tenant policies.
- Native offline desktop assistant and richer sync reconciliation.

No release dates are committed by this repository.

## Suggested production exit criteria

- Reproducible green CI and a complete smoke test on deployment environment.
- Approved security threat model and independent review for personnel data.
- Verified backups, restores, secrets, session control and incident runbooks.
- Company-specific leave, timekeeping and privacy policy review.
- Provider agreements and explicit permission for any personal data passed to an LLM.
- Pilot user acceptance for each supported role and measurable operational outcomes.

See [testing]({{ '/testing/' | relative_url }}) and [security]({{ '/security-and-privacy/' | relative_url }}).
