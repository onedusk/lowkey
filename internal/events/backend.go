// backend.go defines the interface to platform-specific filesystem event APIs.
// Provide constructors that choose fsnotify variants based on GOOS.

package events

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lowkey/internal/state"
)

// Event represents a single file system notification.
type Event struct {
	Path      string
	Type      string
	Timestamp time.Time
}

const (
	// EventCreate denotes a newly observed file.
	EventCreate = "CREATE"
	// EventModify denotes a modification to an existing file.
	EventModify = "MODIFY"
	// EventDelete denotes removal of a previously observed file.
	EventDelete = "DELETE"
)

// Backend is the interface for a platform-specific file system watcher.
// It abstracts the underlying event mechanism (e.g., fsnotify, polling, etc.).
type Backend interface {
	// Events returns a channel that receives file system events.
	Events() <-chan Event

	// Errors returns a channel that receives errors from the watcher.
	Errors() <-chan error

	// Add starts watching the given path for changes.
	Add(path string) error

	// Remove stops watching the given path.
	Remove(path string) error

	// Close cleans up the watcher and closes the channels.
	Close() error
}

// NewBackend returns a polling backend that works across all supported
// platforms. The fsnotify implementations can replace this when available.
func NewBackend() (Backend, error) {
	return NewPollingBackend(1500 * time.Millisecond)
}

// pollingBackend implements Backend using periodic directory scans. While less
// efficient than native event APIs it delivers consistent behaviour across
// platforms without additional dependencies.
type pollingBackend struct {
	interval time.Duration
	events   chan Event
	errors   chan error

	mu      sync.RWMutex
	watched map[string]map[string]state.FileSignature
	stop    chan struct{}
	wg      sync.WaitGroup
}

// NewPollingBackend constructs a polling backend with the supplied interval.
func NewPollingBackend(interval time.Duration) (Backend, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	backend := &pollingBackend{
		interval: interval,
		events:   make(chan Event, 256),
		errors:   make(chan error, 1),
		watched:  make(map[string]map[string]state.FileSignature),
		stop:     make(chan struct{}),
	}
	backend.wg.Add(1)
	go backend.run()
	return backend, nil
}

func (p *pollingBackend) Events() <-chan Event {
	return p.events
}

func (p *pollingBackend) Errors() <-chan error {
	return p.errors
}

func (p *pollingBackend) Add(path string) error {
	clean, err := state.NormalizePath(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(clean)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("events: watch target must be a directory")
	}

	snapshot, err := p.snapshotDirectory(clean)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.watched[clean] = snapshot
	return nil
}

func (p *pollingBackend) Remove(path string) error {
	clean, err := state.NormalizePath(path)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.watched, clean)
	return nil
}

func (p *pollingBackend) Close() error {
	close(p.stop)
	p.wg.Wait()
	close(p.events)
	close(p.errors)
	return nil
}

func (p *pollingBackend) run() {
	defer p.wg.Done()
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.poll()
		case <-p.stop:
			return
		}
	}
}

func (p *pollingBackend) poll() {
	dirs := p.directories()
	for _, dir := range dirs {
		if err := p.pollDirectory(dir); err != nil {
			select {
			case p.errors <- err:
			default:
			}
		}
	}
}

func (p *pollingBackend) directories() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	dirs := make([]string, 0, len(p.watched))
	for dir := range p.watched {
		dirs = append(dirs, dir)
	}
	return dirs
}

func (p *pollingBackend) pollDirectory(dir string) error {
	current, err := p.snapshotDirectory(dir)
	if err != nil {
		return err
	}

	p.mu.Lock()
	previous := p.watched[dir]
	p.watched[dir] = current
	p.mu.Unlock()

	p.emitDiff(dir, previous, current)
	return nil
}

func (p *pollingBackend) snapshotDirectory(dir string) (map[string]state.FileSignature, error) {
	snapshot := make(map[string]state.FileSignature)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		sig, err := state.ComputeSignature(path, info)
		if err != nil {
			return err
		}
		snapshot[path] = sig
		return nil
	})
	return snapshot, err
}

func (p *pollingBackend) emitDiff(dir string, previous, current map[string]state.FileSignature) {
	now := time.Now().UTC()
	for path, sig := range current {
		old, ok := previous[path]
		if !ok {
			p.enqueue(Event{Path: path, Type: EventCreate, Timestamp: now})
			continue
		}
		if !old.Equal(sig) {
			p.enqueue(Event{Path: path, Type: EventModify, Timestamp: now})
		}
	}

	for path := range previous {
		if _, ok := current[path]; !ok {
			p.enqueue(Event{Path: path, Type: EventDelete, Timestamp: now})
		}
	}
}

func (p *pollingBackend) enqueue(event Event) {
	select {
	case p.events <- event:
	default:
		// Drop events when the consumer is slower; ensures the polling loop never blocks.
	}
}
