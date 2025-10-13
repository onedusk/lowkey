// Package watcher provides logging functionality for the watch command to
// create and maintain .lowlog directories with change event logs.
package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lowkey/internal/reporting"
)

// WatchLogger handles logging of file system changes to .lowlog directories
// within each watched directory. It creates date-based log files and ensures
// thread-safe writes.
type WatchLogger struct {
	baseDir     string
	logDir      string
	currentFile *os.File
	currentDate string
	lastLogTime *time.Time
	mu          sync.Mutex
}

// NewWatchLogger creates a new logger for the specified directory.
// It initializes the .lowlog directory structure if it doesn't exist.
func NewWatchLogger(dir string) (*WatchLogger, error) {
	logDir := filepath.Join(dir, ".lowlog")
	logger := &WatchLogger{
		baseDir: dir,
		logDir:  logDir,
	}

	if err := logger.ensureLogDir(); err != nil {
		return nil, fmt.Errorf("watch logger: create log dir: %w", err)
	}

	return logger, nil
}

// LogChange writes a formatted change event to the current log file.
// It handles date-based rotation automatically.
func (wl *WatchLogger) LogChange(change reporting.Change) error {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	// Ensure we have the right log file for today
	if err := wl.ensureCurrentLogFile(); err != nil {
		return fmt.Errorf("watch logger: ensure log file: %w", err)
	}

	// Check if we need to add a gap (9 empty lines) for 1+ hour difference
	now := change.Timestamp
	if wl.lastLogTime != nil {
		timeSinceLastLog := now.Sub(*wl.lastLogTime)
		if timeSinceLastLog >= time.Hour {
			// Insert 9 empty lines to visually separate events with 1+ hour gap
			for i := 0; i < 9; i++ {
				if _, err := wl.currentFile.WriteString("\n"); err != nil {
					return fmt.Errorf("watch logger: write gap: %w", err)
				}
			}
		}
	}

	// Format the log entry
	entry := wl.formatLogEntry(change)

	// Write to file
	if _, err := wl.currentFile.WriteString(entry); err != nil {
		return fmt.Errorf("watch logger: write entry: %w", err)
	}

	// Ensure data is flushed
	if err := wl.currentFile.Sync(); err != nil {
		return fmt.Errorf("watch logger: sync file: %w", err)
	}

	// Update last log time
	wl.lastLogTime = &now

	return nil
}

// Close closes the current log file if open.
func (wl *WatchLogger) Close() error {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	if wl.currentFile != nil {
		err := wl.currentFile.Close()
		wl.currentFile = nil
		return err
	}
	return nil
}

// ensureLogDir creates the .lowlog directory if it doesn't exist.
func (wl *WatchLogger) ensureLogDir() error {
	return os.MkdirAll(wl.logDir, 0o755)
}

// ensureCurrentLogFile ensures the correct date-based log file is open.
// It handles rotation when the date changes.
func (wl *WatchLogger) ensureCurrentLogFile() error {
	today := time.Now().Format("2006-01-02")

	// If date hasn't changed and file is open, nothing to do
	if wl.currentDate == today && wl.currentFile != nil {
		return nil
	}

	// Close previous file if open
	if wl.currentFile != nil {
		wl.currentFile.Close()
		wl.currentFile = nil
	}

	// Open new file for today
	logPath := filepath.Join(wl.logDir, fmt.Sprintf("%s.log", today))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	wl.currentFile = file
	wl.currentDate = today
	// Reset lastLogTime when switching to a new day to avoid gaps at day boundaries
	wl.lastLogTime = nil
	return nil
}

// formatLogEntry formats a change event for logging.
func (wl *WatchLogger) formatLogEntry(change reporting.Change) string {
	timestamp := change.Timestamp.Format("2006-01-02 15:04:05")

	// Make the path relative to the base directory for cleaner logs
	relPath, err := filepath.Rel(wl.baseDir, change.Path)
	if err != nil {
		relPath = change.Path // Fall back to absolute path
	}

	// Format size information
	sizeInfo := ""
	switch change.Type {
	case "CREATE", "NEW":
		if change.Size > 0 {
			sizeInfo = fmt.Sprintf(" (%d bytes)", change.Size)
		} else {
			sizeInfo = " (0 bytes)"
		}
	case "MODIFY", "MODIFIED":
		if change.SizeDelta != 0 {
			sign := "+"
			delta := change.SizeDelta
			if delta < 0 {
				sign = ""
			}
			sizeInfo = fmt.Sprintf(" (%s%d bytes)", sign, delta)
		} else {
			sizeInfo = " (0 bytes)"
		}
	case "DELETE", "DELETED":
		// No size info for deletions
	}

	// Map change types to match expected format
	changeType := change.Type
	switch change.Type {
	case "CREATE":
		changeType = "NEW"
	case "MODIFY":
		changeType = "MODIFIED"
	case "DELETE":
		changeType = "DELETED"
	}

	return fmt.Sprintf("[%s] [%s] %s%s\n", timestamp, changeType, relPath, sizeInfo)
}

// WatchLoggerPool manages multiple WatchLogger instances for different directories.
// This is useful when watching multiple directories simultaneously.
type WatchLoggerPool struct {
	loggers map[string]*WatchLogger
	mu      sync.RWMutex
	enabled bool
}

// NewWatchLoggerPool creates a new pool for managing multiple watch loggers.
func NewWatchLoggerPool(enabled bool) *WatchLoggerPool {
	return &WatchLoggerPool{
		loggers: make(map[string]*WatchLogger),
		enabled: enabled,
	}
}

// LogChange logs a change to the appropriate directory's logger.
// It automatically creates a logger for new directories.
func (p *WatchLoggerPool) LogChange(change reporting.Change) error {
	if !p.enabled {
		return nil
	}

	// Determine which directory this change belongs to
	dir := p.findWatchedDirectory(change.Path)
	if dir == "" {
		return nil // Not in a watched directory
	}

	// Get or create logger for this directory
	logger, err := p.getOrCreateLogger(dir)
	if err != nil {
		return err
	}

	return logger.LogChange(change)
}

// findWatchedDirectory determines which watched directory a path belongs to.
func (p *WatchLoggerPool) findWatchedDirectory(path string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simple implementation - in practice, this would need to track
	// which directories are being watched
	for dir := range p.loggers {
		if filepath.HasPrefix(path, dir) {
			return dir
		}
	}

	// Check parent directories up to a reasonable depth
	current := filepath.Dir(path)
	for i := 0; i < 10; i++ {
		if _, exists := p.loggers[current]; exists {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return ""
}

// getOrCreateLogger gets an existing logger or creates a new one for a directory.
func (p *WatchLoggerPool) getOrCreateLogger(dir string) (*WatchLogger, error) {
	p.mu.RLock()
	if logger, exists := p.loggers[dir]; exists {
		p.mu.RUnlock()
		return logger, nil
	}
	p.mu.RUnlock()

	// Need to create a new logger
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if logger, exists := p.loggers[dir]; exists {
		return logger, nil
	}

	logger, err := NewWatchLogger(dir)
	if err != nil {
		return nil, err
	}

	p.loggers[dir] = logger
	return logger, nil
}

// AddDirectory adds a directory to be logged.
func (p *WatchLoggerPool) AddDirectory(dir string) error {
	if !p.enabled {
		return nil
	}

	_, err := p.getOrCreateLogger(dir)
	return err
}

// Close closes all loggers in the pool.
func (p *WatchLoggerPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for _, logger := range p.loggers {
		if err := logger.Close(); err != nil {
			lastErr = err
		}
	}

	p.loggers = make(map[string]*WatchLogger)
	return lastErr
}
