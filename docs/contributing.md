---
title: Contributing and project structure
description: Repository layout, contribution workflow, issue reports and public-source licensing notes.
section: docs
permalink: /contributing/
---

# Contributing and project structure

## Repository structure

```text
Advance-HIRS-System/
├── apps/
│   ├── api/                 Go HRIS business API
│   ├── agent/               Optional AI tool orchestration
│   ├── web/                 Next.js employee/manager/HR portal
│   └── desktop/             Wails + React native HR console
├── db/
│   ├── schema.sql
│   ├── migrate.sh
│   └── migrations/
├── docs/                    GitHub Pages product site + handbook
├── scripts/                 Repository utility scripts
├── docker-compose.yml
└── docker-compose.agentic.yml
```

## Before creating a change

1. Review the [architecture]({{ '/architecture/' | relative_url }}) and [security]({{ '/security-and-privacy/' | relative_url }}) boundaries.
2. Open an issue or draft PR describing the expected behavior, data model and required permissions.
3. Work on a feature branch and use synthetic HR records.
4. Add tests for business rules, role boundaries and tenant isolation.
5. Run the test commands in [Testing]({{ '/testing/' | relative_url }}).
6. Describe migrations, environment changes, operational impact and any incomplete pieces in the PR.

## Documentation changes

Documentation lives in `docs/`. Markdown pages use Jekyll front matter with `title`, `description`, `section: docs` and a stable `permalink`. Update the sidebar, handbook index, and homepage when you add a new major guide.

The public documentation deployment uses `.github/workflows/pages.yml`. The authenticated Next.js application is **not** statically exported to GitHub Pages.

Run `python3 scripts/check_docs_site.py` and inspect navigation at desktop and mobile sizes. Avoid copying real customer HR documents or private screenshots into public pages.

## Issue report template

Provide:

- expected behavior and actual behavior;
- the affected app/module and browser/OS where relevant;
- step-by-step reproduction with synthetic data;
- logs after removing tokens, names, emails and personal information;
- whether the problem occurs only offline, only with the AI overlay, or in core HRIS.

## License note

This repository was publicly visible but contained **no explicit LICENSE file** when these docs were prepared. Public access alone does not grant downstream rights to use, redistribute or modify the code. Maintainers should select and add an explicit license if open-source distribution is intended. This documentation does not select a license on their behalf.

## Security issues

Do not post credentials, cross-tenant data or detailed exploit steps in a public issue. Seek a private maintainer contact route for responsible reporting.

Source: [GitHub repository](https://github.com/SilverLine-Ravija/Advance-HIRS-System).
