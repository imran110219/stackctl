#!/usr/bin/env bash
# VM Installation Test Script for stackctl
# Run this on a clean Ubuntu 22.04 or 24.04 VM

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test function wrapper
run_test() {
    local test_name="$1"
    local test_cmd="$2"

    log_info "Testing: $test_name"
    if eval "$test_cmd" > /dev/null 2>&1; then
        log_success "$test_name"
        return 0
    else
        log_error "$test_name"
        return 1
    fi
}

# Header
echo "=========================================="
echo "stackctl VM Installation Test"
echo "=========================================="
echo ""

# System info
log_info "System Information"
echo "  OS: $(lsb_release -d | cut -f2-)"
echo "  Kernel: $(uname -r)"
echo "  Architecture: $(uname -m)"
echo ""

# Step 1: Pre-installation checks
log_info "Step 1: Pre-installation Checks"
echo ""

run_test "curl is installed" "command -v curl"
run_test "bash is installed" "command -v bash"
run_test "Internet connectivity" "curl -s --max-time 5 https://github.com"

echo ""

# Step 2: Install stackctl
log_info "Step 2: Installing stackctl"
echo ""

log_info "Running installer..."
if curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/install.sh | bash; then
    log_success "Installation completed"
else
    log_error "Installation failed"
    exit 1
fi

echo ""

# Step 3: Verify installation
log_info "Step 3: Verify Installation"
echo ""

run_test "stackctl command exists" "command -v stackctl"
run_test "stackctl is executable" "test -x /usr/local/bin/stackctl"

# Check version
if command -v stackctl >/dev/null 2>&1; then
    VERSION=$(stackctl version 2>&1 || echo "unknown")
    echo "  Installed version: $VERSION"

    if [[ "$VERSION" == *"v0.1.0"* ]]; then
        log_success "Version check (v0.1.0-rc1 or later)"
    else
        log_warning "Version is $VERSION (expected v0.1.0+)"
    fi
fi

echo ""

# Step 4: Test basic commands
log_info "Step 4: Testing Basic Commands"
echo ""

run_test "stackctl --help" "stackctl --help"
run_test "stackctl version" "stackctl version"

echo ""

# Step 5: Test doctor (pre-flight checks)
log_info "Step 5: Testing Pre-flight Checks"
echo ""

log_info "Running stackctl doctor..."
if stackctl doctor; then
    log_success "Pre-flight checks passed"
else
    log_warning "Pre-flight checks failed (expected if Docker not installed)"
    echo "  Note: Docker may need to be installed manually"
fi

echo ""

# Step 6: Check if Docker is available
log_info "Step 6: Docker Check"
echo ""

if command -v docker >/dev/null 2>&1; then
    log_success "Docker is installed"

    if docker info >/dev/null 2>&1; then
        log_success "Docker daemon is running"
        DOCKER_AVAILABLE=true
    else
        log_warning "Docker daemon is not running"
        echo "  Try: sudo systemctl start docker"
        DOCKER_AVAILABLE=false
    fi
else
    log_warning "Docker is not installed"
    echo "  stackctl requires Docker. Install with:"
    echo "  curl -fsSL https://get.docker.com | bash"
    DOCKER_AVAILABLE=false
fi

echo ""

# Step 7: Test initialization (if Docker is available)
if [ "$DOCKER_AVAILABLE" = true ]; then
    log_info "Step 7: Testing Environment Initialization"
    echo ""

    log_info "Creating test environment..."
    if sudo stackctl init --env test --domain test.local --email test@test.local; then
        log_success "Environment initialization"

        # Check created files
        run_test "compose.yml created" "test -f /srv/stack/test/compose.yml"
        run_test ".env created" "test -f /srv/stack/test/.env"
        run_test "enabled.yml created" "test -f /srv/stack/test/enabled.yml"

        # Test status command
        if sudo stackctl status --env test; then
            log_success "Status command"
        else
            log_warning "Status command (environment may not be running yet)"
        fi

        # Cleanup
        log_info "Cleaning up test environment..."
        sudo rm -rf /srv/stack/test /srv/data/test /srv/backups/test
        log_success "Cleanup completed"
    else
        log_error "Environment initialization"
    fi
else
    log_warning "Step 7: Skipping initialization tests (Docker not available)"
fi

echo ""

# Step 8: Test TUI commands (basic check)
log_info "Step 8: Testing TUI Commands"
echo ""

# Can't fully test interactive TUI in script, but can check if commands run
run_test "stackctl setup --help" "stackctl setup --help 2>&1 | grep -q 'setup'"
run_test "stackctl modules --help" "stackctl modules --help 2>&1 | grep -q 'modules'"
run_test "stackctl dash --help" "stackctl dash --help 2>&1 | grep -q 'dash'"
run_test "stackctl config --help" "stackctl config --help 2>&1 | grep -q 'config'"

echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo "Total tests: $TESTS_TOTAL"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
else
    echo "Failed: $TESTS_FAILED"
fi
echo ""

# Final verdict
if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "stackctl is ready to use on this system."
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed${NC}"
    echo ""
    echo "Review the failures above. Common issues:"
    echo "  - Docker not installed: curl -fsSL https://get.docker.com | bash"
    echo "  - Docker not running: sudo systemctl start docker"
    echo "  - Permission issues: Add user to docker group or use sudo"
    exit 1
fi
