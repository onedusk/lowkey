//go:build windows

// Package events provides a platform-agnostic interface for file system
// notifications. It abstracts the underlying mechanism for watching file and
// directory changes, allowing the application to consume a consistent stream of
// events regardless of the operating system.
//
// The package defines a Backend interface, which can be implemented by different
// watchers (e.g., inotify, kqueue, polling). A polling-based backend is
// provided as a universal fallback.
package events

import "errors"

// newFSNotifyBackend is a placeholder for a native fsnotify-based backend on
// Windows. It currently returns an error, indicating that the native
// implementation is not yet available and the polling backend should be used.
func newFSNotifyBackend() (Backend, error) {
	return nil, errors.New("events: native fsnotify backend not available; using polling backend")
}
