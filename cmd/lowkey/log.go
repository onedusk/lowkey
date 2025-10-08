package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"lowkey/internal/state"
)

// newLogCmd creates the `log` command, which is used to print the contents of
// the daemon's log file. This provides a simple way to inspect the daemon's
// past activity and diagnose issues.
func newLogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Print daemon logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := state.DefaultStateDir()
			if err != nil {
				return err
			}
			logPath := filepath.Join(stateDir, "lowkey.log")
			file, err := os.Open(logPath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("log: no log file found at %s\n", logPath)
					return nil
				}
				return err
			}
			defer file.Close()
			if _, err := io.Copy(os.Stdout, file); err != nil {
				return err
			}
			return nil
		},
	}
}
