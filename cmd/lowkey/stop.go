package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"lowkey/internal/state"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := state.DefaultStateDir()
			if err != nil {
				return err
			}
			store, err := state.NewManifestStore(stateDir)
			if err != nil {
				return err
			}
			if err := store.Clear(); err != nil {
				return err
			}
			// TODO: Implement the logic to find and signal the running daemon process to stop.
			// This should not just clear the manifest but actively stop the process.
			// - Find the PID of the running daemon (e.g., from a PID file).
			// - Send a SIGTERM or SIGINT signal to the process.
			// - Wait for the process to exit gracefully.
			// - Clearing the manifest should happen after the daemon confirms shutdown.
			fmt.Println("stop: cleared persisted manifest; daemon shutdown logic pending")
			return nil
		},
	}
}
