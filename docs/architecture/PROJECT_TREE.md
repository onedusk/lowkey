# Planned Project Tree

```text
lowkey/
├── cmd/
│   └── lowkey/
│       ├── main.go
│       ├── root.go
│       ├── watch.go
│       ├── start.go
│       ├── stop.go
│       ├── status.go
│       ├── log.go
│       ├── tail.go
│       ├── summary.go
│       └── clear.go
├── internal/
│   ├── daemon/
│   │   ├── manager.go
│   │   ├── supervisor.go
│   │   └── watch_manifest.go
│   ├── events/
│   │   ├── backend.go
│   │   ├── fsnotify_unix.go
│   │   └── fsnotify_windows.go
│   ├── filters/
│   │   ├── bloom_filter.go
│   │   └── ignore_tokens.go
│   ├── logging/
│   │   ├── rotator.go
│   │   └── writer.go
│   ├── state/
│   │   ├── cache.go
│   │   ├── persistence.go
│   │   └── manifest_store.go
│   ├── reporting/
│   │   ├── aggregator.go
│   │   └── summaries.go
│   └── watcher/
│       ├── controller.go
│       └── hybrid_monitor.go
├── pkg/
│   ├── config/
│   │   ├── loader.go
│   │   └── schema.go
│   ├── output/
│   │   ├── formatter.go
│   │   └── table.go
│   └── telemetry/
│       ├── metrics.go
│       └── tracing.go
├── scripts/
│   ├── install.sh
│   ├── package.ps1
│   └── services/
│       ├── lowkey.plist
│       └── lowkey.service
├── testdata/
│   ├── fixtures/
│   │   ├── nested_changes
│   │   └── large_directory
│   └── manifests/
│       └── sample_daemon.json
├── docs/
│   ├── CHANGELOG.md
│   ├── CODE_OF_CONDUCT.md
│   ├── CONTRIBUTING.md
│   ├── LICENSE
│   ├── VERSION
│   ├── security.md
│   ├── adrs/
│   │   └── 0001-use-go-for-development.md
│   ├── prds/
│   │   ├── PRD.md
│   │   └── algorithm_design.md
│   ├── guides/
│   │   └── daemon.md
│   └── api/
│       └── cli_reference.md
├── build/
│   ├── artifacts/
│   └── release_notes/
├── .lowkey
├── Makefile
├── go.mod
└── go.sum
```
