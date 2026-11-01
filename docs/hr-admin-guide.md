---
title: HR administrator guide
description: Manage workforce records, policies, schedules, invitations, corrections and audits.
section: docs
permalink: /hr-admin-guide/
---

# HR administrator guide

HR and admin users manage organizational HR workflows through the web portal and the optional native desktop console. The Go API remains authoritative for permissions and data changes.

## Administration areas

| Area | Route / capability | Typical task |
| --- | --- | --- |
| People | Portal directory and protected employee endpoints | Create, update or archive employee profiles |
| Access | `/access` | Invite users and revoke pending invitations |
| Leave | `/leave/admin` | Configure policies, holidays, employee assignments and balance adjustments |
| Attendance | `/attendance/admin` | Configure schedules and employee assignments, review corrections, generate period summaries |
| Audit | `GET /api/v1/audit` | Review scoped HR workflow activity |
| Assistant | `/assistant` | Retrieve authorized insights and prepare approved actions |
| Desktop | Wails HR console | Make local-first employee updates and sync them securely |

## Employee lifecycle

1. Create an employee record with a unique employee number and work email **within the organization**.
2. Keep the employee profile and actual job/department/location details accurate.
3. If the employee needs portal access, issue an invitation linked to their employee profile and work email.
4. Assign appropriate leave policies and work schedules.
5. For departures, archive the employee via the supported API rather than directly deleting PostgreSQL records.

Employee IDs are scoped to the organization. Do not construct cross-organization employee references.

## Invitations and role boundaries

Invitations are one-time and expire after **48 hours**. Only a SHA-256 token digest is stored on the server. Share the raw invitation link through a trusted channel.

- **HR** can invite `employee` and `manager` users.
- **Admin** can also grant `hr` and `admin` access.
- Employee and manager invitation email must match the linked employee work email.
- Reissuing an invitation invalidates earlier pending invitations for that email.

These rules are enforced on the backend; the assistant cannot bypass them. See [access and tenancy]({{ '/access-and-tenancy/' | relative_url }}).

## Leave policy administration

Create or update tracked/untracked policies, entitlement amounts, capped carry-over, company holidays and employee-specific policy assignments. Manual balance adjustments are auditable and require HR/admin permissions.

Do not use unrestricted SQL to repair a leave balance. Ledger entries provide the balance history. See [leave management]({{ '/leave-management/' | relative_url }}).

## Scheduling and timekeeping

Configure work schedules with timezone, active workdays, break minutes, grace period and overtime thresholds. Employee schedule assignments are effective-dated.

When an employee checks in, the effective schedule is snapshotted on that row. Later policy changes must not retroactively change prior payroll-period calculations.

Approve legitimate attendance correction requests with sufficient evidence. Use `GET /api/v1/attendance/timesheet` to obtain payroll-ready time summaries; **this is not a complete payroll module**.

See [attendance management]({{ '/attendance-management/' | relative_url }}).

## Desktop synchronization

The desktop console is limited to HR/admin. It stores encrypted employee payloads locally, queues offline changes and pushes through authenticated tenant-scoped API endpoints when connected. A conflict is reported rather than silently overwriting the remote record. A full conflict-resolution interface has not shipped.

See [desktop and offline sync]({{ '/desktop-guide/' | relative_url }}).

## Production responsibilities

Before operating with real employee data, provision dedicated secrets, TLS, backups, session controls and monitoring. The current repo is an engineering foundation, **not** a production-readiness certification. See [security]({{ '/security-and-privacy/' | relative_url }}) and [production readiness]({{ '/production-readiness/' | relative_url }}).
