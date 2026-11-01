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

GitHub Pages only hosts the **public documentation**. The complete guides cover [employees](employee-guide.md), [managers](manager-guide.md), [HR administrators](hr-admin-guide.md), [leave](leave-management.md), [attendance](attendance-management.md), [agent usage](ai-assistant-guide.md), [desktop sync](desktop-guide.md), [access and tenancy](access-and-tenancy.md), [API examples](api-recipes.md), [configuration](configuration.md), [data model](data-model.md), [security](security-and-privacy.md), [testing](testing.md), [troubleshooting](troubleshooting.md), [FAQ](faq.md), and [contributing](contributing.md).  It does not run the HRIS, connect to customer databases, or expose a production AI assistant.
