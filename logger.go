package main

import (
	"fmt"
	"io"
)

// info prints informational messages only when verbose mode is enabled.
func info(w io.Writer, config *Config, format string, args ...any) {
	if config != nil && config.Verbose {
		_, _ = fmt.Fprintf(w, "• "+format+"\n", args...)
	}
}

// warn prints warning messages - always shown.
func warn(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, "Warning: "+format+"\n", args...)
}
