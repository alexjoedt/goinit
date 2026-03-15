package main

import (
	"fmt"
	"io"
	"os"
)

// ANSI color codes for terminal output.
const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorCyan  = "\033[36m"
	colorGray  = "\033[37m"
	colorBold  = "\033[1m"
)

// colorEnabled reports whether ANSI color output should be emitted to stdout.
// It returns false when NO_COLOR is set, TERM is "dumb", or stdout is not
// connected to an interactive terminal.
func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// colorize wraps text in ANSI color codes when color output is enabled.
func colorize(color, text string) string {
	if !colorEnabled() {
		return text
	}
	return color + text + colorReset
}

// step represents a single named progress step.
type step struct {
	name string
	done bool
}

// ProgressTracker tracks and displays progress through a sequence of steps.
type ProgressTracker struct {
	steps   []step
	current int
	w       io.Writer
}

// NewProgressTracker creates a new ProgressTracker for the given step names.
func NewProgressTracker(steps []string, w io.Writer) *ProgressTracker {
	s := make([]step, len(steps))
	for i, name := range steps {
		s[i] = step{name: name}
	}
	return &ProgressTracker{steps: s, w: w}
}

// Start prints the initializing header when verbose mode is active.
func (p *ProgressTracker) Start(config *Config) {
	if config.Verbose {
		_, _ = fmt.Fprintln(p.w, colorize(colorBold+colorCyan, "Initializing Go project..."))
		_, _ = fmt.Fprintln(p.w)
	}
}

// NextStep marks the current step as complete and advances the tracker.
func (p *ProgressTracker) NextStep(config *Config) {
	if p.current < len(p.steps) {
		p.steps[p.current].done = true
		if config.Verbose {
			_, _ = fmt.Fprintf(p.w, "   %s %s\n",
				colorize(colorGreen, "✓"),
				colorize(colorGray, p.steps[p.current].name))
		}
		p.current++
	}
}

// Complete prints the final success message.
func (p *ProgressTracker) Complete(config *Config) {
	_, _ = fmt.Fprintf(p.w, "Project '%s' created at %s\n",
		colorize(colorGreen+colorBold, config.ProjectName),
		config.TargetDir)
	if config.StandardLayout {
		_, _ = fmt.Fprintf(p.w, "%s\n", colorize(colorGray, "  cmd/"+config.ProjectName+"/main.go"))
		_, _ = fmt.Fprintf(p.w, "%s\n", colorize(colorGray, "  internal/"))
	}
}
