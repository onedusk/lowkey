package daemon

// manager.go coordinates daemon lifecycle: spawning watchers, persisting
// manifests, and updating heartbeat files. Wire this into Cobra in start.go and
// validate via `go test ./internal/daemon`.
