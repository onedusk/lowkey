// Package telemetry provides observability features for the lowkey daemon,
// including Prometheus-style metrics and lightweight tracing. It is designed to
// help monitor the performance and behavior of the file system watcher, offering
// insights into event throughput, latency, and errors.
//
// The components in this package are optional and can be enabled through
// configuration to minimize overhead when not in use.
package telemetry

import (
	"context"
	"log"
	"sync"
	"time"
)

// SpanSnapshot captures the metadata of an exported tracing span. It includes
// the span's name, timing information, attributes, and any associated error.
// This struct is used by SpanExporters to process and store trace data.
type SpanSnapshot struct {
	Name       string
	StartTime  time.Time
	Duration   time.Duration
	Attributes map[string]string
	Error      string
}

// SpanExporter defines the interface for consuming and exporting completed spans.
// Implementations of this interface can be used to send trace data to various
// backends, such as logging systems or distributed tracing platforms.
type SpanExporter interface {
	ExportSpan(snapshot SpanSnapshot)
}

// TracerOptions configures a Tracer instance. It allows enabling or disabling
// tracing and specifying a custom SpanExporter for processing completed spans.
type TracerOptions struct {
	Enabled  bool
	Exporter SpanExporter
}

// Tracer provides lightweight, OpenTelemetry-inspired tracing capabilities.
// It allows for the creation of spans to measure the duration of operations
// and to record contextual attributes. When disabled, the tracer and its
// spans are no-ops, incurring minimal performance overhead.
type Tracer struct {
	enabled  bool
	exporter SpanExporter
}

// NewTracer constructs a new tracer based on the provided options. If tracing
// is disabled in the options, a no-op tracer is returned. If no exporter is
// specified, a default logging exporter is used.
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

// Enabled reports whether the tracer is active. Spans will only be created and
// exported if this method returns true.
func (t *Tracer) Enabled() bool {
	return t != nil && t.enabled
}

// StartSpan creates a new tracing span and embeds it within the returned context.
// This allows the span to be accessed by downstream functions for recording
// attributes or ending the span. If the tracer is disabled, a no-op span is
// returned.
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

// Span represents an in-flight tracing span. It tracks the duration of an
// operation and allows for the attachment of key-value attributes. Spans should
// be ended by calling the End method.
type Span struct {
	noop   bool
	tracer *Tracer
	name   string
	start  time.Time
	attrs  map[string]string
	mu     sync.Mutex
}

// SetAttribute records a key-value pair as an attribute on the span. This can
// be used to add contextual information to a trace. This method is safe for
// concurrent use.
func (s *Span) SetAttribute(key, value string) {
	if s == nil || s.noop {
		return
	}
	s.mu.Lock()
	s.attrs[key] = value
	s.mu.Unlock()
}

// End completes the span, calculates its duration, and forwards it to the
// tracer's exporter. If a non-nil error is provided, it is recorded as part of
// the span's snapshot.
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

// SpanFromContext extracts a span from the given context, if one exists.
// This allows functions to access and interact with the current span without
// needing to pass it as an explicit argument.
func SpanFromContext(ctx context.Context) (*Span, bool) {
	span, ok := ctx.Value(spanKey{}).(*Span)
	return span, ok
}

type spanKey struct{}

type loggingExporter struct{}

func (loggingExporter) ExportSpan(snapshot SpanSnapshot) {
	log.Printf("trace span=%s duration=%s attrs=%v err=%s", snapshot.Name, snapshot.Duration, snapshot.Attributes, snapshot.Error)
}
