---
title: Employee guide
description: Employee self-service for leave, attendance, corrections and HR assistant.
section: docs
permalink: /employee-guide/
---

# Employee guide

Advance HRIS gives employees access to their own leave and attendance workflows through the web portal. The employee role is intentionally **not** a company-wide directory or HR administration account.

## Sign in

1. Obtain an invitation or account from your HR team. Accept a one-time invitation from its trusted link if your organization uses invitations.
2. Open the organization's deployed web portal, or `http://localhost:3000` for local development.
3. Sign in with your account and the correct organization slug.
4. Use **Leave**, **Attendance**, or **AI Assistant** from the portal navigation.

The public GitHub Pages website is documentation, not a place to log in.

## Request time off

1. Open **Leave** and inspect your available policy balances.
2. Choose the leave type and dates. Select a half-day only for a single date.
3. Include an accurate reason where appropriate and submit.
4. Track the pending request. Depending on organizational configuration, an authorized manager or HR reviewer must approve it.
5. If you cancel approved tracked leave, the system records a reversing ledger entry rather than silently changing history.

Days are calculated against the policy's working days and applicable company holidays. A request can be rejected when its dates overlap existing pending or approved leave. An employee cannot approve their own request through the employee role.

Read the [leave workflow explanation]({{ '/leave-management/' | relative_url }}) for policy details.

## Record attendance

1. Open **Attendance** and use check-in at the start of the workday.
2. Select the actual work mode (office, remote, hybrid or field).
3. Check out when finished.
4. Review timekeeping information and any late, early-leave or overtime metrics.

Your recorded work schedule is effective-dated and snapshotted for historical accuracy. A missing clock event should be corrected through the normal request workflow, **not** by inventing a timestamp.

## Submit an attendance correction

1. In **Attendance**, find the relevant date.
2. Supply the corrected check-in and/or checkout timestamps and a factual reason.
3. Submit a correction request.
4. Wait for manager or HR review; your own submission does not automatically rewrite the approved attendance record.

The API expects correction timestamps in **RFC3339 including an offset** (for example `2026-11-01T09:00:00+05:30`). See [attendance workflows]({{ '/attendance-management/' | relative_url }}).

## Ask the AI assistant

At `/assistant`, try questions such as:

- "How much annual leave is available to me?"
- "Which company holidays are configured?"
- "Show my attendance for today."
- "Prepare annual leave from 2026-11-16 to 2026-11-20."

The assistant reads only information your role is allowed to access. A proposed change is **not submitted** until you review and confirm it. The proposal expires after ten minutes.

If the assistant reports an uncertain outcome after an API timeout, check the actual HRIS record before attempting another action. See [AI assistant user guide]({{ '/ai-assistant-guide/' | relative_url }}).

## Privacy and help

Never share passwords, invitation tokens or other employees' sensitive data in a chat. If your account has the wrong role or employee link, contact an HR administrator. The assistant is not a substitute for your employer's HR policies or a final payroll record.

Related: [Roles and permissions]({{ '/access-and-tenancy/' | relative_url }}) · [Troubleshooting]({{ '/troubleshooting/' | relative_url }}).
