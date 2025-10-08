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
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Collector publishes counters and summaries for watcher activity. It exposes
// Prometheus-style metrics over an HTTP endpoint, tracking the number of file
// system events, errors, and event processing latency. The collector is safe
// for concurrent use.
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

// NewCollector constructs an idle metrics collector. The collector does not
// start serving metrics until the Start method is called.
func NewCollector() *Collector {
	return &Collector{}
}

// Start begins serving Prometheus metrics on the supplied TCP address (e.g.,
// "127.0.0.1:9600"). The metrics are exposed at the `/metrics` endpoint. This
// method is safe to call multiple times, but it will only start the server once.
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

// Stop gracefully shuts down the HTTP server that serves the Prometheus
// metrics. It waits for active connections to finish before returning.
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

// IncEvent increments the total number of processed file system events.
// This method is safe for concurrent use.
func (c *Collector) IncEvent() {
	atomic.AddUint64(&c.events, 1)
}

// IncError increments the total number of errors encountered during file
// system monitoring. This method is safe for concurrent use.
func (c *Collector) IncError() {
	atomic.AddUint64(&c.errors, 1)
}

// ObserveLatency records a single event processing duration. This data is used
// to calculate the average event latency. This method is safe for concurrent use.
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
