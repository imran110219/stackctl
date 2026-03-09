# VM Testing Plan for stackctl

This document outlines the complete testing plan for stackctl on clean Ubuntu VMs.

## Objective

Verify that stackctl v0.1.0-rc1 installs and works correctly on:
- Ubuntu 22.04 LTS (Jammy Jellyfish)
- Ubuntu 24.04 LTS (Noble Numbat)
- Both amd64 and arm64 architectures (if available)

---

## Test Environments

### Required Tests
- [ ] Ubuntu 22.04 LTS (amd64)
- [ ] Ubuntu 24.04 LTS (amd64)

### Optional Tests (if available)
- [ ] Ubuntu 22.04 LTS (arm64)
- [ ] Ubuntu 24.04 LTS (arm64)

---

## VM Setup Options

Choose one of these options to get test VMs:

### Option 1: Local VMs (Multipass) - **Recommended for macOS**

Fast and easy on macOS:

```bash
# Install Multipass (if not installed)
brew install multipass

# Create Ubuntu 22.04 VM
multipass launch 22.04 --name stackctl-test-22 --cpus 2 --memory 2G --disk 10G

# Create Ubuntu 24.04 VM
multipass launch 24.04 --name stackctl-test-24 --cpus 2 --memory 2G --disk 10G

# Shell into VM
multipass shell stackctl-test-22

# When done, cleanup
multipass delete stackctl-test-22
multipass purge
```

### Option 2: Docker Containers - **Quick Testing**

Fast but limited (systemd won't work):

```bash
# Ubuntu 22.04
docker run -it --rm ubuntu:22.04 bash

# Ubuntu 24.04
docker run -it --rm ubuntu:24.04 bash

# Inside container, install prerequisites:
apt update && apt install -y curl sudo
```

**Note**: Docker containers can't test systemd features, but good for quick installer validation.

### Option 3: Cloud VMs - **Production-like**

Most realistic testing:

**DigitalOcean**:
```bash
# Create droplet via web UI or CLI
doctl compute droplet create stackctl-test \
  --image ubuntu-22-04-x64 \
  --size s-1vcpu-1gb \
  --region nyc1

# SSH in
doctl compute ssh stackctl-test
```

**AWS EC2**:
```bash
# Launch instance via web console or CLI
# AMI: ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*
# Instance type: t2.micro (free tier)
```

**Linode**:
```bash
# Create Linode via web UI
# Image: Ubuntu 22.04 LTS
# Plan: Nanode 1GB
```

### Option 4: VirtualBox/VMware - **Local Full VMs**

Download Ubuntu ISOs and install:
- [Ubuntu 22.04](https://releases.ubuntu.com/22.04/)
- [Ubuntu 24.04](https://releases.ubuntu.com/24.04/)

---

## Test Script

We've created an automated test script: `tests/vm-test.sh`

### Quick Test (Automated)

```bash
# On the VM:
curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/tests/vm-test.sh | bash
```

This script automatically tests:
1. ✅ Pre-installation checks
2. ✅ Installation process
3. ✅ Binary verification
4. ✅ Basic commands
5. ✅ Pre-flight checks (doctor)
6. ✅ Docker availability
7. ✅ Environment initialization (if Docker available)
8. ✅ TUI command checks

### Manual Test (Comprehensive)

Follow the manual steps below for thorough testing.

---

## Manual Test Steps

### Phase 1: Clean System

**Verify clean state:**
```bash
# Check OS version
lsb_release -a

# Verify stackctl not installed
which stackctl  # Should not exist

# Check architecture
uname -m
```

**Record results:**
- OS: _________________
- Architecture: _________________
- Kernel: _________________

---

### Phase 2: Installation

**Run installer:**
```bash
curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/install.sh | bash
```

**Verify installation:**
```bash
# Check binary exists
which stackctl

# Check version
stackctl version

# Check help
stackctl --help
```

**Expected results:**
- Binary at `/usr/local/bin/stackctl`
- Version shows `v0.1.0-rc1` or later
- Help text displays correctly

**Record results:**
- Installation time: _________________
- Version installed: _________________
- Any errors: _________________

---

### Phase 3: Pre-flight Checks

**Run doctor:**
```bash
stackctl doctor
```

**Expected results:**
- All checks pass (green ✓), or
- Docker check fails (expected if Docker not installed)

**Record results:**
- Docker check: ⬜ Pass ⬜ Fail
- Port checks: ⬜ Pass ⬜ Fail
- Permission checks: ⬜ Pass ⬜ Fail

---

### Phase 4: Install Docker (if needed)

If Docker is not installed:

```bash
# Install Docker
curl -fsSL https://get.docker.com | bash

# Add user to docker group (optional, to avoid sudo)
sudo usermod -aG docker $USER

# Logout and login again, then verify
docker info
```

**Re-run doctor:**
```bash
stackctl doctor
```

**Record results:**
- Docker installation: ⬜ Success ⬜ Failed
- Doctor checks after Docker: ⬜ All Pass ⬜ Some Fail

---

### Phase 5: CLI Commands

**Test initialization:**
```bash
sudo stackctl init --env dev --domain dev.example.com --email test@example.com
```

**Verify created files:**
```bash
ls -la /srv/stack/dev/
ls -la /srv/data/dev/
```

**Expected files:**
- `/srv/stack/dev/compose.yml`
- `/srv/stack/dev/compose.override.yml`
- `/srv/stack/dev/.env`
- `/srv/stack/dev/enabled.yml`
- `/srv/stack/dev/nginx/conf.d/`

**Test status:**
```bash
sudo stackctl status --env dev
```

**Test module management:**
```bash
# Enable a module
sudo stackctl enable prometheus --env dev

# Check enabled modules
cat /srv/stack/dev/enabled.yml

# Disable module
sudo stackctl disable prometheus --env dev
```

**Record results:**
- Init command: ⬜ Success ⬜ Failed
- Files created: ⬜ All present ⬜ Missing: _________________
- Status command: ⬜ Success ⬜ Failed
- Enable/disable: ⬜ Success ⬜ Failed

---

### Phase 6: TUI Commands

**Test setup wizard:**
```bash
sudo stackctl setup
```

**Steps:**
1. Navigate through wizard
2. Select environment (qa)
3. Enter domain: qa.example.com
4. Enter email: test@example.com
5. Select some modules (prometheus, grafana)
6. Complete setup

**Test module manager:**
```bash
sudo stackctl modules --env qa
```

**Actions:**
1. Browse modules
2. Enable/disable a module
3. View module details
4. Press `?` for help

**Test dashboard:**
```bash
sudo stackctl dash
```

**Actions:**
1. View environment overview
2. Select an environment
3. View container status
4. Wait for auto-refresh (5 seconds)

**Test config editor:**
```bash
sudo stackctl config --env dev
```

**Actions:**
1. Edit a value
2. Verify secret masking
3. Generate a password
4. Save changes

**Record results:**
- Setup wizard: ⬜ Success ⬜ Failed ⬜ Notes: _________________
- Module manager: ⬜ Success ⬜ Failed ⬜ Notes: _________________
- Dashboard: ⬜ Success ⬜ Failed ⬜ Notes: _________________
- Config editor: ⬜ Success ⬜ Failed ⬜ Notes: _________________
- Help overlay (`?`): ⬜ Works ⬜ Doesn't work

---

### Phase 7: Apply and Verify

**Apply configuration:**
```bash
sudo stackctl apply --env dev
```

**Verify containers:**
```bash
docker compose -f /srv/stack/dev/compose.yml ps
```

**Check logs:**
```bash
docker compose -f /srv/stack/dev/compose.yml logs nginx
```

**Test backup (if modules enabled):**
```bash
sudo stackctl backup --env dev
ls -lh /srv/backups/dev/
```

**Record results:**
- Apply command: ⬜ Success ⬜ Failed
- Containers running: ⬜ Yes ⬜ No ⬜ Count: _________________
- Logs accessible: ⬜ Yes ⬜ No
- Backup command: ⬜ Success ⬜ Failed ⬜ Skipped

---

### Phase 8: Systemd (Optional)

If systemd units were generated:

```bash
# Check generated units
ls /srv/stack/dev/systemd/

# Test systemd integration (if units installed)
sudo systemctl status stackctl-dev
```

**Record results:**
- Systemd units created: ⬜ Yes ⬜ No
- Systemd status: ⬜ Active ⬜ Inactive ⬜ Not tested

---

### Phase 9: Cleanup

**Remove test environments:**
```bash
sudo rm -rf /srv/stack/dev /srv/data/dev /srv/backups/dev
sudo rm -rf /srv/stack/qa /srv/data/qa /srv/backups/qa
```

**Uninstall (optional):**
```bash
sudo rm /usr/local/bin/stackctl
```

---

## Test Result Template

Copy this template for each VM tested:

```markdown
### Test Results: Ubuntu [22.04/24.04] [amd64/arm64]

**Date**: YYYY-MM-DD
**Tester**: Your Name
**VM Provider**: [Multipass/Docker/DigitalOcean/AWS/Other]

#### Environment
- OS Version:
- Kernel:
- Architecture:
- stackctl Version:

#### Installation
- ⬜ Installer completed successfully
- ⬜ Binary installed at /usr/local/bin/stackctl
- ⬜ Version command works
- Time taken: ___ seconds

#### Basic Commands
- ⬜ `stackctl --help` works
- ⬜ `stackctl version` works
- ⬜ `stackctl doctor` works

#### CLI Functionality
- ⬜ `stackctl init` works
- ⬜ Files created correctly
- ⬜ `stackctl status` works
- ⬜ `stackctl enable/disable` works
- ⬜ `stackctl apply` works

#### TUI Functionality
- ⬜ `stackctl setup` wizard works
- ⬜ `stackctl modules` manager works
- ⬜ `stackctl dash` dashboard works
- ⬜ `stackctl config` editor works
- ⬜ Help overlay (`?`) works

#### Docker Integration
- ⬜ Containers start successfully
- ⬜ Logs are accessible
- ⬜ Compose profiles work

#### Issues Found
- Issue 1:
- Issue 2:
- Issue 3:

#### Overall Result
⬜ PASS - Ready for production
⬜ PASS WITH ISSUES - Works but has minor issues
⬜ FAIL - Critical issues found

#### Notes

```

---

## Reporting Results

After testing, create a test report:

1. **Create report file**: `tests/results/v0.1.0-rc1-ubuntu-22.04-amd64.md`
2. **Fill in all sections** from template above
3. **Document all issues** with reproduction steps
4. **Take screenshots** if UI issues found
5. **Update CHANGELOG.md** with any discovered issues

---

## Success Criteria

For v0.1.0-rc1 to pass testing:

### Must Pass (Critical)
- [ ] Installer completes without errors
- [ ] Binary is executable and shows version
- [ ] `stackctl doctor` runs without crashes
- [ ] `stackctl init` creates environment
- [ ] `stackctl apply` starts containers (with Docker)
- [ ] TUI wizards don't crash

### Should Pass (Important)
- [ ] All CLI commands work as documented
- [ ] All TUI screens are navigable
- [ ] Help overlay displays correctly
- [ ] Module enable/disable works
- [ ] Config editor saves changes

### Nice to Have
- [ ] Performance is acceptable
- [ ] Error messages are clear
- [ ] Documentation matches behavior

---

## Next Steps After Testing

1. **All tests pass**:
   - Update CHANGELOG.md: Mark v0.1.0-rc1 as "tested on Ubuntu 22.04/24.04"
   - Proceed to v0.1.0 final release
   - Update documentation

2. **Minor issues found**:
   - Document in CHANGELOG.md "Known Issues" section
   - Create GitHub issues
   - Decide if issues are blocking for v0.1.0

3. **Critical issues found**:
   - Fix issues
   - Release v0.1.0-rc2
   - Re-test

---

## Automation Future

Future improvements to this testing process:

- [ ] GitHub Actions workflow for automated VM testing
- [ ] Terraform/Vagrant scripts for VM provisioning
- [ ] Automated test suite (Go tests)
- [ ] CI/CD integration tests
- [ ] Screenshot capture automation

---

## References

- Install script: https://raw.githubusercontent.com/imran110219/stackctl/main/install.sh
- Release: https://github.com/imran110219/stackctl/releases/tag/v0.1.0-rc1
- Documentation: https://github.com/imran110219/stackctl/blob/main/USER_GUIDE.md
