package watcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"lowkey/internal/logging"
	"lowkey/internal/reporting"
)

// Controller drives the hybrid monitoring loop, coordinating events,
// incremental scans, and batching.
type Controller struct {
	config ControllerConfig
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// ControllerConfig contains dependencies required to run a watcher instance.
type ControllerConfig struct {
    Directories []string
    IgnoreGlobs []string
    Aggregator  *reporting.Aggregator
    Logger      *logging.Logger
}

// NewController validates configuration and returns a ready-to-start controller.
func NewController(config ControllerConfig) (*Controller, error) {
	if len(config.Directories) == 0 {
		return nil, errors.New("watcher: controller requires at least one directory")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Controller{config: config, ctx: ctx, cancel: cancel}, nil
}

// Start boots goroutines required to watch directories. The current stub simply
// records that the watcher is “running”; future iterations will wire fsnotify
// backends and polling loops.
func (c *Controller) Start() error {
    if c.ctx.Err() != nil {
        return fmt.Errorf("watcher: controller closed")
    }
    if len(c.config.IgnoreGlobs) > 0 && c.config.Logger != nil {
        c.config.Logger.Infof("watcher ignoring %d patterns", len(c.config.IgnoreGlobs))
    }
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
	// TODO: spawn event + polling goroutines; integrate filters and reporters.
	// - Create a new event backend (from internal/events).
	// - Start a goroutine to listen for events from the backend and dispatch them.
	// - Start a polling goroutine for the incremental safety scan.
	// - This polling loop should use the cache (from internal/state) to check for
	//   missed events.
	// - Integrate the filter chain (from internal/filters) to ignore irrelevant events.
	// - Pass the filtered events to the aggregator (from internal/reporting).
	return nil
}

// Stop cancels active goroutines and waits for them to finish.
func (c *Controller) Stop() {
	c.cancel()
	c.wg.Wait()
	if c.config.Logger != nil {
		c.config.Logger.Info("watcher controller stopped")
	}
}
