# stackctl

`stackctl` turns a fresh Ubuntu VM (22.04/24.04) into a reproducible Docker Compose platform with environment-specific, toggleable modules.

- Environments: `dev`, `qa`, `prod`
- Module toggles: Docker Compose profiles
- Generated paths: `/srv/stack`, `/srv/data`, `/srv/backups`
- Secure-by-default networking: only nginx binds public `80/443`; admin tools bind `127.0.0.1`

## Goal

Use one CLI to bootstrap and operate three isolated stack environments on the same host:

- `dev`: fast iteration and feature testing
- `qa`: integration and release validation
- `prod`: production runtime

Each environment gets its own config, data, backups, and module toggles under `/srv/*/<env>`.

## Quickstart

### Interactive (recommended)

```bash
# 1) Install stackctl safely (installs binary + templates only)
./install.sh

# 2) Launch the interactive setup wizard
stackctl setup
```

The setup wizard walks through environment selection, domain/email configuration, module selection, pre-flight checks, and applies everything automatically.

### CLI

```bash
# 1) Install stackctl safely (installs binary + templates only)
./install.sh

# 2) Initialize environment layout (repeat for dev, qa, prod)
stackctl init --env dev --domain dev.example.com --email admin@example.com
stackctl init --env qa --domain qa.example.com --email admin@example.com
stackctl init --env prod --domain example.com --email admin@example.com

# 3) Enable modules per environment
stackctl enable jaeger --env qa
stackctl enable backup --env prod

# 4) Reconcile running services
stackctl apply --env qa
stackctl apply --env prod

# 5) Check status
stackctl status --env prod
```

## Environment workflow

1. Initialize each environment once with `stackctl init --env <env>`.
2. Set secrets in `/srv/stack/<env>/.env`.
3. Toggle modules using `stackctl enable/disable <module> --env <env>`.
4. Reconcile state using `stackctl apply --env <env>`.
5. Use `stackctl backup --env <env>` and migration scripts per environment.

## Commands

### CLI commands

```bash
stackctl init --env dev|qa|prod [--domain example.com] [--email admin@example.com]
stackctl enable <module> --env dev|qa|prod
stackctl disable <module> --env dev|qa|prod
stackctl status --env dev|qa|prod
stackctl apply --env dev|qa|prod
stackctl backup --env dev|qa|prod
stackctl doctor
```

### Interactive TUI commands

```bash
stackctl setup                        # interactive setup wizard
stackctl modules [--env dev|qa|prod]  # module manager
stackctl dash [--env dev|qa|prod]     # status dashboard
stackctl config [--env dev|qa|prod]   # configuration editor
```

- **`setup`** — Step-by-step wizard: environment selection, domain/email input, module selection, pre-flight system checks, then init + enable + apply. Supports setting up multiple environments in one session.
- **`modules`** — Browse, enable/disable, search, and apply modules with a detail pane showing ports, dependencies, and running status.
- **`dash`** — Live dashboard with auto-refresh showing all environments, per-environment container tables with CPU/memory, and quick actions (restart, logs, shell).
- **`config`** — Edit `.env` variables with secret masking, password generation, validation checks, and automatic service restart detection.

Press `?` in any TUI screen to view keyboard shortcuts.

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
3. Run `stackctl apply --env dev`, `stackctl apply --env qa`, and/or `stackctl apply --env prod`.
4. Validate with `stackctl doctor`.

See `docs/migration.md` for full flow and `.tar.zst` helpers.

## Manual test plan (v1)

1. `stackctl doctor` on fresh host.
2. `stackctl init --env dev ...`, `stackctl init --env qa ...`, `stackctl init --env prod ...`.
3. Set secrets in `/srv/stack/dev/.env`, `/srv/stack/qa/.env`, `/srv/stack/prod/.env`.
4. `stackctl apply --env <env>` for each target environment; confirm core containers healthy.
5. Enable module (`jaeger`) in `qa`, apply, and confirm service starts.
6. Disable module in `qa`, apply, and confirm service is removed.
7. Run `stackctl backup --env prod` and validate backup files in `/srv/backups/prod/`.
8. Reboot VM and verify systemd auto-start behavior.

## Manual test plan (TUI)

1. `stackctl setup` — complete full wizard flow for `dev` environment; verify pre-flight checks run and environment is created.
2. On the completion screen, choose "Setup Another Environment" and set up `qa`; verify the wizard resets correctly.
3. `stackctl modules --env dev` — toggle a module on/off, verify save and apply work, verify search filters modules.
4. `stackctl dash` — verify overview shows all environments, select one to view containers, verify auto-refresh updates.
5. `stackctl dash --env dev` — verify it opens directly to the selected environment.
6. `stackctl config --env dev` — edit a variable, verify secret masking, generate a password, run validation, save and verify restart prompt.
7. Press `?` from any TUI screen — verify help overlay shows and dismisses correctly.

## Troubleshooting

- `permission denied` on `/srv/...`: run with `sudo` or fix directory ownership.
- Docker unavailable: start daemon (`systemctl start docker`) and re-run `stackctl doctor`.
- Module appears enabled but not running: run `stackctl apply --env <env>` after toggles.
- Missing template path: set `STACKCTL_TEMPLATES=/path/to/templates`.

## License

MIT. See `LICENSE`.
