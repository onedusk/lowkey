// Package daemon implements the core logic for the lowkey background process.
// It manages the lifecycle of the file system watcher, handles manifest
// persistence and reconciliation, and coordinates with other components like
// logging and telemetry.
//
// The central component is the Manager, which orchestrates the daemon's
// operations. It is supervised by a Supervisor that ensures the daemon
// remains running and automatically restarts it on failure.
package daemon

import (
	"fmt"
	"sort"
	"time"

	"lowkey/internal/watcher"
	"lowkey/pkg/config"
)

// ManifestDiff describes the additions and removals discovered during a manifest
// reconciliation. It provides a structured way to represent changes between
// two versions of a daemon manifest.
type ManifestDiff struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
}

// IsEmpty reports whether the diff contains any changes. This is a convenient
// way to check if a reconciliation resulted in any modifications.
func (d ManifestDiff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0
}

// DiffManifests computes the delta between the current and desired manifests.
// It identifies which directories have been added or removed, returning a
// ManifestDiff that represents these changes.
func DiffManifests(current, desired *config.Manifest) ManifestDiff {
	diff := ManifestDiff{}

	currentSet := make(map[string]struct{})
	if current != nil {
		for _, dir := range current.Directories {
			currentSet[dir] = struct{}{}
		}
	}

	desiredSet := make(map[string]struct{})
	if desired != nil {
		for _, dir := range desired.Directories {
			desiredSet[dir] = struct{}{}
		}
	}

	for dir := range desiredSet {
		if _, ok := currentSet[dir]; !ok {
			diff.Added = append(diff.Added, dir)
		}
	}
	for dir := range currentSet {
		if _, ok := desiredSet[dir]; !ok {
			diff.Removed = append(diff.Removed, dir)
		}
	}

	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)
	return diff
}

// ReconcileManifest reloads the persisted manifest from the store and applies
// any changes to the running daemon. If the on-disk manifest differs from the
// in-memory one, this function will reconfigure the watcher to match the new
// desired state. It returns a diff of the changes and any error encountered.
func (m *Manager) ReconcileManifest() (ManifestDiff, error) {
	if m == nil {
		return ManifestDiff{}, fmt.Errorf("daemon: manager is nil")
	}

	desired, err := m.store.Load()
	if err != nil {
		return ManifestDiff{}, err
	}
	if desired == nil {
		return ManifestDiff{}, nil
	}

	diff := DiffManifests(m.manifest, desired)
	if diff.IsEmpty() {
		return diff, nil
	}

	if err := m.applyManifest(desired, diff); err != nil {
		return diff, err
	}
	return diff, nil
}

func (m *Manager) applyManifest(manifest *config.Manifest, diff ManifestDiff) error {
	if manifest == nil {
		return fmt.Errorf("daemon: manifest cannot be nil")
	}

	ignorePatterns, err := resolveIgnorePatterns(manifest)
	if err != nil {
		return err
	}

	ctrl, err := watcher.NewController(watcher.ControllerConfig{
		Directories:  manifest.Directories,
		IgnoreGlobs:  ignorePatterns,
		Aggregator:   m.aggregator,
		Logger:       m.logger,
		PollInterval: 30 * time.Second,
		OnChange:     m.handleChange,
	})
	if err != nil {
		return err
	}

	m.mux.Lock()
	oldController := m.controller
	oldManifest := m.manifest
	wasRunning := m.running
	m.controller = ctrl
	m.manifest = manifest
	m.mux.Unlock()

	if oldController != nil {
		oldController.Stop()
	}

	if wasRunning {
		if err := ctrl.Start(); err != nil {
			m.mux.Lock()
			m.controller = oldController
			m.manifest = oldManifest
			m.mux.Unlock()
			if oldController != nil {
				if restartErr := oldController.Start(); restartErr != nil && m.logger != nil {
					m.logger.Errorf("daemon: failed to restart previous controller after reconciliation error: %v", restartErr)
				}
			}
			return err
		}
	}

	if err := m.store.Save(manifest); err != nil {
		return err
	}

	if m.logger != nil {
		m.logger.Infof("daemon reconciled manifest: added=%d removed=%d", len(diff.Added), len(diff.Removed))
	}
	return nil
}
