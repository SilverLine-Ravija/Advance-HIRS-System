---
title: Glossary
description: Definitions for HRIS, tenancy, leave accounting, agent proposals and desktop sync.
section: docs
permalink: /glossary/
---

# Glossary

| Term | Meaning |
| --- | --- |
| HRIS | Human Resources Information System; the application storing and operating personnel workflows |
| Organization / tenant | Company-level isolation boundary for records, memberships and permissions |
| RBAC | Role-based access control, enforced by the Go API |
| BFF | Backend for Frontend; Next.js route handlers that hold the browser session and proxy authenticated calls |
| Employee ID | Tenant-scoped record identifier; unique together with `organization_id` |
| Policy-backed leave | Requests validated against a configured leave type, entitlement and rules |
| Leave ledger | Append-only history of entitlement, carry-over, usage and adjustments |
| Carry-over | Allowed unused leave moved from an earlier policy year, subject to a cap |
| Partial-day leave | Half-day request for one calendar date |
| Effective-dated assignment | Employee policy or schedule assignment active only for selected dates |
| Schedule snapshot | Copy of the applicable work schedule stored with a time record |
| Attendance correction | Employee-supplied change request to an attendance entry, requiring review |
| Payroll-ready summary | Attendance hours and exception totals; does not calculate pay, tax or payslips |
| HRIS API | Go business API enforcing authentication, role checks, tenancy and domain rules |
| Agent tool | Allowlisted function by which the AI assistant reads or prepares an HRIS operation |
| Proposed action | Durable, user/tenant-bound record of a mutation not yet executed |
| Human confirmation | Explicit action by the authenticated user to execute a staged mutation |
| Idempotency / single-use | Protection against running the same confirmed agent proposal twice |
| Outcome uncertain | Network failure after a request might already have reached the HRIS; requires reconciliation |
| Local-first | Desktop writes an encrypted local record and queues cloud synchronization |
| Sync conflict | A remote version changed after the local base version; requires explicit reconciliation |
| OpenChoreo | Intended platform for deploying and operating components, separate from the HRIS business logic |
| GitHub Pages | Static public documentation hosting; not the authenticated HRIS runtime |

Next: [Quick start]({{ '/quickstart/' | relative_url }}) or [API reference]({{ '/api-reference/' | relative_url }}).
