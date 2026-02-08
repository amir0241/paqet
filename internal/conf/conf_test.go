package conf

import (
	"os"
	"testing"
)

// TestPerformancePointerSetup verifies that the Performance pointer is correctly
// set in the Network configuration after loading from file.
func TestPerformancePointerSetup(t *testing.T) {
	// Create a temporary config file with performance settings
	configContent := `role: "client"

log:
  level: "info"

socks5:
  - listen: "127.0.0.1:1080"

network:
  interface: "lo"
  ipv4:
    addr: "127.0.0.1:0"
    router_mac: "00:00:00:00:00:00"

server:
  addr: "127.0.0.1:9999"

transport:
  protocol: "kcp"
  kcp:
    key: "test-key-here"

performance:
  max_concurrent_streams: 5000
  packet_workers: 4
`

	// Create temporary file
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load configuration
	cfg, err := LoadFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test 1: Performance pointer should not be nil
	if cfg.Network.Performance == nil {
		t.Fatal("Network.Performance pointer is nil")
	}

	// Test 2: Performance pointer should point to the same object as cfg.Performance
	if cfg.Network.Performance != &cfg.Performance {
		t.Fatal("Network.Performance does not point to cfg.Performance")
	}

	// Test 3: Verify values are accessible through the pointer
	if cfg.Network.Performance.MaxConcurrentStreams != 5000 {
		t.Errorf("Expected MaxConcurrentStreams=5000, got %d", cfg.Network.Performance.MaxConcurrentStreams)
	}

	if cfg.Network.Performance.PacketWorkers != 4 {
		t.Errorf("Expected PacketWorkers=4, got %d", cfg.Network.Performance.PacketWorkers)
	}

	// Test 4: Defaults should have been applied to unset fields
	if cfg.Network.Performance.StreamWorkerPoolSize != 1000 {
		t.Errorf("Expected default StreamWorkerPoolSize=1000, got %d", cfg.Network.Performance.StreamWorkerPoolSize)
	}
}

// TestPerformancePointerWithDefaults verifies that Performance pointer works
// correctly when no performance section is specified in the config.
func TestPerformancePointerWithDefaults(t *testing.T) {
	// Create a temporary config file WITHOUT performance settings
	configContent := `role: "client"

log:
  level: "info"

socks5:
  - listen: "127.0.0.1:1080"

network:
  interface: "lo"
  ipv4:
    addr: "127.0.0.1:0"
    router_mac: "00:00:00:00:00:00"

server:
  addr: "127.0.0.1:9999"

transport:
  protocol: "kcp"
  kcp:
    key: "test-key-here"
`

	// Create temporary file
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load configuration
	cfg, err := LoadFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test 1: Performance pointer should not be nil (even without explicit config)
	if cfg.Network.Performance == nil {
		t.Fatal("Network.Performance pointer is nil")
	}

	// Test 2: Performance pointer should point to the same object as cfg.Performance
	if cfg.Network.Performance != &cfg.Performance {
		t.Fatal("Network.Performance does not point to cfg.Performance")
	}

	// Test 3: Defaults should have been applied
	// For client, default MaxConcurrentStreams should be 5000
	if cfg.Network.Performance.MaxConcurrentStreams != 5000 {
		t.Errorf("Expected default MaxConcurrentStreams=5000 for client, got %d", cfg.Network.Performance.MaxConcurrentStreams)
	}
}
