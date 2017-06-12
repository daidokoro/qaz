package logger

import (
	"runtime"
	"strings"

	"github.com/ttacon/chalk"
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
		result = chalk.Green.Color(s)
	case "PROGRESS":
		result = chalk.Yellow.Color(s)
	case "FAILED":
		result = chalk.Red.Color(s)
	case "SKIPPED":
		result = chalk.Blue.Color(s)
	default:
		// Unidentified, just returns the same string
		return strings.ToLower(s)
	}
	return strings.ToLower(result)
}

// ColorString - Returns colored string
func (l *Logger) ColorString(s, color string) string {

	// If Windows, disable colorS
	if runtime.GOOS == "windows" || *l.Colors {
		return s
	}

	var result string
	switch strings.ToLower(color) {
	case "green":
		result = chalk.Green.Color(s)
	case "yellow":
		result = chalk.Yellow.Color(s)
	case "red":
		result = chalk.Red.Color(s)
	case "magenta":
		result = chalk.Magenta.Color(s)
	case "cyan":
		result = chalk.Cyan.Color(s)
	default:
		// Unidentified, just returns the same string
		return s
	}

	return result
}
