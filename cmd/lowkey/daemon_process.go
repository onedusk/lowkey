package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"lowkey/internal/daemon"
	"lowkey/internal/state"
	"lowkey/pkg/config"
	"lowkey/pkg/telemetry"
)

func runDaemonProcess() error {
	manifestPath := os.Getenv(daemonManifestEnv)
	if manifestPath == "" {
		return fmt.Errorf("daemon: manifest path not provided")
	}

	stateDir := filepath.Dir(manifestPath)
	store, err := state.NewManifestStore(stateDir)
	if err != nil {
		return err
	}

	manifest, err := config.LoadManifest(manifestPath)
	if err != nil {
		return err
	}

	if existing, ok := readPID(stateDir); ok && processAlive(existing) {
		return fmt.Errorf("daemon: process already running with pid %d", existing)
	}

	metricsAddr := os.Getenv(daemonMetricsEnv)
	var metrics *telemetry.Collector
	if metricsAddr != "" {
		collector := telemetry.NewCollector()
		if err := collector.Start(metricsAddr); err != nil {
			return fmt.Errorf("daemon: start metrics server: %w", err)
		}
		metrics = collector
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = collector.Stop(ctx)
		}()
	}

	traceEnabled := os.Getenv(daemonTraceEnv) == "1"
	tracer := telemetry.NewTracer(telemetry.TracerOptions{Enabled: traceEnabled})

	cleanupPID, err := writePIDFile(stateDir)
	if err != nil {
		return err
	}
	defer cleanupPID()

	manager, err := daemon.NewManager(store, manifest)
	if err != nil {
		return err
	}
	manager.SetTelemetry(metrics, tracer)
	if err := manager.Start(); err != nil {
		return err
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()

	done := make(chan struct{})
	go func() {
		manager.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Duration(daemonShutdownGrace) * time.Second):
	}
	return nil
}

func writePIDFile(stateDir string) (func(), error) {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, err
	}
	path := pidFilePath(stateDir)
	record := []byte(strconv.Itoa(os.Getpid()))
	if err := os.WriteFile(path, record, 0o644); err != nil {
		return nil, err
	}
	return func() {
		_ = os.Remove(path)
	}, nil
}

func pidFilePath(stateDir string) string {
	return filepath.Join(stateDir, daemonPIDFilename)
}

func readPID(stateDir string) (int, bool) {
	data, err := os.ReadFile(pidFilePath(stateDir))
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	return pid, true
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		// Windows does not support signal 0 reliably; assume alive if process handle resolves.
		return true
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}
