package main

import (
	"fmt"
	"os"
)

// Color codes for terminal output
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Purple  = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	Bold    = "\033[1m"
)

// isTerminal checks if the output is being written to a terminal
func isTerminal() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	// Simple check - could be improved
	return true
}

// colorize applies color to text if terminal supports it
func colorize(color, text string) string {
	if !isTerminal() {
		return text
	}
	return color + text + Reset
}

// Step represents a progress step
type Step struct {
	Name        string
	Description string
	Done        bool
}

// ProgressTracker tracks and displays progress
type ProgressTracker struct {
	Steps   []Step
	Current int
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(steps []string) *ProgressTracker {
	var progSteps []Step
	for _, step := range steps {
		progSteps = append(progSteps, Step{Name: step, Done: false})
	}
	return &ProgressTracker{Steps: progSteps, Current: 0}
}

// Start begins tracking progress
func (p *ProgressTracker) Start(config *Config) {
	if !isTerminal() {
		return
	}
	if config.Verbose {
		fmt.Println(colorize(Bold+Cyan, "Initializing Go project..."))
		fmt.Println()
	}
}

// NextStep marks the current step as done and moves to the next
func (p *ProgressTracker) NextStep(config *Config) {
	if p.Current < len(p.Steps) {
		p.Steps[p.Current].Done = true
		if config.Verbose {
			fmt.Printf("   %s %s\n", 
				colorize(Green, "✓"), 
				colorize(Gray, p.Steps[p.Current].Name))
		}
		p.Current++
	}
}

// Complete shows the final success message
func (p *ProgressTracker) Complete(config *Config) {
	fmt.Printf("Project '%s' created at %s\n", 
		colorize(Green+Bold, config.ProjectName), 
		config.TargetDir)
}
