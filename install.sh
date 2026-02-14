#!/usr/bin/env bash
set -euo pipefail

REPO_URL="${STACKCTL_REPO:-https://github.com/example/stackctl.git}"
STACKCTL_HOME="${STACKCTL_HOME:-$HOME/.stackctl}"
INSTALL_BIN_DIR="${STACKCTL_BIN_DIR:-$HOME/.local/bin}"
CLONE_DIR="${STACKCTL_HOME}/repo"
BIN_PATH="${INSTALL_BIN_DIR}/stackctl"

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require git
require go

mkdir -p "${STACKCTL_HOME}" "${INSTALL_BIN_DIR}"

if [[ -d "${CLONE_DIR}/.git" ]]; then
  echo "updating existing clone in ${CLONE_DIR}"
  git -C "${CLONE_DIR}" fetch --tags --prune
  git -C "${CLONE_DIR}" pull --ff-only
else
  echo "cloning ${REPO_URL} into ${CLONE_DIR}"
  git clone --depth 1 "${REPO_URL}" "${CLONE_DIR}"
fi

echo "building stackctl binary"
go build -o "${BIN_PATH}" "${CLONE_DIR}/cmd/stackctl"
chmod +x "${BIN_PATH}"

mkdir -p "${STACKCTL_HOME}/templates"
cp -a "${CLONE_DIR}/templates/." "${STACKCTL_HOME}/templates/"

echo "installed: ${BIN_PATH}"
if [[ ":${PATH}:" != *":${INSTALL_BIN_DIR}:"* ]]; then
  echo "add to PATH: export PATH=\"${INSTALL_BIN_DIR}:\$PATH\""
fi

echo "template dir: ${STACKCTL_HOME}/templates"
echo "tip: export STACKCTL_TEMPLATES=${STACKCTL_HOME}/templates"
