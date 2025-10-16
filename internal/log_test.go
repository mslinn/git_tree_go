package internal

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestLogger_Initialization tests logger initialization
func TestLogger_Initialization(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	if logger.verbosity != LogNormal {
		t.Errorf("Expected default verbosity to be LogNormal (%d), got %d", LogNormal, logger.verbosity)
	}

	if logger.queue == nil {
		t.Error("Expected queue to be initialized")
	}

	if logger.stdaux == nil {
		t.Error("Expected stdaux to be initialized")
	}

	// Clean up
	logger.Shutdown()
}

// TestLogger_SetVerbosity tests setting verbosity level
func TestLogger_SetVerbosity(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	testLevels := []int{LogQuiet, LogNormal, LogVerbose, LogDebug}

	for _, level := range testLevels {
		logger.SetVerbosity(level)
		if logger.GetVerbosity() != level {
			t.Errorf("Expected verbosity to be %d, got %d", level, logger.GetVerbosity())
		}
	}
}

// TestLogger_GetVerbosity tests getting verbosity level
func TestLogger_GetVerbosity(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Default should be LogNormal
	if logger.GetVerbosity() != LogNormal {
		t.Errorf("Expected default verbosity to be LogNormal (%d), got %d", LogNormal, logger.GetVerbosity())
	}

	// Set and get
	logger.SetVerbosity(LogVerbose)
	if logger.GetVerbosity() != LogVerbose {
		t.Errorf("Expected verbosity to be LogVerbose (%d), got %d", LogVerbose, logger.GetVerbosity())
	}
}

// TestLogger_Log_WithQuietVerbosity tests that quiet verbosity suppresses logs
func TestLogger_Log_WithQuietVerbosity(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	// Set to quiet
	logger.SetVerbosity(LogQuiet)

	// Try to log at normal level
	logger.Log(LogNormal, "This should not appear", "")

	// Give time for the queue to process
	time.Sleep(50 * time.Millisecond)

	// Buffer should be empty
	if buf.Len() > 0 {
		t.Errorf("Expected no output in quiet mode, got: %s", buf.String())
	}
}

// TestLogger_Log_WithNormalVerbosity tests normal verbosity logging
func TestLogger_Log_WithNormalVerbosity(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	// Set to normal
	logger.SetVerbosity(LogNormal)

	message := "Test message at normal level"
	logger.Log(LogNormal, message, "")

	// Give time for the queue to process
	time.Sleep(50 * time.Millisecond)

	// Buffer should contain the message
	if !strings.Contains(buf.String(), message) {
		t.Errorf("Expected output to contain '%s', got: %s", message, buf.String())
	}
}

// TestLogger_Log_WithVerboseVerbosity tests verbose logging
func TestLogger_Log_WithVerboseVerbosity(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	// Set to verbose
	logger.SetVerbosity(LogVerbose)

	message := "Verbose message"
	logger.Log(LogVerbose, message, "")

	// Give time for the queue to process
	time.Sleep(50 * time.Millisecond)

	// Buffer should contain the message
	if !strings.Contains(buf.String(), message) {
		t.Errorf("Expected output to contain '%s', got: %s", message, buf.String())
	}
}

// TestLogger_Log_WithDebugVerbosity tests debug logging
func TestLogger_Log_WithDebugVerbosity(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	// Set to debug
	logger.SetVerbosity(LogDebug)

	message := "Debug message"
	logger.Log(LogDebug, message, "")

	// Give time for the queue to process
	time.Sleep(50 * time.Millisecond)

	// Buffer should contain the message
	if !strings.Contains(buf.String(), message) {
		t.Errorf("Expected output to contain '%s', got: %s", message, buf.String())
	}
}

// TestLogger_Log_WithColor tests logging with color
func TestLogger_Log_WithColor(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	logger.SetVerbosity(LogNormal)

	message := "Colored message"
	logger.Log(LogNormal, message, "green")

	// Give time for the queue to process
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Should contain the message
	if !strings.Contains(output, message) {
		t.Errorf("Expected output to contain '%s', got: %s", message, output)
	}
}

// TestLogger_Log_MultilineMessage tests logging multiline messages
func TestLogger_Log_MultilineMessage(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	logger.SetVerbosity(LogNormal)

	message := "Line 1\nLine 2\nLine 3"
	logger.Log(LogNormal, message, "")

	// Give time for the queue to process
	time.Sleep(50 * time.Millisecond)

	output := buf.String()

	// Should contain all lines
	if !strings.Contains(output, "Line 1") {
		t.Error("Expected output to contain 'Line 1'")
	}
	if !strings.Contains(output, "Line 2") {
		t.Error("Expected output to contain 'Line 2'")
	}
	if !strings.Contains(output, "Line 3") {
		t.Error("Expected output to contain 'Line 3'")
	}
}

// TestLogger_Log_EmptyMessage tests logging empty messages
func TestLogger_Log_EmptyMessage(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	logger.SetVerbosity(LogNormal)

	logger.Log(LogNormal, "", "")

	// Give time for the queue to process
	time.Sleep(50 * time.Millisecond)

	// Should handle empty messages gracefully
	// (implementation may or may not output anything)
}

// TestLogger_ConcurrentLogging tests thread-safe concurrent logging
func TestLogger_ConcurrentLogging(t *testing.T) {
	logger := NewLogger()
	defer logger.Shutdown()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	logger.SetVerbosity(LogNormal)

	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Log(LogNormal, "Message", "")
			}
		}(i)
	}

	wg.Wait()

	// Give time for the queue to process
	time.Sleep(100 * time.Millisecond)

	// Should have logged all messages without crashing
	output := buf.String()
	messageCount := strings.Count(output, "Message")
	expected := numGoroutines * messagesPerGoroutine

	if messageCount != expected {
		t.Errorf("Expected %d messages, got %d", expected, messageCount)
	}
}

// TestLogger_Shutdown tests graceful shutdown
func TestLogger_Shutdown(t *testing.T) {
	logger := NewLogger()

	// Capture output
	var buf bytes.Buffer
	logger.stdaux = &buf

	logger.SetVerbosity(LogNormal)

	// Log some messages
	logger.Log(LogNormal, "Message 1", "")
	logger.Log(LogNormal, "Message 2", "")

	// Shutdown and wait
	logger.Shutdown()

	// All messages should be processed
	output := buf.String()
	if !strings.Contains(output, "Message 1") || !strings.Contains(output, "Message 2") {
		t.Error("Expected all messages to be processed before shutdown")
	}
}

// TestGetLogger_Singleton tests that GetLogger returns singleton
func TestGetLogger_Singleton(t *testing.T) {
	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("Expected GetLogger to return the same instance")
	}
}

// TestConvenienceFunctions tests the global convenience functions
func TestConvenienceFunctions_SetGetVerbosity(t *testing.T) {
	// Save original
	original := GetVerbosity()
	defer SetVerbosity(original)

	SetVerbosity(LogQuiet)
	if GetVerbosity() != LogQuiet {
		t.Errorf("Expected verbosity to be LogQuiet (%d), got %d", LogQuiet, GetVerbosity())
	}

	SetVerbosity(LogVerbose)
	if GetVerbosity() != LogVerbose {
		t.Errorf("Expected verbosity to be LogVerbose (%d), got %d", LogVerbose, GetVerbosity())
	}
}

// TestConvenienceFunctions_Log tests the global Log function
func TestConvenienceFunctions_Log(t *testing.T) {
	// Save original
	original := GetVerbosity()
	defer SetVerbosity(original)

	logger := GetLogger()
	var buf bytes.Buffer
	logger.stdaux = &buf

	SetVerbosity(LogNormal)
	Log(LogNormal, "Test message", "")

	// Give time for processing
	time.Sleep(50 * time.Millisecond)

	if !strings.Contains(buf.String(), "Test message") {
		t.Error("Expected Log function to work")
	}
}
