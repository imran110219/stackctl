# VM Testing Quick Start

**Goal**: Test stackctl v0.1.0-rc1 on clean Ubuntu VMs.

## Option 1: Automated (Multipass on macOS) ⚡

**Fastest way to test!**

```bash
# 1. Install multipass (one-time)
brew install multipass

# 2. Test Ubuntu 22.04
./tests/multipass-test.sh 22.04

# 3. Test Ubuntu 24.04
./tests/multipass-test.sh 24.04
```

**That's it!** The script will:
- Create a clean VM
- Run all tests automatically
- Show results
- Clean up VM

**Time**: ~5 minutes per OS version

---

## Option 2: Manual (Any VM) 🔧

**Most flexible, works anywhere:**

```bash
# 1. Create any Ubuntu 22.04 or 24.04 VM
#    (DigitalOcean, AWS, VirtualBox, etc.)

# 2. SSH into the VM

# 3. Run test script
curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/tests/vm-test.sh | bash
```

**Time**: ~10 minutes (plus VM creation time)

---

## What Gets Tested?

Both methods test:
- ✅ Installation (curl | bash)
- ✅ Binary verification
- ✅ `stackctl version`
- ✅ `stackctl doctor`
- ✅ `stackctl init`
- ✅ TUI commands
- ✅ Docker integration (if installed)

---

## Recording Results

After testing, record results:

```bash
# Create result file
vim tests/results/v0.1.0-rc1-ubuntu-22.04-amd64.md

# Use template from tests/VM_TEST_PLAN.md
```

---

## Next Steps

After successful testing:

1. **Update CHANGELOG.md**:
   ```markdown
   ## [0.1.0-rc1] - 2026-03-09

   ### Tested On
   - ✅ Ubuntu 22.04 LTS (amd64)
   - ✅ Ubuntu 24.04 LTS (amd64)
   ```

2. **Update ROADMAP.md**:
   - Mark task 5 as ✅ Done

3. **Prepare v0.1.0 release**:
   ```bash
   make changelog-prepare
   # Enter: 0.1.0
   ```

---

## Troubleshooting

### Multipass not working?
```bash
# Check status
multipass list

# Restart
multipass stop --all && multipass start --all
```

### Test script fails?
```bash
# Run with debug
bash -x tests/vm-test.sh
```

### Need help?
See full test plan: `tests/VM_TEST_PLAN.md`

---

## Time Estimate

- **Multipass (automated)**: 10 minutes total (both OS versions)
- **Manual testing**: 30-60 minutes (depends on VM setup)
- **Recording results**: 10 minutes

**Total**: 20-70 minutes depending on method chosen.

---

## Recommendation

**Use Multipass if you have macOS** - it's by far the fastest and easiest!

```bash
brew install multipass
./tests/multipass-test.sh 22.04
./tests/multipass-test.sh 24.04
```

Done! ✅
