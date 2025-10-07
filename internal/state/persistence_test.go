package state

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	cache := NewCache()
	sig := FileSignature{Size: 42, ModTime: time.Now().UTC(), Hash: "abc"}
	cache.Set(filepath.Join(dir, "file.txt"), sig)

	if err := Save(cache, path); err != nil {
		t.Fatalf("save cache: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}

	if loaded.Len() != cache.Len() {
		t.Fatalf("expected len %d, got %d", cache.Len(), loaded.Len())
	}

	if got, ok := loaded.Get(filepath.Join(dir, "file.txt")); !ok || !got.Equal(sig) {
		t.Fatalf("unexpected loaded signature: %+v ok=%v", got, ok)
	}
}

func TestLoadMissingCacheReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.json")

	cache, err := Load(path)
	if err != nil {
		t.Fatalf("load missing cache: %v", err)
	}
	if cache.Len() != 0 {
		t.Fatalf("expected empty cache, got len=%d", cache.Len())
	}
}

func TestSaveRejectsNilCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	if err := Save(nil, path); err == nil {
		t.Fatalf("expected error when saving nil cache")
	}
}
