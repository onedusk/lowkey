// backend.go defines the interface to platform-specific filesystem event APIs.
// Provide constructors that choose fsnotify variants based on GOOS.

package events

import "time"

// Event represents a single file system notification.
type Event struct {
	Path      string
	Type      string
	Timestamp time.Time
}

// Backend is the interface for a platform-specific file system watcher.
// It abstracts the underlying event mechanism (e.g., fsnotify, kqueue, etc.).
// TODO: Implement a constructor `NewBackend()` that uses build tags or runtime
// checks on GOOS to initialize the appropriate platform-specific backend
// (e.g., the one from fsnotify_unix.go or fsnotify_windows.go).
type Backend interface {
	// Events returns a channel that receives file system events.
	Events() <-chan Event

	// Errors returns a channel that receives errors from the watcher.
	Errors() <-chan error

	// Add starts watching the given path for changes.
	Add(path string) error

	// Remove stops watching the given path.
	Remove(path string) error

	// Close cleans up the watcher and closes the channels.
	Close() error
}