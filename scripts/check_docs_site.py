#!/usr/bin/env python3
"""Static checks for the Jekyll product site and handbook (stdlib only)."""
from __future__ import annotations

import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"
LINK_LIQUID = re.compile(r"\{\{\s*['\"](/[^'\"]+)['\"]\s*\|\s*(?:relative_url|absolute_url)\s*\}\}")
LINK_MARKDOWN = re.compile(r"(?<!\!)\[[^]]+\]\(([^\s)]+)(?:\s+[^)]*)?\)")
LINK_HTML = re.compile(r"""(?:href|src)=['\"]([^'\"]+)['\"]""")
EXPECTED = {
    "/docs/", "/quickstart/", "/capabilities/", "/workflows/",
    "/employee-guide/", "/manager-guide/", "/hr-admin-guide/",
    "/leave-management/", "/attendance-management/", "/ai-assistant-guide/",
    "/desktop-guide/", "/access-and-tenancy/", "/architecture/",
    "/api-reference/", "/api-recipes/", "/data-model/", "/agentic-hris/",
    "/configuration/", "/development/", "/openchoreo/", "/testing/",
    "/security-and-privacy/", "/production-readiness/", "/troubleshooting/",
    "/docker-deployment/",
    "/roadmap/", "/faq/", "/glossary/", "/contributing/"
}

def frontmatter(source: str) -> dict[str, str]:
    if not source.startswith("---\n"):
        return {}
    end = source.find("\n---\n", 4)
    if end < 0:
        return {}
    pairs = {}
    for line in source[4:end].splitlines():
        if ":" in line and not line[0].isspace():
            key, value = line.split(":", 1)
            pairs[key.strip()] = value.strip().strip("'\"")
    return pairs

def check() -> int:
    errors: list[str] = []
    documents: dict[str, Path] = {}
    files = sorted(
        path for path in DOCS.rglob("*")
        if path.is_file() and path.suffix in {".md", ".html", ".txt", ".xml"}
        and "_layouts" not in path.parts and path.name != "README.md"
    )
    for path in files:
        content = path.read_text(encoding="utf-8")
        if path.suffix in {".md", ".html"}:
            meta = frontmatter(content)
            if not meta.get("title") or not meta.get("description") and path.name != "404.html":
                errors.append(f"{path}: title and description front matter required")
            permalink = meta.get("permalink")
            if not permalink:
                errors.append(f"{path}: permalink is missing")
            elif permalink in documents:
                errors.append(f"{path}: duplicate permalink {permalink}")
            else:
                documents[permalink] = path
            if path.suffix == ".md" and meta.get("section") != "docs":
                errors.append(f"{path}: markdown guide must declare section: docs")
        for destination in LINK_LIQUID.findall(content):
            if destination.startswith("/assets/"):
                if not (DOCS / destination.lstrip("/")).is_file():
                    errors.append(f"{path}: missing asset {destination}")
                continue
            local = destination.split("#", 1)[0].split("?", 1)[0]
            if local and local not in EXPECTED and local not in {"/", "/404.html", "/sitemap.xml"}:
                errors.append(f"{path}: unrecognized Liquid destination {destination}")

        # Check ordinary relative .md links in the source handbook (not Liquid).
        in_code = False
        for line in content.splitlines():
            if line.lstrip().startswith(("\x60\x60\x60", "~~~")):
                in_code = not in_code
            if in_code:
                continue
            for raw in LINK_MARKDOWN.findall(line):
                if "{{" in raw or raw.startswith(("#", "http:", "https:", "mailto:")):
                    continue
                target = unquote(urlsplit(raw).path)
                if target.endswith(".md") and not (path.parent / target).is_file():
                    errors.append(f"{path}: missing relative Markdown target {raw}")
        if path.name == "sitemap.xml" and "<urlset" not in content:
            errors.append("sitemap.xml is invalid")

    missing = sorted(EXPECTED - set(documents))
    if missing:
        errors.append("Missing routes: " + ", ".join(missing))
    if "/" not in documents:
        errors.append("No homepage permalink found")
    layout = (DOCS / "_layouts" / "default.html").read_text(encoding="utf-8")
    for dest in LINK_LIQUID.findall(layout):
        if dest.startswith("/assets/") and not (DOCS / dest.lstrip("/")).is_file():
            errors.append(f"layout: missing {dest}")
        elif dest not in EXPECTED | {"/"} and not dest.startswith("/assets/"):
            errors.append(f"layout: unknown route {dest}")
    for required in (
        "/employee-guide/", "/manager-guide/", "/hr-admin-guide/",
        "/api-recipes/", "/security-and-privacy/", "/troubleshooting/", "/docker-deployment/"
    ):
        if required not in layout:
            errors.append(f"Sidebar missing {required}")
    if errors:
        print("Documentation checks failed:")
        for issue in errors:
            print(" -", issue)
        return 1
    print(f"OK: {len(documents)} published pages, {len(files)} docs sources, all known route and asset references found.")
    return 0

if __name__ == "__main__":
    sys.exit(check())
