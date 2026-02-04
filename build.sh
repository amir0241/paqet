#!/usr/bin/env bash
set -e

# Build script for paqet
# This script provides an easy way to build paqet for the current platform or create releases

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage
show_usage() {
    cat << EOF
Paqet Build Script

Usage:
  ./build.sh [command]

Commands:
  build       Build for current platform (default)
  release     Build and create release archives for all platforms
  linux       Build for Linux (amd64 and arm64)
  darwin      Build for macOS (amd64 and arm64)
  windows     Build for Windows (amd64)
  clean       Clean build artifacts
  help        Show this help message

Examples:
  ./build.sh              # Build for current platform
  ./build.sh release      # Create release archives
  ./build.sh linux        # Build for Linux only

For more detailed build instructions, see RELEASE.md
EOF
}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check for Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.25 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Go version: $GO_VERSION"
    
    # Check for libpcap
    if [ "$(uname)" = "Linux" ]; then
        if ! ldconfig -p | grep -q libpcap; then
            print_warning "libpcap may not be installed. Install with:"
            print_warning "  sudo apt-get install libpcap-dev  # Debian/Ubuntu"
            print_warning "  sudo yum install libpcap-devel    # RHEL/CentOS"
        fi
    elif [ "$(uname)" = "Darwin" ]; then
        print_info "macOS detected - libpcap is pre-installed"
    fi
    
    # Check for make
    if ! command -v make &> /dev/null; then
        print_error "make is not installed. Please install make."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Main script
main() {
    local command="${1:-build}"
    
    case "$command" in
        help|--help|-h)
            show_usage
            exit 0
            ;;
        build|linux|darwin|windows|release|clean)
            check_prerequisites
            print_info "Running: make $command"
            make "$command"
            print_success "Build complete!"
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
