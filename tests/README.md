# stackctl Testing

This directory contains testing scripts and documentation for stackctl.

## Quick Start

### Automated Testing with Multipass (macOS)

```bash
# Install multipass
brew install multipass

# Test on Ubuntu 22.04 (creates VM, runs tests, cleans up)
./tests/multipass-test.sh 22.04

# Test on Ubuntu 24.04
./tests/multipass-test.sh 24.04

# Test without cleanup (keep VM for inspection)
./tests/multipass-test.sh 22.04 false
```

### Manual Testing on Any VM

```bash
# On a clean Ubuntu VM, run:
curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/tests/vm-test.sh | bash
```

## Test Scripts

### `vm-test.sh`
Automated test script that runs on the target VM. Tests:
- Installation process
- Binary verification
- Basic commands
- Pre-flight checks
- Docker integration
- TUI commands

**Usage:**
```bash
curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/tests/vm-test.sh | bash
```

### `multipass-test.sh`
Wrapper script for macOS that creates VMs with Multipass, runs tests, and cleans up.

**Usage:**
```bash
./tests/multipass-test.sh [22.04|24.04] [true|false]
```

**Arguments:**
- `22.04` or `24.04`: Ubuntu version (default: 22.04)
- `true` or `false`: Cleanup VM after test (default: true)

**Examples:**
```bash
# Test Ubuntu 22.04, cleanup after
./tests/multipass-test.sh 22.04

# Test Ubuntu 24.04, keep VM for inspection
./tests/multipass-test.sh 24.04 false
```

## Test Plan

See [VM_TEST_PLAN.md](./VM_TEST_PLAN.md) for:
- Complete manual testing procedures
- VM setup options
- Test result template
- Success criteria
- Reporting guidelines

## Test Results

Store test results in `results/` directory using this naming convention:

```
v{version}-ubuntu-{os_version}-{arch}.md
```

**Examples:**
- `v0.1.0-rc1-ubuntu-22.04-amd64.md`
- `v0.1.0-rc1-ubuntu-24.04-amd64.md`

Use the template from VM_TEST_PLAN.md.

## VM Options

### Multipass (Recommended for macOS)
- **Pros**: Fast, easy, clean VMs
- **Cons**: macOS/Linux only
- **Install**: `brew install multipass`

### Docker Containers (Quick Tests)
- **Pros**: Very fast, no VM overhead
- **Cons**: No systemd, limited testing
- **Use for**: Quick installer validation

### Cloud VMs (Production-like)
- **Pros**: Most realistic
- **Cons**: Costs money, slower setup
- **Providers**: DigitalOcean, AWS, Linode

### VirtualBox/VMware (Full Local VMs)
- **Pros**: Complete control
- **Cons**: Slow setup, manual process

## Running Tests

### Quick Test (Multipass)

```bash
# Test both versions
./tests/multipass-test.sh 22.04
./tests/multipass-test.sh 24.04
```

### Manual Test (Any VM)

1. Create a clean VM
2. SSH into it
3. Run test script:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/tests/vm-test.sh | bash
   ```

### Comprehensive Test (Manual Steps)

Follow the detailed manual test plan in [VM_TEST_PLAN.md](./VM_TEST_PLAN.md).

## Test Coverage

Current test coverage:

- ✅ Installation (curl | bash)
- ✅ Binary verification
- ✅ Basic CLI commands
- ✅ Pre-flight checks (doctor)
- ✅ Environment initialization
- ✅ Module management
- ✅ TUI command checks
- ⬜ Full TUI interaction (manual only)
- ⬜ Systemd integration (manual only)
- ⬜ Multi-environment setup
- ⬜ Backup/restore

## CI/CD Integration (Future)

Future GitHub Actions workflow:

```yaml
name: VM Tests
on: [push, pull_request]
jobs:
  test-ubuntu:
    strategy:
      matrix:
        os: [ubuntu-22.04, ubuntu-24.04]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: ./tests/vm-test.sh
```

## Troubleshooting

### Multipass issues

```bash
# List VMs
multipass list

# Delete stuck VM
multipass delete <name> --purge

# Restart multipass
multipass stop --all
multipass start --all
```

### Test script issues

```bash
# Run with debug output
bash -x ./tests/vm-test.sh

# Test on existing VM
multipass shell stackctl-test-22.04
curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/tests/vm-test.sh | bash
```

### Docker issues

```bash
# Install Docker
curl -fsSL https://get.docker.com | bash

# Start Docker
sudo systemctl start docker

# Add user to docker group
sudo usermod -aG docker $USER
# Logout/login to apply
```

## Contributing Tests

When adding new tests:

1. Update `vm-test.sh` with new test cases
2. Update `VM_TEST_PLAN.md` with manual procedures
3. Document in this README
4. Test on both Ubuntu 22.04 and 24.04

## References

- [VM_TEST_PLAN.md](./VM_TEST_PLAN.md) - Comprehensive test plan
- [Multipass](https://multipass.run/) - VM management
- [Docker](https://docs.docker.com/) - Container platform
- [Ubuntu Releases](https://releases.ubuntu.com/) - OS images
