package flog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestLoggingUnderLoad verifies that log messages are not dropped under high load
func TestLoggingUnderLoad(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set log level to Info
	SetLevel(int(Info))

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Send many log messages concurrently
	const numMessages = 100
	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				Infof("Test message from goroutine %d: %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	// Give time for all messages to be processed
	time.Sleep(200 * time.Millisecond)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Count how many messages were logged
	lines := strings.Split(strings.TrimSpace(output), "\n")
	messageCount := 0
	for _, line := range lines {
		if strings.Contains(line, "Test message from goroutine") {
			messageCount++
		}
	}

	expectedMessages := numMessages * numGoroutines
	// We expect all or nearly all messages (allow for very small loss due to timing)
	if messageCount < expectedMessages-5 {
		t.Errorf("Expected at least %d messages, but got %d (loss: %d)", 
			expectedMessages-5, messageCount, expectedMessages-messageCount)
	}
}

// TestLogLevels verifies that different log levels work correctly
func TestLogLevels(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set log level to Warn (should not log Debug or Info)
	SetLevel(int(Warn))
	time.Sleep(50 * time.Millisecond)

	Debugf("Debug message")
	Infof("Info message")
	Warnf("Warn message")
	Errorf("Error message")

	time.Sleep(100 * time.Millisecond)

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Debug and Info should not appear
	if strings.Contains(output, "Debug message") {
		t.Error("Debug message should not be logged when level is Warn")
	}
	if strings.Contains(output, "Info message") {
		t.Error("Info message should not be logged when level is Warn")
	}

	// Warn and Error should appear
	if !strings.Contains(output, "Warn message") {
		t.Error("Warn message should be logged when level is Warn")
	}
	if !strings.Contains(output, "Error message") {
		t.Error("Error message should be logged when level is Warn")
	}
}

// TestCompleteLogMessage verifies that log messages are not truncated
func TestCompleteLogMessage(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetLevel(int(Info))
	time.Sleep(50 * time.Millisecond)

	// Simulate the client startup log message
	ipv4Addr := "192.168.1.100"
	ipv6Addr := "<nil>"
	serverAddr := "10.0.0.100:9999"
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

	// Verify all parts of the message are present
	requiredParts := []string{
		"Client started:",
		fmt.Sprintf("IPv4:%s", ipv4Addr),
		fmt.Sprintf("IPv6:%s", ipv6Addr),
		fmt.Sprintf("-> %s", serverAddr),
		fmt.Sprintf("(%d connections)", connCount),
	}

	for _, part := range requiredParts {
		if !strings.Contains(output, part) {
			t.Errorf("Log output missing expected part: %q\nFull output: %s", part, output)
		}
	}

	// Verify the message is on a single line (not truncated)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 log line, got %d lines:\n%s", len(lines), output)
	}
}
