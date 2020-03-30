// Package log contains a simple logging mechanism for qaz
package log

// Simple logging and printing mechanisms

func init() {
	// set default logger
	SetDefault(NewDefaultLogger(false, true))
}

// Logger log interface
type Logger interface {
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Debug(string, ...interface{})
	Colors() bool
}

var df Logger = new(DefaultLogger)

// SetDefault - sets default logger
func SetDefault(logger Logger) {
	df = logger
}

// Default - returns the default logger
func Default() Logger {
	return df
}

// Info - calls default logger Info method
func Info(m string, args ...interface{}) {
	df.Info(m, args...)
}

// Warn - calls default logger Warn method
func Warn(m string, args ...interface{}) {
	df.Warn(m, args...)
}

// Error - calls default logger Error method
func Error(m string, args ...interface{}) {
	df.Error(m, args...)
}

// Debug - calls default logger Debug method
func Debug(m string, args ...interface{}) {
	df.Debug(m, args...)
}
