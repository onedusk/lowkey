package main

import (
	"fmt"
	"os"
)

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
