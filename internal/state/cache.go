package state

// cache.go tracks file signatures for incremental scanning. Implement thread-safe
// read/write helpers as described in algorithm_design.md.

// TODO: Implement the file signature cache.
// This cache will store file metadata (e.g., mtime, size) to avoid redundant
// processing during incremental scans.
// - Define a `FileSignature` struct.
// - Implement a thread-safe map or a more sophisticated cache (e.g., LRU) to store
//   signatures, keyed by file path.
// - Implement `Get(path string) (*FileSignature, bool)` and
//   `Set(path string, sig *FileSignature)` methods.
// - This cache will be used by the hybrid monitor to compare file states.
