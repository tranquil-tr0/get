#!/usr/bin/env bash

# Get - Simple Installation Script
# Downloads and installs the Get CLI tool to /usr/local/bin

set -e

# Configuration
BINARY_NAME="get"
GITHUB_REPO="tranquil-tr0/get"
INSTALL_DIR="/usr/local/bin"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo "→ $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# Detect system architecture
detect_arch() {
    case $(uname -m) in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
}

# Detect operating system
detect_os() {
    case "$OSTYPE" in
        linux-gnu*) echo "linux" ;;
        darwin*) echo "darwin" ;;
        *) log_error "Unsupported OS: $OSTYPE"; exit 1 ;;
    esac
}

# Get latest release version
get_latest_version() {
    local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        curl -s "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        log_error "curl or wget required"
        exit 1
    fi
}

# Download and install binary
install_binary() {
    local os=$(detect_os)
    local arch=$(detect_arch)
    local version=$(get_latest_version)
    
    log_info "Detected: $os/$arch"
    log_info "Latest version: $version"
    
    # Create temp directory
    local temp_dir=$(mktemp -d)
    trap "rm -rf '$temp_dir'" EXIT
    
    # Construct download URL
    local binary_name="${BINARY_NAME}-${os}-${arch}"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${binary_name}"
    local temp_file="${temp_dir}/${binary_name}"
    
    log_info "Downloading ${binary_name}..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_file" "$download_url" || {
            log_error "Download failed"
            exit 1
        }
    else
        wget -O "$temp_file" "$download_url" || {
            log_error "Download failed"
            exit 1
        }
    fi
    
    # Install to /usr/local/bin
    log_info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    
    if [ -w "$INSTALL_DIR" ]; then
        cp "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo cp "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    log_success "Installed successfully!"
}

# Main
main() {
    log_info "Installing Get CLI from GitHub..."
    install_binary
    log_success "Installation complete. Try: get --help"
}

# Run main if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi