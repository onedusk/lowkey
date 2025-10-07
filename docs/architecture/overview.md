# Lowkey Architecture Overview

Lowkey is a Cobra-driven CLI that can run in foreground or as a background
daemon. The runtime is split into clearly scoped packages so that commands,
monitoring logic, and shared utilities evolve in lockstep without leaking
internal details.

## Runtime Flow

1. **CLI entry** – `cmd/lowkey` wires every Cobra command. Foreground
   operations (e.g. `watch`, `tail`) execute controllers directly, while
   lifecycle verbs (`start`, `stop`, `status`, `clear`) coordinate with the
   daemon state directory.
2. **Daemon manager** – `internal/daemon.Manager` owns the long-lived watcher
   controller, manifest persistence, and telemetry hooks. A `Supervisor`
   monitors the manager and attempts restarts when the watcher stops
   unexpectedly.
3. **Watcher controller** – `internal/watcher.Controller` spins up the hybrid
   monitor which blends fsnotify-style events (currently surfaced through a
   portable polling backend) with periodic safety scans.
4. **State & reporting** – `internal/state` persists manifests and cache
   snapshots, while `internal/reporting` aggregates recent events for status and
   summary commands.
5. **Telemetry** – `pkg/telemetry` exposes lightweight Prometheus and tracing
   stubs that can be toggled via `lowkey start --metrics/--trace`.

## Package Responsibilities

- **`cmd/lowkey`** – Cobra commands, argument parsing, and wiring to internal
  packages. Recently implemented commands include:
  - `watch` for foreground monitoring with live change streaming.
  - `start`/`stop` for daemon lifecycle management, leveraging a re-exec model
    so the daemon runs as a dedicated process.
  - `tail`, `status`, and `clear` for operational visibility and maintenance.
- **`internal/daemon`** – Manager, supervisor, and manifest reconciliation
  helpers. The reconciler can reload manifests from disk, diff directories, and
  rebuild watcher controllers without restarting the CLI.
- **`internal/watcher`** – The `HybridMonitor` consumes filesystem events from
  `internal/events`, merges them with periodic scans that consult the signature
  cache, and feeds results into the reporting aggregator.
- **`internal/events`** – A portable polling backend that mimics fsnotify until
  platform-specific integrations are introduced. The interface is compatible
  with native backends, which can later replace the polling implementation.
- **`internal/filters`** – Bloom filter and tokenisation helpers used to apply
  `.lowkey` ignore patterns efficiently.
- **`internal/state`** – Thread-safe file signature cache, atomic persistence
  helpers, and manifest storage rooted in `$XDG_STATE_HOME/lowkey` (with
  platform fallbacks).
- **`pkg/config`** – Manifest loading, validation, and normalisation
  (directories, ignores, log paths).
- **`pkg/output`** – Plain-text and JSON renderers for daemon status
  (including the new supervisor heartbeat metadata).
- **`pkg/telemetry`** – Embeddable metrics/tracing hooks that can run over HTTP
  or log spans for debugging.

## Event Monitoring Pipeline

1. The controller registers watch roots with the events backend.
2. Each backend emission is filtered against Bloom-filtered ignore tokens
   before being recorded.
3. Periodic scans walk the directory tree, compute cached signatures, and
   synthesise `CREATE`, `MODIFY`, or `DELETE` events for anything the real-time
   channel missed.
4. Events feed the reporting aggregator (for CLI output) and optional telemetry
   emitters.

## Persistence & Maintenance

- **Manifest** – Stored atomically at `$XDG_STATE_HOME/lowkey/daemon.json` via
  `state.ManifestStore`.
- **Cache** – `state.Cache` snapshots can be saved and loaded as
  JSON-serialised signature maps, enabling quick recovery after restarts.
- **Logs** – `internal/logging` rotates `lowkey.log` once it reaches 10 MB and
  retains five archives.
- **Maintenance** – `lowkey clear` offers selective removal of logs and cached
  state. `lowkey status` surfaces supervisor heartbeats, restart counts, and
  log summaries.

This architecture keeps the CLI, daemon orchestration layer, and reusable
utilities decoupled so future work—like swapping the polling backend for real
fsnotify integrations or expanding telemetry—can be delivered incrementally.
