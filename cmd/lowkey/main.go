package main

import (
	"fmt"
	"os"
)

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "lowkey: %v\n", err)
		os.Exit(1)
	}
}
