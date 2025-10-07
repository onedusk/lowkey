package telemetry

import (
	"context"
	"log"
	"sync"
	"time"
)

// tracing.go hosts OpenTelemetry hooks for cross-component tracing. Optional but
// useful when diagnosing watcher bottlenecks.

// SpanSnapshot captures exported span metadata.
type SpanSnapshot struct {
	Name       string
	StartTime  time.Time
	Duration   time.Duration
	Attributes map[string]string
	Error      string
}

// SpanExporter consumes completed spans.
type SpanExporter interface {
	ExportSpan(snapshot SpanSnapshot)
}

// TracerOptions configures a Tracer instance.
type TracerOptions struct {
	Enabled  bool
	Exporter SpanExporter
}

// Tracer emits lightweight spans without depending on a full OpenTelemetry stack.
type Tracer struct {
	enabled  bool
	exporter SpanExporter
}

// NewTracer constructs a tracer. When Disabled it returns a no-op instance.
func NewTracer(opts TracerOptions) *Tracer {
	tracer := &Tracer{enabled: opts.Enabled}
	if !opts.Enabled {
		return tracer
	}
	if opts.Exporter != nil {
		tracer.exporter = opts.Exporter
	} else {
		tracer.exporter = &loggingExporter{}
	}
	return tracer
}

// Enabled reports whether tracing is active.
func (t *Tracer) Enabled() bool {
	return t != nil && t.enabled
}

// StartSpan creates a new span. The returned context embeds the span for
// downstream lookups.
func (t *Tracer) StartSpan(ctx context.Context, name string) (*Span, context.Context) {
	if t == nil || !t.enabled {
		return &Span{noop: true}, ctx
	}
	span := &Span{
		tracer: t,
		name:   name,
		start:  time.Now(),
		attrs:  make(map[string]string),
	}
	return span, context.WithValue(ctx, spanKey{}, span)
}

// Span represents an in-flight tracing span.
type Span struct {
	noop   bool
	tracer *Tracer
	name   string
	start  time.Time
	attrs  map[string]string
	mu     sync.Mutex
}

// SetAttribute records a key/value pair on the span.
func (s *Span) SetAttribute(key, value string) {
	if s == nil || s.noop {
		return
	}
	s.mu.Lock()
	s.attrs[key] = value
	s.mu.Unlock()
}

// End completes the span and forwards it to the exporter. If err is non-nil it
// is recorded as part of the snapshot.
func (s *Span) End(err error) {
	if s == nil || s.noop || s.tracer == nil || !s.tracer.enabled {
		return
	}
	s.mu.Lock()
	attrs := make(map[string]string, len(s.attrs))
	for k, v := range s.attrs {
		attrs[k] = v
	}
	s.mu.Unlock()

	snapshot := SpanSnapshot{
		Name:       s.name,
		StartTime:  s.start,
		Duration:   time.Since(s.start),
		Attributes: attrs,
	}
	if err != nil {
		snapshot.Error = err.Error()
	}
	s.tracer.exporter.ExportSpan(snapshot)
}

// SpanFromContext extracts a span from the context when available.
func SpanFromContext(ctx context.Context) (*Span, bool) {
	span, ok := ctx.Value(spanKey{}).(*Span)
	return span, ok
}

type spanKey struct{}

type loggingExporter struct{}

func (loggingExporter) ExportSpan(snapshot SpanSnapshot) {
	log.Printf("trace span=%s duration=%s attrs=%v err=%s", snapshot.Name, snapshot.Duration, snapshot.Attributes, snapshot.Error)
}
