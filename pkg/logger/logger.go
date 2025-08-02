package logger

import (
	"log"
	"os"

	"github.com/fatih/color"
)

// ====================================================================================================
// DATA STRUCTURES AND GLOBAL VARIABLES
// ====================================================================================================

// Logger provides a simple, color-coded logging interface.
// It wraps standard log.Logger instances for different log levels.
type Logger struct {
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
	debug *log.Logger
}

// Log is the global logger instance used throughout the application.
// This provides a single, easy-to-use logging interface.
var Log *Logger

// debugEnabled is a flag to control whether debug logs are printed.
// It can be toggled via the SetDebug function.
var debugEnabled bool

// ====================================================================================================
// INITIALIZATION
// ====================================================================================================

// init sets up the global logger instance with color-coded output.
// This function is automatically called by the Go runtime at startup.
func init() {
	Log = NewLogger(os.Stdout)
}

// NewLogger creates a new Logger instance.
// It configures different loggers for info, warn, error, and debug levels,
// each with its own color-coded prefix.
func NewLogger(out *os.File) *Logger {
	// Define color-coded sprint functions for each log level.
	info := color.New(color.FgGreen).SprintFunc()
	warn := color.New(color.FgYellow).SprintFunc()
	err := color.New(color.FgRed, color.Bold).SprintFunc() // Errors are bold red for emphasis.
	debug := color.New(color.FgHiBlack).SprintFunc()       // Debug logs are a subtle, high-intensity black.

	return &Logger{
		// Info logs are prefixed with "INFO:  " and are green.
		info: log.New(out, info("INFO:  "), log.Ldate|log.Ltime),
		// Warn logs are prefixed with "WARN:  " and are yellow.
		warn: log.New(out, warn("WARN:  "), log.Ldate|log.Ltime),
		// Error logs are prefixed with "ERROR: " and are bold red.
		err: log.New(out, err("ERROR: "), log.Ldate|log.Ltime|log.Lshortfile),
		// Debug logs are prefixed with "DEBUG: " and are high-intensity black.
		debug: log.New(out, debug("DEBUG: "), log.Ldate|log.Ltime),
	}
}

// ====================================================================================================
// PUBLIC METHODS
// ====================================================================================================

// SetDebug enables or disables debug logging.
// This function is typically called based on a command-line flag.
func SetDebug(enabled bool) {
	debugEnabled = enabled
}

// Info logs an informational message.
func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

// Infof logs a formatted informational message.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// Warn logs a warning message.
func (l *Logger) Warn(v ...interface{}) {
	l.warn.Println(v...)
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.warn.Printf(format, v...)
}

// Error logs an error message.
func (l *Logger) Error(v ...interface{}) {
	l.err.Println(v...)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.err.Printf(format, v...)
}

// Debug logs a debug message.
// The message is only printed if debug logging is enabled.
func (l *Logger) Debug(v ...interface{}) {
	if debugEnabled {
		l.debug.Println(v...)
	}
}

// Debugf logs a formatted debug message.
// The message is only printed if debug logging is enabled.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if debugEnabled {
		l.debug.Printf(format, v...)
	}
}
