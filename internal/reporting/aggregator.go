package reporting

import (
	"path/filepath"
	"sync"
	"time"
)

// Change describes a single filesystem change event.
type Change struct {
	Path      string
	Type      string
	Timestamp time.Time
}

// Snapshot summarises recent activity.
type Snapshot struct {
	Count        int
	LastChange   *Change
	PerDirectory map[string]int
}

// Aggregator collects change events for later reporting.
type Aggregator struct {
	mu       sync.Mutex
	snapshot Snapshot
}

// NewAggregator constructs a new aggregator instance.
func NewAggregator() *Aggregator {
	return &Aggregator{snapshot: Snapshot{PerDirectory: make(map[string]int)}}
}

// Record adds a change to the rolling snapshot.
func (a *Aggregator) Record(change Change) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.snapshot.Count++
	copyChange := change
	a.snapshot.LastChange = &copyChange
	dir := filepath.Dir(change.Path)
	a.snapshot.PerDirectory[dir]++
}

// Snapshot returns a copy of the aggregate state for safe consumption.
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
