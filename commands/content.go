package commands

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/daidokoro/qaz/bucket"
	"github.com/daidokoro/qaz/utils"
)

type contentHandlers map[string]func(string) (string, error)

func (c *contentHandlers) init() {
	if (*c) == nil {
		(*c) = make(map[string]func(string) (string, error))
	}

	// add fnctions
	(*c)["http"] = handleHTTTP
	(*c)["https"] = handleHTTTP
	(*c)["s3"] = handleS3
	(*c)["lambda"] = handleLambda
	(*c)["default"] = handleDefault
}

func (c *contentHandlers) execute(n string) (f func(string) (string, error)) {
	if v, ok := (*c)[n]; ok {
		f = v
		return
	}

	f = (*c)["default"]
	return
}

func handleHTTTP(source string) (string, error) {
	log.Debug("Source Type: [http] Detected, Fetching Source: [%s]", source)
	resp, err := utils.Get(source)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func handleS3(source string) (string, error) {
	log.Debug("Source Type: [s3] Detected, Fetching Source: [%s]", source)
	sess, err := manager.GetSess(run.profile)
	utils.HandleError(err)

	resp, err := bucket.S3Read(source, sess)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func handleLambda(source string) (string, error) {
	log.Debug("Source Type: [lambda] Detected, Fetching Source: %s", source)
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
}

func handleDefault(source string) (string, error) {
	if gitrepo.URL != "" {
		log.Debug("Source Type: [git-repo file] Detected, Fetching Source: %s", source)
		out, ok := gitrepo.Files[source]
		if ok {
			return out, nil
		} else if !ok {
			log.Warn("config [%s] not found in git repo - checking local file system", source)
		}

	}

	log.Debug("Source Type: [file] Detected, Fetching Source: [%s]", source)
	b, err := ioutil.ReadFile(source)
	if err != nil {
		return "", err
	}

	log.Debug("config file read: %s", string(b))
	return string(b), nil
}

// fetchContent - checks the source type, url/s3/file and calls the corresponding function
func fetchContent(source string) (string, error) {
	src, err := url.Parse(source)
	if err != nil {
		return "", err
	}

	var handler contentHandlers
	handler.init()

	return handler.execute(src.Scheme)(source)
}
