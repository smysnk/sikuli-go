#!/usr/bin/env bash
set -euo pipefail

THIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${THIS_DIR}/../.." && pwd)"

DOCS_SOURCE_DIR="${DOCS_LOCAL_SOURCE_DIR:-${ROOT_DIR}/docs}"
BENCH_SOURCE_DIR="${DOCS_LOCAL_BENCH_SOURCE:-${ROOT_DIR}/.test-results/bench}"
STAGE_DIR="${DOCS_LOCAL_STAGE_DIR:-${ROOT_DIR}/.test-results/docs-pages-local}"
SOURCE_DIR="${STAGE_DIR}/source"
SITE_DIR="${STAGE_DIR}/site"
SERVE_ROOT="${STAGE_DIR}/serve-root"
MOUNT_DIR="${SERVE_ROOT}/projects/sikuli-go/docs"

HOST="${DOCS_LOCAL_HOST:-127.0.0.1}"
PORT="${DOCS_LOCAL_PORT:-4000}"
SERVE="${DOCS_LOCAL_SERVE:-1}"
OPEN_BROWSER="${DOCS_LOCAL_OPEN_BROWSER:-1}"
INCLUDE_BENCH="${DOCS_LOCAL_INCLUDE_BENCH:-1}"
BUILD_MODE="${DOCS_LOCAL_BUILD_MODE:-auto}" # auto|docker|native|copy

is_true() {
  local raw="${1:-}"
  local norm
  norm="$(printf '%s' "${raw}" | tr '[:upper:]' '[:lower:]')"
  [[ "${norm}" =~ ^(1|true|yes|on)$ ]]
}

echo "[docs-local] source=${DOCS_SOURCE_DIR}"
echo "[docs-local] bench_source=${BENCH_SOURCE_DIR}"
echo "[docs-local] stage=${STAGE_DIR}"
echo "[docs-local] build_mode=${BUILD_MODE}"
echo "[docs-local] include_bench=${INCLUDE_BENCH}"
echo "[docs-local] serve=${SERVE} host=${HOST} port=${PORT}"

if [[ ! -d "${DOCS_SOURCE_DIR}" ]]; then
  echo "[docs-local] docs source directory not found: ${DOCS_SOURCE_DIR}" >&2
  exit 1
fi

rm -rf "${SOURCE_DIR}" "${SITE_DIR}" "${SERVE_ROOT}"
mkdir -p "${SOURCE_DIR}" "${SITE_DIR}"

rsync -a --delete \
  --exclude '.DS_Store' \
  --exclude '.tmp-*' \
  --exclude 'go-test.log' \
  --exclude 'attempts/' \
  --exclude 'strategy-visuals-*/' \
  "${DOCS_SOURCE_DIR}/" "${SOURCE_DIR}/"

if is_true "${INCLUDE_BENCH}"; then
  if [[ -d "${BENCH_SOURCE_DIR}" ]]; then
    mkdir -p "${SOURCE_DIR}/bench/reports"
    rsync -a --delete \
      --exclude '.DS_Store' \
      --exclude '.tmp*' \
      --exclude 'go-test.log' \
      --exclude 'attempts/' \
      --exclude 'strategy-visuals-*/' \
      "${BENCH_SOURCE_DIR}/" "${SOURCE_DIR}/bench/reports/"
    echo "[docs-local] copied benchmark artifacts into /bench/reports"
  else
    echo "[docs-local] benchmark source not found, skipping copy: ${BENCH_SOURCE_DIR}"
  fi
fi

build_copy() {
  rsync -a --delete --exclude '.DS_Store' "${SOURCE_DIR}/" "${SITE_DIR}/"
  echo "[docs-local] build mode=copy (markdown served through local preview renderer)"
}

build_native() {
  if command -v bundle >/dev/null 2>&1 && [[ -f "${SOURCE_DIR}/Gemfile" || -d "${SOURCE_DIR}/.bundle" ]]; then
    if ! (cd "${SOURCE_DIR}" && bundle exec jekyll build --source "${SOURCE_DIR}" --destination "${SITE_DIR}"); then
      return 1
    fi
  elif command -v jekyll >/dev/null 2>&1; then
    if ! jekyll build --source "${SOURCE_DIR}" --destination "${SITE_DIR}"; then
      return 1
    fi
  else
    return 1
  fi
  echo "[docs-local] build mode=native"
}

build_docker() {
  command -v docker >/dev/null 2>&1 || return 1
  docker info >/dev/null 2>&1 || return 1
  if ! docker run --rm \
    -v "${SOURCE_DIR}:/srv/jekyll" \
    -v "${SITE_DIR}:/site" \
    jekyll/jekyll:pages \
    jekyll build --source /srv/jekyll --destination /site >/dev/null; then
    return 1
  fi
  echo "[docs-local] build mode=docker"
}

case "${BUILD_MODE}" in
  docker)
    build_docker || { echo "[docs-local] docker build failed" >&2; exit 1; }
    ;;
  native)
    build_native || { echo "[docs-local] native jekyll build failed" >&2; exit 1; }
    ;;
  copy)
    build_copy
    ;;
  auto)
    if ! build_docker; then
      if ! build_native; then
        build_copy
      fi
    fi
    ;;
  *)
    echo "[docs-local] invalid DOCS_LOCAL_BUILD_MODE=${BUILD_MODE} (expected auto|docker|native|copy)" >&2
    exit 1
    ;;
esac

mkdir -p "${MOUNT_DIR}"
rsync -a --delete "${SITE_DIR}/" "${MOUNT_DIR}/"

url="http://${HOST}:${PORT}/projects/sikuli-go/docs/"
echo "[docs-local] site_dir=${SITE_DIR}"
echo "[docs-local] serve_root=${SERVE_ROOT}"
echo "[docs-local] url=${url}"

if is_true "${OPEN_BROWSER}"; then
  (
    sleep 1
    if command -v open >/dev/null 2>&1; then
      open "${url}"
    elif command -v xdg-open >/dev/null 2>&1; then
      xdg-open "${url}"
    elif command -v powershell.exe >/dev/null 2>&1; then
      powershell.exe -NoProfile -Command "Start-Process '${url}'"
    fi
  ) >/dev/null 2>&1 &
fi

if is_true "${SERVE}"; then
  exec python3 -m http.server "${PORT}" --bind "${HOST}" --directory "${SERVE_ROOT}"
fi
