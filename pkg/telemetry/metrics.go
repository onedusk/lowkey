package telemetry

// metrics.go exports Prometheus-style collectors tracking filesystem events,
// latency, and errors. Wire into daemon startup when metrics are enabled.

// TODO: Implement Prometheus metrics for observability.
// - Define metrics for events, errors, and processing latency.
//   (e.g., `lowkey_events_total`, `lowkey_errors_total`, `lowkey_event_processing_latency_seconds`).
// - Create a registry and expose it via an HTTP endpoint.
// - Add an option to the `start` command to enable/disable the metrics endpoint.
