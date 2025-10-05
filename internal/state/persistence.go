package state

// persistence.go handles durable storage for the cache (e.g., boltDB or JSON).
// Ensure writes are atomic so crash recovery honors the PRD.
