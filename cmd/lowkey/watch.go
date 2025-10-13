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
	"lowkey/pkg/colors"
	"lowkey/pkg/config"
)

// newWatchCmd creates the `watch` command, which runs the file system watcher
// in the foreground. This provides a direct way to monitor directories without
// starting a background daemon.
func newWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch [--log] [dir ...]",
		Short: "Run Lowkey in foreground for the supplied directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse the --log flag from arguments
			enableLogging, args := parseWatchFlags(args)
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

			// Initialize the logger pool for .lowlog directories if enabled
			loggerPool := watcher.NewWatchLoggerPool(enableLogging)
			if enableLogging {
				// Add directories to logger pool
				for _, dir := range manifest.Directories {
					if err := loggerPool.AddDirectory(dir); err != nil {
						fmt.Printf("warning: failed to initialize logger for %s: %v\n", dir, err)
					}
				}
			}
			defer loggerPool.Close()

			onChange := func(change reporting.Change) {
				select {
				case <-signalCtx.Done():
					return
				default:
				}

				// Log to .lowlog directory if enabled
				if enableLogging {
					if err := loggerPool.LogChange(change); err != nil {
						// Don't fail on logging errors, just warn
						fmt.Printf("warning: failed to log change: %v\n", err)
					}
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
			if enableLogging {
				fmt.Println("logging changes to .lowlog directories")
			}
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
						// Print with color based on event type
						eventType := strings.ToUpper(change.Type)
						switch eventType {
						case "CREATE":
							eventType = "NEW"
						case "MODIFY":
							eventType = "MODIFIED"
						case "DELETE":
							eventType = "DELETED"
						}
						coloredType := colors.ColorizeEventType(eventType)
						fmt.Printf("[%s] %s\n", coloredType, change.Path)
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

// parseWatchFlags processes the command-line arguments for the `watch` command,
// extracting the --log flag if present.
func parseWatchFlags(args []string) (enableLogging bool, remaining []string) {
	remaining = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--log":
			enableLogging = true
		case strings.HasPrefix(arg, "--log="):
			val := strings.ToLower(arg[len("--log="):])
			enableLogging = val != "false" && val != "0"
		default:
			remaining = append(remaining, arg)
		}
	}
	return enableLogging, remaining
}

// discoverIgnoreFiles searches for `.lowkey` ignore files in the specified
// directories and aggregates their patterns. This allows for per-directory
// ignore rules in addition to a global ignore file.
func discoverIgnoreFiles(dirs []string) []string {
	patterns := make([]string, 0)
	seen := make(map[string]struct{})

	// Always ignore .lowlog directories to prevent recursive logging
	patterns = append(patterns, ".lowlog")
	seen[".lowlog"] = struct{}{}
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
