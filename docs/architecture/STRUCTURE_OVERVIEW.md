# Code Structure & Dependency Map

## Top-Level Modules
- `cmd/` – Entry points compiled into binaries. `cmd/lowkey` wires CLI commands to runtime services.
- `internal/` – Core implementation details hidden from external consumers. Houses daemon orchestration, event ingestion, state management, reporting, and watcher logic.
- `pkg/` – Reusable utilities with stable APIs (configuration loading, output rendering, telemetry) that `cmd/` and `internal/` packages can import.
- `scripts/` – Operational tooling (installers, packaging, service manifests) invoked from CI or release pipelines.
- `testdata/` – Deterministic fixtures and manifests supporting integration tests.
- `docs/` – Product, architecture, and governance documentation.
- `build/` – Generated artifacts and release notes.

## Directory Relationships
- `cmd/lowkey/*` imports `internal/daemon`, `internal/watcher`, `internal/reporting`, `internal/state`, `pkg/config`, and `pkg/output` to implement CLI verbs. Each command file maps to a PRD-defined command.
- `internal/daemon` depends on:
  - `internal/events` for platform-specific fsnotify backends.
  - `internal/watcher` to start/stop hybrid monitors.
  - `internal/state` to persist manifests and cached file signatures.
  - `internal/logging` for structured JSON log writers.
  - `pkg/telemetry` to emit metrics/tracing when running as a daemon.
- `internal/watcher` consumes `internal/events`, `internal/filters`, and `internal/logging`. It supplies change notifications back to `internal/daemon` and `internal/reporting`.
- `internal/reporting` aggregates state and recent events pulled from `internal/state` and `internal/logging`, exposing summaries for CLI consumption.
- `internal/filters` provides Bloom-filter helpers used by `internal/watcher` and `internal/events` to skip ignored paths.
- `internal/state` is the single source of truth for cached signatures and daemon manifests; it exports read/write methods to `internal/daemon`, `internal/reporting`, and CLI commands that need inspection (`status`, `summary`, `clear`).
- `pkg/config` loads `.lowkey` patterns and daemon JSON manifests on behalf of `cmd/lowkey/start.go` and `internal/daemon/watch_manifest.go`.
- `pkg/output` offers tabular/JSON formatting consumed by CLI commands and optionally by daemon status endpoints.

## Imports & Exports
- Packages under `internal/` expose only package-level constructors (e.g., `daemon.NewManager`, `watcher.NewController`, `state.NewCache`). They are not importable by external modules.
- Packages under `pkg/` expose Go modules with clear interfaces:
  - `pkg/config` exports `Loader` and `Manifest` types.
  - `pkg/output` exports `Renderer` implementations for text table and JSON modes.
  - `pkg/telemetry` exports instrumentation helpers (`MetricsPublisher`, `Tracer`).
- `cmd/lowkey` defines Cobra-style command structs (`watchCmd`, `startCmd`, etc.) that register with `rootCmd`. Each command imports only the interfaces it needs (e.g., `cmd/lowkey/status.go` depends on `internal/daemon.Manager` and `pkg/output.Renderer`).

## Key Types & Responsibilities
- `daemon.Manager` – Starts/stops the background process, maintains PID/heartbeat files, coordinates watchers per manifest.
- `daemon.Supervisor` – Oversees the long-running goroutines, restarts watchers, and exposes health checks.
- `watcher.Controller` – Implements the hybrid event/poll loop described in the algorithm PRD.
- `events.Backend` – Abstract interface wrapping `fsnotify`/platform specifics.
- `filters.IgnoreBloom` – Probabilistic filter for ignore patterns, reducing expensive glob checks.
- `state.Cache` – Signature store backing incremental scans and deletion detection.
- `state.ManifestStore` – Handles atomic JSON updates for daemon configuration.
- `reporting.Aggregator` / `reporting.Summaries` – Generates the data structures consumed by `summary`, `log`, and `status` commands.
- `logging.Rotator` – Implements log rotation thresholds (10 MB × 5 files).
- `output.Renderer` – Abstract writer for human vs machine-readable CLI responses.
- `config.Loader` – Reads `.lowkey` and daemon manifest files with validation against schemas.
- `telemetry.Metrics` / `telemetry.Tracing` – Optional instrumentation hooks.

## External Dependencies
- Go standard library plus `fsnotify` (declared in PRD/ADR) for filesystem events.
- CLI built with `spf13/cobra`, declared in `go.mod` and used by command files under `cmd/lowkey`.
- Config management powered by `spf13/viper`, referenced by `pkg/config` and CLI command setup.
- Additional optional dependencies: `uber-go/zap` or `rs/zerolog` (logging), `prometheus/client_golang` (metrics). Introduce as implementation work starts.

## Data & Control Flow Summary
1. CLI command parses flags via `cmd/lowkey/*`, loads manifests using `pkg/config`, and resolves output format with `pkg/output`.
2. `start` command initializes `daemon.Manager`, which wires `internal/events`, `internal/watcher`, `internal/state`, and `internal/logging` together, then persists PID/manifest info (`state.ManifestStore`).
3. `watcher.Controller` pulls event streams from `events.Backend`, filters them (`filters.IgnoreBloom`), updates `state.Cache`, and emits batches to `reporting.Aggregator` and `logging.Rotator`.
4. `status`, `summary`, and `log` commands read from `internal/state` + `internal/reporting` to present current health via `pkg/output.Renderer`.
5. `stop` terminates the daemon through `daemon.Manager`, flushes logs, and clears manifests (`state.ManifestStore.Clear`).

## Testing & Fixtures
- `testdata/fixtures/*` simulates directory structures for integration tests verifying watcher correctness.
- `testdata/manifests/sample_daemon.json` provides a canonical manifest validating configuration loaders and state persistence.
- Go packages use standard `_test.go` files colocated with sources; integration suites may reside under `internal/watcher` and `internal/daemon` using the fixtures above.

## Outstanding Decisions
- Confirm CLI/daemon packages will use Cobra vs a home-grown parser; adjust builder files accordingly.
- Decide on telemetry/logging libraries to lock down the `pkg/telemetry` and `internal/logging` interfaces.
- Determine whether to expose a machine-readable HTTP status endpoint; if so, add an `internal/api` package.
