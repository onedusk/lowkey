package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"lowkey/internal/reporting"
	"lowkey/internal/watcher"
	"lowkey/pkg/config"
)

func newWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch [dir ...]",
		Short: "Run Lowkey in foreground for the supplied directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = loadWatchTargetsFromConfig()
			}
			if len(args) == 0 {
				return errors.New("provide at least one directory to watch")
			}
			cwd, _ := os.Getwd()
			manifest, err := config.BuildManifestFromArgs(cwd, args)
			if err != nil {
				return err
			}

			signalCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
			defer stopSignals()

			changes := make(chan reporting.Change, 256)
			aggregator := reporting.NewAggregator()

			onChange := func(change reporting.Change) {
				select {
				case <-signalCtx.Done():
					return
				default:
				}
				select {
				case changes <- change:
				default:
				}
			}

			ignorePatterns := discoverIgnoreFiles(manifest.Directories)

			controller, err := watcher.NewController(watcher.ControllerConfig{
				Directories:  manifest.Directories,
				IgnoreGlobs:  ignorePatterns,
				Aggregator:   aggregator,
				PollInterval: 20 * time.Second,
				OnChange:     onChange,
			})
			if err != nil {
				return err
			}

			if err := controller.Start(); err != nil {
				return err
			}
			defer controller.Stop()

			fmt.Printf("watching %s\n", strings.Join(manifest.Directories, ", "))
			fmt.Println("press Ctrl+C to stop")

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-signalCtx.Done():
						return
					case change := <-changes:
						fmt.Printf("[%s] %s\n", change.Type, change.Path)
					}
				}
			}()

			<-signalCtx.Done()
			fmt.Println("stopping watcher...")
			wg.Wait()
			return nil
		},
	}
}

func discoverIgnoreFiles(dirs []string) []string {
	patterns := make([]string, 0)
	seen := make(map[string]struct{})
	for _, dir := range dirs {
		candidate := filepath.Join(dir, ".lowkey")
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		loaded, err := config.LoadIgnorePatterns(candidate)
		if err != nil {
			continue
		}
		for _, pattern := range loaded {
			if _, ok := seen[pattern]; ok {
				continue
			}
			seen[pattern] = struct{}{}
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}
