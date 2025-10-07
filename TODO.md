# Project TODOs

This file lists all the `TODO` items found in the project, generated from the source code comments.

## Core Implementation

- [ ] **`internal/daemon/supervisor.go`**: Implement the supervisor loop that monitors watcher health, restarts failed goroutines, and surfaces heartbeat metadata.
- [ ] **`internal/daemon/watch_manifest.go`**: Reconcile persisted manifests with the active watcher and support dynamic directory updates.

## CLI Commands

- [ ] **`cmd/lowkey/clear.go`**: Wire pruning logic for logs and/or cached state with suitable confirmations.

## Documentation & Scripts

- [ ] **`docs/architecture/overview.md`**: Refresh the architecture overview to reflect the implemented hybrid monitoring pipeline.
- [ ] **`scripts/services/lowkey.plist`**: Fill in `ProgramArguments`, `RunAtLoad`, and `WorkingDirectory` for macOS launch agent support.
- [ ] **`scripts/setup_project_tree.sh`**: Replace placeholder TODOs with actual CLI bootstrap (e.g., invoking `rootCmd.Execute()`).
