package commands

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/ttacon/chalk"
)

// Simple logging and printing mechanisms

// Used for mapping log level, may or may not expand in the future..
var level = struct {
	debug string
	warn  string
	err   string
	info  string
}{"debug", "warn", "error", "info"}

// handleError - handleError the err and exits the app if err not nil
func handleError(e error) {
	if e != nil {
		fmt.Printf("%s: %s\n", colorString(level.err, "red"), e.Error())
		return
	}
	return
}

// Log - Handles all logging accross app
func Log(msg, lvl string) {

	// NOTE do nothin with debug msgs if not set to debug
	if lvl == level.debug && !run.debug {
		return
	}

	switch lvl {
	case "debug":
		// l.Debugln(msg)
		fmt.Printf("%s: %s\n", colorString("debug", "magenta"), msg)
	case "warn":
		// l.Warnln(msg)
		fmt.Printf("%s: %s\n", colorString("warn", "yellow"), msg)
	case "error":
		// l.Errorln(msg)
		fmt.Printf("%s: %s\n", colorString("error", "red"), msg)
	default:
		// l.Infoln(msg)
		fmt.Printf("%s: %s\n", colorString("info", "green"), msg)
	}
}

// colorMap - Used to map a particular color to a cf status phrase - returns lowercase strings in color.
func colorMap(s string) string {

	// If Windows, disable colorS
	if runtime.GOOS == "windows" || run.colors {
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

// colorString - Returns colored string
func colorString(s string, color string) string {

	// If Windows, disable colorS
	if runtime.GOOS == "windows" || run.colors {
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
	default:
		// Unidentified, just returns the same string
		return s
	}

	return result
}
