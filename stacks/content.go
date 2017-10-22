package stacks

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/daidokoro/qaz/bucket"
	"github.com/daidokoro/qaz/utils"
)

type contentHandlers map[string]func(s *Stack) error

func (c *contentHandlers) init() {
	if (*c) == nil {
		(*c) = make(map[string]func(s *Stack) error)
	}

	// add fnctions
	(*c)["http"] = handleHTTTP
	(*c)["https"] = handleHTTTP
	(*c)["s3"] = handleS3
	(*c)["lambda"] = handleLambda
	(*c)["default"] = handleDefault
}

func (c *contentHandlers) execute(n string) (f func(s *Stack) error) {
	if v, ok := (*c)[n]; ok {
		f = v
		return
	}

	f = (*c)["default"]
	return
}

func handleHTTTP(s *Stack) (err error) {
	Log.Debug("Source Type: [http] Detected, Fetching Source: %s ", s.Source)
	resp, err := utils.Get(s.Source)
	if err != nil {
		return
	}
	s.Template = resp
	return
}

func handleS3(s *Stack) (err error) {
	Log.Debug("Source Type: [s3] Detected, Fetching Source: %s", s.Source)
	resp, err := bucket.S3Read(s.Source, s.Session)
	if err != nil {
		return err
	}

	s.Template = resp
	return
}

func handleLambda(s *Stack) (err error) {
	Log.Debug("Source Type: [lambda] Detected, Fetching Source: %s", s.Source)
	src, err := url.Parse(s.Source)
	if err != nil {
		return err
	}

	lambdaSrc := strings.Split(src.Opaque, "@")

	var raw interface{}
	if err = json.Unmarshal([]byte(lambdaSrc[0]), &raw); err != nil {
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
	return
}

func handleDefault(s *Stack) (err error) {
	if Git.URL != "" {
		Log.Debug("Source Type: [git-repo file] Detected, Fetching Source: %s", s.Source)
		out, ok := Git.Files[s.Source]
		if ok {
			s.Template = out
			return nil
		} else if !ok {
			Log.Warn("config [%s] not found in git repo - checking local file system", s.Source)
		}
	}

	Log.Debug("Source Type: [file] Detected, Fetching Source: %s", s.Source)
	b, err := ioutil.ReadFile(s.Source)
	if err != nil {
		return err
	}
	s.Template = string(b)
	return
}

// FetchContent - checks the s.Source type, url/s3/file and calls the corresponding function
func (s *Stack) FetchContent() error {
	src, err := url.Parse(s.Source)
	if err != nil {
		return err
	}

	var handler contentHandlers
	handler.init()

	if err := handler.execute(src.Scheme)(s); err != nil {
		return err
	}

	return nil
}
