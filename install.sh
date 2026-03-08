#!/usr/bin/env bash
# stackctl installer - downloads pre-built binaries from GitHub Releases
# Usage: curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/stackctl/main/install.sh | bash
set -euo pipefail

# Configuration
GITHUB_REPO="${STACKCTL_REPO:-example/stackctl}"
INSTALL_DIR="${STACKCTL_INSTALL_DIR:-/usr/local/bin}"
VERSION="${STACKCTL_VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}==>${NC} $*"
}

success() {
    echo -e "${GREEN}==>${NC} $*"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $*" >&2
}

error() {
    echo -e "${RED}Error:${NC} $*" >&2
    exit 1
}

# Detect OS
detect_os() {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux*)
            echo "linux"
            ;;
        *)
            error "Unsupported operating system: $os. stackctl only supports Linux."
            ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            error "Unsupported architecture: $arch. stackctl supports amd64 and arm64."
            ;;
    esac
}

# Check if running as root (for /usr/local/bin installation)
check_sudo() {
    if [[ "$INSTALL_DIR" == "/usr/local/bin" ]] && [[ $EUID -ne 0 ]]; then
        if ! command -v sudo >/dev/null 2>&1; then
            error "sudo is required to install to $INSTALL_DIR"
        fi
        SUDO="sudo"
    else
        SUDO=""
    fi
}

# Get latest release version from GitHub
get_latest_version() {
    local url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    local version

    if command -v curl >/dev/null 2>&1; then
        version=$(curl -fsSL "$url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        version=$(wget -qO- "$url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "curl or wget is required"
    fi

    if [[ -z "$version" ]]; then
        error "Failed to fetch latest version from GitHub"
    fi

    echo "$version"
}

# Download and extract archive
download_and_install() {
    local version="$1"
    local os="$2"
    local arch="$3"

    local archive="stackctl_${version}_${os}_${arch}.tar.gz"
    local url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${archive}"
    local tmp_dir
    tmp_dir="$(mktemp -d)"

    info "Downloading stackctl ${version} for ${os}/${arch}..."

    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "$url" -o "${tmp_dir}/${archive}"; then
            error "Failed to download from $url"
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -q "$url" -O "${tmp_dir}/${archive}"; then
            error "Failed to download from $url"
        fi
    else
        error "curl or wget is required"
    fi

    info "Extracting archive..."
    tar -xzf "${tmp_dir}/${archive}" -C "${tmp_dir}"

    info "Installing to ${INSTALL_DIR}..."
    $SUDO mkdir -p "$INSTALL_DIR"
    $SUDO install -m 755 "${tmp_dir}/stackctl" "${INSTALL_DIR}/stackctl"

    # Cleanup
    rm -rf "$tmp_dir"
}

# Verify installation
verify_installation() {
    if ! command -v stackctl >/dev/null 2>&1; then
        warn "stackctl was installed but is not in your PATH"
        echo ""
        echo "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        echo "Then reload your shell or run:"
        echo "  source ~/.bashrc  # or ~/.zshrc"
        return
    fi

    local installed_version
    installed_version=$(stackctl version 2>/dev/null || echo "unknown")
    success "stackctl installed successfully!"
    echo ""
    echo "Version: $installed_version"
    echo "Location: $(command -v stackctl)"
    echo ""
    echo "Get started:"
    echo "  stackctl setup    # Interactive setup wizard"
    echo "  stackctl --help   # View all commands"
}

# Main installation flow
main() {
    echo ""
    echo "╔═══════════════════════════════════════╗"
    echo "║     stackctl Installer                ║"
    echo "║     Community Docker Platform         ║"
    echo "╚═══════════════════════════════════════╝"
    echo ""

    # Detect system
    local os arch version
    os=$(detect_os)
    arch=$(detect_arch)

    info "Detected: ${os}/${arch}"

    # Check for sudo if needed
    check_sudo

    # Determine version
    if [[ "$VERSION" == "latest" ]]; then
        version=$(get_latest_version)
        info "Latest version: ${version}"
    else
        version="$VERSION"
        info "Requested version: ${version}"
    fi

    # Download and install
    download_and_install "$version" "$os" "$arch"

    # Verify
    echo ""
    verify_installation

    echo ""
    echo "📚 Documentation: https://github.com/${GITHUB_REPO}"
    echo "🐛 Issues: https://github.com/${GITHUB_REPO}/issues"
    echo ""
}

# Run main function
main "$@"
