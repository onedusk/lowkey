#!/usr/bin/env bash

# setup_project_tree.sh - Materializes the project hierarchy described in PROJECT_TREE.md.
# Usage: run from repo root -> `bash setup_project_tree.sh`
# The script is idempotent and will not overwrite existing files.

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

directories=(
  "cmd/lowkey"
  "internal/daemon"
  "internal/events"
  "internal/filters"
  "internal/logging"
  "internal/state"
  "internal/reporting"
  "internal/watcher"
  "pkg/config"
  "pkg/output"
  "pkg/telemetry"
  "scripts/services"
  "testdata/fixtures/nested_changes"
  "testdata/fixtures/large_directory"
  "testdata/manifests"
  "docs/adrs"
  "docs/prds"
  "docs/guides"
  "docs/api"
  "build/artifacts"
  "build/release_notes"
)

for dir in "${directories[@]}"; do
  mkdir -p "${ROOT_DIR}/${dir}"
done

write_file() {
  local relative_path="$1"
  local full_path="${ROOT_DIR}/${relative_path}"

  if [[ -f "${full_path}" ]]; then
    return
  fi

  case "${relative_path}" in
    cmd/lowkey/main.go)
      cat <<'EOF' >"${full_path}"
package main

// main boots the Cobra CLI so you can run `go run ./cmd/lowkey`. Once commands
// are wired, replace the TODO below with a call to rootCmd.Execute().
func main() {
    // TODO: invoke rootCmd.Execute() after defining commands in this package.
}
EOF
      ;;
    cmd/lowkey/root.go)
      cat <<'EOF' >"${full_path}"
package main

// root.go hosts the root Cobra command shared by every sub-command. Run
// `go build ./cmd/lowkey` while you iterate to ensure flags and wiring compile.
// Define persistent flags and configuration bootstrapping here.
EOF
      ;;
    cmd/lowkey/watch.go)
      cat <<'EOF' >"${full_path}"
package main

// watch.go describes the `lowkey watch` Cobra command for foreground monitoring.
// Use this file to bind CLI flags to the watcher controller in internal/watcher.
// Reload with `go run ./cmd/lowkey watch <dir>` during development.
EOF
      ;;
    cmd/lowkey/start.go)
      cat <<'EOF' >"${full_path}"
package main

// start.go defines the `lowkey start` command that launches the daemon manager.
// Call into internal/daemon.Manager and persist manifests via pkg/config.
// Test with `go run ./cmd/lowkey start --config <manifest>`.
EOF
      ;;
    cmd/lowkey/stop.go)
      cat <<'EOF' >"${full_path}"
package main

// stop.go implements graceful shutdown for the background daemon. Ensure this
// command reads the PID/heartbeat files internal/daemon writes and removes
// residual state. Verify via `go run ./cmd/lowkey stop`.
EOF
      ;;
    cmd/lowkey/status.go)
      cat <<'EOF' >"${full_path}"
package main

// status.go exposes runtime information (PID, directories, heartbeat).
// Combine internal/daemon.Manager lookups with pkg/output renderers to present
// human/JSON output. Iterate using `go run ./cmd/lowkey status`.
EOF
      ;;
    cmd/lowkey/log.go)
      cat <<'EOF' >"${full_path}"
package main

// log.go streams or filters daemon logs. Use internal/logging readers together
// with pkg/output to shape the presentation. Exercise with `go run ./cmd/lowkey log`.
EOF
      ;;
    cmd/lowkey/tail.go)
      cat <<'EOF' >"${full_path}"
package main

// tail.go follows the daemon log in real time. Leverage the rotator pipe in
// internal/logging to deliver incremental updates to stdout.
EOF
      ;;
    cmd/lowkey/summary.go)
      cat <<'EOF' >"${full_path}"
package main

// summary.go queries internal/reporting for aggregated change statistics and
// renders them via pkg/output. Useful for PRD `lowkey summary` behavior.
EOF
      ;;
    cmd/lowkey/clear.go)
      cat <<'EOF' >"${full_path}"
package main

// clear.go handles `lowkey clear` for logs/state. Use internal/reporting and
// internal/state to drop data, ensuring confirmation prompts mirror the PRD.
EOF
      ;;
    internal/daemon/manager.go)
      cat <<'EOF' >"${full_path}"
package daemon

// manager.go coordinates daemon lifecycle: spawning watchers, persisting
// manifests, and updating heartbeat files. Wire this into Cobra in start.go and
// validate via `go test ./internal/daemon`.
EOF
      ;;
    internal/daemon/supervisor.go)
      cat <<'EOF' >"${full_path}"
package daemon

// supervisor.go supervises goroutines for watchers and logging. Implement
// restart logic and health checks here, exercising them with integration tests.
EOF
      ;;
    internal/daemon/watch_manifest.go)
      cat <<'EOF' >"${full_path}"
package daemon

// watch_manifest.go parses and reconciles persisted manifests. Collaborates with
// pkg/config and internal/state to reconcile desired vs actual watch targets.
EOF
      ;;
    internal/events/backend.go)
      cat <<'EOF' >"${full_path}"
package events

// backend.go defines the interface to platform-specific filesystem event APIs.
// Provide constructors that choose fsnotify variants based on GOOS.
EOF
      ;;
    internal/events/fsnotify_unix.go)
      cat <<'EOF' >"${full_path}"
package events

// fsnotify_unix.go wraps Darwin/Linux watchers. Use build tags later to restrict
// compilation and expose a constructor used by backend.go.
EOF
      ;;
    internal/events/fsnotify_windows.go)
      cat <<'EOF' >"${full_path}"
package events

// fsnotify_windows.go contains the Windows implementation. Mirror the Unix API
// so backend.go can return a consistent interface across GOOS values.
EOF
      ;;
    internal/filters/bloom_filter.go)
      cat <<'EOF' >"${full_path}"
package filters

// bloom_filter.go builds the ignore Bloom filter discussed in algorithm_design.
// Implement Add/Contains helpers tuned for CLI patterns. Benchmark with `go test`.
EOF
      ;;
    internal/filters/ignore_tokens.go)
      cat <<'EOF' >"${full_path}"
package filters

// ignore_tokens.go extracts tokens from glob patterns and paths. Keep it in sync
// with the Bloom filter heuristics; add table-driven tests alongside.
EOF
      ;;
    internal/logging/rotator.go)
      cat <<'EOF' >"${full_path}"
package logging

// rotator.go enforces the 10MB × 5 log rotation strategy. Integrate with
// reporting so CLI commands stream archived segments.
EOF
      ;;
    internal/logging/writer.go)
      cat <<'EOF' >"${full_path}"
package logging

// writer.go exposes structured logging utilities (likely wrapping zap/zerolog).
// Provide options so both daemon and CLI share consistent log formats.
EOF
      ;;
    internal/state/cache.go)
      cat <<'EOF' >"${full_path}"
package state

// cache.go tracks file signatures for incremental scanning. Implement thread-safe
// read/write helpers as described in algorithm_design.md.
EOF
      ;;
    internal/state/persistence.go)
      cat <<'EOF' >"${full_path}"
package state

// persistence.go handles durable storage for the cache (e.g., boltDB or JSON).
// Ensure writes are atomic so crash recovery honors the PRD.
EOF
      ;;
    internal/state/manifest_store.go)
      cat <<'EOF' >"${full_path}"
package state

// manifest_store.go persists daemon manifests in `$XDG_STATE_HOME/lowkey`. It is
// invoked by start/stop commands and the daemon manager.
EOF
      ;;
    internal/reporting/aggregator.go)
      cat <<'EOF' >"${full_path}"
package reporting

// aggregator.go collects change events from the watcher pipeline and builds
// aggregates consumed by summary/status commands.
EOF
      ;;
    internal/reporting/summaries.go)
      cat <<'EOF' >"${full_path}"
package reporting

// summaries.go formats aggregated data for CLI output. Keep logic separate from
// rendering so pkg/output can focus on presentation.
EOF
      ;;
    internal/watcher/controller.go)
      cat <<'EOF' >"${full_path}"
package watcher

// controller.go drives the hybrid monitoring loop, coordinating events,
// incremental scans, and batching. Unit test with simulated fixtures.
EOF
      ;;
    internal/watcher/hybrid_monitor.go)
      cat <<'EOF' >"${full_path}"
package watcher

// hybrid_monitor.go implements the algorithm described in docs/prds/algorithm_design.md,
// blending fsnotify events with safety scans. Profile for large directories.
EOF
      ;;
    pkg/config/loader.go)
      cat <<'EOF' >"${full_path}"
package config

// loader.go initializes Viper, loads `.lowkey` files, and reads daemon manifests.
// Export helpers consumed by cmd/lowkey and internal/daemon.
EOF
      ;;
    pkg/config/schema.go)
      cat <<'EOF' >"${full_path}"
package config

// schema.go defines validation rules for manifests and CLI configuration. Keep
// this aligned with PRD expectations for watch directories and logging.
EOF
      ;;
    pkg/output/formatter.go)
      cat <<'EOF' >"${full_path}"
package output

// formatter.go provides interfaces for rendering tables/JSON. Commands depend on
// this package to emit consistent human and machine-readable responses.
EOF
      ;;
    pkg/output/table.go)
      cat <<'EOF' >"${full_path}"
package output

// table.go implements a textual table renderer for CLI results. Add tests to
// ensure wide inputs wrap cleanly.
EOF
      ;;
    pkg/telemetry/metrics.go)
      cat <<'EOF' >"${full_path}"
package telemetry

// metrics.go exports Prometheus-style collectors tracking filesystem events,
// latency, and errors. Wire into daemon startup when metrics are enabled.
EOF
      ;;
    pkg/telemetry/tracing.go)
      cat <<'EOF' >"${full_path}"
package telemetry

// tracing.go hosts OpenTelemetry hooks for cross-component tracing. Optional but
// useful when diagnosing watcher bottlenecks.
EOF
      ;;
    scripts/install.sh)
      cat <<'EOF' >"${full_path}"
#!/usr/bin/env bash

# install.sh installs the lowkey binary locally. Use it during development to
# copy builds into your PATH. Run `bash scripts/install.sh` after `go build`.
EOF
      ;;
    scripts/package.ps1)
      cat <<'EOF' >"${full_path}"
# package.ps1 bundles Windows releases. Execute from PowerShell after running
# `go build ./cmd/lowkey`. Add signing/zip steps here when the release pipeline
# takes shape.
EOF
      ;;
    scripts/services/lowkey.plist)
      cat <<'EOF' >"${full_path}"
<?xml version="1.0" encoding="UTF-8"?>
<!-- launchd plist placeholder for running lowkey as a user agent. Populate
     ProgramArguments and KeepAlive once the daemon is ready. -->
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- TODO: define ProgramArguments, RunAtLoad, and WorkingDirectory. -->
</dict>
</plist>
EOF
      ;;
    scripts/services/lowkey.service)
      cat <<'EOF' >"${full_path}"
# systemd service template for lowkey. Copy to /etc/systemd/system and adjust
# ExecStart once the daemon binary path is finalized.
# [Unit]
# Description=Lowkey filesystem watcher daemon
#
# [Service]
# ExecStart=/usr/local/bin/lowkey daemon
# Restart=on-failure
#
# [Install]
# WantedBy=multi-user.target
EOF
      ;;
    testdata/manifests/sample_daemon.json)
      cat <<'EOF' >"${full_path}"
{
  "_comment": "Sample manifest for tests. Adjust directories before running `go test`.",
  "directories": [
    "/tmp/example"
  ],
  "log_path": "/tmp/lowkey.log"
}
EOF
      ;;
    docs/guides/daemon.md)
      cat <<'EOF' >"${full_path}"
# Daemon Guide (Placeholder)

This guide will explain how to run the background daemon, configure manifests,
and debug lifecycle issues. Expand it alongside the implementation of
`internal/daemon`.
EOF
      ;;
    docs/api/cli_reference.md)
      cat <<'EOF' >"${full_path}"
# CLI Reference (Placeholder)

Document each Cobra command (`watch`, `start`, `status`, etc.) here once their
flags and behaviors solidify. Link back to the PRD for acceptance criteria.
EOF
      ;;
    .lowkey)
      cat <<'EOF' >"${full_path}"
# .lowkey defines ignore patterns for the watcher. Populate this file with glob
# patterns (one per line) to exclude artifacts during development.
EOF
      ;;
    go.sum)
      cat <<'EOF' >"${full_path}"
# go.sum will be generated by `go mod tidy` once dependencies such as cobra and
# viper are added. This placeholder exists so the project tree matches the plan.
EOF
      ;;
    *)
      echo "Unhandled file template for ${relative_path}" >&2
      ;;
  esac
}

files=(
  "cmd/lowkey/main.go"
  "cmd/lowkey/root.go"
  "cmd/lowkey/watch.go"
  "cmd/lowkey/start.go"
  "cmd/lowkey/stop.go"
  "cmd/lowkey/status.go"
  "cmd/lowkey/log.go"
  "cmd/lowkey/tail.go"
  "cmd/lowkey/summary.go"
  "cmd/lowkey/clear.go"
  "internal/daemon/manager.go"
  "internal/daemon/supervisor.go"
  "internal/daemon/watch_manifest.go"
  "internal/events/backend.go"
  "internal/events/fsnotify_unix.go"
  "internal/events/fsnotify_windows.go"
  "internal/filters/bloom_filter.go"
  "internal/filters/ignore_tokens.go"
  "internal/logging/rotator.go"
  "internal/logging/writer.go"
  "internal/state/cache.go"
  "internal/state/persistence.go"
  "internal/state/manifest_store.go"
  "internal/reporting/aggregator.go"
  "internal/reporting/summaries.go"
  "internal/watcher/controller.go"
  "internal/watcher/hybrid_monitor.go"
  "pkg/config/loader.go"
  "pkg/config/schema.go"
  "pkg/output/formatter.go"
  "pkg/output/table.go"
  "pkg/telemetry/metrics.go"
  "pkg/telemetry/tracing.go"
  "scripts/install.sh"
  "scripts/package.ps1"
  "scripts/services/lowkey.plist"
  "scripts/services/lowkey.service"
  "testdata/manifests/sample_daemon.json"
  "docs/guides/daemon.md"
  "docs/api/cli_reference.md"
  ".lowkey"
  "go.sum"
)

for file in "${files[@]}"; do
  write_file "${file}"
done

chmod +x "${ROOT_DIR}/scripts/install.sh" 2>/dev/null || true

echo "Project hierarchy ensured." >&2
