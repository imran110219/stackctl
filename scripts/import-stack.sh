#!/usr/bin/env bash
set -euo pipefail

ARCHIVE="${1:-}"

if [[ -z "${ARCHIVE}" ]]; then
  echo "usage: $0 <input.tar.zst>" >&2
  exit 1
fi

if ! command -v zstd >/dev/null 2>&1; then
  echo "zstd is required" >&2
  exit 1
fi

zstd -dc "${ARCHIVE}" | tar -C / -xf -

echo "import completed from ${ARCHIVE}"
