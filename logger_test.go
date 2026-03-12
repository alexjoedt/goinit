package main

import (
	"os"
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	config := &Config{Verbose: true}

	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = origStderr
	}()

	info(config, "test message")

	w.Close()
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain %q, got %q", "test message", output)
	}
	if !strings.Contains(output, "•") {
		t.Errorf("expected output to contain %q, got %q", "•", output)
	}
}

func TestInfoVerboseOff(t *testing.T) {
	config := &Config{Verbose: false}

	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = origStderr
	}()

	info(config, "test message")

	w.Close()
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if output != "" {
		t.Errorf("expected no output, got %q", output)
	}
}

func TestWarn(t *testing.T) {
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = origStderr
	}()

	warn("test warning")

	w.Close()
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Warning:") {
		t.Errorf("expected output to contain %q, got %q", "Warning:", output)
	}
	if !strings.Contains(output, "test warning") {
		t.Errorf("expected output to contain %q, got %q", "test warning", output)
	}
}
