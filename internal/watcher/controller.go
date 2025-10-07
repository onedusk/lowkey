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

// Controller drives the hybrid monitoring loop, coordinating events,
// incremental scans, and batching.
type Controller struct {
	config  ControllerConfig
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	backend events.Backend
	monitor *HybridMonitor
}

// ControllerConfig contains dependencies required to run a watcher instance.
type ControllerConfig struct {
	Directories  []string
	IgnoreGlobs  []string
	Aggregator   *reporting.Aggregator
	Logger       *logging.Logger
	PollInterval time.Duration
	OnChange     func(reporting.Change)
}

// NewController validates configuration and returns a ready-to-start controller.
func NewController(config ControllerConfig) (*Controller, error) {
	if len(config.Directories) == 0 {
		return nil, errors.New("watcher: controller requires at least one directory")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Controller{config: config, ctx: ctx, cancel: cancel}, nil
}

// Start boots goroutines required to watch directories using the configured
// hybrid monitor.
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

// Stop cancels active goroutines and waits for them to finish.
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
