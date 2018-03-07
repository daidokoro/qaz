package logger

import (
	"fmt"
	"os"
)

// Simple logging and printing mechanisms

// Logger contains logging flags, colors, debug
type Logger struct {
	Colors    *bool
	DebugMode *bool
}

// Info - Prints info level log statments
func (l *Logger) Info(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "%s: %s\n", l.ColorString("info", "green"), fmt.Sprintf(msg, args...))
}

// Warn - Prints warn level log statments
func (l *Logger) Warn(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, "%s: %s\n", l.ColorString("warn", "yellow"), fmt.Sprintf(msg, args...))
}

// Error - Prints error level log statements
func (l *Logger) Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", l.ColorString("error", "red"), fmt.Sprintf(msg, args...))
}

// Debug - Prints debug level log statements
func (l *Logger) Debug(msg string, args ...interface{}) {
	if *l.DebugMode {
		fmt.Fprintf(os.Stdout, "%s: %s\n", l.ColorString("debug", "magenta"), fmt.Sprintf(msg, args...))
	}
}

// New creates a Logger Object
func New(debug, colors bool) *Logger {
	return &Logger{
		Colors:    &colors,
		DebugMode: &debug,
	}
}
