#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLIENT_DIR="$ROOT_DIR/packages/client-node"
NPM_CACHE_DIR="${NPM_CONFIG_CACHE:-$ROOT_DIR/.test-results/npm-cache}"

mkdir -p "$NPM_CACHE_DIR"
export NPM_CONFIG_CACHE="$NPM_CACHE_DIR"

cd "$CLIENT_DIR"
if [[ "${SKIP_INSTALL:-0}" != "1" ]]; then
  npm install --include=dev
fi

missing_tools=()
for tool in tsc grpc_tools_node_protoc grpc_tools_node_protoc_plugin protoc-gen-ts; do
  if [[ ! -x "node_modules/.bin/$tool" ]]; then
    missing_tools+=("$tool")
  fi
done
if [[ ${#missing_tools[@]} -gt 0 ]]; then
  echo "Missing node_modules tooling: ${missing_tools[*]}" >&2
  echo "Run: (cd $CLIENT_DIR && npm install --include=dev) or set SKIP_INSTALL=1 if already installed." >&2
  exit 1
fi

npm run build
npm pack --dry-run

if [[ "${NPM_PUBLISH:-0}" == "1" ]]; then
  if [[ -z "${NPM_TOKEN:-}" ]]; then
    echo "Missing NPM_TOKEN for publish" >&2
    exit 1
  fi
  npm config set //registry.npmjs.org/:_authToken="${NPM_TOKEN}"
  npm publish --access public
else
  echo "Node package scaffold validated (publish skipped; set NPM_PUBLISH=1)"
fi
