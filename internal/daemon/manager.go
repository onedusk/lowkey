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
	"context"
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
	"lowkey/pkg/telemetry"
)

// Manager coordinates the watcher lifecycle, manifest persistence, and logging.
// It acts as the central orchestrator for the daemon, handling the startup and
// shutdown of the file system monitoring process. It is safe for concurrent use.
type Manager struct {
	store      *state.ManifestStore
	manifest   *config.Manifest
	controller *watcher.Controller
	aggregator *reporting.Aggregator
	logger     *logging.Logger
	mux        sync.Mutex
	running    bool
	metrics    *telemetry.Collector
	tracer     *telemetry.Tracer
	supervisor *Supervisor
}

// NewManager creates a new Manager for the provided manifest and store.
// It initializes all necessary components, including the logger, aggregator,
// and watcher controller, preparing the manager to start monitoring.
func NewManager(store *state.ManifestStore, manifest *config.Manifest) (*Manager, error) {
	if store == nil {
		return nil, errors.New("daemon: manifest store is required")
	}
	if manifest == nil {
		return nil, errors.New("daemon: manifest is required")
	}

	logDir := filepath.Dir(store.Path())
	logName := "lowkey.log"
	if manifest.LogPath != "" {
		logDir = filepath.Dir(manifest.LogPath)
		logName = filepath.Base(manifest.LogPath)
	}
	rotator, err := logging.NewRotator(logDir, logName, 10*1024*1024, 5)
	if err != nil {
		return nil, err
	}
	logger := logging.New(rotator)
	aggregator := reporting.NewAggregator()
	ignorePatterns, err := resolveIgnorePatterns(manifest)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		store:      store,
		manifest:   manifest,
		aggregator: aggregator,
		logger:     logger,
	}

	ctrl, err := watcher.NewController(watcher.ControllerConfig{
		Directories:  manifest.Directories,
		IgnoreGlobs:  ignorePatterns,
		Aggregator:   aggregator,
		Logger:       logger,
		PollInterval: 30 * time.Second,
		OnChange:     m.handleChange,
	})
	if err != nil {
		return nil, err
	}
	m.controller = ctrl
	m.supervisor = NewSupervisor(m, 5*time.Second)
	return m, nil
}

func resolveIgnorePatterns(manifest *config.Manifest) ([]string, error) {
	if manifest == nil || manifest.IgnoreFile == "" {
		return nil, nil
	}
	patterns, err := config.LoadIgnorePatterns(manifest.IgnoreFile)
	if err != nil {
		return nil, fmt.Errorf("daemon: load ignore patterns: %w", err)
	}
	return patterns, nil
}

// Start persists the manifest and launches the watcher controller and supervisor.
// This method is idempotent and will not restart the manager if it is already
// running. It is the primary entry point for activating the daemon's monitoring
// functionality.
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
	if m.supervisor != nil {
		m.supervisor.Start()
	}

	m.running = true
	return nil
}

// Stop halts the watcher and supervisor, marking the manager as idle.
// This method provides a graceful shutdown of the daemon's monitoring activities.
func (m *Manager) Stop() {
	m.mux.Lock()
	if !m.running {
		m.mux.Unlock()
		return
	}
	m.running = false
	m.mux.Unlock()

	m.controller.Stop()
	if m.supervisor != nil {
		m.supervisor.Stop()
	}
	if m.logger != nil {
		m.logger.Info("daemon stopped")
	}
}

// SetTelemetry attaches metrics and tracer instances to the manager, enabling
// observability features. This allows the manager to report performance
// metrics and trace information.
func (m *Manager) SetTelemetry(metrics *telemetry.Collector, tracer *telemetry.Tracer) {
	m.metrics = metrics
	m.tracer = tracer
}

// Status reports the current run state, tracked directories, and other
// diagnostic information. It provides a snapshot of the manager's current
// operational status, which is useful for CLI commands and health checks.
func (m *Manager) Status() ManagerStatus {
	m.mux.Lock()
	defer m.mux.Unlock()

	dirs := make([]string, len(m.manifest.Directories))
	copy(dirs, m.manifest.Directories)

	snapshot := reporting.Snapshot{}
	if m.aggregator != nil {
		snapshot = m.aggregator.Snapshot()
	}

	heartbeat := Heartbeat{}
	if m.supervisor != nil {
		heartbeat = m.supervisor.Snapshot()
	}

	return ManagerStatus{
		Running:      m.running,
		Directories:  dirs,
		ManifestPath: m.store.Path(),
		Summary:      reporting.BuildSummary(snapshot, 5*time.Minute),
		Heartbeat:    heartbeat,
	}
}

func (m *Manager) handleChange(change reporting.Change) {
	if m.metrics != nil {
		m.metrics.IncEvent()
	}
	if m.tracer != nil && m.tracer.Enabled() {
		span, _ := m.tracer.StartSpan(context.Background(), "watcher.change")
		span.SetAttribute("path", change.Path)
		span.SetAttribute("type", change.Type)
		span.End(nil)
	}
}

// ManagerStatus summarises the daemon's state for CLI commands. It provides a
// snapshot of the daemon's operational status, including its running state,
// watched directories, and performance metrics.
type ManagerStatus struct {
	Running      bool
	Directories  []string
	ManifestPath string
	Summary      reporting.Summary
	Heartbeat    Heartbeat
}
