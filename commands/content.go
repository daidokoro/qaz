package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// global variables
var (
	region  string
	project string
	stacks  map[string]*stack
)

// fetchContent - checks the source type, url/s3/file and calls the corresponding function
func fetchContent(source string) (string, error) {
	switch strings.Split(strings.ToLower(source), ":")[0] {
	case "http", "https":
		Log(fmt.Sprintln("Source Type: [http] Detected, Fetching Source: ", source), level.debug)
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
	case "lambda":
		Log(fmt.Sprintln("Source Type: [lambda] Detected, Fetching Source: ", source), level.debug)
		lambdaSrc := strings.Split(strings.Replace(source, "lambda:", "", -1), "@")

		var raw interface{}
		if err := json.Unmarshal([]byte(lambdaSrc[0]), &raw); err != nil {
			return "", err
		}

		event, err := json.Marshal(raw)
		if err != nil {
			return "", err
		}

		reg, err := regexp.Compile("[^A-Za-z0-9_-]+")
		if err != nil {
			return "", err
		}

		lambdaName := reg.ReplaceAllString(lambdaSrc[1], "")

		f := awsLambda{
			name:    lambdaName,
			payload: event,
		}

		// using default profile
		sess := manager.sessions[run.profile]
		if err := f.Invoke(sess); err != nil {
			return "", err
		}

		return f.response, nil

	default:
		if gitrepo.URL != "" {
			Log(fmt.Sprintln("Source Type: [git-repo file] Detected, Fetching Source: ", source), level.debug)
			out, ok := gitrepo.files[source]
			if ok {
				return out, nil
			} else if !ok {
				Log(fmt.Sprintf("config [%s] not found in git repo - checking local file system", source), level.warn)
			}

		}

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
