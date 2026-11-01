---
title: Leave management
description: Leave types, policy assignments, balances, approvals, holidays and ledger semantics.
section: docs
permalink: /leave-management/
---

# Leave management

Leave in Advance HRIS is **policy-backed**, rather than a standalone spreadsheet of requests.

## Supported leave types

The current API supports `annual`, `sick`, `parental`, `unpaid`, `remote` and `other`. An organization can customize policy details. Local demo defaults are illustrative, not legal entitlements for any country.

The initial defaults described by the repository include annual leave (20 days with up to 5 days carry-over), sick leave (10 days), and untracked parental/unpaid/remote/other categories. HR should configure policies to match the employer's actual rules.

## Policy and balance model

Tracked leave uses an **append-only ledger** for yearly entitlement, carry-over, adjustments and used days. A current balance is derived from the ledger rather than written as an arbitrary mutable number.

- Yearly entitlement and capped carry-over are provisioned idempotently.
- Approved tracked leave is debited inside the same database transaction as the decision.
- Cancelling approved tracked leave posts a reversing credit.
- Policies can be assigned to individual employees and have effective dates.
- Manual adjustments must be recorded by HR/admin and leave an audit trail.

## Date calculations and overlap

The request engine excludes weekends and matching company/location holidays from chargeable days. A half-day counts as **0.5** and must be a single calendar date. Overlapping pending or approved requests for the same employee are blocked at the PostgreSQL boundary, including concurrent requests.

Example: an annual leave request running Monday through Friday ordinarily charges five working days if no configured holidays or location-specific rules intervene. The actual count is computed by the backend.

## Request lifecycle

1. **Submit:** employee, manager acting for self, or authorized HR creates a request.
2. **Pending:** the request is visible to the permitted reviewer.
3. **Approve or reject:** manager (direct reports only) or HR/admin decides.
4. **Approved:** tracked balance is debited transactionally.
5. **Cancel:** a cancellation path is available with ownership/scope checks. Approved tracked leave receives a ledger reversal.

The AI assistant may **prepare** a request but does not grant approval. Confirmation submits it through the same API as the form.

## API recipes

```http
GET /api/v1/leave/policies
GET /api/v1/leave/balances?year=2026
GET /api/v1/leave/holidays
GET /api/v1/leave/requests
```

For HR/admin balance reads, specify `employee_id`. Employees can omit the ID for self-service; managers may access self and permitted direct reports.

```http
POST /api/v1/leave/requests
Content-Type: application/json
Authorization: Bearer YOUR_SESSION_TOKEN

{
  "leave_type": "annual",
  "start_date": "2026-11-16",
  "end_date": "2026-11-20",
  "partial_day": "full",
  "reason": "Planned vacation"
}
```

When submitting on behalf of someone else, authorized HR must include `employee_id`.

```http
POST /api/v1/leave/requests/{id}/decision
Content-Type: application/json

{"decision":"approved","note":"Reviewed by manager"}
```

## Common error conditions

| Failure | Meaning / next step |
| --- | --- |
| `leave_overlap` | Date range overlaps pending or approved leave. Review existing requests. |
| `no_working_days` | Selected dates contain no chargeable working days. |
| `leave_already_decided` | Request was already processed; refresh before retrying. |
| Insufficient balance | Approval cannot debit the tracked policy; consult HR. |
| `forbidden` | Scope or role is not authorized. |

Related: [Employee guide]({{ '/employee-guide/' | relative_url }}) · [Manager guide]({{ '/manager-guide/' | relative_url }}) · [API reference]({{ '/api-reference/' | relative_url }}).
