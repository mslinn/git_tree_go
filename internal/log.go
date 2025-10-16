package internal

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// Verbosity levels
const (
	LogQuiet   = 0
	LogNormal  = 1
	LogVerbose = 2
	LogDebug   = 3
)

// ANSI color codes
const (
	ColorReset  = ""
	ColorRed    = "red"
	ColorGreen  = "green"
	ColorYellow = "yellow"
	ColorCyan   = "cyan"
)

// Logger provides thread-safe logging with verbosity control and color support.
type Logger struct {
	verbosity int
	queue     chan string
	wg        sync.WaitGroup
	mu        sync.Mutex
	stdaux    io.Writer // Typically stderr
	closed    bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// GetLogger returns the singleton logger instance.
func GetLogger() *Logger {
	once.Do(func() {
		defaultLogger = NewLogger()
	})
	return defaultLogger
}

// NewLogger creates a new Logger instance.
func NewLogger() *Logger {
	l := &Logger{
		verbosity: LogNormal,
		queue:     make(chan string, 100),
		stdaux:    os.Stderr,
	}
	l.wg.Add(1)
	go l.processQueue()
	return l
}

// SetVerbosity sets the verbosity level.
func (l *Logger) SetVerbosity(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbosity = level
}

// GetVerbosity returns the current verbosity level.
func (l *Logger) GetVerbosity() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.verbosity
}

// Log logs a message with the specified verbosity level and optional color.
func (l *Logger) Log(level int, message string, colorName string) {
	if l.GetVerbosity() < level {
		return
	}

	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if line == "" && len(lines) > 1 {
			continue
		}
		var coloredLine string
		switch colorName {
		case ColorRed:
			coloredLine = color.RedString(line)
		case ColorGreen:
			coloredLine = color.GreenString(line)
		case ColorYellow:
			coloredLine = color.YellowString(line)
		case ColorCyan:
			coloredLine = color.CyanString(line)
		default:
			coloredLine = line
		}
		l.queue <- coloredLine
	}
}

// LogStdout logs a message directly to stdout (for pipes).
func (l *Logger) LogStdout(message string) {
	fmt.Fprintln(os.Stdout, message)
}

// Shutdown gracefully shuts down the logger.
func (l *Logger) Shutdown() {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return
	}
	l.closed = true
	l.mu.Unlock()
	close(l.queue)
	l.wg.Wait()
}

// processQueue processes the log queue in a separate goroutine.
func (l *Logger) processQueue() {
	defer l.wg.Done()
	for message := range l.queue {
		fmt.Fprintln(l.stdaux, message)
	}
}

// Convenience functions using the default logger

// SetVerbosity sets the verbosity level on the default logger.
func SetVerbosity(level int) {
	GetLogger().SetVerbosity(level)
}

// GetVerbosity returns the current verbosity level from the default logger.
func GetVerbosity() int {
	return GetLogger().GetVerbosity()
}

// Log logs a message using the default logger.
func Log(level int, message string, color string) {
	GetLogger().Log(level, message, color)
}

// LogStdout logs to stdout using the default logger.
func LogStdout(message string) {
	GetLogger().LogStdout(message)
}

// ShutdownLogger shuts down the default logger.
func ShutdownLogger() {
	if defaultLogger != nil {
		defaultLogger.Shutdown()
	}
}

// ResetLogger resets the default logger (primarily for testing).
func ResetLogger() {
	if defaultLogger != nil {
		defaultLogger.Shutdown()
	}
	defaultLogger = nil
	once = sync.Once{}
}
