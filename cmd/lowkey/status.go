package main

import "github.com/spf13/cobra"

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("status: daemon not yet implemented (stub)")
			return nil
		},
	}
}
