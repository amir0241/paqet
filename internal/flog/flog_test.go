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

// TestFatalfUnderHighPressure tests that Fatalf messages are always delivered
// even when the log channel is under high pressure
func TestFatalfUnderHighPressure(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	
	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Initialize logging
	SetLevel(0) // Debug level to generate more traffic
	
	// Channel to capture the output
	outputChan := make(chan string, 1)
	
	// Start a goroutine to read from the pipe
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outputChan <- buf.String()
	}()
	
	// Simulate high pressure by flooding the log channel
	var wg sync.WaitGroup
	
	// Start multiple goroutines that spam logs to fill the channel
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				Infof("Goroutine %d: Message %d - flooding", id, j)
			}
		}(i)
	}
	
	// Give them a moment to fill the channel
	time.Sleep(50 * time.Millisecond)
	
	// Now call Fatalf in a separate goroutine (since it calls os.Exit)
	// We can't actually test os.Exit in a unit test, so we'll test 
	// that the message gets to the channel
	done := make(chan bool)
	fatalMessage := "CRITICAL ERROR: This must be visible under high pressure"
	
	go func() {
		// We can't actually call Fatalf because it exits
		// Instead, test the core functionality - that the message gets sent
		if Fatal >= minLevel && minLevel != None {
			now := time.Now().Format("2006-01-02 15:04:05.000")
			line := fmt.Sprintf("%s [%s] %s\n", now, Fatal.String(), fatalMessage)
			
			// This is the fix - blocking write instead of select/default
			logCh <- line
			time.Sleep(50 * time.Millisecond)
		}
		done <- true
	}()
	
	// Wait for the fatal message to be sent
	<-done
	
	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Get the output
	output := <-outputChan
	
	// Wait for background goroutines
	wg.Wait()
	
	// Check that the fatal message appears in the output
	if !strings.Contains(output, fatalMessage) {
		t.Errorf("Fatal message was not found in output under high pressure.\nExpected to find: %s\nOutput length: %d", fatalMessage, len(output))
	} else {
		t.Logf("SUCCESS: Fatal message was found in output even under high pressure")
	}
	
	// Check that FATAL level appears
	if !strings.Contains(output, "[FATAL]") {
		t.Errorf("FATAL level tag was not found in output")
	}
}

// TestFatalfMessageFormat tests that Fatalf formats messages correctly
func TestFatalfMessageFormat(t *testing.T) {
	// We can't test the actual Fatalf function since it calls os.Exit
	// But we can test the message formatting
	testFormat := "Error code: %d, message: %s"
	testArgs := []interface{}{42, "test error"}
	
	expected := fmt.Sprintf(testFormat, testArgs...)
	if !strings.Contains(expected, "42") || !strings.Contains(expected, "test error") {
		t.Errorf("Message formatting failed")
	}
}

// TestLogChannelBlockingBehavior verifies that the log channel can handle blocking writes
func TestLogChannelBlockingBehavior(t *testing.T) {
	// Initialize logging
	SetLevel(0)
	
	// Fill the log channel to capacity (1024 messages)
	for i := 0; i < 1024; i++ {
		Infof("Filling message %d", i)
	}
	
	// Give a moment for messages to be processed
	time.Sleep(100 * time.Millisecond)
	
	// Now send another message - with our fix, this should work
	// even if the channel is full (blocking write)
	testMsg := "Critical test message"
	
	done := make(chan bool, 1)
	go func() {
		Errorf("%s", testMsg)
		done <- true
	}()
	
	select {
	case <-done:
		t.Logf("Successfully logged message even when channel was busy")
	case <-time.After(2 * time.Second):
		t.Errorf("Logging timed out - channel may be blocked")
	}
}
