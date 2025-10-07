package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"lowkey/internal/daemon"
)

// Renderer emits formatted output for CLI commands.
type Renderer interface {
	Status(status daemon.ManagerStatus) error
}

// NewRenderer returns an implementation based on the format keyword ("plain" or "json").
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

// WithWriter allows tests to inject a custom writer.
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

// tableRenderer renders human readable text.
type tableRenderer struct {
	writer io.Writer
}

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

// jsonRenderer emits JSON payloads.
type jsonRenderer struct {
	encoder *json.Encoder
}

func (j *jsonRenderer) Status(status daemon.ManagerStatus) error {
	if j.encoder == nil {
		return errors.New("output: json encoder missing")
	}
	j.encoder.SetIndent("", "  ")
	return j.encoder.Encode(status)
}
