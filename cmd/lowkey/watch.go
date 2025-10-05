package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
			fmt.Printf("watch: would start monitoring %s (foreground)\n", strings.Join(args, ", "))
			return nil
		},
	}
}
