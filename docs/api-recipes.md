---
title: API examples and recipes
description: Authenticated curl examples for HRIS endpoints and agent-assisted workflows.
section: docs
permalink: /api-recipes/
---

# API examples and recipes

This page supplements the complete [endpoint catalog]({{ '/api-reference/' | relative_url }}) with practical request shapes. Examples are for **local development with synthetic users**. Never commit tokens or customer information.

## Authentication

The HRIS API listens at `http://localhost:8080` in the standard local stack. Log in with the organization slug and your **own** development credentials.

```sh
API='http://localhost:8080'
curl -sS -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"your-user@example.invalid","password":"YOUR_PASSWORD","organization_slug":"northstar"}'
```

The response includes an `access_token`, `expires_at` and `user`. Set `TOKEN` from the response in your shell without printing or committing it. Then inspect the active principal:

```sh
curl -sS "$API/api/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN"
```

The browser web app **does not** store the bearer token in client-side JavaScript. Its Next.js BFF uses a protected session cookie and proxies requests under `/api/hr/*`.

## Leave: balances and policy

```sh
curl -sS "$API/api/v1/leave/policies" -H "Authorization: Bearer $TOKEN"
curl -sS "$API/api/v1/leave/holidays" -H "Authorization: Bearer $TOKEN"
curl -sS "$API/api/v1/leave/balances?year=2026" -H "Authorization: Bearer $TOKEN"
```

For HR/admin, the balance endpoint requires `employee_id` in the query. Managers can request permitted direct reports; employees are limited to their own records.

## Leave: request and decision

```sh
curl -sS -X POST "$API/api/v1/leave/requests" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"leave_type":"annual","start_date":"2026-11-16","end_date":"2026-11-20","partial_day":"full","reason":"Planned vacation"}'
```

The request becomes pending; submission does not mean approval. HR/admin submitting on behalf of an employee includes an authorized `employee_id`.

An authorized reviewer can decide the request:

```sh
curl -sS -X POST "$API/api/v1/leave/requests/REQUEST_ID/decision" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"decision":"approved","note":"Reviewed against team coverage"}'
```

Managers can decide only direct-report requests. The server calculates policy-backed debits within the approval transaction.

## Attendance: check-in and correction

```sh
curl -sS -X POST "$API/api/v1/attendance/check-in" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"work_mode":"remote"}'
```

```sh
curl -sS -X POST "$API/api/v1/attendance/corrections" \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"work_date":"2026-11-01","requested_check_out":"2026-11-01T17:30:00+05:30","work_mode":"office","reason":"Missing checkout from approved time sheet"}'
```

The correction remains pending until the manager/HR review endpoint processes it.

## Attendance: period reporting

```sh
curl -sS "$API/api/v1/attendance/timesheet?from=2026-11-01&to=2026-11-30" \
  -H "Authorization: Bearer $TOKEN"
```

HR/admin must supply `employee_id`. The response includes `scheduled_days`, `recorded_days`, `worked_minutes`, `overtime_minutes`, `late_minutes`, `early_leave_minutes`, `worked_hours` and `overtime_hours`.

**This is a timekeeping summary, not a salary calculation.**

## AI assistant: chat and proposal confirmation

The assistant is accessed through the **signed-in Next.js website**:

```http
POST /api/agent/chat
Content-Type: application/json

{
  "message": "How much annual leave do I have left?",
  "history": []
}
```

The response includes `reply` and possibly `pending_actions` containing `id`, `tool`, `preview` and `expires_at`.

For a proposed mutation, the same authenticated session must explicitly confirm:

```http
POST /api/agent/confirm
Content-Type: application/json

{"action_id":"THE_UUID_FROM_THE_PROPOSAL"}
```

The private agent service should **not** be exposed directly as a public API. The Next.js BFF attaches the existing session credentials server-side. Confirmation is single-use, actor-bound and expires after ten minutes.

## Error response

Business endpoints generally return:

```json
{
  "error": {
    "code": "validation_error",
    "message": "Explanation of the invalid input"
  }
}
```

Use HTTP status for broad handling and `error.code` for workflow-specific recovery. A 403 can reflect a direct-report or tenant restriction, not just a general missing role. Agent confirmation errors can describe an uncertain execution; always verify the HRIS record before retrying.

Related: [API reference]({{ '/api-reference/' | relative_url }}) · [Access and tenancy]({{ '/access-and-tenancy/' | relative_url }}) · [AI assistant guide]({{ '/ai-assistant-guide/' | relative_url }}).
