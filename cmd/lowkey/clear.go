package main

import "github.com/spf13/cobra"

func newClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Prune logs and/or cached state",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("clear: would delete logs or state based on flags")
			return nil
		},
	}
}
