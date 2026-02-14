# Modules

Modules are overlays under `templates/modules/<module>/` and each service is attached to a Compose profile of the same module name.

## Base stack (always available)

- `nginx` (public `80/443`)
- `frontend` (internal)
- `backend` (internal)
- `keycloak` (internal)
- `postgres` (internal)
- `mariadb` (internal)

## Optional modules

- `socket-proxy`: Docker socket proxy, `127.0.0.1:2375`
- `dozzle`: container logs UI, `127.0.0.1:9999` (auto-pulls `socket-proxy` dependency)
- `node-exporter`: host metrics, `127.0.0.1:9100`
- `prometheus`: metrics store/scrape, `127.0.0.1:9090`
- `alertmanager`: alert routing, `127.0.0.1:9093`
- `grafana`: dashboards, `127.0.0.1:3000`
- `loki`: log backend, `127.0.0.1:3100`
- `jaeger`: tracing UI + OTLP, `127.0.0.1:16686`, `127.0.0.1:4317`, `127.0.0.1:4318`
- `kuma`: uptime checks, `127.0.0.1:3001`
- `certbot`: optional cert management helper (no public bind)
- `backup`: backup helper sidecar (no public bind)

## Enable/disable workflow

```bash
stackctl enable jaeger --env prod
stackctl disable jaeger --env prod
stackctl apply --env prod
```

`apply` re-renders generated files and runs `docker compose up -d --remove-orphans` with enabled profile flags.
