# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Build & Development
```bash
# Build the CLI binary (outputs ./lowkey)
make build

# Run tests with verbose output
make test
# Or with timeout protection (30s limit)
./scripts/run_make_test.sh

# Format code (Go 1.22+ target)
gofmt -w .
# Or with import ordering
goimports -w .

# Clean build artifacts
make clean

# Quick smoke test with dev defaults
make run
```

### Running Lowkey
```bash
# Foreground monitoring (streams events to stdout)
./lowkey watch <dirs...>

# Start daemon with telemetry
./lowkey start --metrics 127.0.0.1:9600 --trace <dirs...>

# Daemon management
./lowkey status    # View daemon status and aggregated events
./lowkey tail      # Follow rotated daemon log
./lowkey stop      # Graceful shutdown
./lowkey clear --logs --state --yes  # Clean up logs and state
```

### Testing Patterns
- Tests follow `TestFeature_Scenario` naming convention
- Table-driven tests are preferred for comprehensive coverage
- Test files are co-located with source files (`*_test.go`)
- Integration tests capture multi-platform edge cases

## Architecture Overview

Lowkey uses a **dual-mode execution model**: it can run as a foreground CLI tool or as a background daemon. The daemon uses a re-exec pattern where `LOWKEY_DAEMON=1` environment variable determines the execution mode.

### Core Components & Interactions

1. **CLI Entry (`cmd/lowkey/main.go`)**: Checks `LOWKEY_DAEMON` env var to determine whether to run as daemon or CLI client. Commands are wired through Cobra in separate files (watch.go, start.go, stop.go, etc.).

2. **Daemon Manager (`internal/daemon/Manager`)**: Orchestrates the entire daemon lifecycle. It's supervised by a `Supervisor` that monitors health and performs automatic restarts with exponential backoff on failures.

3. **Hybrid Monitor (`internal/watcher/HybridMonitor`)**: Combines event-driven monitoring with periodic safety scans. This dual approach ensures no changes are missed even if the event backend fails.

4. **Event Backend (`internal/events`)**: Currently uses a portable polling implementation that mimics fsnotify. The interface is designed to be swappable for platform-specific implementations.

5. **State Persistence**:
   - Manifests stored at `$XDG_STATE_HOME/lowkey/daemon.json`
   - File signature cache for change detection
   - Atomic write operations prevent corruption

### Key Design Patterns

**Supervisor Pattern**: The daemon uses a two-layer supervision model:
- `Supervisor` monitors the `Manager` and restarts it on failure
- `Manager` supervises the `Controller` and handles reconciliation
- Both layers track restart counts and use backoff strategies

**Bloom Filter for Ignores**: `.lowkey` patterns are tokenized and loaded into a Bloom filter (`internal/filters/bloom_filter.go`) to efficiently filter events without expensive glob matching on every file.

**Rotated Logging**: Logs rotate at 10MB, keeping 5 archives. The `tail` command handles rotation transparently by watching for file changes.

**Manifest Reconciliation**: The daemon can hot-reload configuration by detecting manifest changes on disk and rebuilding the watcher without restart.

## Development Considerations

### Module Structure
- Uses local vendoring for cobra/viper dependencies (see `third_party/` and `go.mod` replace directives)
- Target Go 1.22 compatibility
- Package boundaries are strictly enforced: `internal/` for implementation, `pkg/` for reusable components

### Thread Safety
- `state.Cache` uses RWMutex for concurrent access
- `daemon.Manager` serializes operations with mutex
- Event aggregation is lock-free using channels

### Platform Considerations
- State directory uses XDG standards with platform fallbacks
- Event backend interface allows platform-specific implementations
- Windows uses different fsnotify backend (see `fsnotify_windows.go`)

### Error Handling Philosophy
- Controllers validate configuration at construction time
- Graceful degradation: polling continues even if event backend fails
- Clear error propagation through context cancellation

## Common Development Tasks

### Adding a New CLI Command
1. Create new file in `cmd/lowkey/` (e.g., `newcmd.go`)
2. Define Cobra command with proper flags and validation
3. Wire to internal packages for implementation
4. Update root command registration

### Modifying the Watcher
1. Changes to monitoring logic go in `internal/watcher/`
2. Event backend modifications in `internal/events/`
3. Remember to handle both event-driven and polling paths
4. Test with forced polling mode to ensure fallback works

### Debugging the Daemon
1. Check daemon status: `./lowkey status`
2. View logs: `./lowkey tail` or directly at `$XDG_STATE_HOME/lowkey/lowkey.log`
3. Enable tracing: Start with `--trace` flag
4. Check metrics endpoint if `--metrics` was used

## Important Notes from AGENTS.md

- Format every change with `gofmt` (tabs, Unix newlines)
- Keep unit tests beside sources named `*_test.go`
- Run `make test` before every commit
- Write imperative, ≤72-character commit subjects
- Document cross-platform assumptions immediately
- Vendor external dependencies under `third_party/` with licenses

---

## Important

- `.ctxt`: houses previous context dumps from earlier development cycles
- `.tks`: houses system for managing tasks related to this projects scope only
- `.lowlog`: house logs from previous development cycles

**BEFORE FEATURE ADDITIONS:** [follow protocol](docs/protocols/ipp/flow.md)

- `guides`: live on [this path](docs/guides/)
