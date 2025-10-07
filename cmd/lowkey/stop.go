package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"lowkey/internal/state"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir, err := state.DefaultStateDir()
			if err != nil {
				return err
			}
			store, err := state.NewManifestStore(stateDir)
			if err != nil {
				return err
			}
			pid, ok := readPID(stateDir)
			if !ok {
				fmt.Println("stop: daemon is not running")
				_ = store.Clear()
				return nil
			}

			if err := signalDaemon(pid); err != nil && !errors.Is(err, os.ErrProcessDone) {
				return err
			}

			deadline := time.Now().Add(time.Duration(daemonShutdownGrace) * time.Second)
			for processAlive(pid) && time.Now().Before(deadline) {
				time.Sleep(200 * time.Millisecond)
			}
			if processAlive(pid) {
				_ = forceKill(pid)
			}

			if err := os.Remove(pidFilePath(stateDir)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := store.Clear(); err != nil {
				return err
			}
			manifestFromConfig = nil
			fmt.Println("daemon stopped")
			return nil
		},
	}
}

func signalDaemon(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return process.Kill()
	}
	return process.Signal(syscall.SIGTERM)
}

func forceKill(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}
