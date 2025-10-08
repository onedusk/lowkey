// Package logging provides a flexible logging framework for the lowkey daemon.
// It includes support for log rotation based on size and a structured logging
// wrapper for consistent log message formatting.
//
// The package is designed to be thread-safe and can be used to log messages
// from multiple goroutines concurrently.
package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Rotator handles log file rotation based on size and the number of backup
// files. It ensures that log files do not grow indefinitely and that a
// configurable amount of log history is preserved. It is safe for concurrent
// use.
type Rotator struct {
	dir        string
	baseName   string
	maxSize    int64
	maxBackups int

	file *os.File
	mux  sync.Mutex
}

// NewRotator creates a new log rotator. It takes the directory and base name
// for the log files, the maximum size of a log file before it is rotated, and
// the maximum number of backup log files to keep.
func NewRotator(dir, baseName string, maxSize int64, maxBackups int) (*Rotator, error) {
	if dir == "" {
		return nil, fmt.Errorf("logging: directory is required")
	}
	if baseName == "" {
		baseName = "lowkey.log"
	}
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024
	}
	if maxBackups <= 0 {
		maxBackups = 5
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("logging: create dir: %w", err)
	}

	rotator := &Rotator{dir: dir, baseName: baseName, maxSize: maxSize, maxBackups: maxBackups}
	if err := rotator.openFile(); err != nil {
		return nil, err
	}
	return rotator, nil
}

func (r *Rotator) openFile() error {
	path := filepath.Join(r.dir, r.baseName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	r.file = file
	return nil
}

// Write writes a byte slice to the current log file. If the write would cause
// the file to exceed its maximum size, the file is rotated before the write
// occurs. This method is safe for concurrent use.
func (r *Rotator) Write(p []byte) (int, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.file == nil {
		if err := r.openFile(); err != nil {
			return 0, err
		}
	}

	info, err := r.file.Stat()
	if err == nil && info.Size()+int64(len(p)) >= r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	return r.file.Write(p)
}

func (r *Rotator) rotate() error {
	if r.file != nil {
		r.file.Close()
		r.file = nil
	}

	timestamp := time.Now().Format("20060102-150405")
	archivedName := fmt.Sprintf("%s.%s", r.baseName, timestamp)
	oldPath := filepath.Join(r.dir, r.baseName)
	newPath := filepath.Join(r.dir, archivedName)
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	entries, err := filepath.Glob(filepath.Join(r.dir, fmt.Sprintf("%s.*", r.baseName)))
	if err == nil && len(entries) > r.maxBackups {
		excess := len(entries) - r.maxBackups
		toRemove := entries[:excess]
		for _, file := range toRemove {
			os.Remove(file)
		}
	}

	return r.openFile()
}

// Path returns the full path to the active log file.
func (r *Rotator) Path() string {
	return filepath.Join(r.dir, r.baseName)
}

// Close flushes any buffered data and closes the current log file. It is
// important to call this method to ensure that all log messages are written to
// disk.
func (r *Rotator) Close() error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

// NewLogger returns a standard `log.Logger` configured to write to the provided
// rotator. This allows the rotator to be used with any code that expects a
// standard logger.
func NewLogger(rotator *Rotator) *log.Logger {
	return log.New(rotator, "", log.LstdFlags|log.LUTC)
}
