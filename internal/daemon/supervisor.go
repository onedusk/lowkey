package daemon

import "sync"

// supervisor.go supervises goroutines for watchers and logging. Implement
// restart logic and health checks here, exercising them with integration tests.

// Heartbeat captures daemon liveness metadata.
type Heartbeat struct {
	Running bool
}

// Supervisor coordinates managers and periodically updates heartbeat state.
type Supervisor struct {
	heartbeat Heartbeat
	mux       sync.RWMutex
}

// SetRunning toggles the heartbeat flag.
func (s *Supervisor) SetRunning(running bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.heartbeat.Running = running
}

// Snapshot returns the latest heartbeat snapshot.
func (s *Supervisor) Snapshot() Heartbeat {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.heartbeat
}
