package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"lowkey/internal/logs"
	"lowkey/pkg/colors"
)

// newLogCmd creates the `log` command, which is used to view logs from .lowlog
// directories in watched directories. It supports optional grep pattern filtering
// and colorized output based on event types.
func newLogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log [PATTERN]",
		Short: "View logs with optional grep pattern",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate args count
			if len(args) > 1 {
				return errors.New("log command accepts at most one argument (pattern)")
			}
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

			// Extract optional grep pattern
			pattern := ""
			if len(args) > 0 {
				pattern = args[0]
			}

			// Read logs with optional filtering
			reader := logs.NewReader(logDir)
			lines, err := reader.ReadLines(pattern)
			if err != nil {
				return err
			}

			if len(lines) == 0 {
				if pattern != "" {
					fmt.Printf("no logs found matching pattern: %s\n", pattern)
				} else {
					fmt.Println("no logs found")
				}
				return nil
			}

			// Print logs with color coding
			for _, line := range lines {
				printColoredLogLine(line)
			}

			return nil
		},
	}
}

// printColoredLogLine prints a log line with appropriate color based on event type
func printColoredLogLine(line string) {
	// Determine color based on event type in the line
	var color string
	if strings.Contains(line, "[NEW]") {
		color = colors.Green
	} else if strings.Contains(line, "[MODIFIED]") {
		color = colors.Yellow
	} else if strings.Contains(line, "[DELETED]") {
		color = colors.Red
	} else {
		// No color for unrecognized format
		fmt.Println(line)
		return
	}

	fmt.Println(colors.Colorize(line, color))
}
