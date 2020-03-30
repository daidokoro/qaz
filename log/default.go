package log

import (
	"fmt"
	"os"
)

// DefaultLogger - default logger type
type DefaultLogger struct {
	colors    *bool
	debugMode *bool
}

// Info - Prints info level log statments
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "%s: %s\n", ColorString("info", "green"), fmt.Sprintf(msg, args...))
}

// Warn - Prints warn level log statments
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "%s: %s\n", ColorString("warn", "yellow"), fmt.Sprintf(msg, args...))
}

// Error - Prints error level log statements
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", ColorString("error", "red"), fmt.Sprintf(msg, args...))
}

// Debug - Prints debug level log statements
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	if *l.debugMode {
		fmt.Fprintf(os.Stdout, "%s: %s\n", ColorString("debug", "magenta"), fmt.Sprintf(msg, args...))
	}
}

// Colors - returns bool indicating whether colors should be applied to logs
func (l *DefaultLogger) Colors() bool {
	return *l.colors
}

// NewDefaultLogger - creates a Logger Object
func NewDefaultLogger(debug, colors bool) Logger {
	return &DefaultLogger{
		colors:    &colors,
		debugMode: &debug,
	}
}
