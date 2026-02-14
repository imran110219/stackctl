#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="{{.Env}}"
STACK_ROOT="{{.StackRoot}}"
DATA_ROOT="{{.DataRoot}}"
BACKUP_ROOT="{{.BackupRoot}}"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
TARGET_DIR="${BACKUP_ROOT}/${ENV_NAME}"
mkdir -p "${TARGET_DIR}"

COMPOSE_ARGS=(-f "${STACK_ROOT}/${ENV_NAME}/compose.yml" -f "${STACK_ROOT}/${ENV_NAME}/compose.override.yml" --env-file "${STACK_ROOT}/${ENV_NAME}/.env" -p "${ENV_NAME}")

if docker compose "${COMPOSE_ARGS[@]}" ps -q postgres >/dev/null 2>&1; then
  docker compose "${COMPOSE_ARGS[@]}" exec -T postgres sh -c 'PGPASSWORD="$POSTGRES_PASSWORD" pg_dumpall -U "$POSTGRES_USER"' | gzip -c > "${TARGET_DIR}/postgres_${TIMESTAMP}.sql.gz"
fi

if docker compose "${COMPOSE_ARGS[@]}" ps -q mariadb >/dev/null 2>&1; then
  docker compose "${COMPOSE_ARGS[@]}" exec -T mariadb sh -c 'mysqldump --all-databases -uroot -p"$MYSQL_ROOT_PASSWORD"' | gzip -c > "${TARGET_DIR}/mariadb_${TIMESTAMP}.sql.gz"
fi

if [[ -f "${STACK_ROOT}/${ENV_NAME}/.env" ]]; then
  set -a
  source "${STACK_ROOT}/${ENV_NAME}/.env"
  set +a
fi

if [[ -n "${RESTIC_REPOSITORY:-}" && -n "${RESTIC_PASSWORD:-}" ]]; then
  restic backup "${TARGET_DIR}" "${STACK_ROOT}/${ENV_NAME}" "${DATA_ROOT}/${ENV_NAME}"
fi
