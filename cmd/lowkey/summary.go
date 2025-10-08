package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"lowkey/internal/daemon"
	"lowkey/internal/state"
)

// newSummaryCmd creates the `summary` command, which displays recent change
// statistics from the daemon. This provides a quick overview of file system
// activity without needing to parse the full log.
func newSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show recent change statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if manifest == nil {
				fmt.Println("summary: no manifest stored; daemon is not configured")
				return nil
			}

			manager, err := daemon.NewManager(store, manifest)
			if err != nil {
				return err
			}
			status := manager.Status()
			return renderStatus(status)
		},
	}
}
