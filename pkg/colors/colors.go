// Package colors provides ANSI color codes and utilities for terminal output.
// It enables colored output for better readability of file system change events
// and command results, matching the Ruby implementation's color scheme.
package colors

import (
	"fmt"
	"os"
)

// ANSI color codes for terminal output
const (
	Red     = "\033[0;31m"
	Green   = "\033[0;32m"
	Yellow  = "\033[0;33m"
	Blue    = "\033[0;34m"
	Magenta = "\033[0;35m"
	Reset   = "\033[0m"
)

// colorEnabled determines whether color output is enabled for the terminal.
// This can be controlled by checking if stdout is a terminal and respecting
// environment variables like NO_COLOR.
var colorEnabled = isTerminal()

// isTerminal checks if stdout is connected to a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// EnableColor forces color output on
func EnableColor() {
	colorEnabled = true
}

// DisableColor forces color output off
func DisableColor() {
	colorEnabled = false
}

// Colorize wraps text in ANSI color codes if color output is enabled
func Colorize(text, color string) string {
	if !colorEnabled {
		return text
	}
	return color + text + Reset
}

// Printf prints formatted text with color support
func Printf(color, format string, args ...interface{}) {
	fmt.Print(Colorize(fmt.Sprintf(format, args...), color))
}

// Println prints text with color support and a newline
func Println(color, text string) {
	fmt.Println(Colorize(text, color))
}

// EventColor returns the appropriate color for a given event type
func EventColor(eventType string) string {
	switch eventType {
	case "NEW", "CREATE":
		return Green
	case "MODIFIED", "MODIFY":
		return Yellow
	case "DELETED", "DELETE":
		return Red
	default:
		return Reset
	}
}

// ColorizeEventType returns a colored event type string
func ColorizeEventType(eventType string) string {
	return Colorize(eventType, EventColor(eventType))
}
