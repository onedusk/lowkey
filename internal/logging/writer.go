package logging

import (
	"fmt"
	"log"
)

// Logger wraps *log.Logger to expose structured helpers.
type Logger struct {
	base *log.Logger
}

// New constructs a Logger using the provided rotator.
func New(rotator *Rotator) *Logger {
	return &Logger{base: NewLogger(rotator)}
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	l.base.Println("INFO", msg)
}

// Infof formats and logs an informational message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.base.Println("INFO", fmt.Sprintf(format, args...))
}

// Error logs an error message.
func (l *Logger) Error(err error, msg string) {
	l.base.Println("ERROR", msg, "err=", err)
}

// Errorf formats and logs an error message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.base.Println("ERROR", fmt.Sprintf(format, args...))
}
