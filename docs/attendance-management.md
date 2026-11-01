---
title: Attendance management
description: Schedule snapshots, check-in, check-out, corrections and timesheet summaries.
section: docs
permalink: /attendance-management/
---

# Attendance management

Attendance is tied to an employee's effective-dated work schedule. The application persists timing metrics for reproducible reporting.

## Work schedules

An authorized HR user can define:

| Setting | Purpose |
| --- | --- |
| Timezone | Interpret scheduled start/end for the workplace |
| Working days | Define which days are scheduled |
| Start and end | Expected shift timestamps |
| Break minutes | Scheduled break deduction |
| Grace minutes | Threshold before lateness begins |
| Overtime threshold | Configure the minimum threshold for counted overtime |
| Effective-dated assignment | Determine which schedule applies to an employee on a given date |

At check-in, the resolved schedule is **snapshotted** on the attendance record. Editing the schedule later does not retroactively change that record's historical interpretation.

## Employee workflow

1. Check in using the appropriate work mode: `office`, `remote`, `hybrid` or `field`.
2. At the end of work, check out.
3. Review worked minutes, late minutes, early-leave minutes and overtime minutes.
4. If a time entry is missing or incorrect, submit a correction with a reason and real timestamps.

```http
POST /api/v1/attendance/check-in
Content-Type: application/json

{"work_mode":"office"}
```

For employee/manager self-service, the backend enforces the linked self identity. HR/admin acting for a particular employee supplies `employee_id`.

## Corrections and review

A correction request must include a work date, a reason, and at least one corrected timestamp. Timestamps use RFC3339, with timezone information. Requests are reviewed by an authorized manager for direct reports or HR/admin for the organization.

```http
POST /api/v1/attendance/corrections
Content-Type: application/json

{
  "work_date":"2026-11-01",
  "requested_check_in":"2026-11-01T09:00:00+05:30",
  "requested_check_out":"2026-11-01T17:30:00+05:30",
  "work_mode":"office",
  "reason":"The attendance device did not register my checkout."
}
```

The reviewer approves or rejects via `POST /api/v1/attendance/corrections/{id}/decision`. Approval rewrites the relevant attendance record and recalculates metrics transactionally. The assistant must not fabricate an attendance correction or decide it without a human confirmation.

## Period summaries

`GET /api/v1/attendance/timesheet?from=2026-11-01&to=2026-11-30` returns a single scoped employee's scheduled days, recorded days, worked minutes/hours, overtime, lateness and early-leave totals.

- HR/admin must include `employee_id`.
- Managers may read their own or permitted direct-report summaries.
- Date range must be valid and no longer than **366 days**.
- This is an **attendance-to-payroll boundary**, not payslip, tax or payment processing.

## Interpreting anomalies

Missing checkouts, unusual overtime and late arrivals can be investigated using the recorded schedule and attendance entries. They are **signals for review**, not proof of misconduct, noncompliance or incorrect compensation.

Related: [HR administration]({{ '/hr-admin-guide/' | relative_url }}) · [API recipes]({{ '/api-recipes/' | relative_url }}).
