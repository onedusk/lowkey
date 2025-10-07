//go:build windows

package events

import "errors"

// fsnotify_windows.go contains the Windows implementation. A future change can
// wire the ReadDirectoryChangesW backend once vendored.
func newFSNotifyBackend() (Backend, error) {
	return nil, errors.New("events: native fsnotify backend not available; using polling backend")
}
