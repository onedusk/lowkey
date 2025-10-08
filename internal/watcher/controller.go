// Package watcher provides the core file system monitoring capabilities for
// lowkey. It is responsible for detecting file changes, handling ignore
// patterns, and reporting events to the rest of the application.
//
// The central component is the Controller, which manages the lifecycle of the
// monitoring process. It uses a HybridMonitor to combine real-time file system
// events with periodic safety scans, ensuring reliable change detection.
package watcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"lowkey/internal/events"
	"lowkey/internal/logging"
	"lowkey/internal/reporting"
	"lowkey/internal/state"
)

// Controller drives the hybrid monitoring loop, coordinating the event backend,
// incremental scans, and event batching. It provides a high-level interface
// for starting and stopping the file system watcher.
type Controller struct {
	config  ControllerConfig
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	backend events.Backend
	monitor *HybridMonitor
}

// ControllerConfig contains the dependencies and configuration required to run
// a watcher controller.
type ControllerConfig struct {
	Directories  []string
	IgnoreGlobs  []string
	Aggregator   *reporting.Aggregator
	Logger       *logging.Logger
	PollInterval time.Duration
	OnChange     func(reporting.Change)
}

// NewController validates the provided configuration and returns a new,
// ready-to-start controller.
func NewController(config ControllerConfig) (*Controller, error) {
	if len(config.Directories) == 0 {
		return nil, errors.New("watcher: controller requires at least one directory")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Controller{config: config, ctx: ctx, cancel: cancel}, nil
}

// Start launches the goroutines required to watch directories using the
// configured hybrid monitor. It initializes the event backend and the monitor,
// and starts the monitoring process.
func (c *Controller) Start() error {
	if c.ctx.Err() != nil {
		return fmt.Errorf("watcher: controller closed")
	}
	if len(c.config.IgnoreGlobs) > 0 && c.config.Logger != nil {
		c.config.Logger.Infof("watcher ignoring %d patterns", len(c.config.IgnoreGlobs))
	}
	backend, err := events.NewBackend()
	if err != nil {
		return err
	}
	cache := state.NewCache()
	monitor, err := NewHybridMonitor(HybridMonitorConfig{
		Backend:        backend,
		Cache:          cache,
		Aggregator:     c.config.Aggregator,
		Logger:         c.config.Logger,
		Directories:    c.config.Directories,
		PollInterval:   c.config.PollInterval,
		IgnorePatterns: c.config.IgnoreGlobs,
		OnChange:       c.config.OnChange,
	})
	if err != nil {
		_ = backend.Close()
		return err
	}
	c.backend = backend
	c.monitor = monitor
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		_ = monitor.Run(c.ctx)
	}()
	if c.config.Aggregator != nil {
		c.config.Aggregator.Record(reporting.Change{
			Path:      "(daemon startup)",
			Type:      "BOOT",
			Timestamp: time.Now().UTC(),
		})
	}
	if c.config.Logger != nil {
		c.config.Logger.Info("watcher controller started")
	}
	return nil
}

// Stop gracefully cancels the active monitoring goroutines and waits for them
// to shut down. This ensures a clean and orderly termination of the watcher.
func (c *Controller) Stop() {
	c.cancel()
	if c.backend != nil {
		_ = c.backend.Close()
	}
	c.wg.Wait()
	if c.config.Logger != nil {
		c.config.Logger.Info("watcher controller stopped")
	}
}
