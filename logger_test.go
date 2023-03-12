package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogInfo(t *testing.T) {
	verbose = true
	defer func() {
		verbose = false
		infoLog.SetOutput(os.Stdout)
	}()
	buf := new(bytes.Buffer)
	infoLog.SetOutput(buf)

	logInfo("test-log")
	assert.Contains(t, buf.String(), "INFO")
	assert.Contains(t, buf.String(), "test-log")
}

func TestLogDebug(t *testing.T) {
	debug = true
	defer func() {
		debug = false
		debugLog.SetOutput(os.Stdout)
	}()
	buf := new(bytes.Buffer)
	debugLog.SetOutput(buf)

	logDebug("test-log")
	assert.Contains(t, buf.String(), "DEBUG")
	assert.Contains(t, buf.String(), "test-log")
}

func TestLogWarn(t *testing.T) {
	buf := new(bytes.Buffer)
	warnLog.SetOutput(buf)
	defer warnLog.SetOutput(os.Stdout)

	logWarn("test-log")
	assert.Contains(t, buf.String(), "WARN")
	assert.Contains(t, buf.String(), "test-log")
}

func TestLogError(t *testing.T) {
	buf := new(bytes.Buffer)
	errorLog.SetOutput(buf)
	defer errorLog.SetOutput(os.Stderr)

	logErr("test-log")
	assert.Contains(t, buf.String(), "ERROR")
	assert.Contains(t, buf.String(), "test-log")
}
