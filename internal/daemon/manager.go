package daemon

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"lowkey/internal/logging"
	"lowkey/internal/reporting"
	"lowkey/internal/state"
	"lowkey/internal/watcher"
	"lowkey/pkg/config"
)

// Manager coordinates watcher lifecycle and manifest persistence.
type Manager struct {
	store      *state.ManifestStore
	manifest   *config.Manifest
	controller *watcher.Controller
	aggregator *reporting.Aggregator
	logger     *logging.Logger
	mux        sync.Mutex
	running    bool
}

// NewManager creates a Manager for the provided manifest/store pair.
func NewManager(store *state.ManifestStore, manifest *config.Manifest) (*Manager, error) {
	if store == nil {
		return nil, errors.New("daemon: manifest store is required")
	}
	if manifest == nil {
		return nil, errors.New("daemon: manifest is required")
	}

	logDir := filepath.Dir(store.Path())
	rotator, err := logging.NewRotator(logDir, "lowkey.log", 10*1024*1024, 5)
	if err != nil {
		return nil, err
	}
	logger := logging.New(rotator)
	aggregator := reporting.NewAggregator()

	ctrl, err := watcher.NewController(watcher.ControllerConfig{
		Directories: manifest.Directories,
		IgnoreGlobs: nil,
		Aggregator:  aggregator,
		Logger:      logger,
	})
	if err != nil {
		return nil, err
	}

	return &Manager{
		store:      store,
		manifest:   manifest,
		controller: ctrl,
		aggregator: aggregator,
		logger:     logger,
	}, nil
}

// Start persists the manifest and launches the watcher controller.
func (m *Manager) Start() error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.running {
		return nil
	}

	if err := m.store.Save(m.manifest); err != nil {
		return fmt.Errorf("daemon: save manifest: %w", err)
	}
	if err := m.controller.Start(); err != nil {
		return err
	}
	if m.logger != nil {
		m.logger.Infof("daemon started with %d directories", len(m.manifest.Directories))
	}

	m.running = true
	return nil
}

// Stop halts the watcher and marks the manager as idle.
func (m *Manager) Stop() {
	m.mux.Lock()
	if !m.running {
		m.mux.Unlock()
		return
	}
	m.running = false
	m.mux.Unlock()

	m.controller.Stop()
	if m.logger != nil {
		m.logger.Info("daemon stopped")
	}
}

// Status reports the current run state and tracked directories.
func (m *Manager) Status() ManagerStatus {
	m.mux.Lock()
	defer m.mux.Unlock()

	dirs := make([]string, len(m.manifest.Directories))
	copy(dirs, m.manifest.Directories)

	snapshot := reporting.Snapshot{}
	if m.aggregator != nil {
		snapshot = m.aggregator.Snapshot()
	}

	return ManagerStatus{
		Running:      m.running,
		Directories:  dirs,
		ManifestPath: m.store.Path(),
		Summary:      reporting.BuildSummary(snapshot, 5*time.Minute),
	}
}

// ManagerStatus summarises daemon state for CLI commands.
type ManagerStatus struct {
	Running      bool
	Directories  []string
	ManifestPath string
	Summary      reporting.Summary
}
