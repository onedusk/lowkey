package main

const (
	daemonEnvKey        = "LOWKEY_DAEMON"
	daemonManifestEnv   = "LOWKEY_MANIFEST"
	daemonPIDFilename   = "daemon.pid"
	daemonShutdownGrace = 5 // seconds to wait for graceful shutdown
	daemonMetricsEnv    = "LOWKEY_METRICS_ADDR"
	daemonTraceEnv      = "LOWKEY_TRACE_ENABLED"
)
