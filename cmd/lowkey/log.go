package main

import "github.com/spf13/cobra"

func newLogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Print daemon logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("log: would display log history")
			return nil
		},
	}
}
