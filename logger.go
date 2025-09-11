package main

import (
	"fmt"
	"os"
)

// info prints informational messages only when verbose mode is enabled
func info(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, "• "+format+"\n", args...)
	}
}

// warn prints warning messages to stderr - always shown
func warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
}

// fatal prints error message to stderr and exits with code 1
func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
