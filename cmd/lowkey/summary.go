package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"lowkey/internal/logs"
	"lowkey/pkg/colors"
)

// newSummaryCmd creates the `summary` command, which displays change
// statistics from .lowlog files. This provides a comprehensive overview of
// file system activity including most active files and hourly activity.
func newSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show change statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the watched directories from config
			dirs := loadWatchTargetsFromConfig()
			if len(dirs) == 0 {
				return errors.New("no watched directories configured; run 'lowkey watch' or 'lowkey start' first")
			}

			// Use the first watched directory's .lowlog
			logDir := filepath.Join(dirs[0], ".lowlog")

			// Check if log directory exists
			if _, err := os.Stat(logDir); os.IsNotExist(err) {
				fmt.Printf("no logs found at %s\n", logDir)
				return nil
			}

			// Get statistics from logs
			reader := logs.NewReader(logDir)
			stats, err := reader.GetStats()
			if err != nil {
				return err
			}

			if stats.TotalEvents == 0 {
				fmt.Println("no logs found")
				return nil
			}

			// Print summary header
			colors.Println(colors.Blue, "=== File Monitor Summary ===")
			colors.Printf(colors.Magenta, "Total events: %d\n", stats.TotalEvents)
			colors.Printf(colors.Green, "  New files:      %d\n", stats.NewCount)
			colors.Printf(colors.Yellow, "  Modified files: %d\n", stats.ModifiedCount)
			colors.Printf(colors.Red, "  Deleted files:  %d\n", stats.DeletedCount)

			// Print most active files
			if len(stats.MostActiveFiles) > 0 {
				colors.Println(colors.Blue, "\nMost active files:")
				for _, file := range stats.MostActiveFiles {
					fmt.Printf("  %d changes: %s\n", file.Count, file.Path)
				}
			}

			// Print activity by hour
			if len(stats.ActivityByHour) > 0 {
				colors.Println(colors.Blue, "\nActivity by hour:")
				for _, hour := range stats.ActivityByHour {
					fmt.Printf("  %s:00  %d events\n", hour.Hour, hour.Count)
				}
			}

			return nil
		},
	}
}
