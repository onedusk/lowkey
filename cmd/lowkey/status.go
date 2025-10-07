package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"lowkey/internal/daemon"
	"lowkey/internal/state"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
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
				fmt.Println("status: no manifest stored; daemon is not configured")
				return nil
			}

			running := false
			if pid, ok := readPID(stateDir); ok && processAlive(pid) {
				running = true
			}

			status := daemon.ManagerStatus{
				Running:      running,
				Directories:  append([]string(nil), manifest.Directories...),
				ManifestPath: store.Path(),
			}
			if err := renderStatus(status); err != nil {
				return err
			}
			return nil
		},
	}
}
