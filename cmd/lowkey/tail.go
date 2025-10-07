package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"lowkey/internal/state"
	"lowkey/pkg/config"
)

func newTailCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tail",
		Short: "Follow daemon logs in real time",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := state.DefaultStateDir()
			if err != nil {
				return err
			}

			logPath := filepath.Join(stateDir, "lowkey.log")
			if stored, ok := readPID(stateDir); ok && processAlive(stored) {
				if manifest, err := loadStoredManifest(stateDir); err == nil && manifest != nil && manifest.LogPath != "" {
					logPath = manifest.LogPath
				}
			}

			signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			fmt.Printf("tailing %s\n", logPath)
			if err := tailFile(signalCtx, logPath); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		},
	}
}

func tailFile(ctx context.Context, path string) error {
	var file *os.File
	var err error

	for {
		file, err = os.Open(path)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	defer file.Close()

	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return err
		}

		if info.Size() < offset {
			file.Close()
			file, err = os.Open(path)
			if err != nil {
				return err
			}
			offset = 0
			continue
		}

		if info.Size() == offset {
			time.Sleep(400 * time.Millisecond)
			continue
		}

		toRead := info.Size() - offset
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			return err
		}
		buffer := make([]byte, toRead)
		if _, err := io.ReadFull(file, buffer); err != nil {
			return err
		}
		offset = info.Size()
		fmt.Print(string(buffer))
	}
}

func loadStoredManifest(stateDir string) (*config.Manifest, error) {
	store, err := state.NewManifestStore(stateDir)
	if err != nil {
		return nil, err
	}
	return store.Load()
}
