#!/usr/bin/env bash
set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${THIS_DIR}/../.." && pwd)"

cd "${ROOT_DIR}"

echo "[docs-open] generating API docs"
./scripts/generate-api-docs.sh

echo "[docs-open] generating parity docs"
./scripts/generate-parity-docs.sh

echo "[docs-open] generating benchmark docs"
python3 ./scripts/generate-benchmark-docs.py \
  --project-root "${ROOT_DIR}" \
  --report-dir "${ROOT_DIR}/.test-results/bench" \
  --published-report-dir "${ROOT_DIR}/docs/bench/reports" \
  --overview-output "${ROOT_DIR}/docs/bench/index.md"

echo "[docs-open] starting local docs preview"
./scripts/docs/publish-pages-local.sh
