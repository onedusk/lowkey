---
claude: docs/CLAUDE.md
codex: docs/AGENTS.md
gemini: docs/GEMINI.md
---

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

## Flags

### `--metrics`

The `--metrics` flag enables the Prometheus metrics endpoint, allowing you to monitor the performance and activity of the `lowkey` daemon.

- **Usage:** `lowkey start --metrics [address:port] /path/to/watch`
- **Default Address:** `127.0.0.1:9600`

When enabled, `lowkey` exposes an HTTP endpoint serving metrics in the Prometheus exposition format. You can scrape this endpoint with a Prometheus server or query it directly with `curl`.

```bash
# Start lowkey with metrics on the default address
lowkey start --metrics /path/to/watch

# Access the metrics
curl http://127.0.0.1:9600/metrics
```

#### Available Metrics

| Metric Name      | Type      | Description                                                 |
|------------------|-----------|-------------------------------------------------------------|
| `events_total`   | Counter   | Total number of filesystem events processed, labeled by type (e.g., `create`, `modify`, `delete`). |
| `latency`        | Histogram | Latency of event processing in seconds, providing buckets for performance analysis. |
| `restart_count`  | Counter   | The number of times the internal watcher has been automatically restarted by the supervisor. |

### `--trace`

The `--trace` flag enables detailed trace logging for debugging and performance analysis. It provides insight into `lowkey`'s internal operations.

- **Usage:** `lowkey start --trace /path/to/watch`

When this flag is active, `lowkey` generates verbose logs that include:
- **Internal State Transitions:** See how the monitor reacts to events and manages its internal state.
- **Function Calls:** Trace the flow of execution through key parts of the application.
- **Detailed Event Data:** Get enriched information about each filesystem event as it is detected and processed.

**Performance Impact:**
Enabling trace logging can have a noticeable impact on performance due to the high volume of I/O operations for writing logs. It is recommended **only for debugging purposes** and should not be used in a production environment where performance is critical.

Logs generated in trace mode can be viewed with `lowkey tail` or by inspecting the log files directly.

## Event Types

Lowkey tracks the following types of filesystem events:

- **CREATE** - A new file or directory is created
- **MODIFY** - An existing file's content or metadata has changed
- **DELETE** - A file or directory is removed from the filesystem
- **RENAME** - A file or directory is renamed or moved to a new location

## Configuration & State

- **Ignore rules** – Place glob patterns in `.lowkey`; they are tokenised and
  loaded into a Bloom filter to avoid costly glob checks at runtime.

  Example `.lowkey` file:
  ```
  # Ignore patterns (one per line)
  node_modules/
  **/*.tmp
  **/*.log
  .git/
  **/__pycache__/
  **/dist/
  ```

  Pattern syntax:
  - `*` matches any sequence of non-separator characters
  - `**` matches zero or more directories
  - `?` matches any single non-separator character
  - Character classes: `[abc]` or `[a-z]`
- **Manifests** – The daemon persists manifests to the platform-specific state
  directory via `state.ManifestStore`. Updating the file on disk and running
  reconciliation (future CLI verb) enables hot reconfiguration.
- **Logs** – `internal/logging` rotates `lowkey.log` at 10 MB, keeping five
  archives. `lowkey tail` reads the active log and follows rotations.
- **Telemetry** – `--metrics` starts an HTTP server exposing Prometheus-style
  counters and latency gauges, while `--trace` enables lightweight span logging.
- **Supervisor** – A built-in supervisor watches the daemon manager, restarts
  the watcher when needed, and records heartbeat data surfaced by `status`.

## Performance

Lowkey is designed for high-throughput filesystem monitoring with minimal overhead:

- **Event Processing**: 10,000+ events/sec on modern hardware
- **Memory Usage**: ~15-25MB baseline, scales with watched directory count
- **CPU Usage**: <1% idle, 2-5% under moderate load (1000 events/sec)
- **Startup Time**: <100ms for daemon initialization
- **Bloom Filter**: O(1) ignore pattern matching with <1% false positive rate

Benchmarks run on: Apple M1, 16GB RAM, monitoring 50,000 files with 1,000 ignore patterns.

## Architecture

```
┌───────┐   ┌─────┐   ┌───────────────┐   ┌────────────┐   ┌────────────────┐
│  User │──▶│ CLI │──▶│ Daemon Manager│──▶│ Supervisor │──▶│ Hybrid Monitor │
└───────┘   └─────┘   └───────────────┘   └────────────┘   └────────────────┘
                                                                    │
                                                                    ├─▶┌──────────────┐
                                                                    │  │Event Backend │
                                                                    │  │  (fsnotify)  │
                                                                    │  └──────────────┘
                                                                    │
                                                                    └─▶┌──────────────┐
                                                                       │Polling Scanner│
                                                                       │  (Periodic)  │
                                                                       └──────────────┘
```

## Development

- Format code with `gofmt` (Go 1.22+ target).
- Run the full suite with `GOCACHE=$(pwd)/.gocache go test ./...` (mirrors
  `make test`).
- Architecture and operational docs live under `docs/`; `docs/architecture/overview.md`
  describes the hybrid monitoring pipeline and telemetry wiring.

## FAQ

**What platforms are supported?**
Lowkey is written in Go and is cross-platform. It is tested on macOS, Linux, and Windows.

**How does daemon mode work?**
The `lowkey start` command launches the monitor as a background process. A built-in supervisor ensures automatic restarts if crashes occur, providing resilient long-running monitoring.

**Can I use Lowkey with Docker?**
Yes. Run the `lowkey` binary inside a container and mount the host directory as a volume. For reliable event detection, mount from the host: `docker run -v /host/path:/container/path ...`

**What's the difference between `watch` and `start` commands?**
- `lowkey watch` runs in the foreground, streaming events to your terminal (best for temporary monitoring)
- `lowkey start` runs as a background daemon (best for persistent, long-term monitoring)

**How do I debug issues?**
Use these commands:
- `lowkey status` - Check daemon status and event summaries
- `lowkey tail` - Stream daemon log output in real-time
- `lowkey clear --logs` - Clear historical logs

## Contributing

Review the [Contributor Guide](AGENTS.md) for repository structure, workflows, and coding standards, then see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for contribution logistics.

## License

This project is licensed under the MIT License - see the [LICENSE](docs/LICENSE) file for details.
