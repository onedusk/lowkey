// Package logging provides a flexible logging framework for the lowkey daemon.
// It includes support for log rotation based on size and a structured logging
// wrapper for consistent log message formatting.
//
// The package is designed to be thread-safe and can be used to log messages
// from multiple goroutines concurrently.
package logging

import (
	"fmt"
	"log"
)

// Logger provides a simple, structured logging interface. It wraps the standard
// `log.Logger` to offer leveled logging methods (e.g., Info, Error) with a
// consistent format.
type Logger struct {
	base *log.Logger
}

// New constructs a new Logger that writes to the provided rotator. It sets up a
// standard `log.Logger` with the rotator and wraps it in the structured Logger.
func New(rotator *Rotator) *Logger {
	return &Logger{base: NewLogger(rotator)}
}

// Info logs an informational message. The message is prefixed with "INFO".
func (l *Logger) Info(msg string) {
	l.base.Println("INFO", msg)
}

// Infof logs a formatted informational message. The message is prefixed with "INFO".
func (l *Logger) Infof(format string, args ...interface{}) {
	l.base.Println("INFO", fmt.Sprintf(format, args...))
}

// Error logs an error message along with the underlying error. The message is
// prefixed with "ERROR".
func (l *Logger) Error(err error, msg string) {
	l.base.Println("ERROR", msg, "err=", err)
}

// Errorf logs a formatted error message. The message is prefixed with "ERROR".
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.base.Println("ERROR", fmt.Sprintf(format, args...))
}
