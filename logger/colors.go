package logger

import (
	"runtime"
	"strings"

	"github.com/fatih/color"
)

// ColorMap - Used to map a particular color to a cf status phrase - returns lowercase strings in color.
func (l *Logger) ColorMap(s string) string {

	// If Windows, disable colorS
	if runtime.GOOS == "windows" || *l.Colors {
		return strings.ToLower(s)
	}

	v := strings.Split(s, "_")[len(strings.Split(s, "_"))-1]

	var result string

	switch v {
	case "COMPLETE":
		result = color.New(color.BgGreen).Add(color.Bold).SprintFunc()(s)
	case "PROGRESS":
		result = color.New(color.BgYellow).Add(color.Bold).SprintFunc()(s)
	case "FAILED":
		result = color.New(color.BgRed).Add(color.Bold).SprintFunc()(s)
	case "SKIPPED":
		result = color.New(color.BgHiBlue).Add(color.Bold).SprintFunc()(s)
	default:
		// Unidentified, just returns the same string
		return strings.ToLower(s)
	}
	return strings.ToLower(result)
}

// ColorString - Returns colored string
func (l *Logger) ColorString(s, col string) string {

	// If Windows, disable colorS
	if runtime.GOOS == "windows" || *l.Colors {
		return s
	}

	var result string
	switch strings.ToLower(col) {
	case "green":
		result = color.New(color.BgGreen).Add(color.Bold).SprintFunc()(s)
	case "yellow":
		result = color.New(color.BgYellow).Add(color.Bold).SprintFunc()(s)
	case "red":
		result = color.New(color.BgRed).Add(color.Bold).SprintFunc()(s)
	case "magenta":
		result = color.New(color.BgMagenta).Add(color.Bold).SprintFunc()(s)
	case "cyan":
		result = color.New(color.BgCyan).Add(color.Bold).SprintFunc()(s)
	default:
		// Unidentified, just returns the same string
		return s
	}

	return result
}
