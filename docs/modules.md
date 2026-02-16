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

## Module categories

Modules are organized into three categories:

- **Observability**: dozzle, node-exporter, prometheus, alertmanager, grafana, loki, jaeger
- **Infrastructure**: socket-proxy, kuma, certbot
- **Utilities**: backup

## Dependencies

Some modules have automatic dependencies:

| Module | Requires |
|---|---|
| `dozzle` | `socket-proxy` |

When enabling a module (via CLI or TUI), its dependencies are automatically resolved.

## Enable/disable workflow

### CLI

```bash
stackctl enable jaeger --env qa
stackctl disable jaeger --env qa
stackctl apply --env qa
```

`apply` re-renders generated files and runs `docker compose up -d --remove-orphans` with enabled profile flags.

### Interactive module manager

```bash
stackctl modules --env qa
```

The module manager provides:
- Browse modules grouped by category with enabled/disabled/running status
- `space` to toggle modules (dependencies auto-resolved)
- `/` to search and filter by name or description
- `d` to toggle a detail pane showing ports, dependencies, reverse dependencies, and running status
- `s` to save changes to `enabled.yml`
- `a` to save and apply in one step
- Unsaved changes warning on quit
