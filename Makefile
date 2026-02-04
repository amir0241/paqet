.PHONY: all build build-all clean linux windows darwin test help release package

# Version information
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "v1.0.0-alpha.13")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

# Build flags
LDFLAGS = -s -w -buildid= \
	-X 'paqet/cmd/version.Version=$(VERSION)' \
	-X 'paqet/cmd/version.GitCommit=$(GIT_COMMIT)' \
	-X 'paqet/cmd/version.GitTag=$(GIT_TAG)' \
	-X 'paqet/cmd/version.BuildTime=$(BUILD_TIME)'

GCFLAGS = all=-l=4

# Directories
DIST_DIR = dist
BUILD_DIR = build

# Binary names
BINARY_NAME = paqet

all: build

help:
	@echo "Paqet Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build for current platform"
	@echo "  make build-all      Build for all platforms (platform-dependent)"
	@echo "  make linux          Build for Linux (amd64 and arm64)"
	@echo "  make darwin         Build for macOS (amd64 and arm64) - macOS host only"
	@echo "  make windows        Build for Windows (amd64) - Windows host only"
	@echo "  make package        Package existing binaries into release archives"
	@echo "  make release        Build all platforms and create release archives"
	@echo "  make clean          Remove build artifacts"
	@echo "  make test           Run tests"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"
	@echo "  GIT_TAG=$(GIT_TAG)"
	@echo "  BUILD_TIME=$(BUILD_TIME)"
	@echo ""
	@echo "Notes:"
	@echo "  - Cross-compilation with CGO has limitations"
	@echo "  - Use GitHub Actions for multi-platform releases"
	@echo "  - Build on native platform for best results"

build:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build -v -a -trimpath \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	# Linux amd64
	@echo "Building Linux amd64..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -a -trimpath \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 ./cmd/main.go
	# Linux arm64 (requires cross-compilation setup)
	@echo "Building Linux arm64 (may require cross-compilation tools)..."
	@if command -v aarch64-linux-gnu-gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -v -a -trimpath \
			-gcflags "$(GCFLAGS)" \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 ./cmd/main.go; \
	else \
		echo "Warning: aarch64-linux-gnu-gcc not found. Skipping arm64 build."; \
		echo "Install with: sudo apt-get install gcc-aarch64-linux-gnu"; \
	fi
	@echo "Linux builds complete"

darwin:
	@echo "Building for macOS..."
	@if [ "$(shell uname)" != "Darwin" ]; then \
		echo "Warning: Cross-compiling to macOS with CGO from non-macOS is not supported."; \
		echo "Skipping macOS builds. Please build on macOS for best results."; \
	else \
		mkdir -p $(BUILD_DIR); \
		echo "Building macOS amd64..."; \
		CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -v -a -trimpath \
			-gcflags "$(GCFLAGS)" \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 ./cmd/main.go; \
		echo "Building macOS arm64..."; \
		CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -v -a -trimpath \
			-gcflags "$(GCFLAGS)" \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 ./cmd/main.go; \
		echo "macOS builds complete"; \
	fi

windows:
	@echo "Building for Windows..."
	@if [ "$(shell uname)" != "MINGW"* ] && [ "$(shell uname)" != "MSYS"* ] && [ "$(shell uname)" != "CYGWIN"* ]; then \
		echo "Warning: Cross-compiling to Windows with CGO is complex."; \
		echo "Skipping Windows build. Please build on Windows or use GitHub Actions for best results."; \
	else \
		mkdir -p $(BUILD_DIR); \
		echo "Building Windows amd64..."; \
		CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -v -a -trimpath \
			-gcflags "$(GCFLAGS)" \
			-ldflags "$(LDFLAGS)" \
			-o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe ./cmd/main.go; \
		echo "Windows builds complete"; \
	fi

build-all: linux darwin windows
	@echo "All requested platform builds complete"

release: build-all package
	@echo "Release complete!"

package:
	@echo "Packaging existing binaries..."
	@mkdir -p $(DIST_DIR)
	# Linux amd64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 ]; then \
		echo "Creating Linux amd64 archive..."; \
		mkdir -p $(DIST_DIR)/tmp-linux-amd64; \
		cp $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(DIST_DIR)/tmp-linux-amd64/$(BINARY_NAME)_linux_amd64; \
		cp README.md $(DIST_DIR)/tmp-linux-amd64/; \
		cp example/client.yaml.example $(DIST_DIR)/tmp-linux-amd64/; \
		cp example/server.yaml.example $(DIST_DIR)/tmp-linux-amd64/; \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz -C $(DIST_DIR)/tmp-linux-amd64 .; \
		rm -rf $(DIST_DIR)/tmp-linux-amd64; \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz"; \
	fi
	# Linux arm64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 ]; then \
		echo "Creating Linux arm64 archive..."; \
		mkdir -p $(DIST_DIR)/tmp-linux-arm64; \
		cp $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 $(DIST_DIR)/tmp-linux-arm64/$(BINARY_NAME)_linux_arm64; \
		cp README.md $(DIST_DIR)/tmp-linux-arm64/; \
		cp example/client.yaml.example $(DIST_DIR)/tmp-linux-arm64/; \
		cp example/server.yaml.example $(DIST_DIR)/tmp-linux-arm64/; \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz -C $(DIST_DIR)/tmp-linux-arm64 .; \
		rm -rf $(DIST_DIR)/tmp-linux-arm64; \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz"; \
	fi
	# macOS amd64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 ]; then \
		echo "Creating macOS amd64 archive..."; \
		mkdir -p $(DIST_DIR)/tmp-darwin-amd64; \
		cp $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(DIST_DIR)/tmp-darwin-amd64/$(BINARY_NAME)_darwin_amd64; \
		cp README.md $(DIST_DIR)/tmp-darwin-amd64/; \
		cp example/client.yaml.example $(DIST_DIR)/tmp-darwin-amd64/; \
		cp example/server.yaml.example $(DIST_DIR)/tmp-darwin-amd64/; \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz -C $(DIST_DIR)/tmp-darwin-amd64 .; \
		rm -rf $(DIST_DIR)/tmp-darwin-amd64; \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz"; \
	fi
	# macOS arm64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 ]; then \
		echo "Creating macOS arm64 archive..."; \
		mkdir -p $(DIST_DIR)/tmp-darwin-arm64; \
		cp $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 $(DIST_DIR)/tmp-darwin-arm64/$(BINARY_NAME)_darwin_arm64; \
		cp README.md $(DIST_DIR)/tmp-darwin-arm64/; \
		cp example/client.yaml.example $(DIST_DIR)/tmp-darwin-arm64/; \
		cp example/server.yaml.example $(DIST_DIR)/tmp-darwin-arm64/; \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz -C $(DIST_DIR)/tmp-darwin-arm64 .; \
		rm -rf $(DIST_DIR)/tmp-darwin-arm64; \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz"; \
	fi
	# Windows amd64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe ]; then \
		if command -v zip >/dev/null 2>&1; then \
			echo "Creating Windows amd64 archive..."; \
			mkdir -p $(DIST_DIR)/tmp-windows-amd64; \
			cp $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(DIST_DIR)/tmp-windows-amd64/; \
			cp README.md $(DIST_DIR)/tmp-windows-amd64/; \
			cp example/client.yaml.example $(DIST_DIR)/tmp-windows-amd64/; \
			cp example/server.yaml.example $(DIST_DIR)/tmp-windows-amd64/; \
			cd $(DIST_DIR)/tmp-windows-amd64 && zip -r ../$(BINARY_NAME)-windows-amd64-$(VERSION).zip . && cd ../../..; \
			rm -rf $(DIST_DIR)/tmp-windows-amd64; \
			echo "Created: $(DIST_DIR)/$(BINARY_NAME)-windows-amd64-$(VERSION).zip"; \
		else \
			echo "Warning: zip command not found. Skipping Windows archive."; \
		fi \
	fi
	@echo ""
	@echo "Release archives created in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "Clean complete"
