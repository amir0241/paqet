.PHONY: all build build-all clean linux windows darwin test help release

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
	@echo "  make build-all      Build for all platforms"
	@echo "  make linux          Build for Linux (amd64 and arm64)"
	@echo "  make darwin         Build for macOS (amd64 and arm64)"
	@echo "  make windows        Build for Windows (amd64)"
	@echo "  make release        Create release archives for all platforms"
	@echo "  make clean          Remove build artifacts"
	@echo "  make test           Run tests"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"
	@echo "  GIT_TAG=$(GIT_TAG)"
	@echo "  BUILD_TIME=$(BUILD_TIME)"

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
	@mkdir -p $(BUILD_DIR)
	# macOS amd64
	@echo "Building macOS amd64..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -v -a -trimpath \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 ./cmd/main.go
	# macOS arm64
	@echo "Building macOS arm64..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -v -a -trimpath \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 ./cmd/main.go
	@echo "macOS builds complete"

windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	# Windows amd64
	@echo "Building Windows amd64..."
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -v -a -trimpath \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe ./cmd/main.go
	@echo "Windows builds complete"

build-all: linux darwin windows
	@echo "All platform builds complete"

release: build-all
	@echo "Creating release archives..."
	@mkdir -p $(DIST_DIR)
	# Linux amd64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 ]; then \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz \
			-C $(BUILD_DIR) $(BINARY_NAME)_linux_amd64 \
			-C .. README.md \
			-C example client.yaml.example server.yaml.example && \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz"; \
	fi
	# Linux arm64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 ]; then \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz \
			-C $(BUILD_DIR) $(BINARY_NAME)_linux_arm64 \
			-C .. README.md \
			-C example client.yaml.example server.yaml.example && \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz"; \
	fi
	# macOS amd64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 ]; then \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz \
			-C $(BUILD_DIR) $(BINARY_NAME)_darwin_amd64 \
			-C .. README.md \
			-C example client.yaml.example server.yaml.example && \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz"; \
	fi
	# macOS arm64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 ]; then \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz \
			-C $(BUILD_DIR) $(BINARY_NAME)_darwin_arm64 \
			-C .. README.md \
			-C example client.yaml.example server.yaml.example && \
		echo "Created: $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz"; \
	fi
	# Windows amd64
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe ]; then \
		if command -v zip >/dev/null 2>&1; then \
			cd $(BUILD_DIR) && zip ../$(DIST_DIR)/$(BINARY_NAME)-windows-amd64-$(VERSION).zip $(BINARY_NAME)_windows_amd64.exe && cd ..; \
			cd . && zip -r $(DIST_DIR)/$(BINARY_NAME)-windows-amd64-$(VERSION).zip README.md example/client.yaml.example example/server.yaml.example; \
			echo "Created: $(DIST_DIR)/$(BINARY_NAME)-windows-amd64-$(VERSION).zip"; \
		else \
			echo "Warning: zip command not found. Skipping Windows archive."; \
		fi \
	fi
	@echo "Release archives created in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "Clean complete"
