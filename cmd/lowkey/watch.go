package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

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
			// TODO: Implement the foreground watching logic.
			// This should create and start a watcher.Controller, running it directly
			// and waiting for an interrupt signal (e.g., Ctrl+C) to stop.
			// - Create a new watcher.Controller with the manifest.
			// - Start the controller.
			// - Set up a signal handler to catch SIGINT and SIGTERM.
			// - Call controller.Stop() when a signal is received.
			// - Wait for the controller to stop gracefully.
			fmt.Printf("watch: would start monitoring %s (foreground)\n", strings.Join(manifest.Directories, ", "))
			return nil
		},
	}
}
