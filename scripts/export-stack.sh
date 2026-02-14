#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-}"
OUT_FILE="${2:-}"

if [[ -z "${ENV_NAME}" || -z "${OUT_FILE}" ]]; then
  echo "usage: $0 <env> <output.tar.zst>" >&2
  exit 1
fi

if ! command -v zstd >/dev/null 2>&1; then
  echo "zstd is required" >&2
  exit 1
fi

tar -C / -cf - \
  "srv/stack/${ENV_NAME}" \
  "srv/data/${ENV_NAME}" \
  "srv/backups/${ENV_NAME}" \
  | zstd -T0 -19 -o "${OUT_FILE}"

echo "exported ${ENV_NAME} to ${OUT_FILE}"
