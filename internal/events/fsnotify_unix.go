package events

// fsnotify_unix.go wraps Darwin/Linux watchers. Use build tags later to restrict
// compilation and expose a constructor used by backend.go.

// TODO: Implement the fsnotify-based watcher for Unix-like systems (Linux, Darwin).
// - Add the `+build darwin linux` build tag.
// - Create a struct that implements the `Backend` interface defined in backend.go.
// - Use a library like `github.com/fsnotify/fsnotify` to handle the underlying
//   filesystem events.
// - Translate the library-specific events into the generic `Event` type.
