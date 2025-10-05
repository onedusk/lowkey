package main

import "github.com/spf13/cobra"

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("stop: would signal the daemon to exit")
			return nil
		},
	}
}
