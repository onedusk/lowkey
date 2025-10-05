package watcher

// hybrid_monitor.go implements the algorithm described in docs/prds/algorithm_design.md,
// blending fsnotify events with safety scans. Profile for large directories.
