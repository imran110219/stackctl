# stackctl

`stackctl` turns a fresh Ubuntu VM (22.04/24.04) into a reproducible Docker Compose platform with environment-specific, toggleable modules.

- Environments: `prod`, `devqa`
- Module toggles: Docker Compose profiles
- Generated paths: `/srv/stack`, `/srv/data`, `/srv/backups`
- Secure-by-default networking: only nginx binds public `80/443`; admin tools bind `127.0.0.1`

## Quickstart

```bash
# 1) Install stackctl safely (installs binary + templates only)
./install.sh

# 2) Initialize prod stack layout
stackctl init --env prod --domain example.com --email admin@example.com

# 3) Enable Jaeger module in prod
stackctl enable jaeger --env prod

# 4) Reconcile running services
stackctl apply --env prod

# 5) Check status
stackctl status --env prod
```

## Commands

```bash
stackctl init --env prod|devqa [--domain example.com] [--email admin@example.com]
stackctl enable <module> --env prod|devqa
stackctl disable <module> --env prod|devqa
stackctl status --env prod|devqa
stackctl apply --env prod|devqa
stackctl backup --env prod|devqa
stackctl doctor
```

## What `init` creates

- `/srv/stack/<env>/compose.yml`
- `/srv/stack/<env>/compose.override.yml`
- `/srv/stack/<env>/enabled.yml`
- `/srv/stack/<env>/.env` (from template, no secrets committed)
- `/srv/stack/<env>/nginx/conf.d/*.conf`
- `/srv/stack/<env>/systemd/*`
- `/srv/data/<env>/<service>/...`
- `/srv/backups/<env>/...`

`init` is idempotent and does not delete data volumes.

## Module toggles

See `docs/modules.md` for complete module list and ports.

Example:

```bash
stackctl enable prometheus --env prod
stackctl enable node-exporter --env prod
stackctl enable grafana --env prod
stackctl apply --env prod
```

## TLS strategy

Two supported approaches:

1. Bring-your-own TLS termination in front of nginx (cloud LB/reverse proxy).
2. Optional `certbot` module for certificate workflows (documented in `docs/security.md`).

## Security notes

- Installer does not modify SSH hardening defaults.
- Never commit real secrets; set strong values in `/srv/stack/<env>/.env`.
- Non-root operation is preferred, but writing `/srv/*` may require `sudo`.
- Docker group grants root-equivalent access; use intentionally.
- Dozzle uses Docker socket proxy (`socket-proxy`) pattern.

## Backups

- `stackctl backup --env <env>` dumps Postgres and MariaDB/MySQL as `.sql.gz`.
- Optional restic offsite push when `.env` includes `RESTIC_REPOSITORY` and `RESTIC_PASSWORD`.
- Systemd timer templates are generated in `/srv/stack/<env>/systemd/`.

## Migration (new VM)

1. Restore `/srv/stack`.
2. Restore `/srv/data`.
3. Run `stackctl apply --env <env>`.
4. Validate with `stackctl doctor`.

See `docs/migration.md` for full flow and `.tar.zst` helpers.

## Manual test plan (v1)

1. `stackctl doctor` on fresh host.
2. `stackctl init --env prod ...` and verify generated files.
3. Set secrets in `/srv/stack/prod/.env`.
4. `stackctl apply --env prod`; confirm core containers healthy.
5. Enable module (`jaeger`), apply, and confirm service starts.
6. Disable module, apply, and confirm service is removed.
7. Run `stackctl backup --env prod` and validate backup files in `/srv/backups/prod/`.
8. Reboot VM and verify systemd auto-start behavior.

## Troubleshooting

- `permission denied` on `/srv/...`: run with `sudo` or fix directory ownership.
- Docker unavailable: start daemon (`systemctl start docker`) and re-run `stackctl doctor`.
- Module appears enabled but not running: run `stackctl apply --env <env>` after toggles.
- Missing template path: set `STACKCTL_TEMPLATES=/path/to/templates`.

## License

MIT. See `LICENSE`.
