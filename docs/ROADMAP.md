# stackctl Roadmap

**Last Updated**: 2026-03-10 (v0.1.0-rc1 released - see CHANGELOG.md)

This document tracks all features from MVP to advanced capabilities. It serves as the single source of truth for project progress and helps coordinate development.

---

## Project Vision

Transform stackctl from a custom tool into a **community-friendly CLI** that turns blank Ubuntu VMs into production-ready Docker Compose platforms with zero friction.

### Core Principles
- **Single binary**: No runtime dependencies beyond Docker
- **SSH-first**: Works perfectly over headless SSH connections
- **Transparent**: All operations produce auditable files and standard Docker commands
- **Multi-environment**: dev/qa/prod isolation on a single host
- **Module-based**: Optional components via Docker Compose profiles

### Target Users
Self-hosting developers who want a minimal, opinionated platform without the complexity of Kubernetes or cloud-specific tooling.

---

## Status Legend

- ✅ **Done**: Implemented and working
- 🚧 **In Progress**: Actively being developed
- 📋 **Planned**: Prioritized for upcoming work
- 💡 **Future**: Ideas for later consideration
- ❌ **Blocked**: Waiting on dependencies or decisions

## Priority Levels

- 🔴 **Must-Do**: Critical for community adoption or core functionality
- 🟡 **Nice-to-Have**: Enhances experience but not blocking

---

## Phase 1: MVP (Minimum Viable Product)

### Core CLI Commands
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| `stackctl init` | ✅ Done | 🔴 Must-Do | Initialize environment with domain/email config |
| `stackctl apply` | ✅ Done | 🔴 Must-Do | Reconcile enabled modules to running state |
| `stackctl status` | ✅ Done | 🔴 Must-Do | Show environment and container status |
| `stackctl enable/disable` | ✅ Done | 🔴 Must-Do | Toggle modules per environment |
| `stackctl backup` | ✅ Done | 🔴 Must-Do | Dump databases + push to restic |
| `stackctl doctor` | ✅ Done | 🔴 Must-Do | Pre-flight system checks (Docker, ports, permissions) |
| `stackctl version` | ✅ Done | 🟡 Nice-to-Have | Show version and build info |
| `stackctl destroy` | 📋 Planned | 🟡 Nice-to-Have | Tear down environment with confirmation |

### Interactive TUI Commands
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| `stackctl setup` | ✅ Done | 🔴 Must-Do | Step-by-step wizard: env → domain → modules → preflight → apply |
| `stackctl modules` | ✅ Done | 🔴 Must-Do | Browse/enable/disable modules with detail pane |
| `stackctl dash` | ✅ Done | 🔴 Must-Do | Live dashboard with container status, auto-refresh |
| `stackctl config` | ✅ Done | 🔴 Must-Do | Interactive .env editor with secret masking + validation |

### Module System
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Module catalog (code-defined) | ✅ Done | 🔴 Must-Do | ModuleCatalog in `internal/stackctl/modules.go` |
| Dependency resolution | ✅ Done | 🔴 Must-Do | Auto-enable dependencies when module enabled |
| Template-based modules | ✅ Done | 🔴 Must-Do | Modules defined in `templates/modules/` |
| Module categories | ✅ Done | 🟡 Nice-to-Have | Monitoring, Observability, Tools, Backup |
| Per-module compose overlays | ✅ Done | 🔴 Must-Do | Deep merge module compose.yml into base |

### Configuration & State
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| `.env` per environment | ✅ Done | 🔴 Must-Do | `/srv/stack/<env>/.env` with smart defaults |
| `enabled.yml` tracking | ✅ Done | 🔴 Must-Do | Track which modules are enabled per env |
| `.env` comment preservation | ✅ Done | 🟡 Nice-to-Have | WriteDotEnv preserves comments/ordering |
| Multi-env detection | ✅ Done | 🔴 Must-Do | DetectEnvironments scans `/srv/stack/*` |

### Infrastructure
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| nginx reverse proxy | ✅ Done | 🔴 Must-Do | Single public-facing service, handles TLS |
| systemd units | ✅ Done | 🔴 Must-Do | Auto-start on boot, journald logging |
| Docker Compose profiles | ✅ Done | 🔴 Must-Do | Enable/disable services without destroying data |
| Template rendering | ✅ Done | 🔴 Must-Do | Go text/template for all config files |

---

## Phase 2: Core (Community Adoption Ready)

### Distribution & Installation
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Embed templates with `go:embed` | ✅ Done | 🔴 Must-Do | Bundle templates in binary, no external files |
| GoReleaser config | ✅ Done | 🔴 Must-Do | `.goreleaser.yml` for linux/amd64 + arm64 |
| GitHub Actions release workflow | ✅ Done | 🔴 Must-Do | Auto-build on `v*` tags, publish to GitHub Releases |
| Binary installer script | ✅ Done | 🔴 Must-Do | Rewrite `install.sh` to download pre-built binary (no go/git needed) |
| Debian package (.deb) | 💡 Future | 🟡 Nice-to-Have | For organizations with APT repos |

**Subtasks for go:embed**:
- [x] Add `//go:embed templates` to main or dedicated package
- [x] Update `findTemplatesDir()` to use embedded FS with `STACKCTL_TEMPLATES` override for dev
- [x] Update `renderFile()` and module asset sync to work with `fs.FS`
- [x] Test all template rendering (nginx, systemd, compose) with embedded FS
- [ ] Update install script to not copy templates (binary is self-contained) - pending Task I

**Subtasks for GoReleaser**:
- [x] Create `.goreleaser.yml` with builds for linux/amd64 and linux/arm64
- [x] Configure archives, checksums, changelog generation
- [x] Add GitHub Actions workflow `.github/workflows/release.yml`
- [x] Add version variables to main.go for build info injection
- [x] Create docs/RELEASE.md with release instructions
- [ ] Test release process with a pre-release tag

**Subtasks for installer**:
- [x] Rewrite `install.sh` to detect OS/arch
- [x] Download binary from GitHub Releases instead of building
- [x] Handle version selection (latest vs specific tag)
- [x] Remove `go` and `git` requirements
- [x] Implement `stackctl version` command for verification
- [x] Create docs/INSTALL_TESTING.md with testing procedures
- [ ] Test installation on clean Ubuntu VMs (pending first release)
- [ ] Update README.md and USER_GUIDE.md (pending successful tests)

### Package Refactoring
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Split `internal/stackctl` into layers | 📋 Planned | 🔴 Must-Do | Cleaner architecture for contributors |
| `internal/config/` package | 📋 Planned | 🔴 Must-Do | EnvConfig, .env parsing, path helpers, server.yml |
| `internal/engine/` package | 📋 Planned | 🔴 Must-Do | Template rendering, compose deep merge |
| `internal/executor/` package | 📋 Planned | 🔴 Must-Do | RunCmdCapture, RunCmdStream |
| `internal/state/` package | 📋 Planned | 🔴 Must-Do | Module catalog, enabled.yml, dependency graph |
| `internal/cli/` package | 📋 Planned | 🔴 Must-Do | Command handlers (thin wrappers) |

**Subtasks**:
- [ ] Create new package directories
- [ ] Move types and functions to appropriate packages
- [ ] Update imports across codebase
- [ ] Ensure all tests still pass
- [ ] Update memory/MEMORY.md with new architecture

### Base Stack Cleanup
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Remove hardcoded app services | 📋 Planned | 🔴 Must-Do | Base compose should only have nginx + networks |
| Postgres as optional module | 📋 Planned | 🟡 Nice-to-Have | Move from base to `modules/postgres/` |
| MariaDB as optional module | 📋 Planned | 🟡 Nice-to-Have | Move from base to `modules/mariadb/` |
| Document custom service pattern | 📋 Planned | 🔴 Must-Do | How to add app services via compose.override.yml |
| Migration guide for existing users | 📋 Planned | 🔴 Must-Do | Update `docs/migration.md` |

### CLI Enhancements
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| `stackctl modules list` | 📋 Planned | 🔴 Must-Do | Tabular output: name, description, ports, status |
| `stackctl modules info <module>` | 📋 Planned | 🔴 Must-Do | Detailed view: description, ports, deps, compose snippet |
| `stackctl logs <service>` | 📋 Planned | 🔴 Must-Do | Wrapper for `docker compose logs` with --follow support |
| `stackctl restart <service>` | 📋 Planned | 🔴 Must-Do | Wrapper for `docker compose restart` |
| `stackctl exec <service> -- <cmd>` | 📋 Planned | 🔴 Must-Do | Wrapper for `docker compose exec` |

### Declarative Configuration
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| `server.yml` schema | 📋 Planned | 🟡 Nice-to-Have | Declarative config per environment |
| `server.yml` parser | 📋 Planned | 🟡 Nice-to-Have | Read from server.yml, fallback to enabled.yml |
| Write `server.yml` on init | 📋 Planned | 🟡 Nice-to-Have | Single source of truth for env config |
| `stackctl config show` | 📋 Planned | 🟡 Nice-to-Have | Print resolved config (server.yml + .env) |

**server.yml example**:
```yaml
env: prod
domain: example.com
email: ops@example.com
modules:
  - node-exporter
  - prometheus
  - grafana
  - backup
```

---

## Phase 3: Advanced Features

### Module System Enhancements
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Per-module directory structure | 💡 Future | 🟡 Nice-to-Have | `modules/<name>/compose.yml`, `modules/<name>/config/`, `modules/<name>/README.md` |
| Module contribution guide | 💡 Future | 🟡 Nice-to-Have | How to add a module to stackctl |
| Module healthchecks | 💡 Future | 🟡 Nice-to-Have | All modules must declare healthcheck |
| Module versioning | 💡 Future | 🟡 Nice-to-Have | Track module schema version for upgrades |

### Observability & Operations
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Unified logging (Loki + Grafana) | 💡 Future | 🟡 Nice-to-Have | Aggregate all container logs |
| Metrics dashboards | 💡 Future | 🟡 Nice-to-Have | Pre-built Grafana dashboards per module |
| Alerting rules | 💡 Future | 🟡 Nice-to-Have | Prometheus + Alertmanager templates |
| Backup verification | 💡 Future | 🟡 Nice-to-Have | Test restore from restic snapshots |
| Disaster recovery docs | 💡 Future | 🟡 Nice-to-Have | Step-by-step restore procedure |

### Developer Experience
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Help overlay in all TUI screens | ✅ Done | 🟡 Nice-to-Have | Press `?` for keybindings |
| TUI search/filter in module list | 💡 Future | 🟡 Nice-to-Have | Quickly find modules by name |
| Auto-update check | 💡 Future | 🟡 Nice-to-Have | Notify when new version available |
| Shell completions | 💡 Future | 🟡 Nice-to-Have | Bash, zsh, fish |

### Security
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Port binding audit | ✅ Done | 🔴 Must-Do | All admin ports bind 127.0.0.1 only |
| Secret detection in .env | 💡 Future | 🟡 Nice-to-Have | Warn before committing secrets |
| TLS certificate automation | ✅ Done | 🔴 Must-Do | certbot module for Let's Encrypt |
| Docker socket proxy | ✅ Done | 🟡 Nice-to-Have | socket-proxy module for safer API access |

### Multi-Host (Future)
| Feature | Status | Priority | Description |
|---------|--------|----------|-------------|
| Remote apply via SSH | 💡 Future | 🟡 Nice-to-Have | `stackctl apply --host user@server` |
| Multi-server orchestration | 💡 Future | 🟡 Nice-to-Have | Manage multiple hosts from one control node |
| Load balancer support | 💡 Future | 🟡 Nice-to-Have | HA nginx setup across hosts |

---

## Current Sprint (Next Up)

### Priority 1: Binary Distribution (Must-Do)
Enable frictionless installation without requiring `go` or `git` on target servers.

**Tasks**:
1. ✅ ~~Assess current state~~ (completed)
2. ✅ ~~Embed templates with `go:embed`~~ (completed)
3. ✅ ~~Add GoReleaser config + GitHub Actions~~ (completed)
4. ✅ ~~Rewrite install.sh for binary downloads~~ (completed)
4.5. ✅ ~~Release v0.1.0-rc1~~ (completed 2026-03-09)
4.6. ✅ ~~Create VM testing infrastructure~~ (completed 2026-03-10)
5. 🚧 Test installation on clean Ubuntu 22.04 and 24.04 VMs (READY - test scripts created!)
6. 📋 Update documentation (README.md, USER_GUIDE.md) (blocked: needs testing)

### Priority 2: Package Refactoring (Must-Do)
Make codebase easier for contributors to understand and extend.

**Tasks**:
1. 📋 Create new package structure
2. 📋 Move config logic to `internal/config/`
3. 📋 Move engine logic to `internal/engine/`
4. 📋 Move executor logic to `internal/executor/`
5. 📋 Move state logic to `internal/state/`
6. 📋 Thin out CLI handlers in `internal/cli/`
7. 📋 Update tests and memory docs

### Priority 3: CLI Feature Parity (Must-Do)
Add missing CLI commands for completeness.

**Tasks**:
1. 📋 Add `stackctl modules list`
2. 📋 Add `stackctl modules info <module>`
3. 📋 Add `stackctl logs <service>`
4. 📋 Add `stackctl restart <service>`
5. 📋 Add `stackctl exec <service> -- <cmd>`
6. 📋 Add `stackctl version`
7. 📋 Add `stackctl destroy --confirm`

---

## How to Use This Document (AI Instructions)

When asked "what should I work on next?", follow this process:

1. **Check Current Sprint** section for prioritized tasks
2. **Prefer 🔴 Must-Do** over 🟡 Nice-to-Have
3. **Complete phases sequentially**: MVP → Core → Advanced
4. **Update status** as you work:
   - Change 📋 Planned → 🚧 In Progress when starting
   - Change 🚧 In Progress → ✅ Done when complete
   - Add ❌ Blocked if dependencies are missing
5. **Update "Last Updated"** date at the top when making changes
6. **Add subtasks** for complex features to track granular progress

---

## Contributing

See [docs/contributing.md](./contributing.md) (coming soon) for:
- How to add a new module
- Code style guidelines
- Testing requirements
- Pull request process
