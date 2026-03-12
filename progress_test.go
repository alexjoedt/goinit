package main

import (
	"bytes"
	"testing"
)

func TestNewProgressTracker(t *testing.T) {
	t.Parallel()

	steps := []string{"step1", "step2", "step3"}
	var buf bytes.Buffer
	p := NewProgressTracker(steps, &buf)

	if len(p.steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(p.steps))
	}
	if p.current != 0 {
		t.Fatalf("expected current to be 0, got %d", p.current)
	}
	for i, s := range p.steps {
		assertEqual(t, steps[i], s.name)
		if s.done {
			t.Errorf("step %d should not be done", i)
		}
	}
}

func TestProgressTracker_Start_Verbose(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewProgressTracker([]string{"a"}, &buf)
	config := &Config{Verbose: true}

	p.Start(config)

	assertContains(t, buf.String(), "Initializing Go project...")
}

func TestProgressTracker_Start_NonVerbose(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewProgressTracker([]string{"a"}, &buf)
	config := &Config{Verbose: false}

	p.Start(config)

	if buf.Len() != 0 {
		t.Errorf("expected no output in non-verbose mode, got %q", buf.String())
	}
}

func TestProgressTracker_NextStep_Verbose(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewProgressTracker([]string{"Creating files", "Init module"}, &buf)
	config := &Config{Verbose: true}

	p.NextStep(config)
	assertContains(t, buf.String(), "Creating files")

	if p.current != 1 {
		t.Errorf("expected current to be 1, got %d", p.current)
	}
	if !p.steps[0].done {
		t.Error("step 0 should be done")
	}
}

func TestProgressTracker_NextStep_NonVerbose(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewProgressTracker([]string{"Creating files"}, &buf)
	config := &Config{Verbose: false}

	p.NextStep(config)

	if buf.Len() != 0 {
		t.Errorf("expected no output in non-verbose mode, got %q", buf.String())
	}
	if p.current != 1 {
		t.Errorf("expected current to be 1, got %d", p.current)
	}
}

func TestProgressTracker_NextStep_BeyondEnd(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewProgressTracker([]string{"only step"}, &buf)
	config := &Config{Verbose: true}

	p.NextStep(config)
	buf.Reset()

	// Calling again after all steps are done should be a no-op
	p.NextStep(config)
	if buf.Len() != 0 {
		t.Errorf("expected no output after all steps done, got %q", buf.String())
	}
	if p.current != 1 {
		t.Errorf("expected current to remain 1, got %d", p.current)
	}
}

func TestProgressTracker_Complete(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewProgressTracker([]string{"a"}, &buf)
	config := &Config{
		ProjectName: "my-project",
		TargetDir:   "/tmp/my-project",
	}

	p.Complete(config)

	output := buf.String()
	assertContains(t, output, "my-project")
	assertContains(t, output, "/tmp/my-project")
}

func TestProgressTracker_FullSequence(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	steps := []string{"step one", "step two", "step three"}
	p := NewProgressTracker(steps, &buf)
	config := &Config{
		ProjectName: "test-proj",
		TargetDir:   "/tmp/test-proj",
		Verbose:     true,
	}

	p.Start(config)
	for range steps {
		p.NextStep(config)
	}
	p.Complete(config)

	output := buf.String()
	assertContains(t, output, "Initializing Go project...")
	assertContains(t, output, "step one")
	assertContains(t, output, "step two")
	assertContains(t, output, "step three")
	assertContains(t, output, "test-proj")
}

func TestColorize_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	result := colorize(colorGreen, "hello")
	assertEqual(t, "hello", result)
}

func TestColorize_DumbTerm(t *testing.T) {
	t.Setenv("TERM", "dumb")
	result := colorize(colorGreen, "hello")
	assertEqual(t, "hello", result)
}
