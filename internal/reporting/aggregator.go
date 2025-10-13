// Package reporting provides data structures and utilities for aggregating and
// summarizing file system change events. It is used to collect metrics about
// watcher activity, which can then be exposed through the CLI or other
// reporting mechanisms.
//
// The core components are the Aggregator, which collects events, and the
// Snapshot and Summary types, which provide different levels of detail about
// the collected data.
package reporting

import (
	"path/filepath"
	"sync"
	"time"
)

// Change describes a single file system change event, including the path, type
// of change, and when it occurred.
type Change struct {
	Path      string
	Type      string
	Timestamp time.Time
	Size      int64 // Size for new files, or new size for modified files
	OldSize   int64 // Previous size for modified files (used to calculate delta)
	SizeDelta int64 // Size change for modified files (positive for growth, negative for shrink)
}

// Snapshot provides a detailed summary of recent watcher activity. It includes
// the total number of changes, details of the last change, and a breakdown of
// changes per directory.
type Snapshot struct {
	Count        int
	LastChange   *Change
	PerDirectory map[string]int
}

// Aggregator collects and summarizes file system change events. It maintains a
// running snapshot of activity, which can be retrieved for reporting. It is
// safe for concurrent use.
type Aggregator struct {
	mu       sync.Mutex
	snapshot Snapshot
}

// NewAggregator constructs a new, empty Aggregator instance, ready to start
// collecting change events.
func NewAggregator() *Aggregator {
	return &Aggregator{snapshot: Snapshot{PerDirectory: make(map[string]int)}}
}

// Record adds a new change event to the aggregator's snapshot. It updates the
// total count, tracks the last change, and increments the count for the
// relevant directory.
func (a *Aggregator) Record(change Change) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.snapshot.Count++
	copyChange := change
	a.snapshot.LastChange = &copyChange
	dir := filepath.Dir(change.Path)
	a.snapshot.PerDirectory[dir]++
}

// Snapshot returns a thread-safe copy of the current aggregate state. This
// allows other parts of the application to access the summary data without
// needing to worry about race conditions.
func (a *Aggregator) Snapshot() Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()

	snapshot := a.snapshot
	if snapshot.PerDirectory != nil {
		perDir := make(map[string]int, len(snapshot.PerDirectory))
		for k, v := range snapshot.PerDirectory {
			perDir[k] = v
		}
		snapshot.PerDirectory = perDir
	}
	if snapshot.LastChange != nil {
		changeCopy := *snapshot.LastChange
		snapshot.LastChange = &changeCopy
	}
	return snapshot
}
