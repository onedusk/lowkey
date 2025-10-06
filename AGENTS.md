# Repository Guidelines

## Project Structure & Module Organization
- `go.mod` anchors the module at the root, and `main.go` is the CLI entry point.
- CLI subcommands and wiring live in `cmd/`, while reusable logic stays under `internal/` and `pkg/`.
- Product docs sit in `docs/` (`docs/adrs/` for architecture choices, `docs/prds/` for requirements); mirror that layout when adding records.
- Generated binaries land in `./lowkey`; keep temporary fixtures in `testdata/` and reference shared assets from `third_party/` as needed.

## Build, Test, and Development Commands
- `make build` compiles the CLI into `./lowkey` and should precede packaging work.
- `make run` executes the CLI with dev-friendly defaults for quick smoke checks.
- `make test` runs `go test -v ./...`; prefer it over ad‑hoc commands so CI mirrors local runs.
- `make clean` clears the compiled binary; rerun before benchmarking to avoid stale artifacts.

## Coding Style & Naming Conventions
- Target Go 1.24.0, format every change with `gofmt`, and rely on tabs plus Unix newlines.
- Keep imports grouped standard/third-party/internal; `goimports` is recommended during save hooks.
- Exported symbols use concise PascalCase, unexported helpers use camelCase, and file names follow the feature (`watcher.go`, `watcher_test.go`).
- Add brief comments only for behavior that is not obvious from the code.

## Testing Guidelines
- Place unit tests beside their sources using the `*_test.go` suffix and Go’s `testing` package.
- Favor table-driven tests named `TestFeature_Scenario` for clarity and coverage tracking.
- Use `make test` before commits; annotate coverage deltas in PRs when behavior changes.
- Capture multi-platform watcher quirks with integration tests and document caveats in `docs/adrs/`.

## Commit & Pull Request Guidelines
- Follow the imperative commit voice seen in history (`added .github`, `first commit`) and stay under 72 characters.
- Provide context in bodies when choices are non-obvious; link ADRs/PRDs updated in the change.
- Open PRs with problem, user-visible impact, and reproduction notes; attach CLI output or logs when valuable.
- Request reviews from maintainers responsible for touched directories and highlight any new configuration knobs.

## Configuration & Operational Tips
- Record new ignore patterns or watch targets in `.lowkey` and justify them in your PR.
- Document cross-platform assumptions immediately in `docs/adrs/` so downstream agents inherit the rationale.
- When adding tooling, prefer Go-native deps; vendored assets belong under `third_party/` with licensing notes.
