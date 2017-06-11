package stacks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"qaz/bucket"
	"qaz/utils"
	"regexp"
	"strings"
)

// FetchContent - checks the s.Source type, url/s3/file and calls the corresponding function
func (s *Stack) fetchContent() error {
	switch strings.Split(strings.ToLower(s.Source), ":")[0] {
	case "http", "https":
		Log.Debug(fmt.Sprintln("Source Type: [http] Detected, Fetching Source: ", s.Source))
		resp, err := utils.Get(s.Source)
		if err != nil {
			return err
		}
		s.Template = resp
	case "s3":
		Log.Debug(fmt.Sprintln("Source Type: [s3] Detected, Fetching Source: ", s.Source))
		resp, err := bucket.S3Read(s.Source, s.Session)
		if err != nil {
			return err
		}

		s.Template = resp

	case "lambda":
		Log.Debug(fmt.Sprintln("Source Type: [lambda] Detected, Fetching Source: ", s.Source))
		lambdaSrc := strings.Split(strings.Replace(s.Source, "lambda:", "", -1), "@")

		var raw interface{}
		if err := json.Unmarshal([]byte(lambdaSrc[0]), &raw); err != nil {
			return err
		}

		event, err := json.Marshal(raw)
		if err != nil {
			return err
		}

		reg, err := regexp.Compile("[^A-Za-z0-9_-]+")
		if err != nil {
			return err
		}

		lambdaName := reg.ReplaceAllString(lambdaSrc[1], "")

		f := awslambda{
			name:    lambdaName,
			payload: event,
		}

		if err := f.Invoke(s.Session); err != nil {
			return err
		}

		s.Template = f.response

	default:
		if Git.URL != "" {
			Log.Debug(fmt.Sprintln("Source Type: [git-repo file] Detected, Fetching Source: ", s.Source))
			out, ok := Git.Files[s.Source]
			if ok {
				s.Template = out
				return nil
			} else if !ok {
				Log.Warn(fmt.Sprintf("config [%s] not found in git repo - checking local file system", s.Source))
			}

		}

		Log.Debug(fmt.Sprintln("Source Type: [file] Detected, Fetching Source: ", s.Source))
		b, err := ioutil.ReadFile(s.Source)
		if err != nil {
			return err
		}
		s.Template = string(b)
	}

	return nil
}
