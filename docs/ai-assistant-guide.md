---
title: AI assistant user guide
description: Use role-aware AI tools, understand proposal confirmation and recover safely from errors.
section: docs
permalink: /ai-assistant-guide/
---

# AI assistant user guide

The AI assistant lives in the **authenticated HRIS web portal**, at `/assistant`. It is not hosted by GitHub Pages. It is powered by the optional private Go agent service and uses model-provider APIs with structured tool calling.

## What the assistant can do

| Read-only task | Example |
| --- | --- |
| Leave balances and policies | "How much annual leave do I have?" |
| Company holidays | "What holidays are configured?" |
| Leave request summaries | "Show the leave requests I am allowed to view." |
| Attendance and corrections | "Check my attendance and outstanding corrections." |
| Work schedules and period totals | "Summarize the permitted employee's recorded hours." |
| HR directory and dashboard | Available only to roles authorized by the HRIS API |

The assistant can also stage **leave submissions**, **attendance corrections**, **manager review decisions** and a limited set of **HR employee updates or creations**. Some tools are unavailable to employee or manager roles.

### A safe example: apply for leave

1. Ask: "Please check my annual leave balance and prepare leave from November 16 to November 20, 2026."
2. The agent retrieves permitted HR data and may ask for missing details.
3. It displays a proposed action with the leave type, dates and other supplied fields.
4. **Review** every detail. The request is still unsubmitted.
5. Click **Confirm action**. The agent revalidates the tool/role and invokes the existing Go business endpoint.
6. Review the resulting leave request in the HRIS. Approval remains a separate manager/HR process.

Never let the assistant invent a leave reason, time record, employee identifier or approval justification.

## How confirmation works

Each proposed write is inserted in the `agent_action_proposals` table, bound to the authenticated user and organization. A proposal expires after **10 minutes** and may be executed at most once. Confirmation atomically changes `pending → running`; the backend then records `completed` or `failed`.

A completed confirmation reports an HRIS response. It does **not** guarantee the final business outcome (for example, a submitted leave request may still be pending manager approval).

| State | Meaning |
| --- | --- |
| `pending` | Prepared only; no HRIS mutation has run |
| `running` | Confirmation was accepted and the backend call is in progress |
| `completed` | HRIS returned a successful response |
| `failed` | Agent or HRIS returned a failure, or final status could not be confirmed |

If the client times out after submitting, the result may be **uncertain**. Do not blindly confirm or submit a replacement. Look up the actual HRIS record, then have an authorized operator reconcile the action.

## Privacy and model-provider limits

The model may receive information returned by **authorized** HRIS tools. That does not mean the information stays solely in your own infrastructure. Check provider data processing, retention, residency, access, and vendor agreements before using real personnel records. This repository does not implement company-handbook RAG, Slack/Teams messages or unattended autonomous HR jobs.

The agent enforces an additional in-process request rate limit of **12 requests per user per minute**; this is not a substitute for gateway-level abuse and cost controls.

## If the assistant is unavailable

- Confirm you are signed into the web portal.
- Check the agent overlay is running and `OPENAI_API_KEY` is configured securely.
- Confirm the private `AGENT_INTERNAL_URL` and backend `HRIS_API_URL` resolve between services.
- Review the agent's health endpoint and the provider's status without exposing keys.
- Use manual HRIS forms while the AI service is unavailable.

Read [configuration]({{ '/configuration/' | relative_url }}), [AI implementation]({{ '/agentic-hris/' | relative_url }}) and [troubleshooting]({{ '/troubleshooting/' | relative_url }}).


## View or dismiss recent actions

Your assistant workspace now shows recent actions with `pending`, `running`, `completed`, `failed`, `dismissed` or `expired` status. This history is tied to your authenticated account and organization.

- Click **Refresh** to check progress, particularly after a page reload.
- To discard a pending proposal, choose **Dismiss**. The server changes the database status atomically, preventing later confirmation.
- To continue an unexpired pending proposal after a reload, review its full preview and use **Review and confirm**.
- After a timeout or `running` status, verify the leave or attendance record before creating another proposal.

History does not contain full submitted HR data or model messages and is not a substitute for organization audit reporting.
