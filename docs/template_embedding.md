# Template Embedding Implementation

## Overview

As of 2026-03-08, stackctl templates are embedded directly in the binary using Go's `embed` package. This eliminates the need for a separate templates directory at runtime.

## Implementation Details

### File Structure

- **`templates_embed.go`** (repo root): Contains the `//go:embed` directive
  ```go
  //go:embed all:templates
  var EmbeddedTemplates embed.FS
  ```

- **`internal/stackctl/render.go`**: Template rendering engine
  - `InitTemplates()`: Called from main to inject embedded FS
  - `getTemplatesFS()`: Returns either embedded FS or DirFS from `STACKCTL_TEMPLATES`
  - `renderFile()`: Renders templates from the FS
  - `readTemplateFile()`: Reads raw template files

### How It Works

1. **At build time**: `go:embed` bundles `templates/` into the binary
2. **At runtime**:
   - `main()` calls `stackctl.InitTemplates(EmbeddedTemplates)`
   - All template operations use `getTemplatesFS()` which:
     - Returns `os.DirFS(STACKCTL_TEMPLATES)` if env var is set (dev mode)
     - Returns embedded FS otherwise (production mode)

### Updated Functions

All template-loading code was migrated from filesystem paths to `fs.FS`:

- **init.go**: `ensureDotEnv()`, `ensureComposeOverride()`
- **compose.go**: `writeCompose()`, `syncModuleAssets()`
- **nginx.go**: `writeNginxConfs()`
- **systemd.go**: `writeSystemdFiles()`, `writeBackupScript()`

### Development Mode

Set `STACKCTL_TEMPLATES` to use external templates during development:

```bash
export STACKCTL_TEMPLATES=/path/to/stackctl/templates
./stackctl init --env dev
```

### Binary Distribution

The binary is now fully self-contained:

- **No** templates directory needed at runtime
- **No** `~/.stackctl/templates` copying required
- Templates are read from embedded FS automatically

### Testing

Verify embedding works:

```bash
# Build binary
go build -o stackctl ./cmd/stackctl

# Test without templates directory
mv templates templates.backup
./stackctl --help  # Should work!
mv templates.backup templates

# Test dev override
STACKCTL_TEMPLATES=./templates ./stackctl --help
```

### Migration Notes

**Before** (file-based):
```go
templates := findTemplatesDir()
path := filepath.Join(templates, "base", "compose.base.yml")
renderFile(path, data)
```

**After** (embedded FS):
```go
renderFile("base/compose.base.yml", data)
```

### Deprecated

- `findTemplatesDir()`: Still exists for backward compat but returns empty string when using embedded FS
- Install script template copying: Will be removed in Task I when install.sh is rewritten

## Next Steps

See `docs/ROADMAP.md` Priority 1 for:
- GoReleaser config (Task I)
- Rewrite install.sh to download pre-built binaries (Task I)
