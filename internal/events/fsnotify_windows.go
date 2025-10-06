package events

// fsnotify_windows.go contains the Windows implementation. Mirror the Unix API
// so backend.go can return a consistent interface across GOOS values.

// TODO: Implement the fsnotify-based watcher for Windows.
// - Add the `+build windows` build tag.
// - Create a struct that implements the `Backend` interface defined in backend.go.
// - Use a library like `github.com/fsnotify/fsnotify` to handle the underlying
//   filesystem events.
// - Ensure the implementation is consistent with the Unix version.
