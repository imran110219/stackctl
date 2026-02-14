# Security

## Defaults

- Only nginx binds `0.0.0.0:80/443` by default.
- Admin and observability UIs bind loopback (`127.0.0.1`) unless you deliberately expose them.
- `.env` files are generated locally; secrets are not committed.

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
