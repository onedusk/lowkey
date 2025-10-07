//go:build darwin || linux

package events

import "errors"

// fsnotify_unix.go wraps Darwin/Linux watchers. A native implementation can
// replace the polling backend once fsnotify is vendored.
func newFSNotifyBackend() (Backend, error) {
	return nil, errors.New("events: native fsnotify backend not available; using polling backend")
}
