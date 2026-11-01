#!/usr/bin/env python3
"""Add an AI Assistant link to the existing HRIS dashboard navigation.

Run from the repository root after applying the agentic HRIS patch. Idempotent.
"""
from pathlib import Path

root = Path.cwd()
page = root / "apps/web/app/page.tsx"
if not page.is_file():
    raise SystemExit("Run this script from Advance-HIRS-System repository root.")
content = page.read_text()
new_label = 'window.location.href = "/assistant";'
if new_label in content:
    print("Assistant navigation link already present")
    raise SystemExit(0)
existing = '<button className="portal-nav" onClick={() => { window.location.href = "/docs"; }}>Documentation</button>'
if existing not in content:
    raise SystemExit("Dashboard nav changed. Manually add a link to /assistant in apps/web/app/page.tsx")
replacement = '<button className="portal-nav" onClick={() => { window.location.href = "/assistant"; }}>AI Assistant</button>' + existing
page.write_text(content.replace(existing, replacement, 1))
print("Added AI Assistant to dashboard navigation")
