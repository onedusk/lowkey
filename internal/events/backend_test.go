package events

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPollingBackendDetectsNewFile(t *testing.T) {
	backend, err := NewPollingBackend(25 * time.Millisecond)
	if err != nil {
		t.Fatalf("new polling backend: %v", err)
	}
	t.Cleanup(func() {
		_ = backend.Close()
	})

	dir := t.TempDir()
	if err := backend.Add(dir); err != nil {
		t.Fatalf("add watch dir: %v", err)
	}

	path := filepath.Join(dir, "sample.txt")

	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	select {
	case event := <-backend.Events():
		if event.Path != path {
			t.Fatalf("unexpected event path: %s", event.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for event")
	}
}
