# Repository Guidelines

## Project Structure & Module Organization
The Go module lives at the repo root (`go.mod`) with `main.go` supplying the current CLI entry point. Supporting product and governance files sit under `docs/`, with architectural decisions in `docs/adrs/` and product requirements in `docs/prds/`; mirror that layout when adding new records. Build artifacts default to `./lowkey`, and ignore patterns belong in a `.lowkey` file alongside the watched directories you configure.

## Build, Test, and Development Commands
Use the Makefile shortcuts while iterating:

```sh
make build   # compile lowkey into ./lowkey
make run     # run the CLI in development mode
make test    # execute go test -v ./...
make clean   # remove the compiled binary
```

Direct `go build ./...` and `go run main.go` work when scripting, but keep Makefile targets as the documented interface.

## Coding Style & Naming Conventions
Target Go 1.24.0 and format every change with `gofmt` (tabs for indentation, Unix newlines). Import ordering should follow `gofmt`/`goimports`, with standard library, third-party, and internal packages separated by blank lines. Expose packages and functions with concise PascalCase names and keep files named after the package or feature (e.g., `watcher.go`, `watcher_test.go`).

## Testing Guidelines
Place tests beside their sources using the `*_test.go` suffix and Go’s `testing` package. Prefer table-driven tests with descriptive `TestFeature_Scenario` names, and add integration tests whenever filesystem watchers span multiple platforms. Run `make test` before opening a pull request and ensure new behavior ships with meaningful coverage notes in the PR body.

## Commit & Pull Request Guidelines
Follow the existing imperative commit voice (`added .github`, `first commit`) and keep subjects under 72 characters; add focused bodies when context is non-obvious. Each pull request should summarize the problem, list user-visible changes, and mention related ADRs/PRDs that were touched. Include CLI output or logs when they help reviewers reproduce results, and request review from maintainers responsible for the impacted directory.

## Configuration & Operational Tips
Document new ignore patterns or watch targets inside `.lowkey` and explain them in the PR when they impact default behavior. When introducing cross-platform logic, capture any platform-specific caveats in `docs/adrs/` so future agents understand the trade-offs.
