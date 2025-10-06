package state

// persistence.go handles durable storage for the cache (e.g., boltDB or JSON).
// Ensure writes are atomic so crash recovery honors the PRD.

// TODO: Implement durable storage for the file signature cache.
// This will allow the watcher to resume from a previous state without a full rescan.
// - Choose a storage format (e.g., JSON, Gob, or a key-value store like BoltDB).
// - Implement `Save(cache *Cache, path string)` and `Load(path string) (*Cache, error)`
//   functions.
// - Ensure that writes are atomic to prevent data corruption in case of a crash.
