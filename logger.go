package main

import (
	"log"
	"os"
	"runtime"
)

var (
	infoLog  *log.Logger
	debugLog *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
)

var resetColor = "\033[0m"
var red = "\033[31m"
var green = "\033[32m"
var yellow = "\033[33m"
var cyan = "\033[36m"

func init() {
	infoLog = log.New(os.Stdout, setColor(green, "INFO "), log.Ldate|log.Ltime|log.Lmsgprefix)
	debugLog = log.New(os.Stdout, setColor(cyan, "DEBUG "), log.Ldate|log.Ltime|log.Lmsgprefix)
	warnLog = log.New(os.Stdout, setColor(yellow, "WARN "), log.Ldate|log.Ltime|log.Lmsgprefix)
	errorLog = log.New(os.Stderr, setColor(red, "ERROR "), log.Ldate|log.Ltime|log.Lmsgprefix)

	if runtime.GOOS == "windows" {
		resetColor = ""
		red = ""
		green = ""
		yellow = ""
		cyan = ""
	}
}

func logInfo(format string, args ...interface{}) {
	if verbose {
		infoLog.Printf(format, args...)
	}
}

func logDebug(format string, args ...interface{}) {
	if debug {
		debugLog.Printf(format, args...)
	}
}

func logWarn(format string, args ...interface{}) {
	warnLog.Printf(format, args...)
}

func logErr(format string, args ...interface{}) {
	errorLog.Printf(format, args...)
}

func setColor(color string, format string) string {
	return color + format + resetColor
}
