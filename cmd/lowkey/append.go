package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"lowkey/internal/logging"
)

// newAppendCmd creates the `append` command, which accepts JSON log entries
// from stdin and appends them to a specified log file with automatic rotation.
// This enables external tools to leverage lowkey's robust logging infrastructure.
func newAppendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "append",
		Short: "Append JSON log entries with rotation support",
		RunE: func(cmd *cobra.Command, args []string) error {
			logFile, maxSize, maxBackups, remaining := parseAppendFlags(args)
			if len(remaining) > 0 {
				return fmt.Errorf("append: unexpected arguments: %v", remaining)
			}
			if logFile == "" {
				return fmt.Errorf("append: --file is required")
			}

			// Ensure absolute path
			absPath, err := filepath.Abs(logFile)
			if err != nil {
				return fmt.Errorf("append: invalid file path: %w", err)
			}

			// Create log directory if it doesn't exist
			logDir := filepath.Dir(absPath)
			if err := os.MkdirAll(logDir, 0o755); err != nil {
				return fmt.Errorf("append: failed to create log directory: %w", err)
			}

			// Set up rotator
			baseName := filepath.Base(absPath)
			rotator, err := logging.NewRotator(logDir, baseName, maxSize, maxBackups)
			if err != nil {
				return fmt.Errorf("append: failed to create rotator: %w", err)
			}
			defer rotator.Close()

			// Read from stdin and append to log
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Bytes()

				// Validate JSON
				var jsonCheck interface{}
				if err := json.Unmarshal(line, &jsonCheck); err != nil {
					fmt.Fprintf(os.Stderr, "append: skipping invalid JSON: %s\n", err)
					continue
				}

				// Write the line with newline
				if _, err := rotator.Write(append(line, '\n')); err != nil {
					return fmt.Errorf("append: write failed: %w", err)
				}
			}

			if err := scanner.Err(); err != nil && err != io.EOF {
				return fmt.Errorf("append: stdin read error: %w", err)
			}

			return nil
		},
	}
}

// parseAppendFlags processes the command-line arguments for the `append` command,
// extracting the log file path and rotation parameters.
func parseAppendFlags(args []string) (logFile string, maxSize int64, maxBackups int, remaining []string) {
	// Set defaults
	maxSize = 10 * 1024 * 1024 // 10MB
	maxBackups = 5

	remaining = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--file" || arg == "-f":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				logFile = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--file="):
			logFile = arg[len("--file="):]
		case strings.HasPrefix(arg, "-f="):
			logFile = arg[len("-f="):]
		case arg == "--max-size":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				if size, err := strconv.ParseInt(args[i+1], 10, 64); err == nil {
					maxSize = size
				}
				i++
			}
		case strings.HasPrefix(arg, "--max-size="):
			if size, err := strconv.ParseInt(arg[len("--max-size="):], 10, 64); err == nil {
				maxSize = size
			}
		case arg == "--max-backups":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				if backups, err := strconv.Atoi(args[i+1]); err == nil {
					maxBackups = backups
				}
				i++
			}
		case strings.HasPrefix(arg, "--max-backups="):
			if backups, err := strconv.Atoi(arg[len("--max-backups="):]); err == nil {
				maxBackups = backups
			}
		default:
			remaining = append(remaining, arg)
		}
	}
	return logFile, maxSize, maxBackups, remaining
}
