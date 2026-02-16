# Security

## Defaults

- Only nginx binds `0.0.0.0:80/443` by default.
- Admin and observability UIs bind loopback (`127.0.0.1`) unless you deliberately expose them.
- `.env` files are generated locally; secrets are not committed.

## Secret management with `stackctl config`

The interactive configuration editor (`stackctl config --env <env>`) provides:

- **Secret masking**: password and key fields are displayed as `********` by default. Press `u` to unmask the value under the cursor.
- **Password generation**: press `g` on any secret field to generate a cryptographically secure 32-character random password using `crypto/rand`.
- **Validation**: press `v` to check for placeholder values still in use (e.g., `example.com`), missing required fields, and passwords shorter than 8 characters.
- **Restart detection**: after saving changes, the editor identifies which services are affected by the changed variables and offers to restart them immediately.

Recognized secret fields: `POSTGRES_PASSWORD`, `MYSQL_ROOT_PASSWORD`, `RESTIC_PASSWORD`, `KC_DB_PASSWORD`, `KEYCLOAK_ADMIN_PASSWORD`, `SECRET_KEY`, `JWT_SECRET`.

## SSH safety

`install.sh` and `stackctl` do not alter SSH daemon settings, firewall policy, or root login behavior.

## Docker privilege model

- Running Docker requires root or membership in `docker` group.
- `docker` group is root-equivalent. Keep membership minimal and audited.

## Docker socket proxy pattern

For tooling like Dozzle, use `socket-proxy` instead of mounting `/var/run/docker.sock` directly into UI containers.

## TLS options

1. BYO TLS termination: cloud load balancer / edge proxy terminates TLS and forwards HTTP to nginx.
2. Optional `certbot` module: use for certificate issue/renew workflows if the VM handles public TLS directly.

For production, prefer centralized TLS termination with strict access controls.
