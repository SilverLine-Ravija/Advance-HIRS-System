---
title: Frequently asked questions
description: What Advance HRIS does, how AI actions work and what is not yet supported.
section: docs
permalink: /faq/
---

# Frequently asked questions

## What exactly is Advance HRIS?

A multi-component HR information system with a Next.js portal, a Wails desktop interface for HR/admins, a Go business API and PostgreSQL persistence. It includes leave, attendance, employee administration, an audit trail and an optional Go-based AI assistant.

## Is GitHub Pages the application itself?

No. The GitHub Pages URL hosts a **static product and documentation website**. To operate the HRIS, deploy the actual web, API, agent and database components.

## Does the agent make HR decisions by itself?

No. The implemented agent can read authorized records and prepare certain writes. **A person must confirm every staged mutation**, and the business API still enforces organization roles, manager/report boundaries and approval rules. It does not autonomously approve leave or make hiring decisions.

## Can I use the HRIS without an AI provider key?

Yes. The core Go API, Next.js portal and desktop workflows do not need the optional agent. The AI assistant requires a compatible model provider and its own private service.

## Is payroll complete?

No. Attendance period summaries can feed a future payroll module, but compensation, taxes, payslips and payment processing have not been implemented.

## Does it include recruitment or performance management?

Not yet. Those are roadmap modules and should not be described to customers as working features.

## Can an employee see colleagues' personal records?

The employee role is restricted to self-service. Managers can see their own and permitted direct-report information only, subject to individual endpoint rules. HR/admin use organization-wide administrative scopes.

## Is the desktop fully offline?

Employee record edits are local-first and queued for synchronization, but an online validated session is required before unlocking employee data. Full offline conflict resolution and a native offline AI assistant are not shipped.

## Is this production-ready for live employee data?

It is an MVP foundation, not a compliance certification. Production still requires security review, model-provider agreements, centralized abuse controls, session revocation/SSO, backups, observability and organizational privacy policies.

## Is the repository open-source?

The repository is publicly readable. **A public repository is not automatically open-source licensed.** No license file was present when this documentation was prepared. Maintainers must choose and add an explicit license before downstream use rights are assumed. See [Contributing]({{ '/contributing/' | relative_url }}).

## Where can I request a feature or report a problem?

Use [GitHub Issues](https://github.com/SilverLine-Ravija/Advance-HIRS-System/issues) without including employee records, credentials or other sensitive content.

## What is OpenChoreo doing here?

OpenChoreo is the intended deployment and operations layer for the multi-component workload. It is not the AI reasoning engine and does not grant HR permissions. Those behaviors are implemented by the HRIS and agent services.

Explore [product capabilities]({{ '/capabilities/' | relative_url }}) · [getting started]({{ '/quickstart/' | relative_url }}) · [roadmap]({{ '/roadmap/' | relative_url }}).
