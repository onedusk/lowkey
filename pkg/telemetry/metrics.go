package telemetry

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// metrics.go exports Prometheus-style collectors tracking filesystem events,
// latency, and errors. Wire into daemon startup when metrics are enabled.

// Collector publishes counters and summaries for watcher activity.
type Collector struct {
	events uint64
	errors uint64

	latencyMu    sync.Mutex
	latencySum   time.Duration
	latencyCount uint64

	server   *http.Server
	listener net.Listener
	startMu  sync.Mutex
}

// NewCollector constructs an idle metrics collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Start begins serving Prometheus metrics on the supplied TCP address (e.g.,
// "127.0.0.1:9600"). The handler is mounted at `/metrics`.
func (c *Collector) Start(addr string) error {
	if addr == "" {
		return fmt.Errorf("telemetry: empty metrics address")
	}

	c.startMu.Lock()
	defer c.startMu.Unlock()
	if c.listener != nil {
		return fmt.Errorf("telemetry: metrics already started")
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", c.handleMetrics)

	server := &http.Server{Handler: mux}
	c.server = server
	c.listener = listener

	go func() {
		_ = server.Serve(listener)
	}()
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (c *Collector) Stop(ctx context.Context) error {
	c.startMu.Lock()
	defer c.startMu.Unlock()
	if c.server == nil {
		return nil
	}
	err := c.server.Shutdown(ctx)
	c.server = nil
	c.listener = nil
	return err
}

// IncEvent increments the events counter.
func (c *Collector) IncEvent() {
	atomic.AddUint64(&c.events, 1)
}

// IncError increments the error counter.
func (c *Collector) IncError() {
	atomic.AddUint64(&c.errors, 1)
}

// ObserveLatency records a single event processing duration.
func (c *Collector) ObserveLatency(d time.Duration) {
	c.latencyMu.Lock()
	defer c.latencyMu.Unlock()
	c.latencySum += d
	c.latencyCount++
}

func (c *Collector) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	events := atomic.LoadUint64(&c.events)
	errors := atomic.LoadUint64(&c.errors)

	avgLatency := 0.0
	c.latencyMu.Lock()
	if c.latencyCount > 0 {
		avgLatency = (c.latencySum.Seconds()) / float64(c.latencyCount)
	}
	count := c.latencyCount
	c.latencyMu.Unlock()

	fmt.Fprintf(w, "# HELP lowkey_events_total Total filesystem change events processed.\n")
	fmt.Fprintf(w, "# TYPE lowkey_events_total counter\n")
	fmt.Fprintf(w, "lowkey_events_total %d\n", events)

	fmt.Fprintf(w, "# HELP lowkey_errors_total Total errors encountered while monitoring.\n")
	fmt.Fprintf(w, "# TYPE lowkey_errors_total counter\n")
	fmt.Fprintf(w, "lowkey_errors_total %d\n", errors)

	fmt.Fprintf(w, "# HELP lowkey_event_latency_seconds Average latency per event.\n")
	fmt.Fprintf(w, "# TYPE lowkey_event_latency_seconds gauge\n")
	fmt.Fprintf(w, "lowkey_event_latency_seconds %.6f\n", avgLatency)

	fmt.Fprintf(w, "# HELP lowkey_event_latency_samples Number of samples contributing to latency metric.\n")
	fmt.Fprintf(w, "# TYPE lowkey_event_latency_samples counter\n")
	fmt.Fprintf(w, "lowkey_event_latency_samples %d\n", count)
}
