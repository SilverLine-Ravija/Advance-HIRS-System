---
title: Manager guide
description: Review team leave and attendance requests with direct-report scoping.
section: docs
permalink: /manager-guide/
---

# Manager guide

Managers can use self-service HR features plus **direct-report** review workflows. The backend verifies report relationships; simply knowing an employee ID does not grant access.

## Day-to-day workflow

1. Sign in with a manager account linked to your employee profile.
2. Use the portal to review **Leave** requests for yourself and your direct reports.
3. Review **Attendance** corrections for direct reports.
4. Check dates, work schedules, reasons and policy requirements before selecting approve or reject.
5. Confirm the result in the UI and verify the request's updated status.

The manager role does not grant HR-wide employee creation, access invitations, policy administration or organization-wide auditing.

## Leave decision process

A manager can decide only requests for **direct reports**. The API rejects an attempt to decide a request outside that scope.

| Check | Why it matters |
| --- | --- |
| Employee and dates | Confirm the request belongs to the intended direct report. |
| Leave type and days | Verify the correct policy and chargeable day count. |
| Overlap or staffing concern | Review potential coverage needs using authorized information. |
| Available balance | A tracked policy can reject approval if entitlement is insufficient. |
| Decision and note | Document the reason accurately; the system retains the action in its workflow audit. |

Approving tracked leave posts a balance debit transactionally. Canceling an approved request can create a reversing credit, rather than editing the historical debit.

See [leave management]({{ '/leave-management/' | relative_url }}).

## Attendance corrections

Employees may submit missing or incorrect clock events. When reviewing:

1. Confirm the employee and work date.
2. Verify the submitted timestamps and work mode against available evidence.
3. Approve or reject with a meaningful note.
4. Check the corrected attendance view after approval.

The approved change recalculates schedule metrics in one backend transaction. The AI assistant must **not invent attendance entries**, approve requests without you, or reinterpret company rules.

## AI-assisted team review

In the assistant, ask for pending leave requests you are permitted to review, or attendance corrections that need a decision. It can help summarize information and stage a decision, but your explicit confirmation is required and the Go API makes the final authorization check.

## Frequent access errors

- **Forbidden:** the request belongs to someone who is not a direct report.
- **Employee link required:** your user account lacks a linked employee profile.
- **Already decided:** the request changed status before you submitted your decision.
- **Insufficient balance:** the transaction was rejected under the employee's leave policy.

For escalation or policy changes, contact an HR or admin user. Related: [permissions]({{ '/access-and-tenancy/' | relative_url }}) and [API reference]({{ '/api-reference/' | relative_url }}).
