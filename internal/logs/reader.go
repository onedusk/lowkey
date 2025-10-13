// Package logs provides utilities for reading and analyzing .lowlog files
// created by the watch command. It supports reading dated log files,
// filtering by patterns, and extracting statistics.
package logs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// LogEntry represents a parsed log entry from a .lowlog file
type LogEntry struct {
	Timestamp time.Time
	Type      string // NEW, MODIFIED, DELETED
	Path      string
	Details   string // Size information or other details
	RawLine   string
}

// Reader provides methods for reading and analyzing .lowlog files
type Reader struct {
	logDir string
}

// NewReader creates a new log reader for the specified .lowlog directory
func NewReader(logDir string) *Reader {
	return &Reader{logDir: logDir}
}

// ReadAll reads all log entries from all .log files in the directory,
// optionally filtering by a grep pattern. Empty lines are excluded.
func (r *Reader) ReadAll(grepPattern string) ([]LogEntry, error) {
	logFiles, err := r.listLogFiles()
	if err != nil {
		return nil, err
	}

	var pattern *regexp.Regexp
	if grepPattern != "" {
		pattern, err = regexp.Compile("(?i)" + grepPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid grep pattern: %w", err)
		}
	}

	entries := make([]LogEntry, 0)
	for _, logFile := range logFiles {
		fileEntries, err := r.readFile(logFile, pattern)
		if err != nil {
			return nil, err
		}
		entries = append(entries, fileEntries...)
	}

	return entries, nil
}

// ReadLines reads all log lines (including raw formatting) from all files,
// optionally filtering by a grep pattern. This preserves the original format.
func (r *Reader) ReadLines(grepPattern string) ([]string, error) {
	logFiles, err := r.listLogFiles()
	if err != nil {
		return nil, err
	}

	var pattern *regexp.Regexp
	if grepPattern != "" {
		pattern, err = regexp.Compile("(?i)" + grepPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid grep pattern: %w", err)
		}
	}

	lines := make([]string, 0)
	for _, logFile := range logFiles {
		file, err := os.Open(logFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Skip empty lines
			if strings.TrimSpace(line) == "" {
				continue
			}
			// Apply pattern filter if specified
			if pattern != nil && !pattern.MatchString(line) {
				continue
			}
			lines = append(lines, line)
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return lines, nil
}

// Stats provides statistics about the logged events
type Stats struct {
	TotalEvents     int
	NewCount        int
	ModifiedCount   int
	DeletedCount    int
	MostActiveFiles []FileActivity
	ActivityByHour  []HourActivity
	FirstEvent      *time.Time
	LastEvent       *time.Time
}

// FileActivity represents change activity for a single file
type FileActivity struct {
	Path  string
	Count int
}

// HourActivity represents the number of events in a specific hour
type HourActivity struct {
	Hour  string // Format: "2006-01-02 15"
	Count int
}

// GetStats analyzes log entries and returns statistics
func (r *Reader) GetStats() (*Stats, error) {
	entries, err := r.ReadAll("")
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return &Stats{}, nil
	}

	stats := &Stats{
		TotalEvents: len(entries),
	}

	// Count by type
	fileCounts := make(map[string]int)
	hourCounts := make(map[string]int)

	for _, entry := range entries {
		switch entry.Type {
		case "NEW":
			stats.NewCount++
		case "MODIFIED":
			stats.ModifiedCount++
		case "DELETED":
			stats.DeletedCount++
		}

		// Track file activity
		fileCounts[entry.Path]++

		// Track hourly activity
		hour := entry.Timestamp.Format("2006-01-02 15")
		hourCounts[hour]++

		// Track first and last events
		if stats.FirstEvent == nil || entry.Timestamp.Before(*stats.FirstEvent) {
			ts := entry.Timestamp
			stats.FirstEvent = &ts
		}
		if stats.LastEvent == nil || entry.Timestamp.After(*stats.LastEvent) {
			ts := entry.Timestamp
			stats.LastEvent = &ts
		}
	}

	// Build most active files (top 5)
	stats.MostActiveFiles = make([]FileActivity, 0, len(fileCounts))
	for path, count := range fileCounts {
		stats.MostActiveFiles = append(stats.MostActiveFiles, FileActivity{Path: path, Count: count})
	}
	sort.Slice(stats.MostActiveFiles, func(i, j int) bool {
		return stats.MostActiveFiles[i].Count > stats.MostActiveFiles[j].Count
	})
	if len(stats.MostActiveFiles) > 5 {
		stats.MostActiveFiles = stats.MostActiveFiles[:5]
	}

	// Build activity by hour (show last 5 hours with activity)
	stats.ActivityByHour = make([]HourActivity, 0, len(hourCounts))
	for hour, count := range hourCounts {
		stats.ActivityByHour = append(stats.ActivityByHour, HourActivity{Hour: hour, Count: count})
	}
	sort.Slice(stats.ActivityByHour, func(i, j int) bool {
		return stats.ActivityByHour[i].Hour < stats.ActivityByHour[j].Hour
	})
	if len(stats.ActivityByHour) > 5 {
		stats.ActivityByHour = stats.ActivityByHour[len(stats.ActivityByHour)-5:]
	}

	return stats, nil
}

// listLogFiles returns all .log files in the directory, sorted by name (date)
func (r *Reader) listLogFiles() ([]string, error) {
	pattern := filepath.Join(r.logDir, "*.log")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// readFile reads and parses a single log file
func (r *Reader) readFile(path string, pattern *regexp.Regexp) ([]LogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entries := make([]LogEntry, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Apply pattern filter if specified
		if pattern != nil && !pattern.MatchString(line) {
			continue
		}

		entry := parseLogLine(line)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// parseLogLine parses a log line into a LogEntry
// Expected format: [2006-01-02 15:04:05] [TYPE] path details
func parseLogLine(line string) *LogEntry {
	// Regular expression to parse the log format
	// [timestamp] [TYPE] path details
	re := regexp.MustCompile(`^\[([^\]]+)\]\s+\[([^\]]+)\]\s+(\S+)\s*(.*)$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) < 4 {
		// Line doesn't match expected format, skip it
		return nil
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		// Invalid timestamp, skip
		return nil
	}

	return &LogEntry{
		Timestamp: timestamp,
		Type:      matches[2],
		Path:      matches[3],
		Details:   matches[4],
		RawLine:   line,
	}
}
