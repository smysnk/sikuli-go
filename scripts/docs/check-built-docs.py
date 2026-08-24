#!/usr/bin/env python3
from __future__ import annotations

import json
import re
import sys
from pathlib import Path


BASE_PATH = "/projects/sikuli-go/docs"
LEGACY_HOST = "smysnk.github.io/sikuli-go"
ATTRIBUTE_RE = re.compile(r'''(?:href|src)=["']([^"']+)["']''')


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: check-built-docs.py <site-dir>", file=sys.stderr)
        return 2

    site_dir = Path(sys.argv[1]).resolve()
    errors: list[str] = []
    index_path = site_dir / "index.html"
    metadata_path = site_dir / "build-metadata.json"

    if not index_path.is_file():
        errors.append("missing index.html")
    if not metadata_path.is_file():
        errors.append("missing build-metadata.json")
    else:
        metadata = json.loads(metadata_path.read_text(encoding="utf-8"))
        if metadata.get("project") != "sikuli-go":
            errors.append("build metadata has the wrong project")
        if metadata.get("basePath") != BASE_PATH:
            errors.append("build metadata has the wrong base path")

    for html_path in site_dir.rglob("*.html"):
        text = html_path.read_text(encoding="utf-8", errors="replace")
        if LEGACY_HOST in text:
            errors.append(f"{html_path.relative_to(site_dir)} contains the legacy docs host")
        if '<link rel="canonical"' not in text:
            errors.append(f"{html_path.relative_to(site_dir)} is missing a canonical link")
        for target in ATTRIBUTE_RE.findall(text):
            if target.startswith("/") and not target.startswith(f"{BASE_PATH}/"):
                errors.append(
                    f"{html_path.relative_to(site_dir)} escapes the docs base path: {target}"
                )

    if errors:
        print("Built documentation validation failed.", file=sys.stderr)
        for error in errors:
            print(error, file=sys.stderr)
        return 1

    print("Built documentation validation passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
