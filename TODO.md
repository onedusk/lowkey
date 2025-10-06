# Project TODOs

This file lists all the `TODO` items found in the project, generated from the source code comments.

## Core Implementation

- [ ] **`internal/daemon/manager.go`**: The `controller.Start()` call is currently a stub. This needs to be replaced with the actual implementation that starts the monitoring loops.
- [ ] **`internal/daemon/supervisor.go`**: Implement the supervisor logic. This should include a main loop to monitor the health of watcher goroutines, a mechanism to restart failed goroutines, health check endpoints, and integration with the daemon manager.
- [ ] **`internal/daemon/watch_manifest.go`**: Implement the manifest reconciliation logic, including loading the manifest, comparing it with the current configuration, and dynamically updating the running watcher.
- [ ] **`internal/events/backend.go`**: Define the event backend interface and implement a constructor `NewBackend()` that uses build tags or runtime checks on GOOS to initialize the appropriate platform-specific backend.
- [ ] **`internal/events/fsnotify_unix.go`**: Implement the fsnotify-based watcher for Unix-like systems (Linux, Darwin).
- [ ] **`internal/events/fsnotify_windows.go`**: Implement the fsnotify-based watcher for Windows.
- [ ] **`internal/filters/bloom_filter.go`**: Implement the Bloom filter for efficient glob pattern matching.
- [ ] **`internal/filters/ignore_tokens.go`**: Implement the logic for extracting tokens from glob patterns and file paths.
- [ ] **`internal/state/cache.go`**: Implement the file signature cache to store file metadata and avoid redundant processing.
- [ ] **`internal/state/persistence.go`**: Implement durable storage for the file signature cache to allow the watcher to resume from a previous state.
- [ ] **`internal/watcher/controller.go`**: Spawn event and polling goroutines, integrate filters and reporters.
- [ ] **`internal/watcher/hybrid_monitor.go`**: Implement the hybrid monitoring algorithm, blending real-time events with periodic polling.

## CLI Commands

- [ ] **`cmd/lowkey/clear.go`**: Implement the logic to prune logs and/or cached state.
- [ ] **`cmd/lowkey/start.go`**: Implement the logic to launch the daemon as a separate background process.
- [ ] **`cmd/lowkey/stop.go`**: Implement the logic to find and signal the running daemon process to stop.
- [ ] **`cmd/lowkey/tail.go`**: Implement the logic to stream live log entries, similar to `tail -f`.
- [ ] **`cmd/lowkey/watch.go`**: Implement the foreground watching logic, running the controller directly and waiting for an interrupt signal.

## Telemetry

- [ ] **`pkg/telemetry/metrics.go`**: Implement Prometheus metrics for observability (e.g., events, errors, latency).
- [ ] **`pkg/telemetry/tracing.go`**: Implement OpenTelemetry for distributed tracing to diagnose performance bottlenecks.

## Scripts & Documentation

- [ ] **`docs/architecture/overview.md`**: This document should be updated as the application evolves.
- [ ] **`scripts/services/lowkey.plist`**: Define `ProgramArguments`, `RunAtLoad`, and `WorkingDirectory`.
- [ ] **`scripts/setup_project_tree.sh`**: Invoke `rootCmd.Execute()` after defining commands.
