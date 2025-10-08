// Package reporting provides data structures and utilities for aggregating and
// summarizing file system change events. It is used to collect metrics about
// watcher activity, which can then be exposed through the CLI or other
// reporting mechanisms.
//
// The core components are the Aggregator, which collects events, and the
// Snapshot and Summary types, which provide different levels of detail about
// the collected data.
package reporting

import "time"

// Summary provides a high-level overview of watcher activity, suitable for
// display in CLI output. It includes the total number of changes and details
// about the most recent event.
type Summary struct {
	TotalChanges int
	LastEvent    *Change
	Window       time.Duration
}

// BuildSummary converts a detailed Snapshot into a high-level Summary. This is
// useful for presenting a concise overview of watcher activity to the user.
func BuildSummary(snapshot Snapshot, window time.Duration) Summary {
	return Summary{
		TotalChanges: snapshot.Count,
		LastEvent:    snapshot.LastChange,
		Window:       window,
	}
}
