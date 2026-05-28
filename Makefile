BINARY      := paqet
MAIN        := ./cmd/main.go
VERSION     := $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
GIT_COMMIT  := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG     := $(shell git describe --tags --exact-match 2>/dev/null || echo "unknown")
BUILD_TIME  := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

LDFLAGS := -s -w -buildid= \
	-X 'paqet/cmd/version.Version=$(VERSION)' \
	-X 'paqet/cmd/version.GitCommit=$(GIT_COMMIT)' \
	-X 'paqet/cmd/version.GitTag=$(GIT_TAG)' \
	-X 'paqet/cmd/version.BuildTime=$(BUILD_TIME)'

.PHONY: all build test clean check-deps

all: build

## check-deps: verify that libpcap development headers are installed (required by gopacket/pcap).
check-deps:
	@if ! pkg-config --exists libpcap 2>/dev/null && ! [ -f /usr/include/pcap.h ] && ! [ -f /usr/local/include/pcap.h ]; then \
		echo ""; \
		echo "ERROR: libpcap development headers not found."; \
		echo ""; \
		echo "  gopacket/pcap requires libpcap to compile.  Install it with:"; \
		echo "    Debian/Ubuntu : sudo apt-get install libpcap-dev"; \
		echo "    RHEL/Fedora   : sudo yum install libpcap-devel"; \
		echo "    macOS         : xcode-select --install  (ships with Xcode CLT)"; \
		echo "    Windows       : install Npcap from https://npcap.com"; \
		echo ""; \
		exit 1; \
	fi

## build: compile the binary for the current platform.
build: check-deps
	CGO_ENABLED=1 go build -trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(BINARY) $(MAIN)

## test: run all unit tests.
test: check-deps
	go test ./...

## clean: remove the compiled binary.
clean:
	rm -f $(BINARY)
