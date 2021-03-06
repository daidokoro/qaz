package log

import (
	"runtime"
	"strings"

	"github.com/fatih/color"
)

// Color type used in ColorString function
type Color uint8

const (
	// GREEN - log color green
	GREEN Color = 0

	// CYAN - log color cyan
	CYAN Color = 1

	// RED - log color red
	RED Color = 2

	// MAGENTA - log color megenta
	MAGENTA Color = 3

	// YELLOW - log color yellow
	YELLOW Color = 4
)

// ColorMap - Used to map a particular color to a cf status phrase - returns lowercase strings in color.
func ColorMap(s string) string {
	// If Windows, disable colorS
	if runtime.GOOS == "windows" || Default().Colors() {
		return strings.ToLower(s)
	}

	var result string

	switch s {
	case
		"CREATE_COMPLETE",
		"DELETE_COMPLETE",
		"UPDATE_COMPLETE",
		"UPDATE_ROLLBACK_COMPLETE":
		result = color.New(color.FgGreen).SprintFunc()(s)
	case
		"CREATE_IN_PROGRESS",
		"DELETE_IN_PROGRESS",
		"REVIEW_IN_PROGRESS",
		"UPDATE_ROLLBACK_IN_PROGRESS",
		"UPDATE_COMPLETE_CLEANUP_IN_PROGRESS",
		"UPDATE_IN_PROGRESS":
		result = color.New(color.FgYellow).SprintFunc()(s)
	default:
		// NOTE: all other status are red
		result = color.New(color.FgRed).SprintFunc()(s)
	}
	return strings.ToLower(result)
}

// ColorString - Returns colored string
func ColorString(s string, col Color) string {
	// If Windows, disable colorS
	if runtime.GOOS == "windows" || Default().Colors() {
		return s
	}

	var result string
	switch col {
	case GREEN:
		result = color.New(color.FgGreen).Add().SprintFunc()(s)
	case YELLOW:
		result = color.New(color.FgYellow).Add().SprintFunc()(s)
	case RED:
		result = color.New(color.FgRed).Add().SprintFunc()(s)
	case MAGENTA:
		result = color.New(color.FgMagenta).Add().SprintFunc()(s)
	case CYAN:
		result = color.New(color.FgCyan).Add().SprintFunc()(s)
	}

	return result
}
