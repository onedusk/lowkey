package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"lowkey/internal/state"
	"lowkey/pkg/config"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [dir ...]",
		Short: "Launch the background daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			metricsAddr, traceEnabled, args := parseStartFlags(args)
			manifestPath, remaining := extractOption(args, "--manifest", "-m")
			manifest, err := resolveManifest(manifestPath, remaining)
			if err != nil {
				return err
			}

			stateDir, err := state.DefaultStateDir()
			if err != nil {
				return err
			}
			store, err := state.NewManifestStore(stateDir)
			if err != nil {
				return err
			}

			if pid, ok := readPID(stateDir); ok && processAlive(pid) {
				return fmt.Errorf("start: daemon already running with pid %d", pid)
			}

			if err := store.Save(manifest); err != nil {
				return err
			}
			manifestFromConfig = manifest

			proc := exec.Command(os.Args[0])
			env := append(os.Environ(),
				fmt.Sprintf("%s=1", daemonEnvKey),
				fmt.Sprintf("%s=%s", daemonManifestEnv, store.Path()),
			)
			if metricsAddr != "" {
				env = append(env, fmt.Sprintf("%s=%s", daemonMetricsEnv, metricsAddr))
			}
			if traceEnabled {
				env = append(env, fmt.Sprintf("%s=1", daemonTraceEnv))
			}
			proc.Env = env
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			if err := proc.Start(); err != nil {
				return fmt.Errorf("start: launch daemon: %w", err)
			}
			fmt.Printf("daemon launching (pid %d)\n", proc.Process.Pid)
			// Give the process a moment to persist its pid file before returning.
			time.Sleep(250 * time.Millisecond)
			return nil
		},
	}

	return cmd
}

func parseStartFlags(args []string) (metricsAddr string, traceEnabled bool, remaining []string) {
	remaining = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--metrics":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				metricsAddr = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--metrics="):
			metricsAddr = arg[len("--metrics="):]
		case arg == "--trace":
			traceEnabled = true
		case strings.HasPrefix(arg, "--trace="):
			val := strings.ToLower(arg[len("--trace="):])
			traceEnabled = val != "false" && val != "0"
		default:
			remaining = append(remaining, arg)
		}
	}
	return metricsAddr, traceEnabled, remaining
}

func resolveManifest(manifestPath string, args []string) (*config.Manifest, error) {
	if manifestPath != "" {
		return config.LoadManifest(manifestPath)
	}
	if manifestFromConfig != nil {
		return manifestFromConfig, nil
	}
	if len(args) == 0 {
		return nil, errors.New("start: provide directories or a manifest via --manifest/--config")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("start: determine working directory: %w", err)
	}
	return config.BuildManifestFromArgs(cwd, args)
}
