---
title: Security and privacy
description: Threat boundaries, model data handling and operational safeguards before production.
section: docs
permalink: /security-and-privacy/
---

# Security and privacy

Advance HRIS processes personnel records, attendance and leave data. Its current implementation includes meaningful controls, but **is not certified or automatically production-ready**.

## Security boundaries

| Layer | Implemented boundary | Remaining production work |
| --- | --- | --- |
| Browser | Server-side Next.js BFF and HttpOnly/SameSite cookie | Session revocation, CSRF review and abuse protection |
| Go HRIS API | JWT validation, role checks and tenant-scoped queries | Enterprise SSO/OIDC, centralized controls and security review |
| PostgreSQL | Organization-scoped records and transactional leave/attendance workflows | Managed backups, least-privilege database users and tested restores |
| Desktop | AES-256-GCM encrypted SQLite payloads, OS keychain token storage | Device posture, lost-device response and mature conflict resolution UX |
| AI service | Authenticated tools, allowed operations, user-bound proposed actions | Model privacy review, prompt-injection evaluation, ingress throttling and ongoing monitoring |
| GitHub Pages | Public static documentation | Never host employee records, API credentials or live backend secrets |

## Protecting employee data

- Collect only data necessary for the requested HR operation.
- Define lawful processing grounds, a privacy notice, retention and access request procedures for the jurisdictions where the system is deployed.
- Set TLS, secure cookies and secret manager injection for all nonlocal environments.
- Disable demo seeding and remove sample passwords on shared hosts.
- Monitor failed logins and excessive API/assistant access.
- Keep employee data out of public issues, example screenshots, build artifacts and analytics.
- Verify email notification, payroll, hiring and employee document functionality **before** marketing it as available.

## Agent and language-model risk

The AI assistant consumes user messages and authorized HRIS tool results, and may send those to the selected model provider. Deploying the agent inside your infrastructure **does not mean model inputs never leave it**. Review the provider's contractual security posture, processing locations, retention policies, subprocessors and regional legal requirements.

The model itself is not a trusted authorization engine. The Go agent checks tool names, argument types and caller role, while the HRIS Go API independently validates each business operation.

Treat both user-entered text and retrieved content as **untrusted instructions**. Evaluate prompt injection such as a leave request attempting to override system policy, exfiltrate another tenant's employee directory, or bypass confirmation.

## Human confirmation and uncertain outcomes

Writes are prepared as actor/tenant-bound proposals with a ten-minute expiration. Confirmation changes the proposal state atomically; the result is stored for audit.

A timeout can occur **after** the HRIS accepted the mutation. In that case, the correct response is to inspect the actual HRIS state, not blindly retry the same operation or assume failure.

High-impact HR decisions (payroll, hiring, disciplinary or termination actions) must not be delegated to an unreviewed model.

## Tenant isolation test cases

Before customer rollout, verify at least:

1. Employee A cannot read Employee B in a different organization.
2. Employee IDs repeated between two organizations cannot leak through joins or API filters.
3. Manager A cannot decide Manager B's direct-report request.
4. HR cannot grant `hr` or `admin` privileges through the invitation API.
5. A staged AI action is inaccessible to a different actor or tenant.
6. A stale/expired action cannot be executed and concurrent confirmations run once.
7. Agent tools fail closed after permission changes or revoked sessions.

The current suite has integration tests for major HR boundaries, but additional enterprise security testing is needed.

## Responsible disclosure

If you believe you've found a vulnerability, avoid posting real data or exploit details publicly. Contact the repository maintainers through an appropriate private security channel if configured, or open a non-sensitive issue asking for a secure reporting channel. The repository currently has no documented formal security response policy.

Related: [Production readiness]({{ '/production-readiness/' | relative_url }}) · [Roles and tenancy]({{ '/access-and-tenancy/' | relative_url }}) · [Agent implementation]({{ '/agentic-hris/' | relative_url }}).
