package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCacheSetGetDelete(t *testing.T) {
	cache := NewCache()

	sig := FileSignature{Size: 10, ModTime: time.Now().UTC()}
	cache.Set("/tmp/file.txt", sig)

	if got, ok := cache.Get("/tmp/file.txt"); !ok || !got.Equal(sig) {
		t.Fatalf("expected signature %+v, got %+v ok=%v", sig, got, ok)
	}

	if cache.Len() != 1 {
		t.Fatalf("expected len 1, got %d", cache.Len())
	}

	cache.Delete("/tmp/file.txt")
	if _, ok := cache.Get("/tmp/file.txt"); ok {
		t.Fatalf("expected delete to remove entry")
	}
}

func TestCacheFilesUnder(t *testing.T) {
	cache := NewCache()
	sig := FileSignature{Size: 1, ModTime: time.Now().UTC()}
	cache.Set("/tmp/a.txt", sig)
	cache.Set("/tmp/sub/b.txt", sig)
	cache.Set("/other/c.txt", sig)

	files := cache.FilesUnder("/tmp")
	if len(files) != 2 {
		t.Fatalf("expected 2 files under /tmp, got %d", len(files))
	}
	if _, ok := files["/other/c.txt"]; ok {
		t.Fatalf("unexpected file from another directory")
	}
}

func TestComputeSignatureSmallFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat temp file: %v", err)
	}

	sig, err := ComputeSignature(path, info)
	if err != nil {
		t.Fatalf("compute signature: %v", err)
	}

	if sig.Size != int64(len("hello")) {
		t.Fatalf("unexpected size: %d", sig.Size)
	}
	if sig.Hash == "" {
		t.Fatalf("expected hash for small file")
	}
}

func TestDetectChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	info, _ := os.Stat(path)

	sig, changed, err := DetectChange(FileSignature{}, false, info, path)
	if err != nil {
		t.Fatalf("detect change (new): %v", err)
	}
	if !changed {
		t.Fatalf("expected change for new file")
	}

	if _, changed, err = DetectChange(sig, true, info, path); err != nil || changed {
		t.Fatalf("expected no change when signature unchanged (changed=%v, err=%v)", changed, err)
	}

	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("update file: %v", err)
	}
	info, _ = os.Stat(path)
	if _, changed, err = DetectChange(sig, true, info, path); err != nil || !changed {
		t.Fatalf("expected change after modification (changed=%v, err=%v)", changed, err)
	}
}

func TestNormalizePath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	normalized, err := NormalizePath("internal")
	if err != nil {
		t.Fatalf("normalize relative: %v", err)
	}
	if !strings.HasPrefix(normalized, cwd) {
		t.Fatalf("expected normalized path to start with cwd: %s", normalized)
	}

	if _, err := NormalizePath(""); err == nil {
		t.Fatalf("expected error for empty path")
	}
}
