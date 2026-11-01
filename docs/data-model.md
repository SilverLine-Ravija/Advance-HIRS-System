---
title: Data model and migrations
description: How tenants, employee records, leave ledgers, attendance and agent actions are stored.
section: docs
permalink: /data-model/
---

# Data model and migrations

PostgreSQL is the authoritative cloud system of record. The Wails desktop uses encrypted SQLite as a local cache and operation queue.

## Core entity groups

| Group | Tables | Purpose |
| --- | --- | --- |
| Tenant and identity | `organizations`, `users`, `organization_memberships`, `employees`, `access_invitations` | Membership, identity, employee relationships and onboarding |
| Leave | `leave_requests`, `leave_policies`, `leave_policy_assignments`, `leave_balance_ledger`, `company_holidays` | Rules, requests, entitlement history and calendar dates |
| Attendance | `attendance_entries`, `work_schedules`, `work_schedule_assignments`, `attendance_correction_requests` | Schedules, clock records, snapshots and corrections |
| Platform | `audit_events`, `jobs` | Audit and supporting operational data |
| AI assistant | `agent_action_proposals` | User/tenant-bound proposed mutations and execution outcomes |

The agent action table is added in migration `005_agent_actions.sql`; it is not a general model-memory database.

## Tenant key

The employee primary key is composite `(organization_id, id)`. Employee numbers and work emails are unique within an organization. Never assume a human-readable employee ID uniquely identifies an employee across all companies.

## Leave ledger integrity

Entitlement and carry-over entries are created idempotently. Approved tracked leave posts a debit in the approval transaction. Cancellation creates a reversing credit. Ledger adjustments are auditable and should not be replaced with direct updates to computed balances.

## Schedule snapshot integrity

Employee schedules can vary over time. On check-in, the resolved schedule is copied to that time record. Reporting on an old month should not recalculate its scheduled hours based on a newer schedule configuration.

## Agent action state

The agent stores:

- `organization_id` and `actor_user_id` tying the action to the authenticated principal;
- the allowlisted tool name and validated arguments;
- a human-readable preview and expiration timestamp;
- status `pending`, `running`, `completed` or `failed`;
- final response or error, when available.

A PostgreSQL update atomically acquires a pending proposal for confirmation. A timeout after a backend call can produce an uncertain outcome, so manual reconciliation is necessary. The agent must not be granted direct read/write access to HR domain tables in production.

## Schema and migration order

- `db/schema.sql` — repeatable current base schema.
- `db/migrations/002_tenant_scoped_employee_ids.sql` — composite identity for tenants.
- `db/migrations/003_leave_management_v2.sql` — leave policy and ledger features.
- `db/migrations/004_attendance_v2.sql` — schedule-based attendance and corrections.
- `db/migrations/005_agent_actions.sql` — agent proposals.

Use the migration wrapper for the local environment:

```sh
DATABASE_URL='postgres://...' sh db/migrate.sh
```

Production must use controlled, audited migration orchestration with backups and a schema-version registry. Do not reset or drop production volumes to resolve a migration failure.

For table definitions, review [db/schema.sql](https://github.com/SilverLine-Ravija/Advance-HIRS-System/blob/main/db/schema.sql) and [migration files](https://github.com/SilverLine-Ravija/Advance-HIRS-System/tree/main/db/migrations).

Related: [Architecture]({{ '/architecture/' | relative_url }}) · [API reference]({{ '/api-reference/' | relative_url }}).
