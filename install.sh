#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
REPO="CaioDGallo/easy-cli"
BINARY_NAME="easy-cli"
INSTALL_DIR="/usr/local/bin"
VERSION="latest"

print_usage() {
    echo "Easy CLI Installer"
    echo "=================="
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --install-dir DIR        Installation directory (default: $INSTALL_DIR)"
    echo "  --version VERSION        Version to install (default: latest)"
    echo "  --help                   Show this help message"
    echo ""
    echo "This script installs the Easy CLI binary to your system."
    echo "After installation, you can use 'easy-cli' from anywhere in your terminal."
    echo ""
    echo "Environment configuration:"
    echo "The CLI will look for a .env file in the following locations:"
    echo "  1. ~/.easy-cli.env (recommended)"
    echo "  2. Same directory as the binary"
    echo "  3. Current working directory"
}

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        CYGWIN*|MINGW*|MSYS*) echo "windows";;
        *)          error "Unsupported operating system: $(uname -s)";;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64";;
        aarch64|arm64) echo "arm64";;
        armv7l) echo "arm";;
        i386|i686) echo "386";;
        *) error "Unsupported architecture: $(uname -m)";;
    esac
}

check_dependencies() {
    # Check if we have curl or wget
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        error "Neither curl nor wget is available. Please install one of them."
    fi
    
    # Check if we have tar
    if ! command -v tar >/dev/null 2>&1; then
        error "tar is not available. Please install tar."
    fi
}

get_download_url() {
    local os="$1"
    local arch="$2"
    local version="$3"
    
    if [ "$version" = "latest" ]; then
        echo "https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}_${os}_${arch}.tar.gz"
    else
        echo "https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}_${os}_${arch}.tar.gz"
    fi
}

download_and_install() {
    local os="$1"
    local arch="$2"
    local version="$3"
    
    local binary_name="${BINARY_NAME}"
    if [ "$os" = "windows" ]; then
        binary_name="${BINARY_NAME}.exe"
    fi
    
    local download_url=$(get_download_url "$os" "$arch" "$version")
    local temp_dir=$(mktemp -d)
    local temp_file="$temp_dir/${BINARY_NAME}.tar.gz"
    
    log "Downloading $BINARY_NAME ($version) for $os/$arch..."
    
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "$download_url" -o "$temp_file"; then
            error "Failed to download binary from $download_url"
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -q "$download_url" -O "$temp_file"; then
            error "Failed to download binary from $download_url"
        fi
    fi
    
    log "Extracting binary..."
    if ! tar -xzf "$temp_file" -C "$temp_dir"; then
        error "Failed to extract binary"
    fi
    
    # Install binary
    log "Installing to $INSTALL_DIR..."
    if [ -w "$INSTALL_DIR" ]; then
        if ! cp "$temp_dir/$binary_name" "$INSTALL_DIR/$BINARY_NAME"; then
            error "Failed to install binary"
        fi
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        log "Requesting sudo privileges for installation..."
        if ! sudo cp "$temp_dir/$binary_name" "$INSTALL_DIR/$BINARY_NAME"; then
            error "Failed to install binary with sudo"
        fi
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    success "Easy CLI installed successfully!"
}

create_env_template() {
    local env_file="$HOME/.easy-cli.env"
    
    if [ -f "$env_file" ]; then
        warn "Environment file already exists at $env_file"
        return
    fi
    
    log "Creating environment template at $env_file"
    
    cat > "$env_file" << 'EOF'
# Easy CLI Environment Configuration
# Copy this file and fill in your actual values

# Database Configuration
DB_HOST=your-database-host.rds.amazonaws.com
DB_USER=postgres
DB_PASSWORD=your_database_password_here
DB_NAME=postgres

# Vercel Configuration
VERCEL_TOKEN=your_vercel_token_here
VERCEL_TEAM_ID=your_vercel_team_id_here
VERCEL_FRONTEND_REPO_UUID={e4839c5c-d412-4c9d-88f7-c6209fef4b6a}

# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_aws_access_key_id_here
AWS_SECRET_ACCESS_KEY=your_aws_secret_access_key_here

# DigitalOcean Configuration
DO_TOKEN=your_digitalocean_token_here

# SMTP Configuration
SMTP_SERVER=your-smtp-server.com
SMTP_USERNAME=your-smtp-username@yourdomain.com
SMTP_DO_NOT_REPLY_EMAIL=noreply@yourdomain.com
SMTP_DEV_EMAIL=developer@yourdomain.com

# Repository Configuration
BACKEND_REPO=your-org/your-backend-repo
FRONTEND_REPO=your-org/your-frontend-repo

# Application Configuration
APP_NAME_PREFIX=your-app-prefix
EOF
    
    chmod 600 "$env_file"  # Make it readable only by owner
    success "Environment template created at $env_file"
    warn "Please edit $env_file with your actual credentials before using easy-cli"
}

main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            --help)
                print_usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    echo ""
    echo -e "${GREEN}🛠️  Easy CLI Installer${NC}"
    echo "======================="
    echo ""
    
    # Check dependencies
    check_dependencies
    
    # Detect system
    OS=$(detect_os)
    ARCH=$(detect_arch)
    log "Detected system: $OS/$ARCH"
    
    # Download and install binary
    download_and_install "$OS" "$ARCH" "$VERSION"
    
    # Create environment template
    create_env_template
    
    echo ""
    echo -e "${GREEN}🎉 Installation Complete!${NC}"
    echo ""
    echo "Easy CLI has been installed to: $INSTALL_DIR/$BINARY_NAME"
    echo ""
    echo "Next steps:"
    echo "  1. Edit your environment file: ~/.easy-cli.env"
    echo "  2. Add your API credentials and configuration"
    echo "  3. Run: easy-cli fresh-install --client-name \"Your Client\""
    echo ""
    echo "For help: easy-cli --help"
    
    # Check if install dir is in PATH
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo ""
        warn "Note: $INSTALL_DIR is not in your PATH."
        warn "You may need to add it to your PATH or use the full path: $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Run main function
main "$@"