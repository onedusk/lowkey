package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatchLoggerCreatesDailyLogFile(t *testing.T) {
	baseDir := t.TempDir()

	logger, err := NewWatchLogger(baseDir)
	if err != nil {
		t.Fatalf("NewWatchLogger returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := logger.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})

	todayLog := filepath.Join(baseDir, ".lowkey", time.Now().Format("2006-01-02")+".log")

	info, err := os.Stat(todayLog)
	if err != nil {
		t.Fatalf("expected log file %s to exist: %v", todayLog, err)
	}

	if !info.Mode().IsRegular() {
		t.Fatalf("expected %s to be a regular file", todayLog)
	}

	if size := info.Size(); size != 0 {
		t.Fatalf("expected log file to be empty, got size %d", size)
	}
}
