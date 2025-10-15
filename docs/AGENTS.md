# Repository Guidelines

## Project Structure & Module Organization

`go.mod` anchors the module at the repo root, and `main.go` wires the CLI entrypoint. Subcommands and Cobra wiring live in `cmd/`, while reusable business logic sits in `internal/` and exported packages in `pkg/`. Generated binaries land in `./lowkey`; keep scratch fixtures under `testdata/` and place third-party assets with licenses in `third_party/`. Product documentation belongs in `docs/`, with decisions in `docs/adrs/` and requirements in `docs/prds/`.

## Build, Test, and Development Commands

Use `make build` to compile the CLI into `./lowkey/lowkey`; run it with `make run` for a smoke test using dev defaults. `make test` executes `go test -v ./...` so local runs mirror CI. When benchmarking or after major refactors, clear artifacts with `make clean` before rebuilding.

## Coding Style & Naming Conventions

Target Go 1.24.0. Format every change with `gofmt` (tabs, Unix newlines) or `goimports` to keep import groups ordered standard/third-party/internal. Exported API types use PascalCase, helpers stay camelCase, and files match their feature (`watcher.go`, `watcher_test.go`). Add comments only where behavior is non-obvious or platform-specific.

## Testing Guidelines

Keep unit tests beside their sources named `*_test.go`, using table-driven `TestFeature_Scenario` patterns. Run `make test` before every commit and call out coverage or behavior deltas in PR notes. Capture multi-platform watcher edge cases with focused integration tests and document any portability caveats in `docs/adrs/`.

## Engineering Requirements Overview

Lowkey must deliver event-driven directory monitoring via `fsnotify` with periodic safety scans, honor `.lowkey` ignore semantics, and supervise a single daemon per user. Persist daemon state in `$XDG_STATE_HOME/lowkey/daemon.json`, emit structured JSON logs rotated at 10 MB (retain five), and ensure parity across macOS, Windows, and Linux for CLI commands such as `watch`, `start`, `stop`, `status`, and `log`.

## Commit & Pull Request Guidelines

Write imperative, ≤72-character commit subjects (`add watcher retry`, `fix cli prompt`). Include rationale in commit bodies when behavior changes, and link ADRs or PRDs you update. PRs should explain the problem, user-visible impact, and reproduction steps; attach CLI output or logs when it helps reviewers. Request reviews from maintainers of the touched directories and flag any new configuration knobs.

## Configuration & Operational Tips

Log new ignore patterns or watch targets in `.lowkey` alongside a short explanation. Prefer Go-native tooling; when outside dependencies are unavoidable, vendor them under `third_party/` with licensing notes. Document cross-platform assumptions immediately so downstream agents inherit the rationale.
