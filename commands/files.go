package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
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
func getSource(s string) (string, string, error) {
	if strings.Contains(s, "::") {
		vals := strings.Split(s, "::")
		if len(vals) < 2 {
			return "", "", errors.New(`Error, invalid url format --> Example: stackname::http://someurl OR stackname::s3://bucket/key`)
		}

		return vals[0], vals[1], nil

	}

	name := filepath.Base(strings.Replace(s, filepath.Ext(s), "", -1))
	return name, s, nil
}

// genTimeParser - Parses templates before deploying them...
func genTimeParser(source string) (string, error) {

	templ, err := fetchContent(source)
	if err != nil {
		return "", err
	}

	// Create template
	t, err := template.New("gen-template").Funcs(genTimeFunctions).Parse(templ)
	if err != nil {
		return "", err
	}

	// so that we can write to string
	var doc bytes.Buffer
	t.Execute(&doc, config.vars())
	return doc.String(), nil
}
