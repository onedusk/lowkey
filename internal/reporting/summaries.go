package reporting

import "time"

// Summary provides high-level metrics for CLI output.
type Summary struct {
	TotalChanges int
	LastEvent    *Change
	Window       time.Duration
}

// BuildSummary converts a snapshot to a Summary.
func BuildSummary(snapshot Snapshot, window time.Duration) Summary {
	return Summary{
		TotalChanges: snapshot.Count,
		LastEvent:    snapshot.LastChange,
		Window:       window,
	}
}
