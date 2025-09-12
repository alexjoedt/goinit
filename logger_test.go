package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	// Test that info shows output when verbose is true
	config := &Config{Verbose: true}
	
	// Capture stderr
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
	
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "•")
}

func TestInfoVerboseOff(t *testing.T) {
	// Test that info shows no output when verbose is false
	config := &Config{Verbose: false}
	
	// Capture stderr
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
	
	assert.Empty(t, output)
}

func TestWarn(t *testing.T) {
	// Capture stderr
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
	
	assert.Contains(t, output, "Warning:")
	assert.Contains(t, output, "test warning")
}
