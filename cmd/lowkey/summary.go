package main

import "github.com/spf13/cobra"

func newSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show recent change statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("summary: would aggregate change statistics")
			return nil
		},
	}
}
