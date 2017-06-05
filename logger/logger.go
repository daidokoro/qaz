package logger

import "fmt"

// Simple logging and printing mechanisms

// Logger contains logging flags, colors, debug
type Logger struct {
	colors    bool
	DebugMode *bool
}

// Info - Prints info level log statments
func (l *Logger) Info(msg string) {
	fmt.Printf("%s: %s\n", l.ColorString("info", "green"), msg)
}

// Warn - Prints warn level log statments
func (l *Logger) Warn(msg string) {
	fmt.Printf("%s: %s\n", l.ColorString("warn", "yellow"), msg)
}

// Error - Prints error level log statements
func (l *Logger) Error(msg string) {
	fmt.Printf("%s: %s\n", l.ColorString("error", "red"), msg)
}

// Debug - Prints debug level log statements
func (l *Logger) Debug(msg string) {
	if *l.DebugMode {
		fmt.Printf("%s: %s\n", l.ColorString("debug", "magenta"), msg)
	}
}

// New creates a Logger Object
func New(debug, colors bool) *Logger {
	return &Logger{
		colors:    colors,
		DebugMode: &debug,
	}
}
