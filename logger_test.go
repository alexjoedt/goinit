package main

import (
	"bytes"
	"testing"
)

func TestInfo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	config := &Config{Verbose: true}

	info(&buf, config, "test message")

	output := buf.String()
	assertContains(t, output, "test message")
	assertContains(t, output, "•")
}

func TestInfoVerboseOff(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	config := &Config{Verbose: false}

	info(&buf, config, "test message")

	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestInfoNilConfig(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	info(&buf, nil, "should not print")

	if buf.Len() != 0 {
		t.Errorf("expected no output for nil config, got %q", buf.String())
	}
}

func TestInfoFormatArgs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	config := &Config{Verbose: true}

	info(&buf, config, "hello %s %d", "world", 42)

	assertContains(t, buf.String(), "hello world 42")
}

func TestWarn(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	warn(&buf, "test warning")

	output := buf.String()
	assertContains(t, output, "Warning:")
	assertContains(t, output, "test warning")
}

func TestWarnFormatArgs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	warn(&buf, "issue %d: %s", 99, "something broke")

	assertContains(t, buf.String(), "issue 99: something broke")
}
