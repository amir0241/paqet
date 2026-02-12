package flog

import (
	"bytes"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// TestClientStartupLogMessage tests the actual log message format used by the client
func TestClientStartupLogMessage(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetLevel(int(Info))
	time.Sleep(50 * time.Millisecond)

	// Simulate the actual client startup log message as it appears in client.go
	ipv4Addr := "217.195.200.98"
	ipv6Addr := "<nil>"
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("10.0.0.100"),
		Port: 9999,
	}
	connCount := 1

	Infof("Client started: IPv4:%s IPv6:%s -> %s (%d connections)",
		ipv4Addr, ipv6Addr, serverAddr, connCount)

	time.Sleep(100 * time.Millisecond)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Log the output for debugging (only shows with -v flag)
	t.Logf("Captured log output:\n%s", output)

	// Verify the complete log message is present
	expectedParts := []string{
		"[INFO]",
		"Client started:",
		"IPv4:217.195.200.98",
		"IPv6:<nil>",
		"-> 10.0.0.100:9999",
		"(1 connections)",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Log output missing expected part: %q\nFull output: %s", part, output)
		}
	}

	// Verify it's a single complete line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 log line, got %d lines. This indicates message truncation.\nLines: %v", len(lines), lines)
	}

	// Verify the line ends with the expected pattern
	if !strings.Contains(output, "connections)") {
		t.Error("Log line doesn't end with 'connections)' - message may be truncated")
	}
}

// TestClientStartupWithIPv6 tests the log when IPv6 is configured
func TestClientStartupWithIPv6(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetLevel(int(Info))
	time.Sleep(50 * time.Millisecond)

	// Simulate with both IPv4 and IPv6
	ipv4Addr := "192.168.1.100"
	ipv6Addr := "2001:db8::1"
	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP("10.0.0.100"),
		Port: 9999,
	}
	connCount := 4

	Infof("Client started: IPv4:%s IPv6:%s -> %s (%d connections)",
		ipv4Addr, ipv6Addr, serverAddr, connCount)

	time.Sleep(100 * time.Millisecond)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify all parts are present
	expectedParts := []string{
		"IPv4:192.168.1.100",
		"IPv6:2001:db8::1",
		"-> 10.0.0.100:9999",
		"(4 connections)",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Log output missing expected part: %q\nFull output: %s", part, output)
		}
	}

	// Verify it's a single complete line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 log line, got %d lines. This indicates message truncation.", len(lines))
	}
}
