#!/usr/bin/env bash
# Automated VM testing with Multipass
# Requires: brew install multipass (macOS)

set -e

# Configuration
VM_NAME_PREFIX="stackctl-test"
VM_CPUS=2
VM_MEMORY="2G"
VM_DISK="10G"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if multipass is installed
if ! command -v multipass >/dev/null 2>&1; then
    log_error "Multipass is not installed"
    echo ""
    echo "Install with:"
    echo "  macOS: brew install multipass"
    echo "  Linux: sudo snap install multipass"
    exit 1
fi

# Parse arguments
OS_VERSION="${1:-22.04}"
CLEANUP="${2:-true}"

if [[ ! "$OS_VERSION" =~ ^(22.04|24.04)$ ]]; then
    log_error "Invalid OS version: $OS_VERSION"
    echo "Usage: $0 [22.04|24.04] [true|false]"
    echo "  22.04 or 24.04: Ubuntu version to test"
    echo "  true or false: Cleanup VM after test (default: true)"
    exit 1
fi

VM_NAME="${VM_NAME_PREFIX}-${OS_VERSION}"

echo "=========================================="
echo "Multipass VM Testing for stackctl"
echo "=========================================="
echo ""
echo "OS Version: Ubuntu $OS_VERSION"
echo "VM Name: $VM_NAME"
echo "Cleanup after test: $CLEANUP"
echo ""

# Cleanup existing VM if present
if multipass list | grep -q "$VM_NAME"; then
    log_warning "VM $VM_NAME already exists, deleting..."
    multipass delete "$VM_NAME" || true
    multipass purge || true
fi

# Create VM
log_info "Creating VM: $VM_NAME"
if multipass launch "$OS_VERSION" \
    --name "$VM_NAME" \
    --cpus "$VM_CPUS" \
    --memory "$VM_MEMORY" \
    --disk "$VM_DISK"; then
    log_success "VM created"
else
    log_error "Failed to create VM"
    exit 1
fi

# Wait for VM to be ready
log_info "Waiting for VM to be ready..."
sleep 5

# Run test script on VM
log_info "Running test script on VM..."
echo ""

if multipass exec "$VM_NAME" -- bash -c "$(curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/tests/vm-test.sh)"; then
    log_success "Tests completed successfully on Ubuntu $OS_VERSION"
    TEST_RESULT="PASS"
else
    log_error "Tests failed on Ubuntu $OS_VERSION"
    TEST_RESULT="FAIL"
fi

echo ""

# Show VM info
log_info "VM Information:"
multipass info "$VM_NAME"

echo ""

# Cleanup
if [ "$CLEANUP" = "true" ]; then
    log_info "Cleaning up VM..."
    multipass delete "$VM_NAME"
    multipass purge
    log_success "VM deleted"
else
    log_info "VM preserved for manual inspection"
    echo ""
    echo "To access the VM:"
    echo "  multipass shell $VM_NAME"
    echo ""
    echo "To delete the VM later:"
    echo "  multipass delete $VM_NAME && multipass purge"
fi

echo ""
echo "=========================================="
echo "Test Result: $TEST_RESULT"
echo "=========================================="

if [ "$TEST_RESULT" = "PASS" ]; then
    exit 0
else
    exit 1
fi
