// Package logging provides application-wide logging functionality.
// Logs are written to .craizy/YYYY-MM-DD.log in append mode.
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Logger writes log entries to a date-based log file.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	logDir   string
	curDate  string
	disabled bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger with the given log directory.
// The directory will be created if it doesn't exist.
// Logs are written to {logDir}/YYYY-MM-DD.log
func Init(logDir string) error {
	var initErr error
	once.Do(func() {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}
		defaultLogger = &Logger{
			logDir: logDir,
		}
		initErr = defaultLogger.rotateIfNeeded()
	})
	return initErr
}

// Close closes the default logger's file handle.
func Close() {
	if defaultLogger != nil {
		defaultLogger.Close()
	}
}

// Disable disables all logging output.
func Disable() {
	if defaultLogger != nil {
		defaultLogger.mu.Lock()
		defaultLogger.disabled = true
		defaultLogger.mu.Unlock()
	}
}

// Enable enables logging output.
func Enable() {
	if defaultLogger != nil {
		defaultLogger.mu.Lock()
		defaultLogger.disabled = false
		defaultLogger.mu.Unlock()
	}
}

// SetOutput sets the output writer for testing purposes.
func SetOutput(w io.Writer) {
	// This is a no-op for file-based logging but allows testing
}

// Entry logs a method entry with the function name and arguments.
func Entry(args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.entry(args...)
}

// Error logs an error with the function name and error details.
func Error(err error, context ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.error(err, context...)
}

// Info logs an informational message.
func Info(msg string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.info(msg, args...)
}

// Debug logs a debug message.
func Debug(msg string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.debug(msg, args...)
}

// Close closes the logger's file handle.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}

// rotateIfNeeded checks if the log file needs to be rotated to a new day.
func (l *Logger) rotateIfNeeded() error {
	today := time.Now().Format("2006-01-02")
	if l.curDate == today && l.file != nil {
		return nil
	}

	// Close old file if open
	if l.file != nil {
		l.file.Close()
	}

	// Open new file for today
	filename := filepath.Join(l.logDir, today+".log")
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	l.curDate = today
	return nil
}

// write writes a log entry to the file.
func (l *Logger) write(level, funcName, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.disabled {
		return
	}

	if err := l.rotateIfNeeded(); err != nil {
		return // Silently fail if we can't rotate
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	entry := fmt.Sprintf("%s [%s] %s: %s\n", timestamp, level, funcName, message)
	l.file.WriteString(entry)
}

// getCallerFunc returns the name of the calling function (2 levels up).
func getCallerFunc(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	name := fn.Name()
	// Extract just the function name from the full path
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

func (l *Logger) entry(args ...interface{}) {
	funcName := getCallerFunc(3)
	var msg string
	if len(args) > 0 {
		parts := make([]string, len(args))
		for i, arg := range args {
			parts[i] = fmt.Sprintf("%v", arg)
		}
		msg = "ENTRY " + strings.Join(parts, ", ")
	} else {
		msg = "ENTRY"
	}
	l.write("INFO", funcName, msg)
}

func (l *Logger) error(err error, context ...interface{}) {
	funcName := getCallerFunc(3)
	var msg string
	if len(context) > 0 {
		parts := make([]string, len(context))
		for i, arg := range context {
			parts[i] = fmt.Sprintf("%v", arg)
		}
		msg = fmt.Sprintf("ERROR %v | context: %s", err, strings.Join(parts, ", "))
	} else {
		msg = fmt.Sprintf("ERROR %v", err)
	}
	l.write("ERROR", funcName, msg)
}

func (l *Logger) info(format string, args ...interface{}) {
	funcName := getCallerFunc(3)
	msg := fmt.Sprintf(format, args...)
	l.write("INFO", funcName, msg)
}

func (l *Logger) debug(format string, args ...interface{}) {
	funcName := getCallerFunc(3)
	msg := fmt.Sprintf(format, args...)
	l.write("DEBUG", funcName, msg)
}
