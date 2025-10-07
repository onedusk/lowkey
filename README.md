# Lowkey

## Overview

Lowkey is a cross-platform filesystem monitor written in Go. It can run in the
foreground for quick ad-hoc monitoring or as a daemon that supervises a hybrid
event/polling watcher. Manifests, telemetry, and maintenance commands keep the
tooling lightweight while still matching the product requirements.

## Quick Start

```bash
# build the CLI (outputs ./lowkey/lowkey)
make build

# run a foreground watch, streaming create/modify/delete events
./lowkey/lowkey watch ./path/to/project

# launch the daemon with metrics + tracing enabled, then inspect status
./lowkey/lowkey start --metrics 127.0.0.1:9600 --trace ./path/to/project
./lowkey/lowkey status

# stop the daemon and prune logs/state
./lowkey/lowkey stop
./lowkey/lowkey clear --logs --state --yes
```

## CLI Commands

- `lowkey watch <dirs...>` – Run the hybrid monitor in the foreground and stream
  change notifications to stdout until interrupted.
- `lowkey start [--metrics addr] [--trace] <dirs...>` – Re-exec the binary as a
  background daemon, persist the manifest to `$XDG_STATE_HOME/lowkey/daemon.json`
  (with platform fallbacks), and optionally expose Prometheus metrics or log
  tracing spans.
- `lowkey stop` – Read the PID file from the state directory, signal the daemon
  to exit, wait for graceful shutdown, and clear the manifest.
- `lowkey status` – Report the active manifest, supervisor heartbeat metadata
  (running flag, restart count, backoff window), and aggregated change summary.
- `lowkey tail` – Follow the rotated daemon log (default `lowkey.log` in the
  state directory or a manifest-specified path).
- `lowkey clear [--logs] [--state] [--yes]` – Delete rotated logs and/or state
  artifacts (manifest, cache snapshot, PID file) after confirmation.
- Additional scaffolding commands (`summary`, `log`, etc.) live under `cmd/`
  and will evolve alongside product requirements.

## Configuration & State

- **Ignore rules** – Place glob patterns in `.lowkey`; they are tokenised and
  loaded into a Bloom filter to avoid costly glob checks at runtime.
- **Manifests** – The daemon persists manifests to the platform-specific state
  directory via `state.ManifestStore`. Updating the file on disk and running
  reconciliation (future CLI verb) enables hot reconfiguration.
- **Logs** – `internal/logging` rotates `lowkey.log` at 10 MB, keeping five
  archives. `lowkey tail` reads the active log and follows rotations.
- **Telemetry** – `--metrics` starts an HTTP server exposing Prometheus-style
  counters and latency gauges, while `--trace` enables lightweight span logging.
- **Supervisor** – A built-in supervisor watches the daemon manager, restarts
  the watcher when needed, and records heartbeat data surfaced by `status`.

## Development

- Format code with `gofmt` (Go 1.22+ target).
- Run the full suite with `GOCACHE=$(pwd)/.gocache go test ./...` (mirrors
  `make test`).
- Architecture and operational docs live under `docs/`; `docs/architecture/overview.md`
  describes the hybrid monitoring pipeline and telemetry wiring.

## Contributing

Review the [Contributor Guide](AGENTS.md) for repository structure, workflows, and coding standards, then see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for contribution logistics.

## License

This project is licensed under the MIT License - see the [LICENSE](docs/LICENSE) file for details.
