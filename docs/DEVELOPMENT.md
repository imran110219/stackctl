# Development Guide

Quick guide for stackctl development.

## Prerequisites

### Required
- **Go 1.24.2+**: [Download](https://go.dev/dl/)
- **Git**: For version control
- **Make**: For build automation (`sudo apt install build-essential` on Ubuntu)

### Optional
- **GoReleaser**: Only needed for local release testing
  - macOS: `brew install goreleaser`
  - Linux: See [goreleaser.com/install](https://goreleaser.com/install/)
- **golangci-lint**: For running linters
  - [Installation guide](https://golangci-lint.run/usage/install/)

**Note**: GoReleaser is NOT required for development! GitHub Actions handles all releases automatically. Install it only if you want to test release builds locally.

---

## Quick Start

```bash
# 1. Clone the repository
git clone git@github.com:imran110219/stackctl.git
cd stackctl

# 2. Build
make build

# 3. Run tests
make test

# 4. Install locally
make install

# 5. Verify
stackctl version
```

---

## Development Workflow

### Fast Iteration

```bash
# Edit code, then:
make dev          # Format + build + install in one command

# Test your changes:
stackctl doctor
stackctl setup
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
# Opens coverage.html in browser
```

### Code Quality

```bash
# Format code
make fmt

# Run linters (requires golangci-lint)
make lint

# Tidy dependencies
make tidy
```

---

## Project Structure

```
stackctl/
├── cmd/
│   └── stackctl/           # CLI entry point
│       └── main.go         # Main function, TUI command routing
├── internal/
│   ├── stackctl/           # Core business logic
│   │   ├── cli.go          # Command handlers
│   │   ├── config.go       # Config management (.env, enabled.yml)
│   │   ├── modules.go      # Module catalog and dependencies
│   │   ├── compose.go      # Docker Compose generation
│   │   ├── render.go       # Template rendering
│   │   ├── exec.go         # Command execution
│   │   ├── doctor.go       # Pre-flight checks
│   │   ├── nginx.go        # Nginx config generation
│   │   ├── systemd.go      # Systemd unit generation
│   │   ├── backup.go       # Backup commands
│   │   └── init.go         # Environment initialization
│   └── tui/                # Bubble Tea TUI
│       ├── setup.go        # Setup wizard (9 screens)
│       ├── modules*.go     # Module manager
│       ├── dash*.go        # Live dashboard
│       ├── config*.go      # Config editor
│       ├── help.go         # Help overlay
│       ├── styles.go       # Theme and styling
│       └── keys.go         # Key bindings
├── templates/              # Embedded templates (go:embed)
│   ├── base/               # Base compose.yml, .env template
│   ├── modules/            # Module-specific compose files
│   ├── nginx/              # Nginx config templates
│   └── systemd/            # Systemd unit templates
├── scripts/                # Helper scripts
│   ├── suggest-version.sh
│   ├── changelog-prepare.sh
│   ├── changelog-add.sh
│   └── release-dry-run.sh
├── docs/                   # Documentation
├── Makefile                # Build automation
├── CHANGELOG.md            # Release history
└── go.mod                  # Go dependencies
```

---

## Making Changes

### Adding a New Command

1. **CLI Command** (e.g., `stackctl logs`):
   - Add handler in `internal/stackctl/cli.go`
   - Update usage in `usage()` function
   - Wire up in `Run()` switch statement

2. **TUI Command** (e.g., `stackctl logs` with TUI):
   - Add case in `cmd/stackctl/main.go`
   - Create new TUI screen in `internal/tui/`
   - Follow Bubble Tea MVC pattern (Model/Update/View)

### Adding a New Module

1. Create module compose file: `templates/modules/<module>/compose.yml`
2. Add to catalog in `internal/stackctl/modules.go`:
   ```go
   "<module>": {
       Name:        "<module>",
       Description: "...",
       Ports:       []string{"8080"},
       Category:    CategoryMonitoring,
       DependsOn:   []string{},
   }
   ```
3. Document in `docs/modules.md`

### Modifying Templates

Templates are embedded at build time using `go:embed`. During development:

1. **Edit templates** in `templates/` directory
2. **Rebuild** with `make build` (embeds templates)
3. **Test** changes

For development without rebuilding:
```bash
export STACKCTL_TEMPLATES=/path/to/stackctl/templates
stackctl apply --env dev
```

---

## Testing

### Manual Testing

Use the test plans in README.md:
- Manual test plan (v1) - CLI commands
- Manual test plan (TUI) - Interactive interfaces

### On a VM

For realistic testing:

```bash
# 1. Build binary
make build

# 2. Copy to test VM
scp bin/stackctl user@testvm:/tmp/

# 3. On the VM
sudo mv /tmp/stackctl /usr/local/bin/
stackctl version
stackctl doctor
stackctl setup
```

### With Docker Compose

Most features require Docker:

```bash
# Ensure Docker is running
docker info

# Test apply
stackctl init --env dev --domain dev.example.com --email test@example.com
stackctl apply --env dev

# Check containers
docker compose -f /srv/stack/dev/compose.yml ps
```

---

## Debugging

### TUI Debugging

Bubble Tea captures stdout/stderr. To debug:

1. **Log to file**:
   ```go
   f, _ := os.OpenFile("/tmp/stackctl-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
   fmt.Fprintf(f, "Debug: %+v\n", someVar)
   f.Close()
   ```

2. **Use tea.LogToFile**:
   ```go
   // In your TUI Init() or main
   f, _ := tea.LogToFile("/tmp/stackctl-tui.log", "tui")
   defer f.Close()
   ```

### Template Debugging

```bash
# Check if templates are embedded
go list -f '{{.EmbedFiles}}' ./cmd/stackctl

# Test template rendering
stackctl init --env test --domain test.local --email test@test.local
cat /srv/stack/test/compose.yml
```

### Build Info

```bash
# Check version variables
go build -ldflags "-X main.version=test" -o /tmp/stackctl ./cmd/stackctl
/tmp/stackctl version
```

---

## Release Process

See [RELEASE_WORKFLOW.md](./RELEASE_WORKFLOW.md) for complete details.

**Quick version:**

```bash
# 1. Check status
make version

# 2. Test build (optional, requires GoReleaser)
make release-dry-run

# 3. Prepare changelog
make changelog-prepare
# Edit CHANGELOG.md manually

# 4. Commit and tag
git add CHANGELOG.md
git commit -m "chore: prepare release v0.1.0"
git tag -a v0.1.0 -m "Release v0.1.0"
git push && git push --tags

# GitHub Actions builds and publishes automatically!
```

---

## Common Issues

### "permission denied" on /srv/

stackctl writes to `/srv/stack`, `/srv/data`, `/srv/backups` which may require sudo:

```bash
# Option 1: Use sudo
sudo stackctl apply --env dev

# Option 2: Fix ownership
sudo chown -R $USER:$USER /srv/stack /srv/data /srv/backups
stackctl apply --env dev
```

### Templates not found

If you see "template not found" errors:

```bash
# Check if templates are embedded
go list -f '{{.EmbedFiles}}' ./cmd/stackctl

# For development, point to local templates
export STACKCTL_TEMPLATES=$PWD/templates
```

### Docker socket permission

```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Logout/login to apply

# Or use sudo
sudo stackctl apply --env dev
```

### Go version mismatch

```bash
# Check Go version
go version

# Update go.mod if needed
go mod edit -go=1.24
go mod tidy
```

---

## Code Style

### General Guidelines

- Follow standard Go style ([Effective Go](https://go.dev/doc/effective_go))
- Run `make fmt` before committing
- Add comments for exported functions
- Keep functions focused and small
- Use meaningful variable names

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: resolve bug in X
docs: update README
refactor: simplify Y logic
test: add tests for Z
chore: update dependencies
```

Benefits:
- Auto-generated changelogs
- Semantic version suggestions
- Clear history

### Error Handling

```go
// Good: Return errors
func doSomething() error {
    if err := step1(); err != nil {
        return fmt.Errorf("step1 failed: %w", err)
    }
    return nil
}

// Avoid: Panic in library code
func doSomething() {
    if err := step1(); err != nil {
        panic(err)  // Only in main, not libraries
    }
}
```

---

## Getting Help

1. **Check documentation**: `docs/` directory
2. **Run help**: `stackctl --help`, `make help`
3. **Open issue**: [GitHub Issues](https://github.com/imran110219/stackctl/issues)
4. **Read code**: Comments in source files

---

## Contributing

(Coming soon: `docs/CONTRIBUTING.md`)

For now:
1. Fork the repository
2. Create a feature branch
3. Make changes following code style
4. Test thoroughly
5. Update documentation
6. Submit PR with clear description

---

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Docker Compose Spec](https://docs.docker.com/compose/compose-file/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)
