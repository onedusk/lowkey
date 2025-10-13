# Telemetry Guide

This guide covers lowkey's telemetry features for monitoring and observability.

## Overview

Lowkey provides optional telemetry features to help monitor the performance and behavior of the filesystem watcher:

- **`--metrics`**: Expose Prometheus-style metrics over HTTP
- **`--trace`**: Enable lightweight span logging for debugging

Both features are disabled by default to minimize overhead.

## Metrics (`--metrics`)

### Usage

Start the daemon with metrics enabled:

```bash
# Basic usage (localhost only)
lowkey start --metrics 127.0.0.1:9600 /path/to/watch

# Allow external access (use with caution)
lowkey start --metrics 0.0.0.0:9600 /path/to/watch

# Combine with tracing
lowkey start --metrics 127.0.0.1:9600 --trace /path/to/watch
```

### Accessing Metrics

Metrics are exposed at the `/metrics` endpoint:

```bash
# View metrics in terminal
curl http://127.0.0.1:9600/metrics

# Save to file
curl http://127.0.0.1:9600/metrics -o metrics.txt

# Watch metrics continuously
watch -n 2 'curl -s http://127.0.0.1:9600/metrics'
```

### Available Metrics

#### `lowkey_events_total`
- **Type**: Counter
- **Description**: Total number of filesystem change events processed
- **Use case**: Monitor event throughput, detect activity spikes

#### `lowkey_errors_total`
- **Type**: Counter
- **Description**: Total errors encountered during monitoring
- **Use case**: Track error rates, trigger alerts on error spikes

#### `lowkey_event_latency_seconds`
- **Type**: Gauge
- **Description**: Average time spent processing each event
- **Use case**: Identify performance degradation, capacity planning

#### `lowkey_event_latency_samples`
- **Type**: Counter
- **Description**: Number of samples contributing to latency calculation
- **Use case**: Validate latency metric significance

### Example Output

```
# HELP lowkey_events_total Total filesystem change events processed.
# TYPE lowkey_events_total counter
lowkey_events_total 15423

# HELP lowkey_errors_total Total errors encountered while monitoring.
# TYPE lowkey_errors_total counter
lowkey_errors_total 2

# HELP lowkey_event_latency_seconds Average latency per event.
# TYPE lowkey_event_latency_seconds gauge
lowkey_event_latency_seconds 0.000312

# HELP lowkey_event_latency_samples Number of samples contributing to latency metric.
# TYPE lowkey_event_latency_samples counter
lowkey_event_latency_samples 15423
```

## Prometheus Integration

### Prometheus Configuration

Add lowkey as a scrape target in `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'lowkey'
    static_configs:
      - targets: ['127.0.0.1:9600']
        labels:
          instance: 'lowkey-production'
          environment: 'prod'
    scrape_interval: 15s
    scrape_timeout: 10s
```

### Example PromQL Queries

```promql
# Event rate (events per second)
rate(lowkey_events_total[5m])

# Error rate (errors per second)
rate(lowkey_errors_total[5m])

# Average latency over time
lowkey_event_latency_seconds

# Error percentage
(rate(lowkey_errors_total[5m]) / rate(lowkey_events_total[5m])) * 100
```

### Grafana Dashboard

Create a Grafana dashboard with these panels:

1. **Event Throughput**
   - Query: `rate(lowkey_events_total[5m])`
   - Visualization: Graph (line)
   - Unit: events/sec

2. **Error Rate**
   - Query: `rate(lowkey_errors_total[5m])`
   - Visualization: Graph (line)
   - Unit: errors/sec
   - Alert: Error rate > threshold

3. **Event Latency**
   - Query: `lowkey_event_latency_seconds * 1000`
   - Visualization: Graph (line)
   - Unit: milliseconds

4. **Total Events**
   - Query: `lowkey_events_total`
   - Visualization: Single stat
   - Unit: events

## Tracing (`--trace`)

### Usage

Enable tracing for detailed execution logs:

```bash
# Enable trace logging
lowkey start --trace /path/to/watch

# Combine with metrics
lowkey start --metrics 127.0.0.1:9600 --trace /path/to/watch
```

### What Gets Traced

When tracing is enabled, lowkey logs detailed span information:

- Event processing pipeline stages
- Watcher lifecycle operations
- Supervisor actions (restarts, backoff)
- State persistence operations

### Viewing Trace Logs

```bash
# Follow trace logs in real-time
lowkey tail

# Search for specific operations
lowkey tail | grep "span:"

# Filter by component
lowkey tail | grep "supervisor"
```

### Example Trace Output

```
2025-01-15T10:23:45.123Z [trace] span: event_process start path=/foo/bar.txt
2025-01-15T10:23:45.125Z [trace] span: event_process end duration=2ms
2025-01-15T10:23:45.130Z [trace] span: supervisor_heartbeat state=running restarts=0
```

## Use Cases

### Development and Debugging

```bash
# Enable full observability during development
lowkey start --metrics 127.0.0.1:9600 --trace ./project

# Monitor in separate terminals
lowkey tail          # Terminal 1: Watch logs
curl -s http://127.0.0.1:9600/metrics  # Terminal 2: Check metrics
```

### Production Monitoring

```bash
# Production daemon with metrics (no trace overhead)
lowkey start --metrics 127.0.0.1:9600 /var/data

# Configure Prometheus alerting rules
# Alert on high error rate or latency spikes
```

### Performance Profiling

```bash
# Profile event processing performance
lowkey start --metrics 127.0.0.1:9600 --trace /large/directory

# Analyze metrics over time
while true; do
  curl -s http://127.0.0.1:9600/metrics | grep latency
  sleep 5
done
```

### Capacity Planning

```bash
# Monitor resource usage under load
lowkey start --metrics 127.0.0.1:9600 /busy/directory

# Track event throughput trends in Prometheus
# Use data to size infrastructure appropriately
```

## Security Considerations

### Metrics Endpoint Security

The metrics endpoint is **unauthenticated** and should only be exposed on trusted networks:

```bash
# Secure: localhost only (recommended)
lowkey start --metrics 127.0.0.1:9600

# Insecure: accessible from network
lowkey start --metrics 0.0.0.0:9600

# Use firewall rules or reverse proxy for external access
```

### Sensitive Data

Metrics do **not** include:
- File paths
- File contents
- User information

Only aggregate counters and timing information are exposed.

## Performance Impact

### Metrics Overhead

- **Memory**: ~500KB for metrics collector
- **CPU**: <0.1% additional overhead
- **Network**: ~2KB per scrape

### Trace Overhead

- **I/O**: Increased log file writes
- **CPU**: ~1-2% additional overhead for logging
- **Disk**: Higher log rotation frequency

**Recommendation**: Use `--metrics` in production, add `--trace` only for debugging.

## Troubleshooting

### Metrics endpoint not accessible

```bash
# Check if daemon is running
lowkey status

# Verify port is listening
lsof -i :9600  # macOS/Linux
netstat -an | grep 9600  # Windows

# Test local connectivity
curl -v http://127.0.0.1:9600/metrics
```

### No metrics appearing

Ensure the `--metrics` flag was passed when starting the daemon:

```bash
# Check process environment
ps aux | grep lowkey  # macOS/Linux
Get-Process lowkey | Format-List *  # Windows

# Restart with metrics enabled
lowkey stop
lowkey start --metrics 127.0.0.1:9600 /path/to/watch
```

### Trace logs not showing detail

Verify `--trace` flag is enabled:

```bash
# Check daemon status
lowkey status

# Look for trace spans in logs
lowkey tail | grep "\[trace\]"
```
