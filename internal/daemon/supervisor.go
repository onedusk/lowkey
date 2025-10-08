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
	"sync"
	"time"
)

// Heartbeat captures daemon liveness metadata for CLI consumers. It includes
// information about the daemon's running state, health checks, and any errors
// or restarts, providing a quick overview of the daemon's stability.
type Heartbeat struct {
	Running      bool      `json:"running"`
	LastCheck    time.Time `json:"last_check"`
	LastChange   time.Time `json:"last_change"`
	Restarts     int       `json:"restarts"`
	LastError    string    `json:"last_error,omitempty"`
	BackoffUntil time.Time `json:"backoff_until,omitempty"`
}

// Supervisor monitors the daemon manager and restarts it if it becomes
// unresponsive or stops unexpectedly. It provides a layer of resilience,
// ensuring that the file monitoring service remains available.
type Supervisor struct {
	manager  *Manager
	interval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mux       sync.RWMutex
	heartbeat Heartbeat
	started   bool
}

// NewSupervisor constructs a supervisor that probes the provided manager at the
// specified interval. If the manager is nil, a no-op supervisor is created.
func NewSupervisor(manager *Manager, interval time.Duration) *Supervisor {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Supervisor{
		manager:   manager,
		interval:  interval,
		heartbeat: Heartbeat{LastCheck: time.Now(), LastChange: time.Now()},
	}
}

// Start launches the supervision loop in a new goroutine. The call is
// idempotent; if the supervisor is already running, this method has no effect.
func (s *Supervisor) Start() {
	if s == nil || s.manager == nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	s.mux.Lock()
	if s.started {
		s.mux.Unlock()
		cancel()
		return
	}
	s.ctx = ctx
	s.cancel = cancel
	s.started = true
	s.mux.Unlock()

	s.wg.Add(1)
	go s.loop(ctx)
}

// Stop halts the supervision loop and waits for it to shut down gracefully.
// This ensures that the supervisor's goroutine is properly terminated.
func (s *Supervisor) Stop() {
	if s == nil {
		return
	}

	s.mux.Lock()
	if !s.started {
		s.mux.Unlock()
		return
	}
	cancel := s.cancel
	s.started = false
	s.cancel = nil
	s.ctx = nil
	s.mux.Unlock()

	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
}

func (s *Supervisor) loop(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			supervisorErr := s.probe()
			if supervisorErr == nil {
				backoff = time.Second
				continue
			}

			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			s.setBackoff(time.Now().Add(backoff))
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

func (s *Supervisor) probe() error {
	s.updateHeartbeat(func(h *Heartbeat) {
		h.LastCheck = time.Now()
		h.LastError = ""
		h.BackoffUntil = time.Time{}
	})

	status := s.manager.Status()
	if status.Running {
		s.updateHeartbeat(func(h *Heartbeat) {
			if !h.Running {
				h.Running = true
				h.LastChange = time.Now()
			}
		})
		return nil
	}

	// Attempt a restart when the manager reports not running.
	if err := s.manager.Start(); err != nil {
		s.updateHeartbeat(func(h *Heartbeat) {
			h.Running = false
			h.LastError = err.Error()
		})
		return err
	}

	s.updateHeartbeat(func(h *Heartbeat) {
		h.Running = true
		h.Restarts++
		h.LastChange = time.Now()
	})
	return nil
}

// Snapshot returns a copy of the latest heartbeat, providing a thread-safe way
// to access the supervisor's status information.
func (s *Supervisor) Snapshot() Heartbeat {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.heartbeat
}

func (s *Supervisor) updateHeartbeat(mutator func(*Heartbeat)) {
	s.mux.Lock()
	defer s.mux.Unlock()
	mutator(&s.heartbeat)
}

func (s *Supervisor) setBackoff(until time.Time) {
	s.updateHeartbeat(func(h *Heartbeat) {
		h.BackoffUntil = until
	})
}
