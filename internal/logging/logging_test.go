package logging

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".craizy")

	// Reset the once for testing
	once = sync.Once{}
	defaultLogger = nil

	err := Init(logDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// Check that log directory was created
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Log directory was not created")
	}

	// Check that log file was created
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, today+".log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestEntry(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".craizy")

	// Reset for testing
	once = sync.Once{}
	defaultLogger = nil

	if err := Init(logDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Entry("arg1", "arg2", 123)

	// Flush and read the log
	Close()

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, today+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[INFO]") {
		t.Error("Log entry missing [INFO] level")
	}
	if !strings.Contains(logContent, "ENTRY") {
		t.Error("Log entry missing ENTRY marker")
	}
	if !strings.Contains(logContent, "arg1") {
		t.Error("Log entry missing argument")
	}
}

func TestError(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".craizy")

	// Reset for testing
	once = sync.Once{}
	defaultLogger = nil

	if err := Init(logDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	testErr := &testError{msg: "test error message"}
	Error(testErr, "context1", "context2")

	// Flush and read the log
	Close()

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, today+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[ERROR]") {
		t.Error("Log entry missing [ERROR] level")
	}
	if !strings.Contains(logContent, "test error message") {
		t.Error("Log entry missing error message")
	}
	if !strings.Contains(logContent, "context1") {
		t.Error("Log entry missing context")
	}
}

func TestInfo(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".craizy")

	// Reset for testing
	once = sync.Once{}
	defaultLogger = nil

	if err := Init(logDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Info("test message %d", 42)

	// Flush and read the log
	Close()

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, today+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "[INFO]") {
		t.Error("Log entry missing [INFO] level")
	}
	if !strings.Contains(logContent, "test message 42") {
		t.Error("Log entry missing formatted message")
	}
}

func TestDisable(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".craizy")

	// Reset for testing
	once = sync.Once{}
	defaultLogger = nil

	if err := Init(logDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	// Disable logging
	Disable()
	Info("this should not appear")

	// Re-enable
	Enable()
	Info("this should appear")

	// Flush and read the log
	Close()

	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, today+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if strings.Contains(logContent, "this should not appear") {
		t.Error("Disabled logging still wrote to log")
	}
	if !strings.Contains(logContent, "this should appear") {
		t.Error("Enabled logging did not write to log")
	}
}

func TestAppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, ".craizy")

	// Reset for testing
	once = sync.Once{}
	defaultLogger = nil

	if err := Init(logDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	Info("first message")
	Close()

	// Re-initialize (simulating app restart)
	once = sync.Once{}
	defaultLogger = nil

	if err := Init(logDir); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	Info("second message")
	Close()

	// Read and verify both messages exist
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, today+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "first message") {
		t.Error("First message missing - file was not appended")
	}
	if !strings.Contains(logContent, "second message") {
		t.Error("Second message missing")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
