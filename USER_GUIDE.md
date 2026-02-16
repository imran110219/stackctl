# stackctl User Guide

This guide is for end users who want to set up their own Ubuntu server with `stackctl`.

## Supported OS

- Ubuntu 22.04 or 24.04

## Requirements

- Docker Engine and Docker Compose v2
- `git` and `go` (required by the current installer script)
- A domain name for each environment (recommended)
- A user with `sudo` access

## Install Option A: Curl Installer

This uses the repo installer script, which builds `stackctl` locally and installs templates.

```bash
curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/install.sh | bash
```

Notes:

- The installer needs `git` and `go`.
- The binary is installed to `~/.local/bin/stackctl` by default.
- Templates are installed to `~/.stackctl/templates` by default.
- If `~/.local/bin` is not in your `PATH`, add it:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Optional environment variables for the installer:

- `STACKCTL_REPO` to override the repo URL.
- `STACKCTL_HOME` to change install root (default `~/.stackctl`).
- `STACKCTL_BIN_DIR` to change the binary install directory.

## Install Option B: Ubuntu Package (.deb)

If your organization provides a `stackctl` `.deb` package, you can install it like this:

```bash
sudo dpkg -i stackctl_*.deb
sudo apt-get -f install
```

If you have an APT repository configured for `stackctl`, use:

```bash
sudo apt-get update
sudo apt-get install stackctl
```

## First-Time Server Setup

1. Install Docker Engine and ensure the daemon is running.
2. Add your user to the `docker` group (optional but common):

```bash
sudo usermod -aG docker "$USER"
```

3. Log out and log back in so group changes take effect.

### Option A: Interactive Setup (recommended)

```bash
stackctl setup
```

The setup wizard guides you through:
- Choosing an environment (`dev`, `qa`, or `prod`) — existing environments are flagged with `[exists]`
- Setting a domain (smart defaults: `dev.example.com` for dev, `qa.example.com` for qa, `example.com` for prod)
- Setting an admin email
- Selecting optional modules (dependencies are auto-resolved)
- Running pre-flight system checks (Docker, disk space, port availability, etc.)
- Initializing, enabling modules, and applying in one step

After completion you can choose "Setup Another Environment" to configure additional environments without restarting the wizard.

### Option B: CLI Setup

4. Initialize each environment:

```bash
stackctl init --env dev --domain dev.example.com --email admin@example.com
stackctl init --env qa --domain qa.example.com --email admin@example.com
stackctl init --env prod --domain example.com --email admin@example.com
```

5. Set secrets in each environment:

```bash
sudo -E $EDITOR /srv/stack/dev/.env
sudo -E $EDITOR /srv/stack/qa/.env
sudo -E $EDITOR /srv/stack/prod/.env
```

6. Apply the configuration:

```bash
stackctl apply --env dev
stackctl apply --env qa
stackctl apply --env prod
```

## Daily Operations

### Module Management

CLI:

```bash
stackctl enable jaeger --env qa
stackctl disable jaeger --env qa
stackctl apply --env qa
```

Interactive module manager:

```bash
stackctl modules --env qa
```

The module manager lets you browse modules grouped by category, toggle them with `space`, search with `/`, view details with `d`, save with `s`, and save + apply with `a`. Dependency resolution is automatic (enabling `dozzle` auto-enables `socket-proxy`).

### Monitoring

CLI:

```bash
stackctl status --env prod
```

Interactive dashboard:

```bash
stackctl dash                  # overview of all environments
stackctl dash --env prod       # jump directly to prod
```

The dashboard provides a live view with 5-second auto-refresh showing:
- **Overview tab**: all environments with container counts and status (OK/DEGRADED/NOT DEPLOYED)
- **Environment tab**: per-container table with service name, state, health, CPU, memory, and ports
- **Detail tab**: full container info with quick actions — `r` to restart, `l` to view logs, `x` to open a shell

### Configuration

Edit `.env` secrets directly:

```bash
sudo -E $EDITOR /srv/stack/prod/.env
```

Interactive config editor:

```bash
stackctl config --env prod
```

The config editor provides:
- Scrollable list of variables grouped by section (Core, Databases, Security, Backup)
- Secret masking — passwords are hidden by default; press `u` to unmask
- Password generation — press `g` to generate a secure 32-character password
- Validation — press `v` to check for missing required fields, placeholder values, and short passwords
- Automatic restart detection — after saving, shows which services need restarting and offers to apply immediately

### Backups

```bash
stackctl backup --env prod
```

## Paths Used by stackctl

- Config and compose: `/srv/stack/<env>`
- Data volumes: `/srv/data/<env>`
- Backups: `/srv/backups/<env>`

## Keyboard Shortcuts

Press `?` in any TUI screen to see a full keyboard shortcut reference.

Common shortcuts across all TUI tools:

| Key | Action |
|---|---|
| `ctrl+c` | Quit immediately |
| `?` | Toggle help overlay |
| `j` / `k` or arrows | Navigate up/down |
| `h` / `l` or arrows | Navigate left/right |
| `enter` | Confirm / select |
| `esc` | Go back / cancel |
| `space` | Toggle selection |
| `tab` | Switch tabs (dashboard) |
| `/` | Search (module manager) |
| `q` | Quit current tool |

## Troubleshooting

- `permission denied` under `/srv`: run the command with `sudo` or fix directory ownership.
- Docker not running: `sudo systemctl start docker`
- Templates not found: set `STACKCTL_TEMPLATES` to the templates directory.
- TUI renders incorrectly: ensure your terminal supports 256 colors and Unicode.

## Uninstall

- Remove the binary and templates:

```bash
rm -f ~/.local/bin/stackctl
rm -rf ~/.stackctl
```

- Remove server data (destructive):

```bash
sudo rm -rf /srv/stack /srv/data /srv/backups
```
