# Advance HRIS documentation

The project documentation is published as a GitHub Pages handbook from the `docs/` directory. The same core technical material is also accessible in the web application's `/docs` route.

## Start here

- [Product capabilities](capabilities.html) — shipped HR and agent functionality versus future modules.
- [Quick start](quickstart.html) — run the local stack and optional AI service.
- [Architecture](architecture.md) — components and tenant boundaries.
- [Agentic implementation](agentic-hris.md) — AI tools, single-use action confirmation and limitations.
- [API reference](api-reference.md) — HRIS and assistant endpoints.
- [Development guide](development.md) — local setup and testing.
- [OpenChoreo deployment](openchoreo.md) — web, API and private agent components.
- [Production readiness](production-readiness.md) — security and operational prerequisites.

## Product status

Implemented: employee management and secure invitations; leave policies, balances and approvals; schedule-aware attendance and corrections; local-first desktop synchronization; and a permission-aware AI assistant with explicit confirmation for staged writes.

Not shipped: payroll and compensation processing, recruiting, performance management, document ingestion/RAG, notification providers, unattended agent jobs, and enterprise-ready SSO/session revocation.

GitHub Pages only hosts the **public documentation**. It does not run the HRIS, connect to customer databases, or expose a production AI assistant.
