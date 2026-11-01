---
title: Access, roles and tenancy
description: JWT sessions, role matrix, invitations and organization isolation rules.
section: docs
permalink: /access-and-tenancy/
---

# Access, roles and tenancy

The core authorization boundary is the **Go HRIS API**, not the browser, desktop UI or language model.

## Roles and scope

| Capability | Employee | Manager | HR | Admin |
| --- | :---: | :---: | :---: | :---: |
| Own leave and attendance | Yes | Yes | Via linked/selected employee | Via linked/selected employee |
| Review direct-report leave/corrections | No | Yes | Organization-wide | Organization-wide |
| Directory view | No | Self and direct reports | Organization | Organization |
| Employee create/update/archive | No | No | Yes | Yes |
| Configure leave/attendance policies | No | No | Yes | Yes |
| Grant employee/manager access | No | No | Yes | Yes |
| Grant HR/admin access | No | No | No | Yes |
| Review organization audit events | No | No | Yes | Yes |
| AI assistant tools | Scoped self-service | Scoped team subset | Scoped HR tools | Scoped admin/HR tools |

This table summarizes the implemented endpoints. Some HR/admin self-service API reads require an explicit `employee_id`. Exact per-operation authorization remains defined in `apps/api/internal/httpapi/routes_v5.go` and its handler/store checks.

## Authentication path

1. Client posts credentials plus `organization_slug` to `POST /api/v1/auth/login`.
2. Server verifies the password using bcrypt and resolves an organization membership.
3. The API issues a signed **HS256 JWT** with user, organization and role claims.
4. Next.js stores the token in an HttpOnly/SameSite cookie behind its BFF. The browser does not need access to the bearer token.
5. On each protected API call, the Go service validates the JWT and enforces role, organization and manager report scope.

The current access token lifetime is **eight hours**. Refresh token revocation, SSO/OIDC and centralized session management are future hardening work.

## Organization isolation

Employee identity is composite **`(organization_id, employee_id)`**. An employee ID may exist in multiple organizations; SQL queries and foreign keys must scope the organization.

The server derives the organization from the authenticated membership. Client-supplied tenant IDs must never be trusted. Manager queries are constrained to linked self + direct reports where the endpoint permits them. Automated assistant tool calls must go through exactly the same checks.

## Invitations

- HR/admin creates a one-time invitation for an allowed role.
- The invitation token expires after **48 hours**.
- Only the token's SHA-256 digest is stored in PostgreSQL.
- For employee/manager roles, invitation email must match the linked employee work email.
- Reissuing a pending invitation invalidates its previous token.
- Users may join another organization without having an existing password overwritten.

Accept invitations through the trusted `/accept?token=...` workflow, not through a publicly shared screenshot or a chat transcript.

## Common access errors

- **401 Unauthorized:** login or token is missing, invalid or expired.
- **403 Forbidden:** authenticated identity lacks role or resource scope.
- **Employee link required:** manager/employee user has no linked employee profile.
- **Not found:** resource may not exist within your organization.
- **Invite expired or consumed:** request a new one-time invitation through HR.

## Production warning

Do not enable demo bootstrap credentials on shared environments. Use TLS, server-side rate limits, secure cookies, secret-manager injection, role audits, session-revocation work and regularly tested backups before production.

Read [security and privacy]({{ '/security-and-privacy/' | relative_url }}) and [production readiness]({{ '/production-readiness/' | relative_url }}).
