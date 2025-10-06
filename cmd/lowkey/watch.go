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
			fmt.Printf("watch: would start monitoring %s (foreground)\n", strings.Join(manifest.Directories, ", "))
			return nil
		},
	}
}
