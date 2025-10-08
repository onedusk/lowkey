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
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"lowkey/internal/events"
	"lowkey/internal/filters"
	"lowkey/internal/logging"
	"lowkey/internal/reporting"
	"lowkey/internal/state"
)

// HybridMonitor coordinates real-time file system events with periodic safety
// scans to provide resilient and reliable change detection. It is designed to
// catch events that might be missed by the real-time event backend.
type HybridMonitor struct {
	backend        events.Backend
	cache          *state.Cache
	aggregator     *reporting.Aggregator
	logger         *logging.Logger
	directories    []string
	pollInterval   time.Duration
	ignorePatterns []string
	ignoreBloom    *filters.BloomFilter
	changeHandler  func(reporting.Change)
}

// HybridMonitorConfig encapsulates the dependencies and configuration required
// to create a HybridMonitor.
type HybridMonitorConfig struct {
	Backend        events.Backend
	Cache          *state.Cache
	Aggregator     *reporting.Aggregator
	Logger         *logging.Logger
	Directories    []string
	PollInterval   time.Duration
	IgnorePatterns []string
	OnChange       func(reporting.Change)
}

// NewHybridMonitor validates the provided configuration and constructs a new
// HybridMonitor. It sets up the necessary components, including the event
// backend, cache, and ignore pattern filters.
func NewHybridMonitor(cfg HybridMonitorConfig) (*HybridMonitor, error) {
	if len(cfg.Directories) == 0 {
		return nil, fmt.Errorf("watcher: hybrid monitor requires directories to watch")
	}

	backend := cfg.Backend
	if backend == nil {
		var err error
		backend, err = events.NewBackend()
		if err != nil {
			return nil, err
		}
	}

	cache := cfg.Cache
	if cache == nil {
		cache = state.NewCache()
	}

	pollInterval := cfg.PollInterval
	if pollInterval <= 0 {
		pollInterval = 30 * time.Second
	}

	patterns := make([]string, 0, len(cfg.IgnorePatterns))
	for _, pattern := range cfg.IgnorePatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern != "" {
			patterns = append(patterns, pattern)
		}
	}
	var bloom *filters.BloomFilter
	if len(patterns) > 0 {
		bloom = filters.NewBloomFilter(len(patterns)*8, 0.01)
		for _, pattern := range patterns {
			for _, token := range filters.ExtractPatternTokens(pattern) {
				bloom.Add(token)
			}
		}
	}

	return &HybridMonitor{
		backend:        backend,
		cache:          cache,
		aggregator:     cfg.Aggregator,
		logger:         cfg.Logger,
		directories:    cfg.Directories,
		pollInterval:   pollInterval,
		ignorePatterns: patterns,
		ignoreBloom:    bloom,
		changeHandler:  cfg.OnChange,
	}, nil
}

// Run starts the hybrid monitoring process and blocks until the provided context
// is canceled. It launches goroutines for consuming real-time events and
// performing periodic safety scans.
func (m *HybridMonitor) Run(ctx context.Context) error {
	for _, dir := range m.directories {
		if err := m.backend.Add(dir); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		m.consumeEvents(ctx)
	}()

	go func() {
		defer wg.Done()
		m.safetyScanLoop(ctx)
	}()

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (m *HybridMonitor) consumeEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-m.backend.Events():
			if !ok {
				return
			}
			m.handleEvent(event)
		case err, ok := <-m.backend.Errors():
			if !ok {
				continue
			}
			if m.logger != nil {
				m.logger.Errorf("event backend error: %v", err)
			}
		}
	}
}

func (m *HybridMonitor) safetyScanLoop(ctx context.Context) {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.performSafetyScan()
		}
	}
}

func (m *HybridMonitor) performSafetyScan() {
	for _, dir := range m.directories {
		if err := m.scanDirectory(dir); err != nil && m.logger != nil {
			m.logger.Errorf("safety scan error: %v", err)
		}
	}
}

func (m *HybridMonitor) handleEvent(event events.Event) {
	if m.shouldIgnore(event.Path) {
		return
	}

	switch event.Type {
	case events.EventDelete:
		m.cache.Delete(event.Path)
		m.recordChange(event.Path, events.EventDelete, event.Timestamp)
	case events.EventCreate, events.EventModify:
		info, err := os.Stat(event.Path)
		if err != nil {
			if os.IsNotExist(err) {
				m.cache.Delete(event.Path)
				m.recordChange(event.Path, events.EventDelete, event.Timestamp)
			}
			return
		}

		sig, err := state.ComputeSignature(event.Path, info)
		if err != nil {
			if m.logger != nil {
				m.logger.Errorf("compute signature: %v", err)
			}
			return
		}

		prev, ok := m.cache.Get(event.Path)
		m.cache.Set(event.Path, sig)
		if !ok {
			m.recordChange(event.Path, events.EventCreate, event.Timestamp)
			return
		}
		if !prev.Equal(sig) {
			m.recordChange(event.Path, events.EventModify, event.Timestamp)
		}
	default:
		m.recordChange(event.Path, event.Type, event.Timestamp)
	}
}

func (m *HybridMonitor) scanDirectory(dir string) error {
	reference := m.cache.FilesUnder(dir)
	seen := make(map[string]struct{}, len(reference))

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if m.shouldIgnore(path) {
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
		seen[path] = struct{}{}

		cached, ok := reference[path]
		m.cache.Set(path, sig)
		if !ok {
			m.recordChange(path, events.EventCreate, time.Now().UTC())
			return nil
		}
		if !cached.Equal(sig) {
			m.recordChange(path, events.EventModify, time.Now().UTC())
		}
		return nil
	})
	if err != nil {
		return err
	}

	for path := range reference {
		if _, ok := seen[path]; ok {
			continue
		}
		m.cache.Delete(path)
		m.recordChange(path, events.EventDelete, time.Now().UTC())
	}

	return nil
}

func (m *HybridMonitor) recordChange(path, changeType string, timestamp time.Time) {
	change := reporting.Change{Path: path, Type: changeType, Timestamp: timestamp}
	if m.aggregator != nil {
		m.aggregator.Record(change)
	}
	if m.logger != nil {
		m.logger.Infof("%s %s", changeType, path)
	}
	if m.changeHandler != nil {
		m.changeHandler(change)
	}
}

func (m *HybridMonitor) shouldIgnore(path string) bool {
	if len(m.ignorePatterns) == 0 {
		return false
	}

	tokens := filters.ExtractPathTokens(path)
	bloomMatch := false
	if m.ignoreBloom == nil {
		bloomMatch = true
	} else {
		for _, token := range tokens {
			if m.ignoreBloom.Contains(token) {
				bloomMatch = true
				break
			}
		}
	}

	if !bloomMatch {
		return false
	}

	normalized := filepath.ToSlash(path)
	base := filepath.Base(normalized)

	for _, pattern := range m.ignorePatterns {
		if matchPattern(pattern, normalized, base) {
			return true
		}
	}

	return false
}

func matchPattern(pattern, fullPath, base string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}

	normPattern := filepath.ToSlash(pattern)

	if strings.Contains(normPattern, "**") {
		prefix := strings.TrimSuffix(normPattern, "**")
		if prefix == "" || strings.HasPrefix(fullPath, prefix) {
			return true
		}
	}

	if ok, _ := pathpkg.Match(normPattern, fullPath); ok {
		return true
	}
	if ok, _ := filepath.Match(pattern, base); ok {
		return true
	}
	return false
}
