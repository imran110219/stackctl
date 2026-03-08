# Install Script Testing Guide

This guide explains how to test the install.sh script before and after creating releases.

## Pre-Release Testing (Before v0.1.0)

Since the install script requires GitHub Releases to exist, you need to test the logic without actually running it end-to-end.

### 1. Syntax Validation

```bash
# Check for syntax errors
bash -n install.sh

# Check with shellcheck (if available)
shellcheck install.sh
```

### 2. Function Testing

Test individual functions by sourcing the script:

```bash
# Create a test script
cat > test_install.sh << 'EOF'
#!/usr/bin/env bash
source ./install.sh

# Test OS detection
echo "OS: $(detect_os)"

# Test arch detection
echo "Arch: $(detect_arch)"

# Test version parsing (requires internet)
# GITHUB_REPO="example/stackctl"
# echo "Latest: $(get_latest_version)"
EOF

chmod +x test_install.sh
./test_install.sh
```

Expected output:
```
OS: linux
Arch: amd64  # or arm64 depending on your system
```

### 3. Dry Run Simulation

Test the download URL format without actually downloading:

```bash
# Simulate what the script would download
VERSION="v0.1.0"
OS="linux"
ARCH="amd64"
REPO="example/stackctl"

echo "Would download:"
echo "https://github.com/${REPO}/releases/download/${VERSION}/stackctl_${VERSION}_${OS}_${ARCH}.tar.gz"
```

Expected:
```
Would download:
https://github.com/example/stackctl/releases/download/v0.1.0/stackctl_v0.1.0_linux_amd64.tar.gz
```

## Post-Release Testing (After v0.1.0)

Once you've created a release (even a pre-release), test the full installation flow.

### 1. Test with Pre-Release Tag

Create a pre-release first to test the entire flow:

```bash
# Create and push a pre-release tag
git tag -a v0.1.0-rc1 -m "Release candidate 1"
git push origin v0.1.0-rc1

# Wait for GitHub Actions to complete
# Check: https://github.com/YOUR_ORG/stackctl/actions
```

### 2. Test Installation from Pre-Release

On a **clean Ubuntu VM** (22.04 or 24.04):

```bash
# Test specific version
export STACKCTL_VERSION=v0.1.0-rc1
export STACKCTL_REPO=YOUR_ORG/stackctl
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash

# Verify installation
stackctl version
stackctl --help
```

### 3. Test Different Installation Directories

#### User-local installation (no sudo)
```bash
export STACKCTL_INSTALL_DIR="$HOME/.local/bin"
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash

# Verify
which stackctl
# Should be: /home/username/.local/bin/stackctl
```

#### System-wide installation (requires sudo)
```bash
export STACKCTL_INSTALL_DIR="/usr/local/bin"
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash

# Verify
which stackctl
# Should be: /usr/local/bin/stackctl
```

### 4. Test Latest Version Detection

```bash
# Don't specify version (should auto-detect latest)
unset STACKCTL_VERSION
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash

# Verify it installed the latest
stackctl version
```

### 5. Test on Different Architectures

#### AMD64 (x86_64)
```bash
# On an AMD64 machine or VM
uname -m  # Should show: x86_64
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash
stackctl version
```

#### ARM64 (aarch64)
```bash
# On an ARM64 machine (Raspberry Pi 4, AWS Graviton, etc.)
uname -m  # Should show: aarch64
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash
stackctl version
```

### 6. Test Error Handling

#### Invalid version
```bash
export STACKCTL_VERSION=v99.99.99
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash
# Should fail with: "Failed to download..."
```

#### Unsupported OS
```bash
# On macOS
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash
# Should fail with: "Unsupported operating system: Darwin"
```

#### Missing curl/wget
```bash
# Remove curl and wget (dangerous - for testing only!)
sudo apt remove curl wget
bash install.sh
# Should fail with: "curl or wget is required"
```

### 7. Test Binary Functionality

After successful installation:

```bash
# Test that embedded templates work
stackctl doctor

# Test that commands are available
stackctl --help
stackctl setup  # Should launch TUI (exit with Ctrl+C)

# Verify no external template directory is needed
ls -la ~/.stackctl/templates  # Should NOT exist
env | grep STACKCTL  # Should NOT require STACKCTL_TEMPLATES
```

## Automated Testing with Docker

Create a test script that uses Docker to simulate clean installations:

```bash
cat > test_install_docker.sh << 'EOF'
#!/usr/bin/env bash
set -e

# Test on Ubuntu 22.04 amd64
docker run --rm -v $(pwd)/install.sh:/install.sh ubuntu:22.04 bash -c "
  apt-get update && apt-get install -y curl
  export STACKCTL_VERSION=v0.1.0-rc1
  export STACKCTL_REPO=YOUR_ORG/stackctl
  export STACKCTL_INSTALL_DIR=/usr/local/bin
  bash /install.sh
  stackctl version
"

# Test on Ubuntu 24.04 amd64
docker run --rm -v $(pwd)/install.sh:/install.sh ubuntu:24.04 bash -c "
  apt-get update && apt-get install -y curl
  export STACKCTL_VERSION=v0.1.0-rc1
  export STACKCTL_REPO=YOUR_ORG/stackctl
  export STACKCTL_INSTALL_DIR=/usr/local/bin
  bash /install.sh
  stackctl version
"

echo "All tests passed!"
EOF

chmod +x test_install_docker.sh
```

## Checklist Before Official Release

Before creating the official v0.1.0 release:

- [ ] Syntax check passes (`bash -n install.sh`)
- [ ] Pre-release tag created (e.g., v0.1.0-rc1)
- [ ] GitHub Actions workflow completed successfully
- [ ] Release assets are present on GitHub
- [ ] Install script works on Ubuntu 22.04 amd64
- [ ] Install script works on Ubuntu 24.04 amd64
- [ ] Install script works on ARM64 (if available)
- [ ] `stackctl version` shows correct version
- [ ] `stackctl --help` works
- [ ] `stackctl doctor` works (implies templates are embedded)
- [ ] No external template directory is required
- [ ] Latest version detection works (no STACKCTL_VERSION set)
- [ ] Specific version selection works (STACKCTL_VERSION=v0.1.0-rc1)
- [ ] Error handling works (invalid version, missing tools)

## Troubleshooting

### "Failed to download from GitHub"

**Cause**: Release doesn't exist or URL is wrong

**Fix**: Verify the release exists:
```bash
GITHUB_REPO="YOUR_ORG/stackctl"
VERSION="v0.1.0-rc1"
curl -I "https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/stackctl_${VERSION}_linux_amd64.tar.gz"
# Should return: HTTP/2 200 (not 404)
```

### "stackctl: command not found"

**Cause**: Install directory not in PATH

**Fix**: Add to shell profile:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### "Permission denied"

**Cause**: Trying to install to /usr/local/bin without sudo

**Fix**: Either use sudo or install to user directory:
```bash
export STACKCTL_INSTALL_DIR="$HOME/.local/bin"
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash
```

## Integration with Documentation

After testing is complete, update:

1. **README.md**: Add the one-liner install command
2. **USER_GUIDE.md**: Add detailed installation instructions
3. **docs/RELEASE.md**: Reference this testing guide

Example for README.md:
```markdown
## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash
```

See [USER_GUIDE.md](./USER_GUIDE.md) for detailed installation options.
```
