#!/usr/bin/env bash
set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${THIS_DIR}/paths.sh"

step() {
  echo "[integration-verify] $1"
}

step "1/6 Run local packaging/CLI verification"
"${THIS_DIR}/local-verify.sh"

step "2/6 Run gRPC RPC-surface integration test (all Sikuli service methods)"
(
  cd "${API_DIR}"
  go test ./internal/grpcv1 -run TestRPCSurfaceIntegrationViaBufconn -count=1
)

step "3/6 Run gRPC image/OCR E2E integration test"
(
  cd "${API_DIR}"
  go test ./internal/grpcv1 -run TestRPCImageAndOCRE2EViaBufconn -count=1
)

step "4/6 Run API package integration flow tests"
(
  cd "${API_DIR}"
  go test ./pkg/sikuli -run TestCrossProtocolIntegrationFlow -count=1
)

step "5/6 Run Node client E2E startup/connect verification"
"${THIS_DIR}/e2e-node-client.sh"

step "6/6 Run Python client E2E startup/connect verification"
"${THIS_DIR}/e2e-python-client.sh"

if [[ "${RUN_REAL_DESKTOP_E2E:-0}" == "1" ]]; then
  step "Optional Run real-desktop E2E verification"
  REAL_DESKTOP_E2E=1 "${THIS_DIR}/real-desktop-e2e.sh"
else
  echo "[integration-verify] optional real-desktop E2E skipped (set RUN_REAL_DESKTOP_E2E=1 to enable)"
fi

step "Done"
echo "[integration-verify] all integration checks passed"
