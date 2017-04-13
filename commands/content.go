package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

// global variables
var (
	region    string
	project   string
	stackname string
	stacks    map[string]*stack
)

// used for stack keyword referencing
var keyword = struct {
	depends        string
	parameters     string
	cloudformation string
}{
	"depends_on",
	"parameters",
	"cf",
}

// fetchContent - checks the source type, url/s3/file and calls the corresponding function
func fetchContent(source string) (string, error) {
	switch strings.Split(strings.ToLower(source), ":")[0] {
	case "http":
		Log(fmt.Sprintln("Source Type: [http] Detected, Fetching Source: ", source), level.debug)
		resp, err := Get(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "https":
		Log(fmt.Sprintln("Source Type: [https] Detected, Fetching Source: ", source), level.debug)
		resp, err := Get(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	case "s3":
		Log(fmt.Sprintln("Source Type: [s3] Detected, Fetching Source: ", source), level.debug)
		resp, err := S3Read(source)
		if err != nil {
			return "", err
		}
		return resp, nil
	default:
		Log(fmt.Sprintln("Source Type: [file] Detected, Fetching Source: ", source), level.debug)
		b, err := ioutil.ReadFile(source)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

// getName  - Checks if arg is url or file and returns stack name and filepath/url
func getSource(src string) (string, string, error) {

	vals := strings.Split(src, "::")
	if len(vals) < 2 {
		return "", "", errors.New(`Error, invalid format - Usage: stackname::http://someurl OR stackname::path/to/template`)
	}

	return vals[0], vals[1], nil
}
