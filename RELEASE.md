# Release Guide

This document explains how to create compiled prebuilt releases for the paqet project.

## Overview

Paqet supports multiple build methods:
- **GitHub Actions (Recommended)**: Automated builds for all platforms via CI/CD
- **Local Makefile**: Build locally for development or custom releases
- **Manual Build**: Direct Go build commands

## Prerequisites

### For All Platforms
- Go 1.25 or later
- libpcap development libraries

### Platform-Specific Requirements

#### Linux
```bash
# Debian/Ubuntu
sudo apt-get install libpcap-dev

# For ARM64 cross-compilation
sudo apt-get install gcc-aarch64-linux-gnu libc6-dev-arm64-cross
```

#### macOS
```bash
# libpcap comes pre-installed with Xcode Command Line Tools
xcode-select --install
```

#### Windows
- Install [Npcap](https://npcap.com/)
- Install MinGW for CGO support: `choco install mingw`

## Method 1: GitHub Actions (Recommended)

The repository includes a comprehensive GitHub Actions workflow that builds for all platforms automatically.

### Creating a Release via GitHub Actions

1. **Tag a release:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **The workflow automatically:**
   - Builds for Linux (amd64, arm64)
   - Builds for macOS (amd64, arm64)
   - Builds for Windows (amd64)
   - Creates release archives (.tar.gz for Linux/macOS, .zip for Windows)
   - Publishes the release to GitHub Releases page

3. **Manual trigger:**
   You can also manually trigger the workflow from the GitHub Actions tab without creating a tag. This is useful for testing builds.

### Build Artifacts

Each platform build includes:
- Compiled binary (e.g., `paqet_linux_amd64`, `paqet_darwin_arm64`, `paqet_windows_amd64.exe`)
- README.md
- example/client.yaml.example
- example/server.yaml.example

## Method 2: Local Makefile

The Makefile provides convenient commands for local builds.

### Available Commands

```bash
# Show all available commands
make help

# Build for current platform only
make build

# Build for specific platforms
make linux      # Build Linux (amd64 and arm64)
make darwin     # Build macOS (amd64 and arm64)
make windows    # Build Windows (amd64)

# Build for all platforms
make build-all

# Create release archives for all platforms
make release

# Clean build artifacts
make clean

# Run tests
make test
```

### Creating a Local Release

1. **Build all platforms:**
   ```bash
   make build-all
   ```

2. **Create release archives:**
   ```bash
   make release
   ```

3. **Find artifacts:**
   - Binaries: `build/` directory
   - Archives: `dist/` directory

### Custom Version Information

You can override version information during build:

```bash
make release VERSION=v1.0.0 GIT_TAG=v1.0.0
```

## Method 3: Manual Build

For custom or one-off builds, you can use direct Go commands.

### Single Platform Build

```bash
# Build for current platform
CGO_ENABLED=1 go build -o paqet ./cmd/main.go
```

### Cross-Platform Build

#### Linux amd64
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -o paqet_linux_amd64 ./cmd/main.go
```

#### Linux arm64
```bash
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc \
  go build -o paqet_linux_arm64 ./cmd/main.go
```

#### macOS amd64
```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
  go build -o paqet_darwin_amd64 ./cmd/main.go
```

#### macOS arm64
```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
  go build -o paqet_darwin_arm64 ./cmd/main.go
```

#### Windows amd64
```bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  go build -o paqet_windows_amd64.exe ./cmd/main.go
```

### Build with Version Information

```bash
VERSION="v1.0.0"
GIT_COMMIT=$(git rev-parse HEAD)
GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

CGO_ENABLED=1 go build -v -a -trimpath \
  -gcflags "all=-l=4" \
  -ldflags "-s -w -buildid= \
    -X 'paqet/cmd/version.Version=${VERSION}' \
    -X 'paqet/cmd/version.GitCommit=${GIT_COMMIT}' \
    -X 'paqet/cmd/version.GitTag=${GIT_TAG}' \
    -X 'paqet/cmd/version.BuildTime=${BUILD_TIME}'" \
  -o paqet ./cmd/main.go
```

## Release Checklist

Before creating a release, ensure:

- [ ] All tests pass: `make test` or `go test ./...`
- [ ] Version number is updated in `cmd/version/version.go` (if not using tag-based versioning)
- [ ] CHANGELOG is updated (if applicable)
- [ ] Documentation is up to date
- [ ] Example configuration files are current

## Testing Releases

After building, test the binary:

```bash
# Check version
./paqet version

# Verify it runs
./paqet --help

# Test commands
./paqet secret
./paqet iface
```

## Distribution

### GitHub Releases
When using GitHub Actions, releases are automatically published to the repository's Releases page.

### Manual Distribution
If distributing manually:

1. Build with `make release`
2. Upload archives from `dist/` directory to your distribution platform
3. Provide checksums for verification:
   ```bash
   cd dist
   sha256sum * > checksums.txt
   ```

## Versioning

Paqet follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality (backwards-compatible)
- **PATCH** version for bug fixes (backwards-compatible)

Development versions use format: `v1.0.0-alpha.X` or `v1.0.0-beta.X`

## Troubleshooting

### Cross-Compilation Issues

**Linux ARM64 on non-ARM systems:**
```bash
# Install cross-compiler
sudo apt-get install gcc-aarch64-linux-gnu
```

**macOS Cross-Compilation:**
Cross-compiling between macOS architectures typically works without additional tools when building on macOS.

**Windows on Linux/macOS:**
Cross-compiling to Windows requires MinGW or a similar cross-compiler:
```bash
# On Ubuntu/Debian
sudo apt-get install mingw-w64
```

### CGO Issues

If you encounter CGO-related errors:
1. Ensure libpcap development libraries are installed
2. Verify CGO_ENABLED=1 is set
3. For cross-compilation, ensure the correct C compiler is set via CC environment variable

### Build Flag Optimization

The default build flags optimize for:
- **Small binary size**: `-s -w` (strip debug info)
- **Security**: `-buildid=` (reproducible builds)
- **Performance**: `-gcflags "all=-l=4"` (aggressive inlining)

To include debug information for development:
```bash
go build -o paqet ./cmd/main.go
```

## Support

For build-related issues:
1. Check this guide
2. Review GitHub Actions workflow logs
3. Open an issue on GitHub with:
   - Your platform and Go version
   - Complete error output
   - Build command used
