package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lowkey/internal/daemon"
	"lowkey/internal/state"
	"lowkey/pkg/config"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start [dir ...]",
		Short: "Launch the background daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath, remaining := extractOption(args, "--manifest", "-m")
			manifest, err := resolveManifest(manifestPath, remaining)
			if err != nil {
				return err
			}

			stateDir, err := state.DefaultStateDir()
			if err != nil {
				return err
			}
			store, err := state.NewManifestStore(stateDir)
			if err != nil {
				return err
			}

			manager, err := daemon.NewManager(store, manifest)
			if err != nil {
				return err
			}

			if err := manager.Start(); err != nil {
				return err
			}

			status := manager.Status()
			if err := renderStatus(status); err != nil {
				return err
			}
			fmt.Println("start: daemon process not yet implemented; controller Start is currently a stub")
			return nil
		},
	}
}

func resolveManifest(manifestPath string, args []string) (*config.Manifest, error) {
	if manifestPath != "" {
		return config.LoadManifest(manifestPath)
	}
	if manifestFromConfig != nil {
		return manifestFromConfig, nil
	}
	if len(args) == 0 {
		return nil, errors.New("start: provide directories or a manifest via --manifest/--config")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("start: determine working directory: %w", err)
	}
	return config.BuildManifestFromArgs(cwd, args)
}
