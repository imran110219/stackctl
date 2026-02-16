# Migration: move stack to another VM

## Standard flow

1. Restore `/srv/stack` (configs, compose files, `.env`, enabled modules).
2. Restore `/srv/data` (volumes).
3. Install Docker + Compose v2 and `stackctl`.
4. Run `stackctl apply --env dev`, `stackctl apply --env qa`, and/or `stackctl apply --env prod`.
5. Validate with `stackctl doctor`.

## Backup/restore

- Online DB dumps: `stackctl backup --env <env>`
- Full filesystem migration: use tar/zstd helpers below.

## Helper scripts

- `scripts/export-stack.sh <env> <output.tar.zst>`
- `scripts/import-stack.sh <input.tar.zst>`

These scripts package or restore `/srv/stack/<env>`, `/srv/data/<env>`, `/srv/backups/<env>`.

## Restore validation checklist

1. `docker compose ps` shows expected services.
2. Nginx routes `app/api/kc` subdomains correctly.
3. Optional modules respond on loopback.
4. Backup timer status is healthy.

You can also use the interactive dashboard to verify restored environments:

```bash
stackctl dash
```

The dashboard overview tab shows all detected environments with status (OK/DEGRADED/NOT DEPLOYED). Select an environment to inspect individual container states, health checks, CPU/memory usage, and port bindings. Use `r` to restart unhealthy services, `l` to view logs, or `x` to open a shell for debugging.
