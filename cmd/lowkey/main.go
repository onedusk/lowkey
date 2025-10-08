package main

import (
	"fmt"
	"os"
)

// main is the entry point for the lowkey application. It determines whether to
// run as a background daemon or as a command-line client and executes the
// appropriate logic.
func main() {
	if os.Getenv(daemonEnvKey) == "1" {
		if err := runDaemonProcess(); err != nil {
			fmt.Fprintf(os.Stderr, "lowkey daemon: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "lowkey: %v\n", err)
		os.Exit(1)
	}
}
