#!/usr/bin/env bash
set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${THIS_DIR}/../.." && pwd)"
SOURCE_DIR="${DOCS_PUBLIC_SOURCE_DIR:-${ROOT_DIR}/docs}"
STAGE_DIR="${DOCS_PUBLIC_STAGE_DIR:-${ROOT_DIR}/.test-results/docs-public-source}"
SOURCE_REVISION="${DOCS_SOURCE_REVISION:-$(git -C "${ROOT_DIR}" rev-parse HEAD)}"
BUILD_TIME="${DOCS_BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

rm -rf "${STAGE_DIR}"
mkdir -p "${STAGE_DIR}"

rsync -a --delete \
  --exclude '.DS_Store' \
  --exclude '.tmp-*' \
  --exclude 'go-test.log' \
  --exclude 'attempts/' \
  --exclude 'strategy-visuals-*/' \
  "${SOURCE_DIR}/" "${STAGE_DIR}/"

python3 - "${STAGE_DIR}/build-metadata.json" "${SOURCE_REVISION}" "${BUILD_TIME}" "${STAGE_DIR}" <<'PY'
import json
import re
import sys
from pathlib import Path

output_path = Path(sys.argv[1])
stage_dir = Path(sys.argv[4])
payload = {
    "project": "sikuli-go",
    "revision": sys.argv[2],
    "builtAt": sys.argv[3],
    "basePath": "/projects/sikuli-go/docs",
}
output_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")

parity_dir = stage_dir / "reference" / "parity"
for markdown_path in parity_dir.glob("*.md"):
    text = markdown_path.read_text(encoding="utf-8")
    if text.startswith("---\n"):
        continue
    heading = re.search(r"^#\s+(.+)$", text, flags=re.MULTILINE)
    title = heading.group(1).strip() if heading else markdown_path.stem.replace("-", " ").title()
    front_matter = (
        "---\n"
        "layout: guide\n"
        f"title: {json.dumps(title)}\n"
        "nav_key: reference\n"
        "kicker: Parity Reference\n"
        "---\n\n"
    )
    markdown_path.write_text(front_matter + text, encoding="utf-8")
PY

if find "${STAGE_DIR}" -type d -name attempts -print -quit | grep -q .; then
  echo "Public docs staging unexpectedly contains benchmark attempts." >&2
  exit 1
fi

echo "Staged public documentation: ${STAGE_DIR}"
