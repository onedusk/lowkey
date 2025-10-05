package main

import "github.com/spf13/cobra"

func newTailCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tail",
		Short: "Follow daemon logs in real time",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("tail: would stream live log entries")
			return nil
		},
	}
}
