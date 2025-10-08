// Package output provides formatted output capabilities for the lowkey command-line
// interface. It defines a flexible rendering system that supports multiple
// output formats, such as plain text and JSON, allowing command results to be
// displayed in either a human-readable or machine-parsable way.
//
// The package is designed around a Renderer interface, which abstracts the
// specific formatting logic. This allows for easy extension with new output
// formats in the future. Concrete implementations for table-based (plain text)
// and JSON rendering are provided.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"lowkey/internal/daemon"
)

// Renderer defines the interface for emitting formatted output for CLI commands.
// It abstracts the underlying output format (e.g., plain text, JSON) and
// provides methods for rendering specific data structures, such as daemon status.
type Renderer interface {
	Status(status daemon.ManagerStatus) error
}

// NewRenderer returns a Renderer implementation based on the specified format
// keyword. It supports "plain" (or "text") for human-readable output and "json"
// for machine-readable output. An error is returned if the format is unsupported.
func NewRenderer(format string) (Renderer, error) {
	switch format {
	case "", "plain", "text":
		return &tableRenderer{writer: os.Stdout}, nil
	case "json":
		return &jsonRenderer{encoder: json.NewEncoder(os.Stdout)}, nil
	default:
		return nil, fmt.Errorf("output: unsupported format %q", format)
	}
}

// WithWriter allows tests to inject a custom io.Writer into a Renderer.
// This is useful for capturing output during testing without writing to stdout.
// It returns a new Renderer configured with the provided writer.
func WithWriter(r Renderer, w io.Writer) Renderer {
	switch r.(type) {
	case *tableRenderer:
		return &tableRenderer{writer: w}
	case *jsonRenderer:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return &jsonRenderer{encoder: enc}
	default:
		panic("output: unknown renderer implementation")
	}
}

// tableRenderer renders daemon status and other command outputs as human-readable
// text. It writes to the configured io.Writer, which is typically os.Stdout.
type tableRenderer struct {
	writer io.Writer
}

// Status formats and prints the daemon's status to the configured writer in a
// table-like, human-readable format.
func (t *tableRenderer) Status(status daemon.ManagerStatus) error {
	if t.writer == nil {
		return errors.New("output: table renderer missing writer")
	}

	fmt.Fprintf(t.writer, "daemon: running=%t\n", status.Running)
	fmt.Fprintf(t.writer, "manifest: %s\n", status.ManifestPath)
	fmt.Fprintf(t.writer, "directories (%d):\n", len(status.Directories))
	for _, dir := range status.Directories {
		fmt.Fprintf(t.writer, "  - %s\n", dir)
	}
	fmt.Fprintf(t.writer, "changes: total=%d window=%s\n", status.Summary.TotalChanges, status.Summary.Window)
	if status.Summary.LastEvent != nil {
		fmt.Fprintf(t.writer, "last change: %s (%s) at %s\n", status.Summary.LastEvent.Path, status.Summary.LastEvent.Type, status.Summary.LastEvent.Timestamp.Format("2006-01-02 15:04:05"))
	}
	if !status.Heartbeat.LastCheck.IsZero() {
		lastChange := "-"
		if !status.Heartbeat.LastChange.IsZero() {
			lastChange = status.Heartbeat.LastChange.Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(t.writer, "heartbeat: running=%t restarts=%d last_change=%s last_error=%s\n",
			status.Heartbeat.Running,
			status.Heartbeat.Restarts,
			lastChange,
			status.Heartbeat.LastError)
		if !status.Heartbeat.BackoffUntil.IsZero() {
			fmt.Fprintf(t.writer, "heartbeat backoff until: %s\n", status.Heartbeat.BackoffUntil.Format("2006-01-02 15:04:05"))
		}
	}
	return nil
}

// jsonRenderer emits command outputs as JSON payloads. This is suitable for
// scripting or integration with other tools that can parse JSON.
type jsonRenderer struct {
	encoder *json.Encoder
}

// Status encodes the daemon's status as a JSON object and writes it to the
// configured writer.
func (j *jsonRenderer) Status(status daemon.ManagerStatus) error {
	if j.encoder == nil {
		return errors.New("output: json encoder missing")
	}
	j.encoder.SetIndent("", "  ")
	return j.encoder.Encode(status)
}
