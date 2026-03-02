#!/usr/bin/env bash
set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${THIS_DIR}/paths.sh"

step() {
  echo "[integration-verify] $1"
}

step "1/3 Run local packaging/CLI verification"
"${THIS_DIR}/local-verify.sh"

step "2/3 Run gRPC RPC-surface integration test (all Sikuli service methods)"
(
  cd "${API_DIR}"
  go test ./internal/grpcv1 -run TestRPCSurfaceIntegrationViaBufconn -count=1
)

step "3/3 Run API package integration flow tests"
(
  cd "${API_DIR}"
  go test ./pkg/sikuli -run TestCrossProtocolIntegrationFlow -count=1
)

step "Done"
echo "[integration-verify] all integration checks passed"
