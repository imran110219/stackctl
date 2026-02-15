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

Enable or disable a module:

```bash
stackctl enable jaeger --env qa
stackctl disable jaeger --env qa
stackctl apply --env qa
```

Check status:

```bash
stackctl status --env prod
```

Run a backup:

```bash
stackctl backup --env prod
```

## Paths Used by stackctl

- Config and compose: `/srv/stack/<env>`
- Data volumes: `/srv/data/<env>`
- Backups: `/srv/backups/<env>`

## Troubleshooting

- `permission denied` under `/srv`: run the command with `sudo` or fix directory ownership.
- Docker not running: `sudo systemctl start docker`
- Templates not found: set `STACKCTL_TEMPLATES` to the templates directory.

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
