package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"lowkey/internal/state"
	"lowkey/pkg/config"
)

// newClearCmd creates the `clear` command, which is responsible for pruning
// logs and cached state. This helps in resetting the daemon's state or freeing
// up disk space.
func newClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Prune logs and/or cached state",
		RunE: func(cmd *cobra.Command, args []string) error {
			clearLogs, clearState, yes, remaining := parseClearArgs(args)
			if len(remaining) > 0 {
				return fmt.Errorf("clear: unexpected arguments: %v", remaining)
			}
			if !clearLogs && !clearState {
				clearLogs, clearState = true, true
			}

			stateDir, err := state.DefaultStateDir()
			if err != nil {
				return err
			}

			store, err := state.NewManifestStore(stateDir)
			if err != nil {
				return err
			}

			manifest, err := store.Load()
			if err != nil {
				return err
			}

			logTargets := collectLogTargets(stateDir, manifest)
			stateTargets := collectStateTargets(stateDir)

			fmt.Println("targets to clear:")
			if clearLogs {
				fmt.Println("  logs:")
				if len(logTargets) == 0 {
					fmt.Println("    (none found)")
				} else {
					for _, path := range logTargets {
						fmt.Printf("    - %s\n", path)
					}
				}
			}
			if clearState {
				fmt.Println("  state:")
				fmt.Printf("    - %s (manifest)\n", store.Path())
				for _, path := range stateTargets {
					fmt.Printf("    - %s\n", path)
				}
			}

			if !yes {
				fmt.Print("proceed? [y/N]: ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.ToLower(strings.TrimSpace(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("clear: aborted")
					return nil
				}
			}

			var errs []error
			if clearLogs {
				if err := removePaths(logTargets); err != nil {
					errs = append(errs, err)
				}
			}
			if clearState {
				if err := store.Clear(); err != nil {
					errs = append(errs, err)
				}
				if err := removePaths(stateTargets); err != nil {
					errs = append(errs, err)
				}
			}

			if len(errs) > 0 {
				return errors.Join(errs...)
			}

			fmt.Println("clear: completed")
			return nil
		},
	}
}

// collectLogTargets identifies all log files that should be considered for
// clearing. It checks for rotated log files based on the base log path.
func collectLogTargets(stateDir string, manifest *config.Manifest) []string {
	base := filepath.Join(stateDir, "lowkey.log")
	if manifest != nil && manifest.LogPath != "" {
		base = manifest.LogPath
	}
	matches, err := filepath.Glob(base + "*")
	if err != nil {
		return []string{base}
	}
	if len(matches) == 0 {
		return []string{base}
	}
	return matches
}

// collectStateTargets gathers the paths of all state files, such as the cache
// and PID file, that should be removed during a state clear operation.
func collectStateTargets(stateDir string) []string {
	return []string{
		filepath.Join(stateDir, "cache.json"),
		pidFilePath(stateDir),
	}
}

// removePaths deletes a list of files. It continues even if some deletions
// fail and returns a consolidated error.
func removePaths(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	var errs []error
	for _, path := range paths {
		if path == "" {
			continue
		}
		if err := os.Remove(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			errs = append(errs, fmt.Errorf("remove %s: %w", path, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// parseClearArgs processes the command-line arguments for the `clear` command,
// identifying which components to clear (logs, state) and whether to bypass
// the confirmation prompt.
func parseClearArgs(args []string) (logs, state, yes bool, remaining []string) {
	remaining = make([]string, 0, len(args))
	for _, arg := range args {
		switch {
		case arg == "--logs":
			logs = true
		case arg == "--state":
			state = true
		case arg == "--yes" || arg == "-y":
			yes = true
		default:
			remaining = append(remaining, arg)
		}
	}
	return logs, state, yes, remaining
}
