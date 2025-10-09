# Lowkey

## Project Overview

This project, named "Lowkey," is a cross-platform filesystem monitor written in Go. It is designed to run either in the foreground for ad-hoc monitoring or as a background daemon. The tool utilizes a hybrid event/polling watcher to supervise file system changes. Key features include manifest-based configuration, telemetry (Prometheus metrics and tracing), and a set of maintenance commands for managing the daemon's state.

The project is structured as a Cobra-driven command-line interface. The core logic is divided into several packages, each with a specific responsibility:

* **`cmd/lowkey`**: Handles command-line argument parsing and wiring to the internal packages.
* **`internal/daemon`**: Manages the lifecycle of the background daemon, including a supervisor for automatic restarts.
* **`internal/watcher`**: Implements the hybrid monitoring logic, combining a polling-based event system with periodic safety scans.
* **`internal/state`**: Manages the persistence of manifests and cache snapshots.
* **`internal/reporting`**: Aggregates events for status and summary commands.
* **`pkg/config`**: Handles loading and validation of configuration manifests.
* **`pkg/output`**: Provides formatting for command output.
* **`pkg/telemetry`**: Exposes Prometheus metrics and tracing capabilities.

## Building and Running

### Building

To build the `lowkey` executable, run the following command:

```bash
make build
```

This will create the binary at `./lowkey/lowkey`.

### Running

**Foreground Monitoring:**

To run a foreground watch and stream create/modify/delete events, use the `watch` command:

```bash
./lowkey/lowkey watch ./path/to/project
```

**Daemon Mode:**

To start the daemon with metrics and tracing enabled, use the `start` command:

```bash
./lowkey/lowkey start --metrics 127.0.0.1:9600 --trace ./path/to/project
```

**Interacting with the Daemon:**

* **Check status:** `./lowkey/lowkey status`
* **Stop the daemon:** `./lowkey/lowkey stop`
* **Tail logs:** `./lowkey/lowkey tail`
* **Clear logs and state:** `./lowkey/lowkey clear --logs --state --yes`

### Testing

To run the full test suite, use the following command:

```bash
GOCACHE=$(pwd)/.gocache go test ./...
```

This is equivalent to running `make test`.

## Development Conventions

* **Code Formatting:** Code should be formatted with `gofmt` (targeting Go 1.22+).
* **Dependencies:** The project uses `cobra` and `viper` for the CLI and configuration, with local copies of these libraries stored in the `third_party` directory.
* **Architecture:** The project follows a modular architecture with a clear separation of concerns between packages. The `docs/architecture/overview.md` file provides a detailed description of the architecture.
* **Configuration:** Ignore rules are defined in a `.lowkey` file, which uses glob patterns. The daemon's configuration is managed through a manifest file.
* **Contributing:** The `AGENTS.md` and `docs/CONTRIBUTING.md` files provide guidance for contributors.
