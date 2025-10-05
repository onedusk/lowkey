package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start [dir ...]",
		Short: "Launch the background daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets := args
			if len(targets) == 0 {
				targets = loadWatchTargetsFromConfig()
			}
			if len(targets) == 0 {
				fmt.Println("start: no directories provided; daemon would use manifest defaults")
			} else {
				fmt.Printf("start: would daemonize watcher for %s\n", strings.Join(targets, ", "))
			}
			return nil
		},
	}
}
