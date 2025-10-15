# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Durable file-signature cache with atomic persistence helpers plus unit coverage.
- Ignore Bloom-filter/tokenization utilities backing the hybrid watcher pipeline.
- Cross-platform polling event backend, hybrid monitor, and controller wiring.
- Watch-mode now initializes `.lowkey/<date>.log` at watch startup so downstream tooling can tail changes immediately.
- Foreground watch streaming, daemon re-exec with `--metrics`/`--trace`, log tailing, and PID-aware stop/status flows.
- Telemetry stubs for Prometheus metrics and lightweight tracing toggled by CLI flags.
- Supervisor with restart/backoff logic and heartbeat telemetry stitched into the daemon manager and status output.
- Manifest reconciliation helpers that diff persisted manifests and rebuild watcher controllers on the fly.
- Fully implemented `lowkey clear` command with selective log/state pruning and safety prompts.
- Updated architecture overview and launchd service template reflecting the new runtime pipeline.

### Changed

- Daemon manager now loads ignore patterns from manifests and routes watcher events into telemetry hooks.
- `lowkey status` renders supervisor heartbeat data; scaffolding scripts emit ready-to-run Cobra stubs.

## [0.1.0] - 2025-10-03

### Added

- Initial project structure and documentation.
- Go module initialization.
- Basic "Hello, world!" main function.
